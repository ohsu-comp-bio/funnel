package papi

import (
  "bytes"
  "fmt"
  "encoding/base64"
	"github.com/golang/protobuf/jsonpb"
  "context"
	"github.com/ohsu-comp-bio/funnel/tes"
	"google.golang.org/api/genomics/v2alpha1"
	"github.com/ohsu-comp-bio/funnel/events"
)

// Submit submits an operation to the Pipelines service.
func (b *Backend) Submit(task *tes.Task) error {
  ctx := context.Background()
  papiID, err := b.submit(task)
	if err != nil {
		b.event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
		b.event.WriteEvent(
			ctx,
			events.NewSystemLog(
				task.Id, 0, 0, "error",
				"error submitting task to Google Pipelines",
				map[string]string{"error": err.Error()},
			),
		)
		return err
	}

  b.event.WriteEvent(
    ctx,
    events.NewSystemLog(
      task.Id, 0, 0, "info",
      "submitted task to Google Pipelines", nil,
    ),
  )

  return b.event.WriteEvent(
		ctx, events.NewMetadata(task.Id, 0, map[string]string{
      "pipeline_operation_id": papiID,
    }),
	)
}

func (b *Backend) submit(task *tes.Task) (string, error) {
  res := task.GetResources()

  // TODO optimize machine type selection with custom machine types.
  var machineType string
  switch {
  // Standard machine types
  case res.CpuCores == 0:
    machineType = "n1-standard-1"
  case res.CpuCores == 1 && res.RamGb < 3.75:
    machineType = "n1-standard-1"
  case res.CpuCores == 2 && res.RamGb < 7.5:
    machineType = "n1-standard-2"
  case res.CpuCores == 3 || res.CpuCores == 4 && res.RamGb < 15:
    machineType = "n1-standard-4"
  case res.CpuCores >= 5 && res.CpuCores <= 8 && res.RamGb < 30:
    machineType = "n1-standard-8"
  case res.CpuCores >= 9 && res.CpuCores <= 16 && res.RamGb < 60:
    machineType = "n1-standard-16"
  case res.CpuCores >= 17 && res.CpuCores <= 32 && res.RamGb < 120:
    machineType = "n1-standard-32"
  case res.CpuCores >= 33 && res.CpuCores <= 64 && res.RamGb < 240:
    machineType = "n1-standard-64"
  case res.CpuCores > 64 && res.RamGb < 360:
    machineType = "n1-standard-96"

  // High-cpu machine types
  case res.CpuCores == 2 && res.RamGb < 1.8:
    machineType = "n1-highcpu-2"
  case res.CpuCores == 3 || res.CpuCores == 4 && res.RamGb < 3.6:
    machineType = "n1-highcpu-4"
  case res.CpuCores >= 5 && res.CpuCores <= 8 && res.RamGb < 7.2:
    machineType = "n1-highcpu-8"
  case res.CpuCores >= 9 && res.CpuCores <= 16 && res.RamGb < 14.4:
    machineType = "n1-highcpu-16"
  case res.CpuCores >= 17 && res.CpuCores <= 32 && res.RamGb < 28.8:
    machineType = "n1-highcpu-32"
  case res.CpuCores >= 33 && res.CpuCores <= 64 && res.RamGb < 57.6:
    machineType = "n1-highcpu-64"
  case res.CpuCores > 64 && res.RamGb < 86.4:
    machineType = "n1-highcpu-96"

  // High-mem machine types
  case res.CpuCores == 2 && res.RamGb < 13:
    machineType = "n1-highmem-2"
  case res.CpuCores == 3 || res.CpuCores == 4 && res.RamGb < 26:
    machineType = "n1-highmem-4"
  case res.CpuCores >= 5 && res.CpuCores <= 8 && res.RamGb < 52:
    machineType = "n1-highmem-8"
  case res.CpuCores >= 9 && res.CpuCores <= 16 && res.RamGb < 104:
    machineType = "n1-highmem-16"
  case res.CpuCores >= 17 && res.CpuCores <= 32 && res.RamGb < 208:
    machineType = "n1-highmem-32"
  case res.CpuCores >= 33 && res.CpuCores <= 64 && res.RamGb < 416:
    machineType = "n1-highmem-64"
  case res.CpuCores > 64 && res.RamGb < 624:
    machineType = "n1-highmem-96"
  default:
    return "", fmt.Errorf("could not find matching machine type")
  }

  if len(res.Zones) == 0 {
    return "", fmt.Errorf("at least one zone is required")
  }


  data := &bytes.Buffer{}
  mar := jsonpb.Marshaler{}
  _ = mar.Marshal(data, task)
  b64 := base64.StdEncoding.EncodeToString(data.Bytes())

  pl := &genomics.Pipeline{
    Environment: map[string]string{
      "FUNNEL_TASK_ID": task.Id,
      "FUNNEL_TASK": b64,
    },
    Resources: &genomics.Resources{
      ProjectId: b.conf.Project,
      Zones: res.GetZones(),
      VirtualMachine: &genomics.VirtualMachine{
        BootDiskSizeGb: int64(res.GetDiskGb()),
        MachineType: machineType,
        Preemptible: res.GetPreemptible(),
      },
    },
    Actions: []*genomics.Action{
      {
        Name: fmt.Sprintf("funnel-task-%s", task.Id),
        Commands: []string{"sh", "-c", "worker run --taskBase64 $FUNNEL_TASK"},
        ImageUri: "ohsucompbio/funnel:SNAPSHOT-4c658ecc",
      },
    },
  }

  call := b.client.Pipelines.Run(&genomics.RunPipelineRequest{
    Pipeline: pl,
  })

  resp, err := call.Do()
  return resp.Name, err

}

// Cancel cancels a running Pipelines Operation.
func (b *Backend) Cancel(ctx context.Context, taskID string) error {
  return fmt.Errorf("Cancel is not implemented")
}

// WriteEvent writes an event to the compute backend.
// Currently, only TASK_CREATED is handled, which calls Submit.
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

