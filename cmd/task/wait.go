package task

import (
	"github.com/ohsu-comp-bio/funnel/cmd/client"
)

func Wait(server string, ids []string) error {
	return client.NewClient(server).WaitForTask(ids...)
}
