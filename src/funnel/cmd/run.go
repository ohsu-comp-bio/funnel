package cmd

import (
	"funnel/logger"
	"github.com/spf13/cobra"
  "fmt"
  "bytes"
  "os"
  "text/template"
  "strings"
)

var log = logger.New("funnel-run")
var zones []string
var name, project, description, tpl string
var dryrun, preemptible, wait, interactive bool
var cpu int
var ram float64
// TODO allow outputs to be defined when they don't fit into the command
// TODO workdir
// TODO with input contents, script could be loaded from file
// TODO what is stdout/err of funnel run?
//      should have job id to access job state later
// TODO load vars from file
// TODO scatter over many vars files
// TODO is there a case for glob inputs?
// TODO volumes? size requirement?

var examples string = `
Examples:
  funnel run ubuntu 'md5sum {in .SRC} > {out .DST}' -- INFILE=~/input.txt OUTFILE=md5sum.txt",
  funnel run ubuntu 'ls {dir .P | name "IN PATH" | desc "IN DESCRIPTION" | path "/tmp/in"  | create } > {out .O | path "/tmp/out" }
`

var cmd = &cobra.Command{
	Use:   "run [flags] IMAGE TEMPLATE -- VARS",
	Short: "Run a task.\n" + examples,
	Long:  ``,
	Run: run,
}

func init() {
	RootCmd.AddCommand(cmd)
  f := cmd.Flags()
  f.StringVar(&name, "name", name, "Task name")
  f.StringVar(&description, "description", description, "Task description")
  f.StringVar(&project, "project", project, "Project")
  f.StringVar(&tpl, "tpl", tpl, "Read task template from the given path")
  f.IntVar(&cpu, "cpu", cpu, "Number of CPUs required")
  f.Float64Var(&ram, "ram", ram, "Amount of RAM required")
  f.BoolVar(&preemptible, "preemptible", preemptible, "Allow task to be scheduled on preemptible workers")
  f.StringSliceVar(&zones, "zones", zones, "Require task be scheduled in certain zones")
  f.BoolVar(&wait, "wait", wait, "Wait for task to complete before exiting")
  f.BoolVar(&dryrun, "dry-run", dryrun, "Print task JSON only, do not run task")
  f.BoolVarP(&interactive, "interactive", "i", interactive, "Interactive prompt helps define the command template, vars, and other config outside of the shell environment.")
}

type builder struct {
}

func (b *builder) In(args ...string) string {
  fmt.Println("INNNNNNNN", args)
  return args[0]
}

func (b *builder) Out(args ...string) string {
  return args[0]
}

func parseVars(args []string) (map[string]string, error) {
  data := map[string]string{}

  if len(args) == 0 {
    return data, nil
  }

  key := ""
  mode := "key"

  for _, arg := range args {
    if mode == "key" {
      if !strings.HasPrefix(arg, "-") {
        return nil, fmt.Errorf("Unexpected value. Expected key (e.g. '-key' or '--key')")
      }
      key = strings.TrimLeft(arg, "-")
      mode = "value"
    } else {
      if strings.HasPrefix(arg, "-") {
        return nil, fmt.Errorf("Unexpected key. Expected value for key '%s'", key)
      }
      data[key] = arg
      mode = "key"
    }
  }

  if mode == "value" {
    return nil, fmt.Errorf("No value found for key: '%s'", key)
  }

  return data, nil
}

func run(cmd *cobra.Command, args []string) {

  before := cmd.Flags().ArgsLenAtDash()
  if before != 2 {
    cmd.Usage()
    os.Exit(2)
    return
  }

  //image := args[0]
  rawcmd := args[1]

  data, _ := parseVars(args[before:])
  fmt.Println("DATA", data)

  b := builder{}
  funcs := template.FuncMap{
    "in": b.In,
    "out": b.Out,
    "name": func(args ...string) string {
      fmt.Println(":LWKJEFLEWJFWELKFJ", args)
      return "TNALKJ"
    },
  }

  t, err := template.New("cmd").
    Delims("{", "}").
    Funcs(funcs).
    Parse(rawcmd)

  if err != nil {
    fmt.Println("Invalid template: %s", err)
    os.Exit(3)
  }

  buf := &bytes.Buffer{}
  eerr := t.Execute(buf, data)
  if eerr != nil {
    fmt.Println("ERROR EXEUC")
    os.Exit(4)
  }
  fmt.Println(buf.String())
}
