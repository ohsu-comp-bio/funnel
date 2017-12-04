package task

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

// NewCommand returns the "task" subcommands.
func NewCommand() *cobra.Command {
	cmd, _ := newCommandHooks()
	return cmd
}

func newCommandHooks() (*cobra.Command, *hooks) {

	h := &hooks{
		Create: Create,
		Get:    Get,
		List:   List,
		Cancel: Cancel,
		Wait:   Wait,
	}

	defaultTesServer := "http://localhost:8000"
	tesServer := defaultTesServer

	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"tasks"},
		Short:   "Make API calls to a TES server.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if tesServer == defaultTesServer {
				if val := os.Getenv("FUNNEL_SERVER"); val != "" {
					tesServer = val
				}
			}
		},
	}
	f := cmd.PersistentFlags()
	f.StringVarP(&tesServer, "server", "S", defaultTesServer, "")

	create := &cobra.Command{
		Use:   "create [task.json ...]",
		Short: "Create one or more tasks to run on the server.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Create(tesServer, args, cmd.OutOrStdout())
		},
	}

	var (
		pageToken string
		pageSize  uint32
		listAll   bool
		listView  string
	)

	list := &cobra.Command{
		Use:   "list",
		Short: "List all tasks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.List(tesServer, listView, pageToken, pageSize, listAll, cmd.OutOrStdout())
		},
	}

	lf := list.Flags()
	lf.StringVarP(&listView, "view", "v", "basic", "Task view")
	lf.StringVarP(&pageToken, "page-token", "p", pageToken, "Page token")
	lf.Uint32VarP(&pageSize, "page-size", "s", pageSize, "Page size")
	lf.BoolVar(&listAll, "all", listAll, "List all tasks")

	var getView string
	get := &cobra.Command{
		Use:   "get [taskID ...]",
		Short: "Get one or more tasks by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Get(tesServer, args, getView, cmd.OutOrStdout())
		},
	}

	gf := get.Flags()
	gf.StringVarP(&getView, "view", "v", "full", "Task view")

	cancel := &cobra.Command{
		Use:   "cancel [taskID ...]",
		Short: "Cancel one or more tasks by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Cancel(tesServer, args, cmd.OutOrStdout())
		},
	}

	wait := &cobra.Command{
		Use:   "wait [taskID...]",
		Short: "Wait for one or more tasks to complete.\n",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Wait(tesServer, args)
		},
	}

	cmd.AddCommand(create, get, list, cancel, wait)
	return cmd, h
}

type hooks struct {
	Create func(server string, messages []string, w io.Writer) error
	Get    func(server string, ids []string, view string, w io.Writer) error
	List   func(server, view, pageToken string, pageSize uint32, all bool, w io.Writer) error
	Cancel func(server string, ids []string, w io.Writer) error
	Wait   func(server string, ids []string) error
}

func getTaskView(taskView string) (int32, error) {
	taskView = strings.ToUpper(taskView)
	var view int32
	var ok bool
	view, ok = tes.TaskView_value[taskView]
	if !ok {
		return view, fmt.Errorf("Unknown task view: %s. Valid task views: ['basic', 'minimal', 'full']", taskView)
	}
	return view, nil
}
