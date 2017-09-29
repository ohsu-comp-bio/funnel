package batch

// Config represents configuration of the AWS proxy, including
// the compute environment, job queue, and base job definition.
type Config struct {
	Region     string
	Key        string
	Secret     string
	ComputeEnv ComputeEnvConfig
	JobQueue   JobQueueConfig
}

// ComputeEnvConfig represents configuration of the AWS Batch
// Compute Environment.
type ComputeEnvConfig struct {
	Name             string
	MinVCPUs         int64
	MaxVCPUs         int64
	SecurityGroupIds []string
	Subnets          []string
	Tags             map[string]string
	ServiceRole      string
	InstanceRole     string
	InstanceTypes    []string
}

// JobQueueConfig represents configuration of the AWS Batch
// Job Queue.
type JobQueueConfig struct {
	Name        string
	ComputeEnvs []string
}

// DefaultConfig returns default configuration of AWS.
func DefaultConfig() Config {
	return Config{
		Region: "us-west-2",
		ComputeEnv: ComputeEnvConfig{
			Name:          "funnel-compute-environment",
			InstanceRole:  "ecsInstanceRole",
			InstanceTypes: []string{"optimal"},
			MinVCPUs:      0,
			MaxVCPUs:      256,
			Tags: map[string]string{
				"Name": "Funnel",
			},
		},
		JobQueue: JobQueueConfig{
			Name: "funnel-job-queue",
			ComputeEnvs: []string{
				"funnel-compute-environment",
			},
		},
	}
}
