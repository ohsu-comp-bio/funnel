package worker

import (
	"context"
	"testing"
)

func TestRunset(t *testing.T) {
	r := runSet{}
	r.Add("foo", func(c context.Context, s string) {
		if s != "foo" {
			t.Fatal("Unexpected task ID")
		}
		if r.Count() != 1 {
			t.Fatal("Unexpected runner count")
		}
	})

	r.Wait()

	if r.Count() != 0 {
		t.Fatal("Unexpected runner count")
	}
}
