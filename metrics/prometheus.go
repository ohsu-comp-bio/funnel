package metrics

import (
	"context"
	"time"

	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/prometheus/client_golang/prometheus"
)

var taskStates = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "funnel",
		Subsystem: "tasks",
		Name:      "state_count_total",
		Help:      "Number of tasks in each state.",
	},
	[]string{"state"},
)

func init() {
	for key := range tes.State_value {
		taskStates.WithLabelValues(key).Set(0)
	}
}

// TaskStateCounter is implemented by database backends which provide
// queries for counting tasks in each state.
type TaskStateCounter interface {
	// TaskStateCounts returns the number of tasks in each state.
	TaskStateCounts(context.Context) (map[string]int32, error)
}

// Register registers a prometheus metric which counts task states,
// and starts a background routine to update the task state every 5 seconds.
func Register(ctx context.Context, counter TaskStateCounter) error {
	err := prometheus.Register(taskStates)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				counts, err := counter.TaskStateCounts(ctx)
				if err != nil {
					break
				}
				for key, count := range counts {
					taskStates.WithLabelValues(key).Set(float64(count))
				}
			}
		}
	}()

	return nil
}
