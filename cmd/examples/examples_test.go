package examples

import (
	"github.com/golang/protobuf/jsonpb"
	ex "github.com/ohsu-comp-bio/funnel/cmd/examples/internal"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"strings"
	"testing"
)

func TestExamplesAreValid(t *testing.T) {
	for _, en := range ex.AssetNames() {
		if strings.HasSuffix(en, ".json") {
			t.Log(en)
			tb, err := ex.Asset(en)
			if err != nil {
				t.Error(err)
			}
			var task tes.Task
			err = jsonpb.UnmarshalString(string(tb), &task)
			if err != nil {
				t.Error(err)
			}
			if err := tes.Validate(&task); err != nil {
				t.Error("Invalid task message:", en, "\n", "error:", err)
			}
		}
	}
}
