package worker

import (
	"errors"
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
	"testing"
)

func TestReconcileSingleJobCompleteFlow(t *testing.T) {

	var err error
	jobs := map[string]*pbr.JobWrapper{}
	w := worker{
		runJob: noopRunJob,
		ctrls:  map[string]JobControl{},
	}

	err = w.reconcile(jobs)

	if err != nil {
		t.Error("Unexpected error on empty reconcile")
	}
	if len(w.ctrls) != 0 {
		t.Error("Unexpected runner created on empty reconcile")
	}

	j := &pbe.Job{
		JobID: "job-1",
		State: pbe.State_Queued,
	}
	addJob(jobs, j)

	w.reconcile(jobs)

	if _, exists := w.ctrls["job-1"]; !exists {
		t.Error("Expected runner to be created for new job")
	}

	ctrl := w.ctrls["job-1"]

	if j.State != pbe.State_Initializing {
		t.Error("Expected job just started to be in initializing state.")
	}

	ctrl.SetRunning()
	w.reconcile(jobs)

	if j.State != pbe.State_Running {
		t.Error("Expected job state to be running")
	}

	ctrl.SetResult(nil)
	w.reconcile(jobs)

	if j.State != pbe.State_Complete {
		t.Error("Expected job state to be complete")
	}
}

func TestReconcileJobError(t *testing.T) {

	jobs := map[string]*pbr.JobWrapper{}
	w := worker{
		runJob: noopRunJob,
		ctrls:  map[string]JobControl{},
	}
	j := &pbe.Job{
		JobID: "job-1",
		State: pbe.State_Queued,
	}
	addJob(jobs, j)
	w.reconcile(jobs)
	ctrl := w.ctrls["job-1"]
	ctrl.SetResult(errors.New("Test job error"))
	w.reconcile(jobs)

	if j.State != pbe.State_Error {
		t.Error("Expected job state to be Error")
	}
}

func TestReconcileCancelJob(t *testing.T) {
	jobs := map[string]*pbr.JobWrapper{}
	w := worker{
		runJob: noopRunJob,
		ctrls:  map[string]JobControl{},
	}
	j := &pbe.Job{
		JobID: "job-1",
		State: pbe.State_Queued,
	}
	addJob(jobs, j)
	w.reconcile(jobs)

	j.State = pbe.State_Canceled
	ctrl := w.ctrls["job-1"]
	w.reconcile(jobs)

	if ctrl.State() != pbe.State_Canceled {
		t.Error("Expected runner state to be canceled")
	}

	if w.ctrls["job-1"] != nil {
		t.Error("Expected job ctrl to be cleaned up")
	}
}

func TestReconcileMultiple(t *testing.T) {

	jobs := map[string]*pbr.JobWrapper{}
	w := worker{
		runJob: noopRunJob,
		ctrls:  map[string]JobControl{},
	}

	w.reconcile(jobs)

	addJob(jobs, &pbe.Job{
		JobID: "job-1",
		State: pbe.State_Queued,
	})

	w.reconcile(jobs)

	if _, exists := w.ctrls["job-1"]; !exists {
		t.Error("Expected runner to be created for new job")
	}

	if jobs["job-1"].Job.State != pbe.State_Initializing {
		t.Error("Expected job just started to be in initializing state.")
	}

	w.ctrls["job-1"].SetRunning()
	w.reconcile(jobs)

	if jobs["job-1"].Job.State != pbe.State_Running {
		t.Error("Expected job state to be running")
	}

	addJob(jobs, &pbe.Job{
		JobID: "job-2",
		State: pbe.State_Queued,
	})
	addJob(jobs, &pbe.Job{
		JobID: "job-3",
		State: pbe.State_Queued,
	})

	w.reconcile(jobs)

	if len(w.ctrls) != 3 {
		t.Error("Expected runner to be created for new job")
	}

	if jobs["job-2"].Job.State != pbe.State_Initializing {
		t.Error("Expected job 2 state to be init")
	}

	if jobs["job-3"].Job.State != pbe.State_Initializing {
		t.Error("Expected job 3 state to be init")
	}

	w.ctrls["job-1"].SetResult(nil)
	jobs["job-2"].Job.State = pbe.State_Canceled
	w.ctrls["job-3"].SetResult(errors.New("Job 3 error"))

	j2ctrl := w.ctrls["job-2"]
	w.reconcile(jobs)

	if jobs["job-1"].Job.State != pbe.State_Complete {
		t.Error("Expected job 1 state to be complete")
	}

	if j2ctrl.State() != pbe.State_Canceled {
		t.Error("Expected job 2 controller to be canceled state")
	}

	if w.ctrls["job-2"] != nil {
		t.Error("Expected job 2 ctrl to be cleaned up")
	}

	if jobs["job-3"].Job.State != pbe.State_Error {
		t.Error("Expected job 3 state to be error")
	}
}

// Tests how the worker handles the case where it finds a job without a controller
// and the job state is not Queued (normal case), but is Initializing or Running
func TestStraightToRunning(t *testing.T) {
	jobs := map[string]*pbr.JobWrapper{}
	w := worker{
		runJob: noopRunJob,
		ctrls:  map[string]JobControl{},
	}

	addJob(jobs, &pbe.Job{
		JobID: "job-1",
		State: pbe.State_Initializing,
	})
	addJob(jobs, &pbe.Job{
		JobID: "job-2",
		State: pbe.State_Running,
	})

	w.reconcile(jobs)

	if _, exists := w.ctrls["job-1"]; !exists {
		t.Error("Expected runner to be created for new job 1")
	}
	if _, exists := w.ctrls["job-2"]; !exists {
		t.Error("Expected runner to be created for new job 2")
	}

	if jobs["job-1"].Job.State != pbe.State_Initializing {
		t.Error("Expected job 1 state to be unchanged.")
	}

	if jobs["job-2"].Job.State != pbe.State_Initializing {
		t.Error("Expected job 2 state to revert to initializing.")
	}
}

// TODO test edge cases
// - missing job
// - missing ctrl
// - complete job, ctrl incomplete
