// Package gcp_batch contains code for accessing compute resources via Google Batch.
// ref: https://cloud.google.com/batch/docs
// ref: https://cloud.google.com/batch/docs/reference/rest
package gcp_batch

import (
	"context"
	"fmt"
	"strings"
	"time"

	batch "cloud.google.com/go/batch/apiv1"
	"cloud.google.com/go/batch/apiv1/batchpb"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tes"
)

type Backend struct {
	client   client
	conf     *config.GCPBatch
	event    events.Writer
	database tes.ReadOnlyServer
	log      *logger.Logger
	events.Backend
}

func NewBackend(ctx context.Context, conf *config.GCPBatch, reader tes.ReadOnlyServer, writer events.Writer, log *logger.Logger) (*Backend, error) {
	client, err := batch.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	b := &Backend{
		client:   client,
		conf:     conf,
		event:    writer,
		database: reader,
		log:      log,
	}

	if !conf.DisableReconciler {
		go b.reconcile(ctx)
	}

	return b, nil
}

func (b *Backend) WriteEvent(ctx context.Context, ev *events.Event) error {
	switch ev.Type {
	case events.Type_TASK_CREATED:
		return b.Submit(ev.GetTask())

	case events.Type_TASK_STATE:
		if ev.GetState() == tes.State_CANCELED {
			return b.Cancel(ctx, ev.Id)
		}
	}
	return nil
}

func (b *Backend) Close() {
	b.database.Close()
	b.event.Close()
}

// Example storage interface
func (b *Backend) NewStorage(conf config.Config) (*storage.GoogleCloud, error) {
	gs, nerr := storage.NewGoogleCloud(conf.GoogleStorage)
	if nerr != nil {
		return nil, fmt.Errorf("failed to configure Google Storage backend: %s", nerr)
	}

	return gs, nil
}

// extractBucketName extracts the bucket name from a gs:// URL
func extractBucketName(url string) string {
	if strings.HasPrefix(url, "gs://") {
		parts := strings.SplitN(strings.TrimPrefix(url, "gs://"), "/", 2)
		return parts[0]
	}
	return ""
}

// extractGCSPath extracts bucket and object path from a gs:// URL
func extractGCSPath(url string) (bucket string, objectPath string) {
	if strings.HasPrefix(url, "gs://") {
		urlPath := strings.TrimPrefix(url, "gs://")
		parts := strings.SplitN(urlPath, "/", 2)
		bucket = parts[0]
		if len(parts) > 1 {
			objectPath = parts[1]
		}
	}
	return bucket, objectPath
}

// validatePath checks if a path is safe to use in shell commands
func validatePath(path string) error {
	if path == "" {
		return nil // Empty paths are handled elsewhere
	}

	// Check for dangerous shell metacharacters
	dangerousChars := ";|&$`\n\r<>()"
	if strings.ContainsAny(path, dangerousChars) {
		return fmt.Errorf("path contains dangerous shell metacharacters: %s", path)
	}

	// Ensure path is absolute (starts with /)
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must be absolute (start with /): %s", path)
	}

	return nil
}

// detectPathCollisions checks for multiple inputs/outputs using the same path
func detectPathCollisions(inputs []*tes.Input, outputs []*tes.Output) error {
	seen := make(map[string]string) // path -> url

	for _, input := range inputs {
		if input.Path == "" {
			continue
		}
		if existingURL, exists := seen[input.Path]; exists && existingURL != input.Url {
			return fmt.Errorf("path collision detected: %s used by both %s and %s",
				input.Path, existingURL, input.Url)
		}
		seen[input.Path] = input.Url
	}

	for _, output := range outputs {
		if output.Path == "" {
			continue
		}
		if existingURL, exists := seen[output.Path]; exists && existingURL != output.Url {
			return fmt.Errorf("path collision detected: %s used by both %s and %s",
				output.Path, existingURL, output.Url)
		}
		seen[output.Path] = output.Url
	}

	return nil
}

// checkMixedStorageBackends warns if task contains both GCS and non-GCS URLs
func checkMixedStorageBackends(task *tes.Task, log *logger.Logger) {
	hasGCS := false
	hasNonGCS := false

	for _, input := range task.Inputs {
		if hasGCS && hasNonGCS {
			break
		}
		if input.Url == "" {
			continue
		}
		if strings.HasPrefix(input.Url, "gs://") {
			hasGCS = true
		} else {
			hasNonGCS = true
		}
	}

	for _, output := range task.Outputs {
		if hasGCS && hasNonGCS {
			break
		}
		if output.Url == "" {
			continue
		}
		if strings.HasPrefix(output.Url, "gs://") {
			hasGCS = true
		} else {
			hasNonGCS = true
		}
	}

	if hasGCS && hasNonGCS {
		log.Warn("Task contains mixed storage backends. Non-GCS URLs will be ignored by GCP Batch symlink mapping.",
			"taskID", task.Id)
	}
}

func (b *Backend) Submit(task *tes.Task) error {
	ctx := context.Background()

	// Validate all input and output paths
	for _, input := range task.Inputs {
		if err := validatePath(input.Path); err != nil {
			return fmt.Errorf("invalid input path: %w", err)
		}
	}
	for _, output := range task.Outputs {
		if err := validatePath(output.Path); err != nil {
			return fmt.Errorf("invalid output path: %w", err)
		}
	}

	// Check for path collisions
	if err := detectPathCollisions(task.Inputs, task.Outputs); err != nil {
		return fmt.Errorf("path collision error: %w", err)
	}

	// Warn about mixed storage backends
	checkMixedStorageBackends(task, b.log)

	// 1. Identify all unique buckets used by the task
	buckets := make(map[string]bool)

	for _, input := range task.Inputs {
		if bucket := extractBucketName(input.Url); bucket != "" {
			buckets[bucket] = true
		}
	}

	for _, output := range task.Outputs {
		if bucket := extractBucketName(output.Url); bucket != "" {
			buckets[bucket] = true
		}
	}

	// Mount all buckets to `/mnt/share/<BUCKET>` as volumes in the GCP Job Request
	var volumes []*batchpb.Volume
	for bucketName := range buckets {
		volumes = append(volumes, &batchpb.Volume{
			Source: &batchpb.Volume_Gcs{
				Gcs: &batchpb.GCS{
					RemotePath: bucketName,
				},
			},
			MountPath: fmt.Sprintf("/mnt/disks/%s", bucketName),
		})
	}

	// Runnables
	var runnables []*batchpb.Runnable

	for _, executor := range task.Executors {
		// Build input symlink commands
		var inputCmds []string
		for _, input := range task.Inputs {
			if input.Url == "" || input.Path == "" {
				continue
			}
			bucket, objectPath := extractGCSPath(input.Url)
			if bucket == "" {
				continue // Skip non-GCS URLs
			}

			mountPath := fmt.Sprintf("/mnt/disks/%s/%s", bucket, objectPath)
			// Shell-escape paths for safety
			escapedInputPath := shellquote.Join(input.Path)
			escapedMountPath := shellquote.Join(mountPath)

			// Create parent directory and symlink
			inputCmds = append(inputCmds, fmt.Sprintf("mkdir -p $(dirname %s)", escapedInputPath))
			inputCmds = append(inputCmds, fmt.Sprintf("ln -sf %s %s", escapedMountPath, escapedInputPath))
		}

		// Build output symlink commands
		var outputCmds []string
		for _, output := range task.Outputs {
			if output.Url == "" || output.Path == "" {
				continue
			}
			bucket, objectPath := extractGCSPath(output.Url)
			if bucket == "" {
				continue // Skip non-GCS URLs
			}

			mountPath := fmt.Sprintf("/mnt/disks/%s/%s", bucket, objectPath)
			// Shell-escape paths for safety
			escapedOutputPath := shellquote.Join(output.Path)
			escapedMountPath := shellquote.Join(mountPath)

			// Create parent directory in mount and symlink output path to it
			outputCmds = append(outputCmds, fmt.Sprintf("mkdir -p $(dirname %s)", escapedMountPath))
			outputCmds = append(outputCmds, fmt.Sprintf("mkdir -p $(dirname %s)", escapedOutputPath))
			outputCmds = append(outputCmds, fmt.Sprintf("ln -sf %s %s", escapedMountPath, escapedOutputPath))
		}

		// Build the full command - executor.Command is already a proper command array
		// We need to wrap it as a subshell to execute properly
		var executorCmd string
		if len(executor.Command) > 0 {
			// Use shellquote.Join to properly escape arguments with spaces/quotes
			executorCmd = shellquote.Join(executor.Command...)
		}

		if executor.Stdout != "" {
			// Redirect command output to the specified file path
			executorCmd = fmt.Sprintf("(%s) | tee %s", executorCmd, executor.Stdout)
		} else {
			// Wrap in subshell for proper execution
			executorCmd = fmt.Sprintf("(%s)", executorCmd)
		}

		// Combine: input setup + output setup + executor command
		var fullCmd string
		allCmds := []string{"set -ex"}

		if len(inputCmds) > 0 {
			allCmds = append(allCmds, "echo '=== Setting up input symlinks ==='")
			allCmds = append(allCmds, inputCmds...)
		}

		if len(outputCmds) > 0 {
			allCmds = append(allCmds, "echo '=== Setting up output symlinks ==='")
			allCmds = append(allCmds, outputCmds...)
		}

		allCmds = append(allCmds, "echo '=== Running executor command ==='")
		allCmds = append(allCmds, executorCmd)

		fullCmd = strings.Join(allCmds, " && ")

		runnable := &batchpb.Runnable{
			Executable: &batchpb.Runnable_Container_{
				Container: &batchpb.Runnable_Container{
					ImageUri: executor.Image,
					Commands: []string{"sh", "-c", fullCmd},
				},
			},
		}

		runnables = append(runnables, runnable)
	}

	// Resources
	resources := &batchpb.ComputeResource{}
	if task.Resources != nil {
		resources = &batchpb.ComputeResource{
			CpuMilli:  int64(task.Resources.CpuCores) * 1000,
			MemoryMib: int64(task.Resources.RamGb) * 1024,
		}
	}

	// TaskSpec
	taskSpec := &batchpb.TaskSpec{
		Runnables:       runnables,
		ComputeResource: resources,
		Environment:     &batchpb.Environment{},
		Volumes:         volumes,
	}

	// TaskGroup
	taskGroup := &batchpb.TaskGroup{
		TaskCount: 1,
		TaskSpec:  taskSpec,
	}

	// JobRequest
	req := &batchpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", b.conf.Project, b.conf.Location),
		JobId:  task.Id,
		Job: &batchpb.Job{
			Name:       task.Id,
			Uid:        task.Id,
			TaskGroups: []*batchpb.TaskGroup{taskGroup},
			LogsPolicy: &batchpb.LogsPolicy{
				Destination: batchpb.LogsPolicy_CLOUD_LOGGING,
			},
		},
	}

	logger.Debug("GCP Batch Job Request", "request", req)

	resp, err := b.client.CreateJob(context.Background(), req)

	if err != nil {
		b.event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
		b.event.WriteEvent(
			ctx,
			events.NewSystemLog(
				task.Id, 0, 0, "error",
				"error submitting task to GCP Batch",
				map[string]string{"error": err.Error()},
			),
		)
		return err
	}

	logger.Debug("Submitted task to GCP Batch",
		"taskID", task.Id,
		"gcpbatch_uid", resp.Uid,
		"gcpbatch_name", resp.Name)

	return b.event.WriteEvent(
		ctx, events.NewMetadata(task.Id, 0, map[string]string{
			"gcpbatch_uid":  resp.Uid,
			"gcpbatch_name": resp.Name,
		}),
	)
}

func (b *Backend) Cancel(ctx context.Context, taskID string) error {
	// TODO: Implement
	return nil
}

// Reconciler adapted from aws_batch/backend.go
//
// The reconciler syncs task states between Funnel and GCP Batch:
//  1. List all tasks in QUEUED, INITIALIZING, RUNNING states from the Funnel Database
//  2. For each task, fetch the corresponding GCP Batch Job details
//  3. Extract the dominant task state from the Job's task group status
//  4. Map the GCP Batch task state to the corresponding TES state
//  5. Update the task state in the Funnel Database if it has changed
//
// State mapping:
// |----------------------|---------------------|
// | GCP Batch State      | TES State           |
// |----------------------|---------------------|
// | STATE_UNSPECIFIED    | UNKNOWN             |
// | PENDING              | QUEUED              |
// | ASSIGNED             | INITIALIZING        |
// | RUNNING              | RUNNING             |
// | FAILED               | EXECUTOR_ERROR      |
// | SUCCEEDED            | COMPLETE            |
// | UNEXECUTED           | PREEMPTED           |
// |----------------------|---------------------|
func (b *Backend) reconcile(ctx context.Context) {
	ticker := time.NewTicker(b.conf.ReconcileRate.AsDuration())

ReconcileLoop:
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			pageToken := ""
			states := []tes.State{tes.Queued, tes.Initializing, tes.Running}

			// For all Task states in QUEUED, INITIALIZING, RUNNING
			for _, s := range states {
				for {
					// List Tasks from Funnel Database
					lresp, err := b.database.ListTasks(ctx, &tes.ListTasksRequest{
						View:      tes.View_BASIC.String(),
						State:     s,
						PageSize:  100,
						PageToken: pageToken,
					})
					if err != nil {
						b.log.Error("Calling ListTasks", err)
						continue ReconcileLoop
					}
					pageToken = lresp.NextPageToken

					// Process each task
					for _, task := range lresp.Tasks {
						// Get GCP Batch Job UID from task metadata
						jobUID := getTaskID(task)
						if jobUID == "" {
							continue
						}

						// Construct the full job name
						jobName := fmt.Sprintf("projects/%s/locations/%s/jobs/%s",
							b.conf.Project, b.conf.Location, task.Id)

						// Fetch job details from GCP Batch
						job, err := b.client.GetJob(ctx, &batchpb.GetJobRequest{
							Name: jobName,
						})
						if err != nil {
							b.log.Error("Calling GetJob",
								"taskID", task.Id,
								"jobName", jobName,
								"error", err)
							continue
						}

						// Get the dominant task state from the job
						gcpTaskState := getDominantTaskState(job)

						// Map GCP Batch task state to TES state
						tesState := mapTaskStateToTES(gcpTaskState)

						// Only update if the state has changed
						if tesState != task.State && tesState != tes.Unknown {
							b.log.Info("Updating task state",
								"taskID", task.Id,
								"oldState", task.State,
								"newState", tesState,
								"gcpTaskState", gcpTaskState)

							// Write state change event
							err = b.event.WriteEvent(ctx, events.NewState(task.Id, tesState))
							if err != nil {
								b.log.Error("Writing state event",
									"taskID", task.Id,
									"state", tesState,
									"error", err)
							}
						}
					}

					// continue to next page from ListTasks or break
					if pageToken == "" {
						break
					}
					time.Sleep(time.Millisecond * 100)
				}
			}
		}
	}
}

// Retreives Batch Job ID from Task metadata (created in #Submit) stored in Funnel Database
func getTaskID(task *tes.Task) string {
	logs := task.GetLogs()
	if len(logs) > 0 {
		metadata := logs[0].GetMetadata()
		if metadata != nil {
			return metadata["gcpbatch_uid"]
		}
	}
	return ""
}

// mapTaskStateToTES maps GCP Batch TaskStatus states to TES states
// according to the mapping defined in the issue:
// https://github.com/ohsu-comp-bio/funnel/issues/XXX
func mapTaskStateToTES(gcpState batchpb.TaskStatus_State) tes.State {
	switch gcpState {
	case batchpb.TaskStatus_STATE_UNSPECIFIED:
		return tes.Unknown
	case batchpb.TaskStatus_PENDING:
		return tes.Queued
	case batchpb.TaskStatus_ASSIGNED:
		return tes.Initializing
	case batchpb.TaskStatus_RUNNING:
		return tes.Running
	case batchpb.TaskStatus_FAILED:
		return tes.ExecutorError
	case batchpb.TaskStatus_SUCCEEDED:
		return tes.Complete
	case batchpb.TaskStatus_UNEXECUTED:
		return tes.State_PREEMPTED
	default:
		return tes.Unknown
	}
}

// getDominantTaskState determines the most relevant task state from GCP Batch job status.
// It prioritizes states in the following order: FAILED > SUCCEEDED > RUNNING > ASSIGNED > PENDING > UNEXECUTED
// This ensures that terminal states (FAILED, SUCCEEDED) are reported before in-progress states.
func getDominantTaskState(job *batchpb.Job) batchpb.TaskStatus_State {
	if job.Status == nil || len(job.Status.TaskGroups) == 0 {
		return batchpb.TaskStatus_STATE_UNSPECIFIED
	}

	// Get counts from the first task group (typically there's only one)
	var counts map[string]int64
	for _, tgStatus := range job.Status.TaskGroups {
		counts = tgStatus.Counts
		break
	}

	if counts == nil || len(counts) == 0 {
		return batchpb.TaskStatus_STATE_UNSPECIFIED
	}

	// Priority order: terminal states first, then active states
	// Check for FAILED first
	if counts["FAILED"] > 0 {
		return batchpb.TaskStatus_FAILED
	}

	// Check for SUCCEEDED
	if counts["SUCCEEDED"] > 0 {
		return batchpb.TaskStatus_SUCCEEDED
	}

	// Check for active states
	if counts["RUNNING"] > 0 {
		return batchpb.TaskStatus_RUNNING
	}

	if counts["ASSIGNED"] > 0 {
		return batchpb.TaskStatus_ASSIGNED
	}

	if counts["PENDING"] > 0 {
		return batchpb.TaskStatus_PENDING
	}

	if counts["UNEXECUTED"] > 0 {
		return batchpb.TaskStatus_UNEXECUTED
	}

	return batchpb.TaskStatus_STATE_UNSPECIFIED
}
