package task

import (
	"fmt"
	"io"
	"os"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/client"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
)

// Create runs the "task create" CLI command, connecting to the server,
// calling CreateTask, and writing output to the given writer.
// Tasks are loaded from the "files" arg. "files" are file paths to JSON objects.
func Create(server string, files []string, writer io.Writer) error {
	cli, err := client.NewClient(server)
	if err != nil {
		return err
	}

	res := []string{}

	for _, taskFile := range files {
		var err error
		var task tes.Task

		f, err := os.Open(taskFile)
		defer f.Close()
		if err != nil {
			return err
		}

		err = jsonpb.Unmarshal(f, &task)
		if err != nil {
			return fmt.Errorf("can't load task: %s", err)
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
