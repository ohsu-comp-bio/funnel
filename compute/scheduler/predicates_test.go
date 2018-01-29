package scheduler

import (
	"runtime/debug"
	"testing"

	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

func TestZonesFitEmptyTask(t *testing.T) {
	testEmptyTask(t, ZonesFit, "ZonesFit")
}

func TestResourcesFitEmptyTask(t *testing.T) {
	testEmptyTask(t, ResourcesFit, "ResourcesFit")
}

func TestCpuResourcesFit(t *testing.T) {
	j := &tes.Task{
		Resources: &tes.Resources{
			CpuCores: 1,
		},
	}

	w := &pbs.Node{
		Id: "test-node",
		Resources: &pbs.Resources{
			Cpus:   1.0,
			RamGb:  1.0,
			DiskGb: 1.0,
		},
		Available: &pbs.Resources{
			Cpus:   1.0,
			RamGb:  1.0,
			DiskGb: 1.0,
		},
	}

	if ResourcesFit(j, w) != nil {
		t.Error("Execpted resources to fit")
	}

	w.Available.Cpus = 0.0

	if ResourcesFit(j, w) == nil {
		t.Error("Execpted resources NOT to fit")
	}

	w.Available.Cpus = 1.0
	j.Resources.CpuCores = 2

	if ResourcesFit(j, w) == nil {
		t.Error("Execpted resources NOT to fit")
	}
}

// testEmptyTask tests that the predicates all handle an empty task.
// Protects against nil-pointer panics.
func testEmptyTask(t *testing.T, p Predicate, name string) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Predicate panic: %s\n%s", name, debug.Stack())
		}
	}()

	j := &tes.Task{}
	w := &pbs.Node{}
	p(j, w)
}
