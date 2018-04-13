package task

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
)

// Create runs the "task create" CLI command, connecting to the server,
// calling CreateTask, and writing output to the given writer.
// Tasks are loaded from the "files" arg. "files" are file paths to JSON objects.
func Create(server string, files []string, writer io.Writer) error {
	cli, err := tes.NewClient(server)
	if err != nil {
		return err
	}

	var reader io.Reader
	reader = os.Stdin

	for _, taskFile := range files {
		f, err := os.Open(taskFile)
		defer f.Close()
		if err != nil {
			return err
		}
		reader = io.MultiReader(reader, f)
	}

	dec := json.NewDecoder(reader)
	for {
		var task tes.Task
		err := jsonpb.UnmarshalNext(dec, &task)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("can't load task: %s", err)
		}

		r, err := cli.CreateTask(context.Background(), &task)
		if err != nil {
			return err
		}
		fmt.Fprintln(writer, r.Id)
	}

	return nil
}
