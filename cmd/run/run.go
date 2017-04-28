package run

import (
	"errors"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
)

var log = logger.New("run")

// TODO figure out a nice default for name
var name = "Funnel Run Task"
var workdir = "/opt/funnel"
var server = "localhost:9090"
var image, project, description, varsFile string
var printTask, preemptible, wait bool
var cliInputs, cliOutputs, cliEnvVars, tags, volumes []string

// TODO validate resources
var cpu int
var ram float64
var disk float64
var zones []string

// TODO with input contents, script could be loaded from file
// TODO load vars from file
// TODO is there a case for glob inputs?
// TODO don't bother with specific outputs? Just upload entire working directory by default?

var examples = `
Run a task.

Example:
  funnel run <<TPL
    'bowtie2 -f $factor -x $other -p1 $pair1 -p2 $pair2 -o $alignments'
    --image opengenomics/bowtie2:latest
    --name 'Bowtie2 test'
    --description 'Testings an example of using 'funnel run' for a bowtie2 command'
    --in pair1=file://~/pair1.fastq
    --in pair2=file://~/pair2.fastq
    --out alignments=gs://bkt/bowtie/alignments.bam
    --env factor=5
    --vol /tmp 
    --vol /opt
    --cpu 8
    --ram 32
    --disk 100
  TPL
`

// Cmd represents the run command
var Cmd = &cobra.Command{
	Use:   "run [flags] --image IMAGE CMD",
	Short: "Run a task.\n",
	Long:  examples,
	Run:   run,
}

func init() {
	f := Cmd.Flags()
	f.StringVar(&name, "name", name, "Task name")
	f.StringVar(&description, "description", description, "Task description")
	f.StringVar(&project, "project", project, "Project")
	f.StringVar(&workdir, "workdir", workdir, "Set the containter working directory")
	f.StringVar(&image, "image", image, "Specify the containter image")
	f.StringSliceVar(&cliInputs, "in", cliInputs, "A key-value map of input files")
	f.StringSliceVar(&cliOutputs, "out", cliOutputs, "A key-value map of output files")
	f.StringSliceVar(&cliEnvVars, "env", cliEnvVars, "A key-value map of enviromental variables")
	f.StringSliceVar(&volumes, "vol", volumes, "Volumes to be defined on the container")
	f.StringSliceVar(&tags, "tag", tags, "A key-value map of arbitrary tags")
	f.IntVar(&cpu, "cpu", cpu, "Number of CPUs requested")
	f.Float64Var(&ram, "ram", ram, "Amount of RAM requested (in GB)")
	f.Float64Var(&disk, "disk", disk, "Amount of disk space requested (in GB)")
	f.BoolVar(&preemptible, "preemptible", preemptible, "Allow task to be scheduled on preemptible workers")
	f.StringSliceVar(&zones, "zones", zones, "Require task be scheduled in certain zones")
	f.StringVar(&varsFile, "vars-file", varsFile, "Read vars a file")
	f.StringVar(&server, "server", server, "Address of Funnel server")
	f.BoolVar(&printTask, "print", printTask, "Print the task, instead of running it")
	f.BoolVar(&wait, "wait", wait, "Wait for the task to finish before exiting")
}

func run(cmd *cobra.Command, args []string) {

	if len(args) != 1 {
		fmt.Printf("ERROR - You must specify a command to run\n\n")
		cmd.Usage()
		os.Exit(2)
		return
	}

	if image == "" {
		fmt.Printf("ERROR - You must specify an image to run your command in\n\n")
		cmd.Usage()
		os.Exit(2)
		return
	}

	// TODO validate cmd
	rawcmd := args[0]
	executorCmd := []string{"bash", "-c", rawcmd}
	log.Debug("Command:", "cmd", executorCmd)

	// Get template variables from the command line.
	// TODO variables file.
	inputsMap, err := parseCliVars(cliInputs)
	checkErr(err)
	outputsMap, err := parseCliVars(cliOutputs)
	checkErr(err)
	envVarsMap, err := parseCliVars(cliEnvVars)
	checkErr(err)
	tagsMap, err := parseCliVars(tags)
	checkErr(err)

	// check for key collisions
	err = compareKeys(inputsMap, outputsMap, envVarsMap)
	checkErr(err)

	// Create map of enviromental variables to be passed to the executor
	inputEnvVars, err := fileMapToEnvVars(inputsMap, "/opt/funnel/inputs/")
	checkErr(err)
	outputEnvVars, err := fileMapToEnvVars(outputsMap, "/opt/funnel/outputs/")
	checkErr(err)
	environ, err := mergeVars(inputEnvVars, outputEnvVars, envVarsMap)
	checkErr(err)

	// Build task input parameters
	inputs, err := createTaskParams(inputsMap, "/opt/funnel/inputs/")
	checkErr(err)

	// Build task output parameters
	outputs, err := createTaskParams(outputsMap, "/opt/funnel/outputs/")
	checkErr(err)

	// Build the task message
	task := &tes.Task{
		Name:        name,
		Project:     project,
		Description: description,
		Inputs:      inputs,
		Outputs:     outputs,
		Resources: &tes.Resources{
			CpuCores:    uint32(cpu),
			RamGb:       ram,
			SizeGb:      disk,
			Zones:       zones,
			Preemptible: preemptible,
		},
		Executors: []*tes.Executor{
			{
				ImageName: image,
				Cmd:       executorCmd,
				Environ:   environ,
				Workdir:   workdir,
				Stdin:     "",
				Stdout:    "/opt/funnel/outputs/stdout",
				Stderr:    "/opt/funnel/outputs/stderr",
				// TODO no ports
				Ports: nil,
			},
		},
		Volumes: volumes,
		Tags:    tagsMap,
	}
	log.Debug("Task", "taskmsg", task)

	cli := client.NewClient(server)
	// Marshal message to JSON
	taskJSON, merr := cli.Marshaler.MarshalToString(task)
	checkErr(merr)

	if printTask {
		fmt.Println(taskJSON)
		return
	}

	resp, rerr := cli.CreateTask([]byte(taskJSON))
	checkErr(rerr)

	taskID := resp.Id
	fmt.Println(taskID)

	if wait {
		werr := cli.WaitForTask(taskID)
		checkErr(werr)
	}
}

func stripStoragePrefix(url string) (string, error) {
	re := regexp.MustCompile("[a-z3]+://")
	if !re.MatchString(url) {
		err := errors.New("File paths must be prefixed with one of:\n file://\n gs://\n s3://")
		return "", err
	}
	path := re.ReplaceAllString(url, "")
	return strings.TrimPrefix(path, "/"), nil
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}
