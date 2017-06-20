package worker

import (
	//"context"
	//"github.com/ohsu-comp-bio/funnel/storage"
	"testing"
)

func TestRunner(t *testing.T) {
	/*
		r := DefaultRunner{
			//Conf:   conf,
			//Mapper: NewFileMapper(baseDir),
			Store:  storage.Storage{},
			//Svc:    svc,
		}
		r.Run(context.Background())
	*/

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
