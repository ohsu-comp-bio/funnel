package storage

import (
	"context"
	"encoding/json"
	"fmt"

	cmdutil "github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/spf13/cobra"
)

// NewCommand returns the "storage" subcommands.
func NewCommand() *cobra.Command {

	configFile := ""
	conf := config.Config{}
	flagConf := config.Config{}

	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Access storage via Funnel's client libraries.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			conf, err = cmdutil.MergeConfigFileWithFlags(configFile, flagConf)
			if err != nil {
				return fmt.Errorf("processing config: %v", err)
			}
			return nil
		},
	}
	cmd.SetGlobalNormalizationFunc(cmdutil.NormalizeFlags)
	f := cmd.PersistentFlags()
	f.StringVarP(&configFile, "config", "c", configFile, "Config File")

	statCmd := &cobra.Command{
		Use:   "stat [url]",
		Short: "Returns information about the object at the given URL.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			store, err := storage.NewMux(conf)
			if err != nil {
				return fmt.Errorf("creating storage clients: %s", err)
			}

			obj, err := store.Stat(context.Background(), args[0])
			if err != nil {
				return err
			}

			b, err := json.Marshal(obj)
			if err != nil {
				return fmt.Errorf("marshaling output: %s", err)
			}
			fmt.Println(string(b))
			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:   "list [url]",
		Short: "List objects at the given URL.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			store, err := storage.NewMux(conf)
			if err != nil {
				return fmt.Errorf("creating storage clients: %s", err)
			}

			objects, err := store.List(context.Background(), args[0])
			if err != nil {
				return err
			}

			b, err := json.Marshal(objects)
			if err != nil {
				return fmt.Errorf("marshaling output: %s", err)
			}
			fmt.Println(string(b))
			return nil
		},
	}

	cmd.AddCommand(listCmd)
	cmd.AddCommand(statCmd)
	return cmd
}
