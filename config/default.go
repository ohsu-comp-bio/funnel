package config

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"os"
	"path"
	"time"
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
		RPCClientTimeout:    time.Second * 60,
		RPCClientMaxRetries: 10,
	}

	c := Config{
		Compute:      "local",
		Database:     "boltdb",
		EventWriters: []string{"log"},
		// funnel components
		Server: server,
		Scheduler: Scheduler{
			ScheduleRate:    time.Second,
			ScheduleChunk:   10,
			NodePingTimeout: time.Minute,
			NodeInitTimeout: time.Minute * 5,
			NodeDeadTimeout: time.Minute * 5,
		},
		Node: Node{
			Timeout:    -1,
			UpdateRate: time.Second * 5,
			Metadata:   map[string]string{},
		},
		Worker: Worker{
			WorkDir:    workDir,
			UpdateRate: time.Second * 5,
			BufferSize: 10000,
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
			Timeout:  time.Minute * 5,
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
			Timeout: time.Second * 60,
		},
		AmazonS3: AmazonS3Storage{
			AWSConfig: AWSConfig{
				MaxRetries: 10,
			},
		},
		Swift: SwiftStorage{
			MaxRetries: 3,
		},
	}

	// compute
	htcondorTemplate, _ := Asset("config/htcondor-template.txt")
	slurmTemplate, _ := Asset("config/slurm-template.txt")
	pbsTemplate, _ := Asset("config/pbs-template.txt")
	geTemplate, _ := Asset("config/gridengine-template.txt")

	c.HTCondor.Template = string(htcondorTemplate)
	c.Slurm.Template = string(slurmTemplate)
	c.PBS.Template = string(pbsTemplate)
	c.GridEngine.Template = string(geTemplate)

	c.AWSBatch.JobDefinition = "funnel-job-def"
	c.AWSBatch.JobQueue = "funnel-job-queue"

	return c
}
