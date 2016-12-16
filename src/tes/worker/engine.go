package tesTaskEngineWorker

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"syscall"
	"tes/scheduler"
	"tes/storage"
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
  "github.com/dgrijalva/jwt-go"
)

// Engine is responsivle for running a job. This includes downloading inputs,
// communicating updates to the scheduler service, running the actual command,
// and uploading outputs.
type Engine interface {
	RunJob(ctx context.Context, job *pbr.JobResponse) error
}

// engine is the internal implementation of a docker job engine.
type engine struct {
  // Address of the scheduler, e.g. "localhost:9090"
	schedAddr string
  // Directory to write job files to
	workDir string
  // Storage client for downloading/uploading files
	storage   *storage.Storage
}

type auth struct {
  Key string
  Secret string
}

// NewEngine returns a new Engine instance configured with a given scheduler address,
// working directory, and storage client.
//
// If the working directory can't be initialized, this returns an error.
func NewEngine(addr string, wDir string, store *storage.Storage) (Engine, error) {
	dir, err := filepath.Abs(wDir)
	if err != nil {
		return nil, err
	}
	ensureDir(dir)
	return &engine{addr, dir, store}, nil
}

// RunJob runs a job.
func (eng *engine) RunJob(ctx context.Context, job *pbr.JobResponse) error {
	// This is essentially a simple helper for runJob() (below).
	// This ensures that the job state is always updated in the scheduler,
	// without having to do it on 15+ different lines in runJob() and others.
	//
	// Please try to keep this function as simple as possible.
	// New code should probably go in runJob()

	// Get a client for the scheduler service
	sched, schederr := scheduler.NewClient(eng.schedAddr)
	defer sched.Close()
	// TODO if we're here then we have a serious problem. We have already
	//      told the scheduler that we're running the job, but now we can't
	//      tell it things are broken, so the job is going to stay running
	//      forever. Possibly the scheduler should have a job timeout.
	if schederr != nil {
		return schederr
	}

	// Tell the scheduler the job is running.
	sched.SetRunning(ctx, job.Job)
	joberr := eng.runJob(ctx, sched, job)

	// Tell the scheduler whether the job failed or completed.
	if joberr != nil {
		sched.SetFailed(ctx, job.Job)
		//BUG: error status not returned to scheduler
		log.Printf("Failed to run job: %s", job.Job.JobID)
		log.Printf("%s", joberr)
	} else {
		sched.SetComplete(ctx, job.Job)
	}
	return joberr
}

func (eng *engine) getAuth(tokenstr string) *auth {
  log.Printf("Parsing token: %s", tokenstr)
  // TODO there was a temporary error about time vs server time.
  //      does that have to do with token parsing?
  // TODO what to do if tokenstr is empty?
  // Parse takes the token string and a function for looking up the key. The latter is especially
  // useful if you use multiple keys for your application.  The standard is to use 'kid' in the
  // head of the token to identify which key to use, but the parsed token (head and claims) is provided
  // to the callback, providing flexibility.
  token, err := jwt.Parse(tokenstr, func(token *jwt.Token) (interface{}, error) {
      // Don't forget to validate the alg is what you expect:
      if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
          return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
      }
      // TODO where does secret key come from?
      //      it's possible we'll have multiple secrets, so need to get the secret
      //      based on some information in the token
      // The key must be []byte for HMAC. The docs/API for the jwt lib are pretty bad about
      // this point, so now you know!
      return []byte("secret"), nil
  })

  if err != nil {
      log.Println("Error parsing auth token")
      log.Println(err)
      return nil
  }

  claims, ok := token.Claims.(jwt.MapClaims)
  if ok && token.Valid {
    // TODO key/error checking here. These claims could be missing.
    key, _ := claims["S3_ACCESS_KEY_ID"].(string)
    sec, _ := claims["S3_ACCESS_SECRET"].(string)
    return &auth{key, sec}
  } else {
    log.Println("Error accessing auth token")
  }
  return nil
}

// runJob calls a series of other functions to process a job:
// 1. set up the file mapping between the host and the container
// 2. set up the storage client
// 3. download the inputs
// 4. run the job steps
// 4a. update the scheduler with job status after each step
// 5. upload the outputs
func (eng *engine) runJob(ctx context.Context, sched *scheduler.Client, jobR *pbr.JobResponse) error {
  auth := eng.getAuth(jobR.Auth)
  job := jobR.Job
	mapper, merr := eng.getMapper(job)
	if merr != nil {
		return merr
	}

	store, serr := eng.getStorage(job, auth)
	if serr != nil {
		return serr
	}

	derr := eng.downloadInputs(mapper, store)
	if derr != nil {
		return derr
	}

	// TODO is it possible to allow context.Done() to kill the current step?
	for stepNum, step := range job.Task.Docker {
		stepLog, err := eng.runStep(mapper, step)

		if stepLog != nil {
			// Send the scheduler service a job status update
			statusReq := &pbr.UpdateStatusRequest{
				Id:   job.JobID,
				Step: int64(stepNum),
				Log:  stepLog,
			}
			sched.UpdateJobStatus(ctx, statusReq)
		}
		if err != nil {
			return err
		}
	}

	uerr := eng.uploadOutputs(mapper, store)
	if uerr != nil {
		return uerr
	}

	return nil
}

// getMapper returns a FileMapper instance with volumes, inputs, and outputs
// configured for the given job.
func (eng *engine) getMapper(job *pbe.Job) (*FileMapper, error) {
	mapper := NewJobFileMapper(job.JobID, eng.workDir)

	// Iterates through job.Task.Resources.Volumes and add the volume to mapper.
	for _, vol := range job.Task.Resources.Volumes {
		err := mapper.AddVolume(vol.Source, vol.MountPoint)
		if err != nil {
			return nil, err
		}
	}

	// Add all the inputs to the mapper
	for _, input := range job.Task.Inputs {
		err := mapper.AddInput(input)
		if err != nil {
			return nil, err
		}
	}

	// Add all the outputs to the mapper
	for _, output := range job.Task.Outputs {
		err := mapper.AddOutput(output)
		if err != nil {
			return nil, err
		}
	}

	return mapper, nil
}

// getStorage returns a Storage instance configured for the given job.
// This will check the user authorization against the supported storage systems.
func (eng *engine) getStorage(job *pbe.Job, au *auth) (*storage.Storage, error) {
	// TODO catch error
  if au != nil {
    return eng.storage.WithS3(
      "192.168.99.101:9000",
      au.Key,
      au.Secret,
      false)
  } else {
    return eng.storage, nil
  }
}

func (eng *engine) downloadInputs(mapper *FileMapper, store *storage.Storage) error {
	// Validate all the input source URLs
	//for _, input := range mapper.Inputs {
	// TODO ?
	//}

	// Download all the inputs from storage
	for _, input := range mapper.Inputs {
		err := store.Get(input.Location, input.Path, input.Class)
		if err != nil {
			return err
		}
	}
	return nil
}

// The bulk of job running happens here.
func (eng *engine) runStep(mapper *FileMapper, step *pbe.DockerExecutor) (*pbe.JobLog, error) {

	dcmd := DockerCmd{
		ImageName: step.ImageName,
		Cmd:       step.Cmd,
		Volumes:   mapper.Volumes,
		Workdir:   step.Workdir,
		// TODO make this configurable
		RemoveContainer: true,
		Stdin:           nil,
		Stdout:          nil,
		Stderr:          nil,
	}

	// Find the path for job stdin
	if step.Stdin != "" {
		f, err := mapper.OpenHostFile(step.Stdin)
		if err != nil {
			return nil, fmt.Errorf("Error setting up job stdin: %s", err)
		}
		defer f.Close()
		dcmd.Stdin = f
	}

	// Create file for job stdout
	if step.Stdout != "" {
		f, err := mapper.CreateHostFile(step.Stdout)
		if err != nil {
			return nil, fmt.Errorf("Error setting up job stdout: %s", err)
		}
		defer f.Close()
		dcmd.Stdout = f
	}

	// Create file for job stderr
	if step.Stderr != "" {
		f, err := mapper.CreateHostFile(step.Stderr)
		if err != nil {
			return nil, fmt.Errorf("Error setting up job stderr: %s", err)
		}
		defer f.Close()
		dcmd.Stderr = f
	}

	cmdErr := dcmd.Run()
	exitCode := getExitCode(cmdErr)
	log.Printf("Exit code: %d", exitCode)

	// TODO rethink these messages. You probably don't want head().
	//      you also don't get this until the step is finished,
	//      when you really want streaming.
	//
	// Get the head of the stdout/stderr files, if they exist.
	stdoutText := ""
	stderrText := ""
	if dcmd.Stdout != nil {
		stdoutText = readFileHead(dcmd.Stdout.Name())
	}
	if dcmd.Stderr != nil {
		stderrText = readFileHead(dcmd.Stderr.Name())
	}

	steplog := &pbe.JobLog{
		Stdout:   stdoutText,
		Stderr:   stderrText,
		ExitCode: exitCode,
	}

	if cmdErr != nil {
		return steplog, cmdErr
	}
	return steplog, nil
}

func (eng *engine) uploadOutputs(mapper *FileMapper, store *storage.Storage) error {
	// Upload all the outputs to storage
	for _, out := range mapper.Outputs {
		err := store.Put(out.Location, out.Path, out.Class)
		if err != nil {
			return err
		}
	}
	return nil
}

// getExitCode gets the exit status (i.e. exit code) from the result of an executed command.
// The exit code is zero if the command completed without error.
func getExitCode(err error) int32 {
	if err != nil {
		if exiterr, exitOk := err.(*exec.ExitError); exitOk {
			if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
				return int32(status.ExitStatus())
			}
		} else {
			log.Printf("Could not determine exit code. Using default -999")
			return -999
		}
	}
	// The error is nil, the command returned successfully, so exit status is 0.
	return 0
}
