package pipelines

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kr/pretty"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tes"
	"google.golang.org/api/genomics/v2alpha1"
)

// Reconcile loops through tasks and checks the status from Funnel's database
// against the status reported by AWS Batch. This allows the backend to report
// system error's that prevented the worker process from running.
//
// Currently this handles a narrow set of cases:
//
// |---------------------|-----------------|--------------------|
// |    Funnel State     |  Backend State  |  Reconciled State  |
// |---------------------|-----------------|--------------------|
// |        QUEUED       |     FAILED      |    SYSTEM_ERROR    |
// |  INITIALIZING       |     FAILED      |    SYSTEM_ERROR    |
// |       RUNNING       |     FAILED      |    SYSTEM_ERROR    |
//
// In this context a "FAILED" state is being used as a generic term that captures
// one or more terminal states for the backend.
func (b *Backend) reconcile(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(b.conf.ReconcileRate))

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			pageToken := ""
			states := []tes.State{tes.Queued, tes.Initializing, tes.Running}
			for _, s := range states {
				for {
					resp, _ := b.database.ListTasks(ctx, &tes.ListTasksRequest{
						View:      tes.TaskView_BASIC,
						State:     s,
						PageSize:  100,
						PageToken: pageToken,
					})
					pageToken = resp.NextPageToken

					for _, task := range resp.Tasks {
						logger.Debug("Reconcile task", task.Id)

						jobid, ok := task.GetTaskLog(0).GetMetadata()["pipeline_operation_id"]
						if !ok {
							continue
						}

						job, err := b.client.Projects.Operations.Get(jobid).Context(ctx).Do()
						if err != nil {
							logger.Debug("ERROR", err)
							continue
						}
						event := events.NewTaskWriter(task.Id, 0, b.event)

						if job.Done {
							if job.Error != nil {
								// TODO distinguish between system vs exec error
								//      worker should be doing this.
								event.State(tes.SystemError)
								event.Error("System error", "error", job.Error.Message)
								for _, raw := range job.Error.Details {
									deet := map[string]interface{}{}
									json.Unmarshal(raw, &deet)
									pretty.Println(deet)
								}

								meta := &genomics.OperationMetadata{}
								json.Unmarshal(job.Metadata, meta)
								for _, ev := range meta.Events {
									event.Info("Pipelines event", "description", ev.Description)
								}

							} else {
								event.State(tes.Complete)
							}
						}
						time.Sleep(time.Millisecond * 100)
					}

					// continue to next page from ListTasks or break
					if pageToken == "" {
						break
					}
				}
			}
		}
	}
}
