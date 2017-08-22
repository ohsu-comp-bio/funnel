package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/worker"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strings"
)

var rawTask string
var rawTaskFile string
var taskID string
var taskSvc string
var loggers = [3]string{"rpc", "in-memory", "log"}

func init() {
	f := runCmd.Flags()
	f.StringVar(&rawTask, "task", "", "Task JSON")
	f.StringVar(&rawTaskFile, "task-file", "", "Task JSON file path")
	f.StringVar(&taskID, "task-id", "", "Task ID")
	f.StringVar(
		&taskSvc,
		"task-logger",
		"",
		fmt.Sprintln("Task logger interface.\n\t",
			"'rpc' - task logs will be sent via gRPC calls to the funnel server.\n\t",
			"'in-memory' - store task logs in memory and print the full logs at the end of the run.\n\t",
			"'log' - task logs are printed to stderr via the logger."),
	)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a task directly, bypassing the server.",
	RunE: func(cmd *cobra.Command, args []string) error {

		if rawTask == "" && rawTaskFile == "" && taskID == "" {
			fmt.Printf("No task was provided.\n\n")
			return cmd.Help()
		}

		if rawTask != "" {
			rawTask = strings.Join(append([]string{rawTask}, args...), " ")

			var anything interface{}
			err := json.Unmarshal([]byte(rawTask), &anything)
			if err != nil {
				return fmt.Errorf("Error parsing Task JSON: %v", err)
			}
			if t, ok := anything.(string); ok {
				rawTask = t
			} else {
				b, err := json.Marshal(&anything)
				if err != nil {
					return fmt.Errorf("Error cleaning Task JSON: %v", err)
				}
				rawTask = string(b)
			}
		}

		if rawTaskFile != "" {
			b, err := ioutil.ReadFile(rawTaskFile)
			if err != nil {
				return err
			}
			rawTask = string(b)
		}

		// Load tes.Task from raw string (comes from CLI flag).
		var task tes.Task
		err := jsonpb.UnmarshalString(rawTask, &task)
		if err != nil {
			return err
		}

		if taskID != "" {
			task.Id = taskID
		}

		// configure worker/runner
		conf := config.DefaultConfig()
		config.ParseFile(configFile, &conf)

		workerDconf := config.WorkerInheritConfigVals(flagConf)

		// file vals <- cli val
		err = mergo.MergeWithOverwrite(&conf.Worker, workerDconf)
		if err != nil {
			return err
		}

		return runTask(&task, conf)
	},
}

// runTask handles TaskRunner startup. No Worker is created,
// the task goes directly to a TaskRunner.
func runTask(task *tes.Task, conf config.Config) error {
	if err := tes.Validate(task); err != nil {
		return fmt.Errorf("Invalid task message: %v", err)
	}

	if conf.Worker.ID == "" && taskSvc == "rpc" {
		conf.Worker.ID = scheduler.GenWorkerID("funnel")
	}

	if task.Id == "" {
		if taskSvc == "rpc" {
			return fmt.Errorf("No task id provided")
		}
		task.Id = util.GenTaskID()
	}

	logger.Configure(conf.Worker.Logger)
	log := logger.Sub("runner", "workerID", conf.Worker.ID, "taskID", task.Id)

	runner := worker.DefaultRunner{
		Conf:   conf.Worker,
		Mapper: worker.NewFileMapper(conf.Worker.WorkDir),
		Store:  storage.Storage{},
		Svc:    nil,
		Log:    log,
	}

	if taskSvc == "" {
		taskSvc = "rpc"
	}

	switch strings.ToLower(taskSvc) {
	case "in-memory":
		runner.Svc = worker.NewInMemoryTaskSvc(task, log)
	case "log":
		runner.Svc = worker.NewLogTaskSvc(task, log)
	case "rpc":
		svc, _ := worker.NewRPCTaskSvc(conf.Worker, task.Id)
		runner.Svc = svc
	default:
		return fmt.Errorf("Unknown task-logger: %s. Must be one of: %s", taskSvc, loggers)
	}

	runner.Run(context.Background())
	if taskSvc == "in-memory" {
		t, _ := runner.Svc.Task()
		log.Info("Task reached terminal state", "state", t.State, "task", t)
	}
	return nil
}
