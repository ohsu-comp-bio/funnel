package run

import (
  "context"
	"funnel/logger"
	"github.com/spf13/cobra"
  "fmt"
  "os"
  "funnel/proto/tes"
  "github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
  "time"
)

var log = logger.New("run")

// TODO figure out a nice default for name
var name = "Funnel run"
var workdir = "/opt/funnel"
var server = "localhost:9090"
var project, description, varsFile string
var printTask, preemptible, wait bool
// TODO validate resources
var cpu int
var ram float64
var disk float64 = 10.0
var zones []string

// TODO allow outputs to be defined when they don't fit into the command
//      such as bam file index secondary file
// TODO with input contents, script could be loaded from file
// TODO load vars from file
// TODO is there a case for glob inputs?
// TODO better volumes? size requirement?
// TODO don't bother with specific outputs? Just upload entire working directory by default?

var examples string
/*
TODO
Examples:
  funnel run ubuntu 'md5sum {in .SRC name:"Input path"} {.DST | stdout}' SRC: ~/input.txt DST: md5sum.txt"

  tes template 'md5sum' --stdin ~/input.txt --stdout gs://bkt/output.txt

  tes tpl --script runscript.sh --in input ~/input.txt --out output gs://bkt/results.txt

  funnel task template  \
    'bowtie2 -f $factor -x $other -p1 $pair1 -p2 $pair2 -o $alignments' \
    --in pair1 ~/pair1.fastq                        \
    --in pair2 ~/pair2.fastq                        \
    --contents other ~/input.txt                    \
    --name pair1 'Pair 1 inputs'                    \
    --desc pair1 'Pair 1 description'               \
    --type pair1 directory                          \
    --out alignments gs://bkt/bowtie/alignments.bam \
    --factor 5                                      \
    --image my-bowtie2                              \
    --name 'Bowtie2 test'                           \
    --volume /tmp                                   \
    --volume /opt                                   \
    --tag 'foo: bar'                                \
    --description 'Testings an example of using `funnel run` for a bowtie2 command'

  funnel run <<TPL
  funnel task template 

  align=$(cat <<TPL
    'bowtie2 -f $factor -x $other -p1 $pair1 -p2 $pair2 -o $alignments' 
    -i pair1 ~/pair1.fastq            
    -i pair2 ~/pair2.fastq           
    -o alignments gs://bkt/bowtie/alignments.bam
    -c other ~/input.txt        
    -n pair1 'Pair 1 inputs'        
    -d pair1 'Pair 1 description'   
    -t pair1 directory             
    -e factor 5
    -m my-bowtie2
    -n 'Bowtie2 test'
    -v /tmp /opt
    -l 'foo: bar'
    -d 'Testings an example of using `funnel run` for a bowtie2 command'
  TPL)



  align_tpl=$(cat <<TPL
    'bowtie2 -f $factor -x $other -p1 $pair1 -p2 $pair2 -o $alignments' 
    -c other ~/input.txt        
    -n pair1 'Pair 1 inputs'        
    -d pair1 'Pair 1 description'   
    -t pair1 directory             
    -e factor 5
    -m my-bowtie2
    -n 'Bowtie2 test'
    -v /tmp /opt
    -l 'foo: bar'
    -d 'Testings an example of using `funnel run` for a bowtie2 command'
  TPL)

  align_sample_1=$(cat <<TPL
    -i pair1 ~/sample1/pair1.fastq            
    -i pair2 ~/sample1/pair2.fastq           
    -o alignments gs://bkt/bowtie/sample1/alignments.bam
  TPL)

  align_sample_2=$(cat <<TPL
    -i pair1 ~/sample2/pair1.fastq            
    -i pair2 ~/sample2/pair2.fastq           
    -o alignments gs://bkt/bowtie/sample2/alignments.bam
  TPL)

  sample_1_task=$(funnel run $align_tpl $align_sample_1)
  sample_2_task=$(funnel run $align_tpl $align_sample_2)

  funnel wait $sample_1_task $sample_2_task

  sample_align_tasks=$( for sample in $(ls -1 samples/); \
    funnel run $align_tpl \
      -i pair1 "samples/$sample/pair1.fastq" \
      -i pair2 "samples/$sample/pair2.fastq" \
      -o alignments "gs://bkt/bowtie/$sample/alignments.bam"; \
  done)

  funnel wait $sample_align_tasks


  funnel run <<TPL
    --cpus 10
    --preemptible
    --ram 150GB
    --disk 1TB
    --zones west,east
  TPL


  funnel task template --yaml <<TPL
    name: My task
    description: My task description
    resources:
    inputs:
      - name: Input one
        url: {{ .input1 }}
        path: /tmp/input.txt
        type: FILE
        contents: $(cat ~/other.txt)
    cmd: md5sum /tmp/input.txt
    stdout: STDOUT
    stderr: STDERR
    outputs:
      - name: Output
        url: {{ .output1 }}
  TPL

  funnel task create md5-task.json
*/

var Cmd = &cobra.Command{
	Use:   "run [flags] IMAGE TEMPLATE -- VARS",
	Short: "Run a task.\n" + examples,
	Long:  ``,
	Run: run,
}

func init() {
  f := Cmd.Flags()
  f.StringVar(&name, "name", name, "Task name")
  f.StringVar(&description, "description", description, "Task description")
  f.StringVar(&project, "project", project, "Project")
  f.StringVar(&workdir, "workdir", workdir, "Set the containter working directory")
  f.IntVar(&cpu, "cpu", cpu, "Number of CPUs requested")
  f.Float64Var(&ram, "ram", ram, "Amount of RAM requested")
  f.Float64Var(&disk, "disk", disk, "Amount of disk space requested (in GB)")
  f.BoolVar(&preemptible, "preemptible", preemptible, "Allow task to be scheduled on preemptible workers")
  f.StringSliceVar(&zones, "zones", zones, "Require task be scheduled in certain zones")
  f.StringVar(&varsFile, "vars-file", varsFile, "Read vars a file")
  f.StringVar(&server, "server", server, "Address of Funnel server")
  f.BoolVar(&printTask, "print", printTask, "Print the task, instead of running it")
  f.BoolVar(&wait, "wait", wait, "Wait for the task to finish before exiting")
}


func run(cmd *cobra.Command, args []string) {

  dashIdx := cmd.Flags().ArgsLenAtDash()
  if dashIdx != -1 && dashIdx != 2 {
    cmd.Usage()
    os.Exit(2)
    return
  }

  // TODO validate
  image := args[0]
  rawcmd := args[1]
  rawCliVars := []string{}

  if dashIdx != -1 {
    rawCliVars = args[dashIdx:]
  }

  log.Debug("CLIVARS", rawCliVars)

  // Get template variables from the command line and variables file.
  cliVars, err := parseCliVars(rawCliVars)
  fileVars, err := parseFileVars(varsFile)
  vars, err := mergeVars(cliVars, fileVars)

  checkErr(err)


  checkErr(err)

  // Build task input parameters
  inputs := []*tes.TaskParameter{}
  for path, url := range res.Inputs {
    input := &tes.TaskParameter{
    }
    inputs = append(inputs, input)
  }

  // Build task output parameters
  outputs := []*tes.TaskParameter{}
  for path, url := range res.Outputs {
    output := &tes.TaskParameter{}
    outputs = append(outputs, output)
  }

  // Build the task message
  task := &tes.Task{
    Name: name,
    Project: project,
    Description: description,
    Inputs: inputs,
    Outputs: outputs,
    Resources: &tes.Resources{
      CpuCores: uint32(cpu),
      RamGb: ram,
      Zones: zones,
      Preemptible: preemptible,
    },
    Executors: []*tes.Executor{
      {
        ImageName: image,
        Cmd: res.Cmd,
        Workdir: workdir,
        // TODO
        Stdin: "",
        Stdout: "",
        Stderr: "",
        // TODO no ports
        Ports: nil,
      },
    },
  }
  log.Debug("TASK", task)

  // Marshal message to JSON and print
  if printTask {
    mar := jsonpb.Marshaler{
      EmitDefaults: true,
      Indent: "\t",
    }
    s, merr := mar.MarshalToString(task)
    checkErr(merr)
    fmt.Println(s)
    return
  }

  c, cerr := newClient(server)
  checkErr(cerr)
  resp, rerr := c.CreateTask(context.TODO(), task)
  checkErr(rerr)

  taskID := resp.Id
  fmt.Println(taskID)

  if wait {
    c.waitForTask(taskID)
    // TODO print/log result
    // TODO stream logs while waiting
  }
}

func checkErr(err error) {
  if err != nil {
    fmt.Println(err)
    os.Exit(3)
  }
}

type client struct {
	tes.TaskServiceClient
	conn *grpc.ClientConn
}

func (c *client) waitForTask(taskID string) {
  for range time.NewTicker(time.Second * 2).C {
    // TODO handle error
    r, _ := c.GetTask(context.TODO(), &tes.GetTaskRequest{Id: taskID})
    switch r.State {
    case tes.State_COMPLETE, tes.State_ERROR, tes.State_SYSTEM_ERROR, tes.State_CANCELED:
      return
    }
  }
}

func newClient(address string) (*client, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		log.Error("Couldn't open RPC connection",
			"error", err,
			"address", address,
		)
		return nil, err
	}

	if err != nil {
		log.Error("Couldn't connect to server", err)
		return nil, err
	}

	s := tes.NewTaskServiceClient(conn)
	return &client{s, conn}, nil
}

// Close closes the client connection.
func (client *client) Close() {
	client.conn.Close()
}
