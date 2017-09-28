package boltdb

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use: "boltdb",
}

func init() {
	Cmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use: "export db-file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		conf := config.Config{}
		conf.Server.DBPath = args[0]

		db, err := server.NewTaskBolt(conf)
		if err != nil {
			return err
		}
		w := &writer{}
		return db.ReadEvents(w)
	},
}

type writer struct{}

func (w *writer) Write(ev *events.Event) error {
	s, err := events.Marshal(ev)
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}
