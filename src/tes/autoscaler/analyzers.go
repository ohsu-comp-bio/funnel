package autoscaler

import (
	"context"
	"io"
	"log"
	"tes/scheduler"
	pbr "tes/server/proto"
)

// DumbAutoscaler is a prototype of what an autoscaler might look like.
// It's "dumb" because it uses dead simple ideas for scaling, e.g.
// "are there more than 10 tasks in the queue?"
type DumbAutoscaler struct {
	sched         *scheduler.Client
	workerFactory WorkerFactory
}

func (da DumbAutoscaler) Analyze() {
	status := GetSchedStatus(da.sched)
	if status.QueueCount > 10 {
		da.workerFactory.AddWorkers(1)
	} else if status.SlotCount == 0 {
		da.workerFactory.AddWorkers(1)
	}
}

func getQueueCount(ctx context.Context, client *scheduler.Client) int {
	count := 0

	queue_req := &pbr.QueuedTaskInfoRequest{1000}
	stream, _ := client.GetQueueInfo(ctx, queue_req)

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		count++
	}
	return count
}

func getSlotSummary(ctx context.Context, sched *scheduler.Client) (*pbr.SlotSummary, error) {
	sumReq := &pbr.SlotSummaryRequest{}
	resp, err := sched.GetSlotSummary(ctx, sumReq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type SchedStatus struct {
	QueueCount int
	SlotCount  int
}

func GetSchedStatus(sched *scheduler.Client) SchedStatus {
	ctx := context.Background()
	log.Println("Checking scale")
	queueCount := getQueueCount(ctx, sched)
	slots, _ := getSlotSummary(ctx, sched)
	status := SchedStatus{queueCount, int(slots.Count)}
	log.Printf("Slot count %d", slots.Count)
	log.Printf("Queue count: %s", queueCount)
	return status
}
