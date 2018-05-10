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
		HostName:            "localhost",
		HTTPPort:            "8000",
		RPCPort:             "9090",
		ServiceName:         "Funnel",
		DisableHTTPCache:    true,
		RPCClientTimeout:    Duration(time.Second * 60),
		RPCClientMaxRetries: 10,
	}

	c := Config{
		Compute:      "local",
		Database:     "boltdb",
		EventWriters: []string{"log"},
		// funnel components
		Server: server,
		Scheduler: Scheduler{
			DBPath:          path.Join(workDir, "scheduler.db"),
			NodePingTimeout: Duration(time.Minute),
			NodeDeadTimeout: Duration(time.Minute * 5),
		},
		Node: Node{
			UpdateRate: Duration(time.Second * 5),
			Metadata:   map[string]string{},
		},
		Worker: Worker{
			WorkDir:       workDir,
			PollingRate:   Duration(time.Second * 5),
			LogUpdateRate: Duration(time.Second * 5),
			LogTailSize:   10000,
		},
		Logger: logger.DefaultConfig(),
		// databases / event handlers
		BoltDB: BoltDB{
			Path: path.Join(workDir, "funnel.db"),
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
		AmazonS3: AmazonS3Storage{
			AWSConfig: AWSConfig{
				MaxRetries: 10,
			},
		},
		Swift: SwiftStorage{
			MaxRetries:     3,
			ChunkSizeBytes: int64(500 * units.MB),
		},
	}

	// compute
	reconcile := Duration(time.Minute * 10)

	htcondorTemplate, _ := intern.Asset("config/htcondor-template.txt")
	c.HTCondor.Template = string(htcondorTemplate)
	c.HTCondor.ReconcileRate = reconcile
	c.HTCondor.DisableReconciler = true

	slurmTemplate, _ := intern.Asset("config/slurm-template.txt")
	c.Slurm.Template = string(slurmTemplate)
	c.Slurm.ReconcileRate = reconcile
	c.Slurm.DisableReconciler = true

	pbsTemplate, _ := intern.Asset("config/pbs-template.txt")
	c.PBS.Template = string(pbsTemplate)
	c.PBS.ReconcileRate = reconcile
	c.PBS.DisableReconciler = true

	geTemplate, _ := intern.Asset("config/gridengine-template.txt")
	c.GridEngine.Template = string(geTemplate)

	c.AWSBatch.JobDefinition = "funnel-job-def"
	c.AWSBatch.JobQueue = "funnel-job-queue"
	c.AWSBatch.ReconcileRate = reconcile
	c.AWSBatch.DisableReconciler = true

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
