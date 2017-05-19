package e2e

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"testing"
)

func TestValidationError(t *testing.T) {
	_, err := runTask(&tes.Task{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
