package cmd

import (
	"fmt"
	"funnel/examples"
	"github.com/spf13/cobra"
	"strings"
)

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Print example task messages.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("Example name required")
		}
		name := args[0]

		if name == "list" {
			for _, n := range examples.AssetNames() {
				n = strings.TrimPrefix(n, "examples/")
				n = strings.TrimSuffix(n, ".json")
				print(n + "\n")
			}
			return nil
		}

		data, err := examples.Asset("examples/" + name + ".json")
		if err != nil {
			return fmt.Errorf("No example by the name of %s", name)
		}

		print(string(data))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(examplesCmd)
}
