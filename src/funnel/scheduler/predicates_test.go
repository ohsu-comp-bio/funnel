package scheduler

import (
	pbe "funnel/ga4gh"
	pbr "funnel/server/proto"
	"runtime/debug"
	"testing"
)

func TestPortsFitEmptyJob(t *testing.T) {
	testEmptyJob(t, PortsFit, "PortsFit")
}

func TestVolumesFitEmptyJob(t *testing.T) {
	testEmptyJob(t, VolumesFit, "VolumesFit")
}

func TestZonesFitEmptyJob(t *testing.T) {
	testEmptyJob(t, ZonesFit, "ZonesFit")
}

func TestResourcesFitEmptyJob(t *testing.T) {
	testEmptyJob(t, ResourcesFit, "ResourcesFit")
}

func TestCpuResourcesFit(t *testing.T) {
	j := &pbe.Job{
		Task: &pbe.Task{
			Resources: &pbe.Resources{
				MinimumCpuCores: 1,
			},
		},
	}

	w := &pbr.Worker{
		Id: "test-worker",
		Resources: &pbr.Resources{
			Cpus: 1.0,
			Ram:  1.0,
			Disk: 1.0,
		},
		Available: &pbr.Resources{
			Cpus: 1.0,
			Ram:  1.0,
			Disk: 1.0,
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
	j.Task.Resources.MinimumCpuCores = 2

	if ResourcesFit(j, w) {
		t.Error("Execpted resources NOT to fit")
	}
}

// testEmptyJob tests that the predicates all handle an empty job.
// Protects against nil-pointer panics.
func testEmptyJob(t *testing.T, p Predicate, name string) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Predicate panic: %s\n%s", name, debug.Stack())
		}
	}()

	j := &pbe.Job{}
	w := &pbr.Worker{}
	p(j, w)
}
