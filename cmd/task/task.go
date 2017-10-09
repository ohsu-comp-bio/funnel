package task

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
	"os"
)

var log = logger.New("task cmd")

type Hooks struct {
	Create func(server string, messages []string) error
	Get    func(server string, ids []string, taskView string) error
	List   func(server, view, state, name string) error
	Cancel func(server string, ids []string) error
	Wait   func(server string, ids []string) error
}

var DefaultHooks = Hooks{
	Create: Create,
	Get:    Get,
	List:   List,
	Cancel: Cancel,
	Wait:   Wait,
}

func NewCommand(hooks Hooks) *cobra.Command {

	defaultTesServer := "http://localhost:8000"
	tesServer := defaultTesServer

	var (
		taskView  string
		taskName  string
		taskState string
	)

	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"tasks"},
		Short:   "Make API calls to a TES server.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			f := cmd.PersistentFlags()
			f.StringVarP(&tesServer, "server", "S", tesServer, "")
			if tesServer == defaultTesServer {
				tesServer = os.Getenv("FUNNEL_SERVER")
			}
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "create [task.json ...]",
		Short: "Create one or more tasks to run on the server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return hooks.Create(tesServer, args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all tasks.",
		PreRun: func(cmd *cobra.Command, args []string) {
			f := cmd.Flags()
			f.StringVarP(&taskView, "view", "v", "BASIC", "Task view")
			f.StringVarP(&taskState, "state", "s", "", "Task state")
			f.StringVarP(&taskName, "name", "n", "", "Task name")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return hooks.List(tesServer, taskView, taskState, taskName)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get [taskID ...]",
		Short: "Get one or more tasks by ID.",
		PreRun: func(cmd *cobra.Command, args []string) {
			cmd.Flags().StringVarP(&taskView, "view", "v", "FULL", "Task view")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return hooks.Get(tesServer, args, taskView)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "cancel [taskID ...]",
		Short: "Cancel one or more tasks by ID.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return hooks.Cancel(tesServer, args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "wait [taskID...]",
		Short: "Wait for one or more tasks to complete.\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return hooks.Wait(tesServer, args)
		},
	})

	return cmd
}
