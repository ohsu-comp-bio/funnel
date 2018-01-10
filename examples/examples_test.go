package examples

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"testing"
)

func TestExamplesAreValid(t *testing.T) {
	for _, en := range AssetNames() {
		tb, err := Asset(en)
		if err != nil {
			t.Fatal(err)
		}
		var task tes.Task
		err = jsonpb.UnmarshalString(string(tb), &task)
		if err != nil {
			t.Fatal("unmarshal failed", en, err)
		}
		if err := tes.Validate(&task); err != nil {
			t.Fatal("Invalid task message:", en, "\n", "error:", err)
		}
	}
}
