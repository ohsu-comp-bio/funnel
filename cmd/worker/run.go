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
			"'rpc' - task logs will be sent via gRPC calls to the funnel server. (default)\n\t",
			"'in-memory' - store task logs in memory and print the full logs at the end of the run.\n\t",
			"'log' - task logs are printed to stderr via the logger."),
	)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a task directly, bypassing the server.",
	RunE: func(cmd *cobra.Command, args []string) error {

		task := &tes.Task{}

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
		if rawTask != "" {
			err := jsonpb.UnmarshalString(rawTask, task)
			if err != nil {
				return err
			}
		}

		if taskID != "" {
			task.Id = taskID
		}

		// configure worker/runner
		conf := config.DefaultConfig()
		config.ParseFile(configFile, &conf)

		workerDconf := config.WorkerInheritConfigVals(flagConf)

		// file vals <- cli val
		err := mergo.MergeWithOverwrite(&conf.Worker, workerDconf)
		if err != nil {
			return err
		}

		// set to 'rpc' by default if unset
		if taskSvc == "" {
			taskSvc = "rpc"
		}

		return run(task, conf.Worker, taskSvc)
	},
}

// run runs a task with the TaskRunner directly. No Worker is created,
// the task goes directly to a TaskRunner.
func run(task *tes.Task, conf config.Worker, taskLogger string) error {
	if conf.ID == "" {
		conf.ID = scheduler.GenWorkerID("funnel")
	}

	logger.Configure(conf.Logger)
	log := logger.Sub("runner", "workerID", conf.ID, "taskID", task.Id)

	runner := worker.DefaultRunner{
		Conf:   conf,
		Mapper: worker.NewFileMapper(conf.WorkDir),
		Store:  storage.Storage{},
		Svc:    nil,
		Log:    log,
	}

	switch strings.ToLower(taskLogger) {
	case "in-memory":
		if err := tes.Validate(task); err != nil {
			return fmt.Errorf("Invalid task message: %v", err)
		}
		task.Id = util.GenTaskID()
		runner.Svc = worker.NewInMemoryTaskSvc(task, log)
	case "log":
		if err := tes.Validate(task); err != nil {
			return fmt.Errorf("Invalid task message: %v", err)
		}
		task.Id = util.GenTaskID()
		runner.Svc = worker.NewLogTaskSvc(task, log)
	case "rpc":
		if task.Id == "" {
			return fmt.Errorf("No task id provided")
		}
		svc, _ := worker.NewRPCTaskSvc(conf, task.Id)
		runner.Svc = svc
	default:
		return fmt.Errorf("Unknown task-logger: %s. Must be one of: %s", taskLogger, loggers)
	}

	runner.Run(context.Background())

	if taskLogger == "in-memory" {
		t, _ := runner.Svc.Task()
		log.Info("Task reached terminal state", "state", t.State)
		m := &jsonpb.Marshaler{
			EnumsAsInts:  false,
			EmitDefaults: true,
			Indent:       "  ",
			OrigName:     true,
		}
		ts, _ := m.MarshalToString(t)
		fmt.Println(ts)
	}

	return nil
}
