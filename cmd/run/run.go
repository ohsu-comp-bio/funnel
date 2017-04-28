package run

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/spf13/cobra"
	"os"
)

var log = logger.New("run")

// TODO figure out a nice default for name
var name = "Funnel Run Task"
var workdir = "/opt/funnel"
var server = "http://localhost:8000"
var image, project, description, stdin, stdout, stderr string
var printTask, preemptible, wait bool
var cliInputs, cliInputDirs, cliOutputs, cliOutputDirs, cliEnvVars, tags, volumes []string

// TODO validate resources
var cpu int
var ram float64
var disk float64
var zones []string

// TODO with input contents, script could be loaded from file
// TODO load vars from file
// TODO is there a case for glob inputs?

var example = `
    funnel run
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
`

// Cmd represents the run command
var Cmd = &cobra.Command{
	Use:     "run [flags] --image IMAGE CMD",
	Short:   "Run a task.\n",
	Long:    ``,
	Example: example,
	Run:     run,
}

func init() {
	f := Cmd.Flags()
	f.StringVarP(&name, "name", "n", name, "Task name")
	f.StringVar(&description, "description", description, "Task description")
	f.StringVar(&project, "project", project, "Project")
	f.StringVarP(&workdir, "workdir", "w", workdir, "Set the containter working directory")
	f.StringVarP(&image, "container", "c", image, "Specify the containter image")
	f.StringSliceVarP(&cliInputs, "in", "i", cliInputs, "A key-value map of input files")
	f.StringSliceVarP(&cliInputDirs, "in-dir", "I", cliInputDirs, "A key-value map of input directories")
	f.StringSliceVarP(&cliOutputs, "out", "o", cliOutputs, "A key-value map of output files")
	f.StringSliceVarP(&cliOutputDirs, "out-dir", "O", cliOutputDirs, "A key-value map of output directories")
	f.StringSliceVarP(&cliEnvVars, "env", "e", cliEnvVars, "A key-value map of enviromental variables")
	f.StringVar(&stdin, "stdin", stdin, "File to pass via stdin to the command")
	f.StringVar(&stdout, "stdout", stdout, "File to write the stdout of the command")
	f.StringVar(&stderr, "stderr", stderr, "File to write the stderr of the command")
	f.StringSliceVar(&volumes, "vol", volumes, "Volumes to be defined on the container")
	f.StringSliceVar(&tags, "tag", tags, "A key-value map of arbitrary tags")
	f.IntVar(&cpu, "cpu", cpu, "Number of CPUs requested")
	f.Float64Var(&ram, "ram", ram, "Amount of RAM requested (in GB)")
	f.Float64Var(&disk, "disk", disk, "Amount of disk space requested (in GB)")
	f.BoolVar(&preemptible, "preemptible", preemptible, "Allow task to be scheduled on preemptible workers")
	f.StringSliceVar(&zones, "zone", zones, "Require task be scheduled in certain zones")
	// TODO
	//f.StringVar(&varsFile, "vars-file", varsFile, "Read vars from file")
	//f.StringVar(&cmdFile, "cmd-file", cmdFile, "Read cmd template from file")
	f.StringVarP(&server, "server", "S", server, "Address of Funnel server")
	f.BoolVarP(&printTask, "print", "p", printTask, "Print the task, instead of running it")
	f.BoolVar(&wait, "wait", wait, "Wait for the task to finish before exiting")
}

func run(cmd *cobra.Command, args []string) {

	if len(args) < 1 {
		fmt.Printf("ERROR - You must specify a command to run\n\n")
		cmd.Usage()
		os.Exit(2)
		return
	}

	if len(args) > 1 {
		fmt.Printf("ERROR - extra arg(s) %s\n\n", args[1:])
		fmt.Printf("--in, --out and --env args should have the form 'KEY=VALUE' not 'KEY VALUE'\n\n")
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

	// Get template variables from the command line.
	// TODO variables file.
	inputFileMap, err := parseCliVars(cliInputs)
	checkErr(err)
	inputDirMap, err := parseCliVars(cliInputDirs)
	checkErr(err)
	outputFileMap, err := parseCliVars(cliOutputs)
	checkErr(err)
	outputDirMap, err := parseCliVars(cliOutputDirs)
	checkErr(err)
	envVarMap, err := parseCliVars(cliEnvVars)
	checkErr(err)
	tagsMap, err := parseCliVars(tags)
	checkErr(err)

	// check for key collisions
	err = compareKeys(inputFileMap, inputDirMap, outputFileMap, outputDirMap, envVarMap)
	checkErr(err)

	// Create map of enviromental variables to be passed to the executor
	inputEnvVars, err := fileMapToEnvVars(inputFileMap, "/opt/funnel/inputs/")
	checkErr(err)
	inputDirEnvVars, err := fileMapToEnvVars(inputDirMap, "/opt/funnel/inputs/")
	checkErr(err)
	outputEnvVars, err := fileMapToEnvVars(outputFileMap, "/opt/funnel/outputs/")
	checkErr(err)
	outputDirEnvVars, err := fileMapToEnvVars(outputDirMap, "/opt/funnel/outputs/")
	checkErr(err)
	environ, err := mergeVars(inputEnvVars, inputDirEnvVars, outputEnvVars, outputDirEnvVars, envVarMap)
	checkErr(err)

	// Build task input parameters
	inputs, err := createTaskParams(inputFileMap, "/opt/funnel/inputs/", tes.FileType_FILE)
	checkErr(err)
	inputDirs, err := createTaskParams(inputDirMap, "/opt/funnel/inputs/", tes.FileType_DIRECTORY)
	checkErr(err)
	inputs = append(inputs, inputDirs...)

	// Build task output parameters
	outputs, err := createTaskParams(outputFileMap, "/opt/funnel/outputs/", tes.FileType_FILE)
	checkErr(err)
	outputDirs, err := createTaskParams(outputDirMap, "/opt/funnel/outputs/", tes.FileType_DIRECTORY)
	checkErr(err)
	outputs = append(outputs, outputDirs...)

	stdinPath := ""
	if stdin != "" {
		stdinPath = stdin
	}

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
				Stdin:     stdinPath,
				Stdout:    "/opt/funnel/outputs/stdout",
				Stderr:    "/opt/funnel/outputs/stderr",
				// TODO no ports
				Ports: nil,
			},
		},
		Volumes: volumes,
		Tags:    tagsMap,
	}

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

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
}
