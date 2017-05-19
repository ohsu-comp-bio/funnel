package run

import (
	"bufio"
	"github.com/kballard/go-shellquote"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/spf13/cobra"
	"os"
)

// *********************************************************************
// IMPORTANT:
// Usage/help docs are defined in usage.go.
// If you're updating flags, you probably need to update that file.
// *********************************************************************

var log = logger.New("run")

// Cmd represents the run command
var Cmd = &cobra.Command{
	Use:   "run 'CMD' [flags]",
	Short: "Run a task.",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Run(args)
		if err != nil {
			//cmd.Usage()
		}
		return err
	},
	DisableFlagParsing: true,
}

func init() {
	Cmd.SetUsageTemplate(usage)
}

// ParseString calls shellquote.Split(s) and passes those args to Parse().
func ParseString(s string) ([]*tes.Task, error) {
	args, err := shellquote.Split(s)
	if err != nil {
		return nil, err
	}
	return Parse(args)
}

// Parse task a list of CLI args/flags, and converts them to tasks.
func Parse(args []string) ([]*tes.Task, error) {

	vals := flagVals{}
	err := parseTopLevelArgs(&vals, args)
	if err != nil {
		return nil, err
	}

	// Scatter all vals into tasks
	tasks := []*tes.Task{}
	for _, v := range scatter(vals) {
		// Parse inputs, outputs, environ, and tags from flagVals
		// and update the task.
		task, err := valsToTask(v)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// Run takes a list of CLI args/flags, converts them to tasks and runs them.
func Run(args []string) error {
	vals := flagVals{}
	err := parseTopLevelArgs(&vals, args)
	if err != nil {
		return err
	}

	// TES HTTP client
	tg := taskGroup{
		printTask: vals.printTask,
		client:    client.NewClient(vals.server),
	}

	// Scatter all vals into tasks
	for _, v := range scatter(vals) {
		// Parse inputs, outputs, environ, and tags from flagVals
		// and update the task.
		task, err := valsToTask(v)
		if err != nil {
			return err
		}

		tg.runTask(task, v.wait, v.waitFor)
	}

	return tg.wait()
}

// scatter reads each line from each scatter file, extending "base" flagVals
// with per-scatter vals from each line.
func scatter(base flagVals) []flagVals {
	if len(base.scatterFiles) == 0 {
		return []flagVals{base}
	}

	out := []flagVals{}

	for _, sc := range base.scatterFiles {
		// Read each line of the scatter file.
		fh, _ := os.Open(sc)
		scanner := bufio.NewScanner(fh)
		for scanner.Scan() {
			// Per-scatter flags
			sp, _ := shellquote.Split(scanner.Text())
			tv := base
			parseTaskArgs(&tv, sp)
			// Parse scatter file flags into new flagVals
			out = append(out, tv)
		}
	}
	return out
}
