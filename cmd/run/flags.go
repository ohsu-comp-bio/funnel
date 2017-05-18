package run

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/spf13/pflag"
	"io/ioutil"
	"os"
)

// *********************************************************************
// IMPORTANT:
// Usage/help docs are defined in usage.go.
// If you're updating flags, you probably need to update that file.
// *********************************************************************

type executor struct {
	cmd    string
	stdin  string
	stdout string
	stderr string
}

// flagVals captures values from CLI flag parsing
type flagVals struct {
	// Top-level flag values. These are not allowed to be redefined
	// by scattered tasks or extra args, to avoid complexity in avoiding
	// circular imports or nested scattering
	printTask    bool
	server       string
	extra        []string
	extraFiles   []string
	scatterFiles []string
	cmds         []string

	// Internal tracking of executors. Not set by flags.
	execs []executor

	// Per-task flag values. These may be overridden by scattered tasks.
	name string
	// TODO all executors share the same container and workdir
	//      but could possibly be separate.
	workdir     string
	container   string
	project     string
	description string
	stdin       []string
	stdout      []string
	stderr      []string
	preemptible bool
	wait        bool
	waitFor     []string
	inputs      []string
	inputDirs   []string
	outputs     []string
	outputDirs  []string
	contents    []string
	environ     []string
	tags        []string
	volumes     []string
	zones       []string
	cpu         int
	ram         float64
	disk        float64
}

func newFlags(v *flagVals) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)

	// These flags are separate because they are not allowed
	// in scattered tasks.
	//
	// Scattering and loading extra args is currently only allowed
	// at the top level in order to avoid any issues with circular
	// includes. If we want this to be per-task, it's possible,
	// but more work.
	f.StringVarP(&v.server, "server", "S", v.server, "")
	f.BoolVarP(&v.printTask, "print", "p", v.printTask, "")
	f.StringSliceVarP(&v.extra, "extra", "x", v.extra, "")
	f.StringSliceVarP(&v.extraFiles, "extra-file", "X", v.extraFiles, "")
	f.StringSliceVar(&v.scatterFiles, "scatter", v.scatterFiles, "")
	f.StringSliceVar(&v.cmds, "cmd", v.cmds, "")

	// Disable sorting in order to visit flags in primordial order below.
	// See buildExecs()
	f.SortFlags = false

	// General
	f.StringVarP(&v.container, "container", "c", v.container, "")
	f.StringVarP(&v.workdir, "workdir", "w", v.workdir, "")

	// Input/output
	f.StringSliceVarP(&v.inputs, "in", "i", v.inputs, "")
	f.StringSliceVarP(&v.inputDirs, "in-dir", "I", v.inputDirs, "")
	f.StringSliceVarP(&v.outputs, "out", "o", v.outputs, "")
	f.StringSliceVarP(&v.outputDirs, "out-dir", "O", v.outputDirs, "")
	f.StringSliceVar(&v.stdin, "stdin", v.stdin, "")
	f.StringSliceVar(&v.stdout, "stdout", v.stdout, "")
	f.StringSliceVar(&v.stderr, "stderr", v.stderr, "")
	f.StringSliceVarP(&v.contents, "contents", "C", v.contents, "")

	// Resoures
	f.IntVar(&v.cpu, "cpu", v.cpu, "")
	f.Float64Var(&v.ram, "ram", v.ram, "")
	f.Float64Var(&v.disk, "disk", v.disk, "")
	f.StringSliceVar(&v.zones, "zone", v.zones, "")
	f.BoolVar(&v.preemptible, "preemptible", v.preemptible, "")

	// Other
	f.StringVarP(&v.name, "name", "n", v.name, "")
	f.StringVar(&v.description, "description", v.description, "")
	f.StringVar(&v.project, "project", v.project, "")
	f.StringSliceVar(&v.volumes, "vol", v.volumes, "")
	f.StringSliceVar(&v.tags, "tag", v.tags, "")
	f.StringSliceVarP(&v.environ, "env", "e", v.environ, "")

	// TODO
	//f.StringVar(&cmdFile, "cmd-file", cmdFile, "Read cmd template from file")
	f.BoolVar(&v.wait, "wait", v.wait, "")
	f.StringSliceVar(&v.waitFor, "wait-for", v.waitFor, "")
	return f
}

// Set default flagVals
func defaultVals(vals *flagVals) {
	if vals.workdir == "" {
		vals.workdir = "/opt/funnel"
	}

	if vals.container == "" {
		vals.container = "alpine"
	}

	// Default name
	if vals.name == "" {
		vals.name = "Funnel run: " + vals.cmds[0]
	}

	if vals.server == "" {
		vals.server = "http://localhost:8000"
	}
}

func parseTopLevelArgs(vals *flagVals, args []string) error {
	args = loadExtras(args)
	flags := newFlags(vals)
	err := flags.Parse(args)

	if err != nil {
		return err
	}

	if len(flags.Args()) > 1 {
		return fmt.Errorf("--in, --out and --env args should have the form 'KEY=VALUE' not 'KEY VALUE'. Extra args: %s", flags.Args()[1:])
	}

	// Prepend command string given as positional argument to the args.
	// Prepend it as a flag so that it works better with parseTaskArgs().
	if len(flags.Args()) == 1 {
		cmd := flags.Args()[0]
		args = append([]string{"--cmd", cmd}, args...)
	}

	if len(vals.cmds) == 0 {
		return fmt.Errorf("you must specify a command to run")
	}

	// Fill in empty values with defaults.
	defaultVals(vals)
	parseTaskArgs(vals, args)

	return nil
}

func parseTaskArgs(vals *flagVals, args []string) {
	fl := newFlags(vals)
	fl.Parse(args)
	buildExecs(fl, vals, args)
}

// Visit flags to determine commands + stdin/out/err
// and build that information into vals.execs
func buildExecs(flags *pflag.FlagSet, vals *flagVals, args []string) {
	vals.execs = nil
	vals.cmds = nil
	var exec *executor
	flags.ParseAll(args, func(f *pflag.Flag, value string) error {
		switch f.Name {
		case "cmd":
			if exec != nil {
				// Append the current executor and start a new one.
				vals.execs = append(vals.execs, *exec)
			}
			exec = &executor{
				cmd: value,
			}
		case "stdout":
			exec.stdout = value
		case "stderr":
			exec.stderr = value
		case "stdin":
			exec.stdin = value
		}
		return nil
	})
	if exec != nil {
		vals.execs = append(vals.execs, *exec)
	}
}

// Load extra arguments from "--extra", "--extra-file", and stdin
func loadExtras(args []string) []string {
	vals := &flagVals{}
	flags := newFlags(vals)
	flags.Parse(args)

	// Load CLI arguments from files, which allows reusing common CLI args.
	for _, xf := range vals.extraFiles {
		b, _ := ioutil.ReadFile(xf)
		vals.extra = append(vals.extra, string(b))
	}

	// Load CLI arguments from stdin, which allows bash heredoc for easily
	// spreading args over multiple lines.
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		b, _ := ioutil.ReadAll(os.Stdin)
		if len(b) > 0 {
			vals.extra = append(vals.extra, string(b))
		}
	}

	// Load and parse all "extra" CLI arguments.
	for _, ex := range vals.extra {
		sp, _ := shellquote.Split(ex)
		args = append(args, sp...)
	}
	return args
}
