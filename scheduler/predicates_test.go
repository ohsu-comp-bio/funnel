package scheduler

import (
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"runtime/debug"
	"testing"
)

func TestPortsFitEmptyTask(t *testing.T) {
	testEmptyTask(t, PortsFit, "PortsFit")
}

func TestVolumesFitEmptyTask(t *testing.T) {
	testEmptyTask(t, VolumesFit, "VolumesFit")
}

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

	w := &pbf.Worker{
		Id: "test-worker",
		Resources: &pbf.Resources{
			Cpus:   1.0,
			RamGb:  1.0,
			DiskGb: 1.0,
		},
		Available: &pbf.Resources{
			Cpus:   1.0,
			RamGb:  1.0,
			DiskGb: 1.0,
		},
	}

	if !ResourcesFit(j, w) {
		t.Error("Execpted resources to fit")
	}

	w.Available.Cpus = 0.0

	if ResourcesFit(j, w) {
		t.Error("Execpted resources NOT to fit")
	}

	w.Available.Cpus = 1.0
	j.Resources.CpuCores = 2

	if ResourcesFit(j, w) {
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
	w := &pbf.Worker{}
	p(j, w)
}
