package task

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
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
	)
	listView := choiceVar{val: "BASIC"}

	list := &cobra.Command{
		Use:   "list",
		Short: "List all tasks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.List(tesServer, listView.val, pageToken, pageSize, listAll, cmd.OutOrStdout())
		},
	}

	lf := list.Flags()
	listView.AddChoices("BASIC", "MINIMAL", "FULL")
	lf.VarP(&listView, "view", "v", "Task view")
	lf.StringVarP(&pageToken, "page-token", "p", pageToken, "Page token")
	lf.Uint32VarP(&pageSize, "page-size", "s", pageSize, "Page size")
	lf.BoolVar(&listAll, "all", listAll, "List all tasks")

	getView := choiceVar{val: "FULL"}
	get := &cobra.Command{
		Use:   "get [taskID ...]",
		Short: "Get one or more tasks by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.Get(tesServer, args, getView.val, cmd.OutOrStdout())
		},
	}

	gf := get.Flags()
	getView.AddChoices("BASIC", "MINIMAL", "FULL")
	gf.VarP(&getView, "view", "v", "Task view")

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

type choiceVar struct {
	choices map[string]bool
	val     string
}

func (c *choiceVar) AddChoices(choices ...string) {
	if c.choices == nil {
		c.choices = map[string]bool{}
	}
	for _, choice := range choices {
		c.choices[choice] = true
	}
}

func (c *choiceVar) String() string {
	return c.val
}

func (c *choiceVar) Set(v string) error {
	if _, ok := c.choices[v]; !ok {
		return fmt.Errorf("invalid choice: %s", v)
	}
	c.val = v
	return nil
}

func (c *choiceVar) Get() interface{} {
	return c.val
}

func (c *choiceVar) Type() string {
	return "string"
}
