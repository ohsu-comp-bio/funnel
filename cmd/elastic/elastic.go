package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/elastic"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/spf13/cobra"
	"os"
)

var Cmd = &cobra.Command{
	Use: "elastic",
}

func init() {
	Cmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use: "import",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Set up database.
		ctx := context.Background()
		c := elastic.DefaultConfig()
		es, err := elastic.NewElastic(c)
		if err != nil {
			return err
		}

		err = es.Init(ctx)
		if err != nil {
			return err
		}

		// Decode a stream of JSON events from stdin.
		dec := json.NewDecoder(os.Stdin)
		for {
			// Read next event from input stream.
			ev := &events.Event{}
			err := jsonpb.UnmarshalNext(dec, ev)
			if err != nil {
				return err
			}

			// Write event to database.
			err = es.Write(ev)
			if err != nil {
				return err
			}
			fmt.Println("Imported", ev.Id)
		}

		return nil
	},
}

func prototype() {

	ctx := context.Background()
	c := elastic.DefaultConfig()
	es, err := elastic.NewElastic(c)
	if err != nil {
		panic(err)
	}

	err = es.Init(ctx)

	if err != nil {
		panic(err)
	}

	id := "b76l27irl6qmadfm7dm0"
	task, err := es.GetTask(ctx, id)
	if err != nil {
		panic(err)
	}
	fmt.Println("task", task)

	/*
	  id := util.GenTaskID()
	  err = es.CreateTask(ctx, &tes.Task{
	    Id: id,
	    Executors: []*tes.Executor{
	      {
	        Cmd: []string{"echo", "hello world"},
	      },
	    },
	  })
	  if err != nil {
	    panic(err)
	  }
	*/

	ev := events.NewState(id, 0, tes.State_QUEUED)
	err = es.Write(ev)
	if err != nil {
		panic(err)
	}

	tasks, err := es.ListTasks(ctx)
	if err != nil {
		panic(err)
	}

	for _, task := range tasks {
		fmt.Println("tasks", task)
	}
}
