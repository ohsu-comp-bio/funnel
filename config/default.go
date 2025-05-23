package config

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/units"
	intern "github.com/ohsu-comp-bio/funnel/config/internal"
	"github.com/ohsu-comp-bio/funnel/logger"
)

// DefaultConfig returns configuration with simple defaults.
func DefaultConfig() Config {
	cwd, _ := os.Getwd()
	workDir := path.Join(cwd, "funnel-work-dir")

	allowedDirs := []string{cwd}
	if os.Getenv("HOME") != "" {
		allowedDirs = append(allowedDirs, os.Getenv("HOME"))
	}
	if os.Getenv("TMPDIR") != "" {
		allowedDirs = append(allowedDirs, os.Getenv("TMPDIR"))
	}

	server := Server{
		HostName:         "localhost",
		HTTPPort:         "8000",
		RPCPort:          "9090",
		ServiceName:      "Funnel",
		DisableHTTPCache: true,
		TaskAccess:       "All",
	}

	c := Config{
		Compute:      "local",
		Database:     "boltdb",
		EventWriters: []string{"log"},
		// funnel components
		Server: server,
		RPCClient: RPCClient{
			ServerAddress: server.RPCAddress(),
			Timeout:       Duration(time.Second * 60),
			MaxRetries:    10,
		},
		Scheduler: Scheduler{
			ScheduleRate:    Duration(time.Second),
			ScheduleChunk:   10,
			NodePingTimeout: Duration(time.Minute),
			NodeInitTimeout: Duration(time.Minute * 5),
			NodeDeadTimeout: Duration(time.Minute * 5),
		},
		Node: Node{
			Timeout:    -1,
			UpdateRate: Duration(time.Second * 5),
			Metadata:   map[string]string{},
		},
		Worker: Worker{
			WorkDir:              workDir,
			PollingRate:          Duration(time.Second * 5),
			LogUpdateRate:        Duration(time.Second * 5),
			LogTailSize:          10000,
			MaxParallelTransfers: 10,
			// `docker run` command flags
			// https://docs.docker.com/reference/cli/docker/container/run/
			Container: ContainerConfig{
				DriverCommand: "docker",
				RunCommand: "run -i --read-only " +
					// Remove container after it exits
					"{{if .RemoveContainer}}--rm{{end}} " +

					// Environment variables
					"{{range $k, $v := .Env}}--env {{$k}}={{$v}} {{end}} " +

					// Tags/Labels
					"{{range $k, $v := .Tags}}--label {{$k}}={{$v}} {{end}} " +

					// Container Name
					"{{if .Name}}--name {{.Name}}{{end}} " +

					// Workdir
					"{{if .Workdir}}--workdir {{.Workdir}}{{end}} " +

					// Volumes
					"{{range .Volumes}}--volume {{.HostPath}}:{{.ContainerPath}}:{{if .Readonly}}ro{{else}}rw{{end}} {{end}} " +

					// Image and Command
					"{{.Image}} {{.Command}}",
				PullCommand: "pull {{.Image}}",
				StopCommand: "rm -f {{.Name}}",
			},
		},
		Logger: logger.DefaultConfig(),
		// databases / event handlers
		BoltDB: BoltDB{
			Path: path.Join(workDir, "funnel.db"),
		},
		Badger: Badger{
			Path: path.Join(workDir, "funnel.badger.db"),
		},
		DynamoDB: DynamoDB{
			TableBasename: "funnel",
		},
		Elastic: Elastic{
			URL:         "http://localhost:9200",
			IndexPrefix: "funnel",
		},
		MongoDB: MongoDB{
			Addrs:    []string{"localhost"},
			Timeout:  Duration(time.Minute * 5),
			Database: "funnel",
		},
		Kafka: Kafka{
			Topic: "funnel",
		},
		// storage
		LocalStorage: LocalStorage{
			AllowedDirs: allowedDirs,
		},
		HTTPStorage: HTTPStorage{
			Timeout: Duration(time.Second * 60),
		},
		FTPStorage: FTPStorage{
			Timeout:  Duration(time.Second * 10),
			User:     "anonymous",
			Password: "anonymous",
		},
		AmazonS3: AmazonS3Storage{
			AWSConfig: AWSConfig{
				MaxRetries: 10,
			},
		},
		Swift: SwiftStorage{
			MaxRetries:     20,
			ChunkSizeBytes: int64(500 * units.MB),
		},
	}

	// compute
	reconcile := Duration(time.Minute * 10)

	htcondorTemplate := intern.MustAsset("config/htcondor-template.txt")
	c.HTCondor.Template = string(htcondorTemplate)
	c.HTCondor.ReconcileRate = reconcile
	c.HTCondor.DisableReconciler = true

	slurmTemplate := intern.MustAsset("config/slurm-template.txt")
	c.Slurm.Template = string(slurmTemplate)
	c.Slurm.ReconcileRate = reconcile
	c.Slurm.DisableReconciler = true

	pbsTemplate := intern.MustAsset("config/pbs-template.txt")
	c.PBS.Template = string(pbsTemplate)
	c.PBS.ReconcileRate = reconcile
	c.PBS.DisableReconciler = true

	geTemplate := intern.MustAsset("config/gridengine-template.txt")
	c.GridEngine.Template = string(geTemplate)

	c.AWSBatch.JobDefinition = "funnel-job-def"
	c.AWSBatch.JobQueue = "funnel-job-queue"
	c.AWSBatch.ReconcileRate = reconcile
	c.AWSBatch.DisableReconciler = true

	kubernetesTemplate := intern.MustAsset("config/kubernetes-template.yaml")
	executorTemplate := intern.MustAsset("config/kubernetes-executor-template.yaml")
	pvTemplate := intern.MustAsset("config/kubernetes-pv.yaml")
	pvcTemplate := intern.MustAsset("config/kubernetes-pvc.yaml")
	c.Kubernetes.Executor = "docker"
	c.Kubernetes.Namespace = "default"
	c.Kubernetes.ServiceAccount = "funnel-sa"
	c.Kubernetes.Template = string(kubernetesTemplate)
	c.Kubernetes.ExecutorTemplate = string(executorTemplate)
	c.Kubernetes.Bucket = ""
	c.Kubernetes.Region = ""
	c.Kubernetes.PVTemplate = string(pvTemplate)
	c.Kubernetes.PVCTemplate = string(pvcTemplate)
	c.Kubernetes.ReconcileRate = reconcile

	return c
}

var examples = buildExamples()

func buildExamples() map[string]string {
	examples := map[string]string{}
	for _, n := range intern.AssetNames() {
		sn := path.Base(n)
		sn = strings.TrimSuffix(sn, path.Ext(sn))
		b := intern.MustAsset(n)
		examples[sn] = string(b)
	}
	return examples
}

// Examples returns a set of example configurations strings.
func Examples() map[string]string {
	return examples
}
