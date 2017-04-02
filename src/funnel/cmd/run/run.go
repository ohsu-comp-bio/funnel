package run

import (
  "context"
	"funnel/logger"
	"github.com/spf13/cobra"
  "fmt"
  "os"
  "funnel/proto/tes"
  "github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
  "time"
)

var log = logger.New("run")

// TODO figure out a nice default for name
var name = "Funnel run"
var workdir = "/opt/funnel"
var server = "localhost:9090"
var project, description, varsFile string
var printTask, preemptible, wait bool
// TODO validate resources
var cpu int
var ram float64
var disk float64 = 10.0
var zones []string

// TODO allow outputs to be defined when they don't fit into the command
//      such as bam file index secondary file
// TODO with input contents, script could be loaded from file
// TODO what is stdout/err of funnel run?
//      should have job id to access job state later
// TODO load vars from file
// TODO is there a case for glob inputs?
// TODO volumes? size requirement?

var examples string = `
Examples:
  funnel run ubuntu 'md5sum {in .SRC} > {out .DST}' -- -SRC ~/input.txt -DST md5sum.txt"
  funnel run ubuntu 'ls {in .P | name "IN PATH" | desc "IN DESCRIPTION" | path "/tmp/in"  | create } > {out .O | path "/tmp/out" }'
`

var Cmd = &cobra.Command{
	Use:   "run [flags] IMAGE TEMPLATE -- VARS",
	Short: "Run a task.\n" + examples,
	Long:  ``,
	Run: run,
}

func init() {
  f := Cmd.Flags()
  f.StringVar(&name, "name", name, "Task name")
  f.StringVar(&description, "description", description, "Task description")
  f.StringVar(&project, "project", project, "Project")
  f.StringVar(&workdir, "workdir", workdir, "Set the containter working directory")
  f.IntVar(&cpu, "cpu", cpu, "Number of CPUs requested")
  f.Float64Var(&ram, "ram", ram, "Amount of RAM requested")
  f.Float64Var(&disk, "disk", disk, "Amount of disk space requested (in GB)")
  f.BoolVar(&preemptible, "preemptible", preemptible, "Allow task to be scheduled on preemptible workers")
  f.StringSliceVar(&zones, "zones", zones, "Require task be scheduled in certain zones")
  f.StringVar(&varsFile, "vars-file", varsFile, "Read vars a file")
  f.StringVar(&server, "server", server, "Address of Funnel server")
  f.BoolVar(&printTask, "print", printTask, "Print the task, instead of running it")
  f.BoolVar(&wait, "wait", wait, "Wait for the task to finish before exiting")
}


func run(cmd *cobra.Command, args []string) {

  dashIdx := cmd.Flags().ArgsLenAtDash()
  if dashIdx != -1 && dashIdx != 2 {
    cmd.Usage()
    os.Exit(2)
    return
  }

  // TODO validate
  image := args[0]
  rawcmd := args[1]
  rawCliVars := []string{}

  if dashIdx != -1 {
    rawCliVars = args[dashIdx:]
  }

  log.Debug("CLIVARS", rawCliVars)

  cliVars, err := parseCliVars(rawCliVars)
  fileVars, err := parseFileVars(varsFile)
  vars, err := mergeVars(cliVars, fileVars)

  checkErr(err)

  res, err := parseTpl(rawcmd, vars)

  checkErr(err)

  // TODO
  res.Volumes[0].SizeGb = disk

  task := &tes.Task{
    Name: name,
    ProjectID: project,
    Description: description,
    Inputs: res.Inputs,
    Outputs: res.Outputs,
    Resources: &tes.Resources{
      MinimumCpuCores: uint32(cpu),
      MinimumRamGb: ram,
      Zones: zones,
      Preemptible: preemptible,
      Volumes: res.Volumes,
    },
    Docker: []*tes.DockerExecutor{
      // TODO allow more than one command?
      {
        ImageName: image,
        Cmd: res.Cmd,
        Workdir: workdir,
        // TODO
        Stdin: "",
        Stdout: "",
        Stderr: "",
        // TODO no ports
        Ports: nil,
      },
    },
  }
  log.Debug("TASK", task)

  // Marshal message to JSON and print
  if printTask {
    mar := jsonpb.Marshaler{
      EmitDefaults: true,
      Indent: "\t",
    }
    s, merr := mar.MarshalToString(task)
    checkErr(merr)
    fmt.Println(s)
    return
  }

  c, cerr := newClient(server)
  checkErr(cerr)
  resp, rerr := c.RunTask(context.TODO(), task)
  checkErr(rerr)

  jobID := resp.Value
  fmt.Println(jobID)

  if wait {
    c.waitForTask(jobID)
    // TODO print/log result
    // TODO stream logs while waiting
  }
}

func checkErr(err error) {
  if err != nil {
    fmt.Println(err)
    os.Exit(3)
  }
}

type client struct {
	tes.TaskServiceClient
	conn *grpc.ClientConn
}

func (c *client) waitForTask(jobID string) {
  for range time.NewTicker(time.Second * 2).C {
    // TODO handle error
    r, _ := c.GetJob(context.TODO(), &tes.JobID{jobID})
    switch r.State {
    case tes.State_Complete, tes.State_Error, tes.State_SystemError, tes.State_Canceled:
      return
    }
  }
}

func newClient(address string) (*client, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		log.Error("Couldn't open RPC connection",
			"error", err,
			"address", address,
		)
		return nil, err
	}

	if err != nil {
		log.Error("Couldn't connect to server", err)
		return nil, err
	}

	s := tes.NewTaskServiceClient(conn)
	return &client{s, conn}, nil
}

// Close closes the client connection.
func (client *client) Close() {
	client.conn.Close()
}
