package worker

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger/logutils"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/worker"
	"github.com/spf13/cobra"
)

var rawTask string
var RunTaskCmd = &cobra.Command{
	Use: "runtask",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(rawTask)
		// Load tes.Task from raw string
		var task tes.Task
		err := jsonpb.UnmarshalString(rawTask, &task)
		if err != nil {
			return err
		}
		task.Id = util.GenTaskID()

		// TODO validate task

		conf := config.DefaultConfig()
		conf.Worker.WorkDir = "direct-" + task.Id
		conf.Worker.ID = scheduler.GenWorkerID("funnel")
		// TODO conf.Worker.Storage =

		return RunTask(&task, conf)
	},
}

func init() {
	f := RunTaskCmd.Flags()
	f.StringVar(&rawTask, "task", "", "Task JSON")
}

func RunTask(task *tes.Task, conf config.Config) error {

	logutils.Configure(conf)
	runner := worker.NewDirectRunner(task, conf.Worker)
	runner.Run(context.Background())

	return nil
}
