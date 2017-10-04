package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/elastic"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/spf13/cobra"
	"os"
)

// Cmd provides the "elastic" command.
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
		c := config.DefaultConfig()
		es, err := elastic.NewElastic(c.Server.Databases.Elastic)
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
	},
}
