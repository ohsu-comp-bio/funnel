package task

import (
	"github.com/ohsu-comp-bio/funnel/client"
	"golang.org/x/net/context"
)

// Wait runs the "task wait" CLI command, which polls the server,
// calling GetTask() for each ID, and blocking until the tasks have
// reached a terminal state.
func Wait(server string, ids []string) error {
	return client.NewClient(server).WaitForTask(context.Background(), ids...)
}
