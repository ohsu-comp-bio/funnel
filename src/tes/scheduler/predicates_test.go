package scheduler

import (
	"runtime/debug"
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
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
