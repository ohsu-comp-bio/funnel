package run

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/client"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"sync"
)

type taskGroup struct {
	wg        sync.WaitGroup
	err       chan error
	printTask bool
	client    *client.Client
}

func (tg *taskGroup) runTask(t *tes.Task, wait bool, waitFor []string) {
	if tg.err == nil {
		tg.err = make(chan error)
	}

	tg.wg.Add(1)
	go func() {
		err := tg._run(t, wait, waitFor)
		if err != nil {
			tg.err <- err
		}
		tg.wg.Done()
	}()
}

func (tg *taskGroup) wait() error {
	done := make(chan struct{})
	go func() {
		tg.wg.Wait()
		close(done)
	}()

	select {
	case err := <-tg.err:
		return err
	case <-done:
		return nil
	}
}

func (tg *taskGroup) _run(task *tes.Task, wait bool, waitFor []string) error {
	// Marshal message to JSON
	taskJSON, merr := tg.client.Marshaler.MarshalToString(task)
	if merr != nil {
		return merr
	}

	if tg.printTask {
		fmt.Println(taskJSON)
		return nil
	}

	if len(waitFor) > 0 {
		for _, tid := range waitFor {
			tg.client.WaitForTask(tid)
		}
	}

	resp, rerr := tg.client.CreateTask([]byte(taskJSON))
	if rerr != nil {
		return rerr
	}

	taskID := resp.Id
	fmt.Println(taskID)

	if wait {
		return tg.client.WaitForTask(taskID)
	}
	return nil
}
