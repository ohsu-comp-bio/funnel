package worker

import (
	"github.com/ohsu-comp-bio/funnel/storage"
	"testing"
)

func TestRunner(t *testing.T) {
	r := taskRunner{
		mapper: NewFileMapper(baseDir),
		store:  storage.Storage{},
		conf:   conf,
		taskID: "task_id",
		svc:    svc,
	}
	r.Run(ctx)

	// Expect logging endpoints to be called
	// Expect task to be set to running
	// Expect task to be set to complete
}

func TestCancelContext(t *testing.T) {
}

func TestExecutorError(t *testing.T) {
}

func TestDownloadError(t *testing.T) {
}

func TestUploadError(t *testing.T) {
}

func TestPanic(t *testing.T) {
}

func TestCancelTask(t *testing.T) {
}
