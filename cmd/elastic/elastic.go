package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/server/elastic"
	"github.com/spf13/cobra"
	"os"
)

var configFile string

// Cmd provides the "elastic" command.
var Cmd = &cobra.Command{
	Use: "elastic",
}

func init() {
	Cmd.AddCommand(importCmd)
	Cmd.AddCommand(countCmd)
	flags := Cmd.PersistentFlags()
	flags.StringVarP(&configFile, "config", "c", "", "Config file")
}

var importCmd = &cobra.Command{
	Use: "import",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Set up database.
		ctx := context.Background()
		c := config.DefaultConfig()
		config.ParseFile(configFile, &c)
		c = config.InheritServerProperties(c)

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

var countCmd = &cobra.Command{
	Use: "count",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Set up database.
		c := config.DefaultConfig()
		config.ParseFile(configFile, &c)
		c = config.InheritServerProperties(c)

		es, err := elastic.NewElastic(c.Server.Databases.Elastic)
		if err != nil {
			return err
		}
		es.Counts()
		return nil
	},
}
