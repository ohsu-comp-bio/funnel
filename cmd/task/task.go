package task

import (
	"bufio"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
	"os"
)

var tesServer string
var log = logger.New("task cmd")

// Cmd represents the task command
var Cmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"tasks"},
	Short:   "Make API calls to a TES server.",
}

func init() {
	Cmd.PersistentFlags().StringVarP(&tesServer, "server", "S", "http://localhost:8000", "")
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(cancelCmd)
	Cmd.AddCommand(waitCmd)
}

// stdinPiped returns true if the cli command has stdin being piped,
// e.g. 'echo $TASK_ID | funnel task get'
func stdinPiped() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

type perIDHandler func(c *client.Client, id string) error

// runPerID helps with commands that can accept task IDs via both CLI args
// and stdin, such as `task get` and `task cancel`. The given handler func
// is called once per task ID. Processing stops on the first error.
func runPerID(cmd *cobra.Command, args []string, h perIDHandler) error {
	// Arguments may be given either via the cli parser,
	// or via stdin.
	if len(args) == 0 && !stdinPiped() {
		return cmd.Help()
	}

	cli := client.NewClient(tesServer)

	// Read arguments from the cli parser
	for _, arg := range args {
		if err := h(cli, arg); err != nil {
			return err
		}
	}

	// Read arguments from stdin
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		arg := s.Text()
		if err := h(cli, arg); err != nil {
			return err
		}
	}

	// Check for scanner error
	if err := s.Err(); err != nil {
		return err
	}

	return nil
}
