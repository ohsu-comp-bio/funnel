package task

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/client"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
)

// Create runs the "task create" CLI command, connecting to the server,
// calling CreateTask, and writing output to the given writer.
// Tasks are loaded from the "files" arg. "files" are file paths to JSON objects.
func Create(server string, files []string, writer io.Writer) error {
	cli := client.NewClient(server)
	res := []string{}

	for _, taskFile := range files {
		var err error
		var taskMessage []byte
		var task tes.Task

		taskMessage, err = ioutil.ReadFile(taskFile)
		if err != nil {
			return err
		}

		err = proto.Unmarshal(taskMessage, &task)
		if err != nil {
			return err
		}

		r, err := cli.CreateTask(context.Background(), &task)
		if err != nil {
			return err
		}
		res = append(res, r.Id)
	}

	for _, x := range res {
		fmt.Fprintln(writer, x)
	}

	return nil
}
