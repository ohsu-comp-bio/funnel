package examples

import (
	"fmt"
	ex "github.com/ohsu-comp-bio/funnel/cmd/examples/internal"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
)

// Cmd represents the examples command
var Cmd = &cobra.Command{
	Use:     "examples [name]",
	Aliases: []string{"example"},
	Short:   "Print example task messages.",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Map example name to asset name
		// e.g. config => examples/config.yml
		byShortName := map[string]string{}
		for _, n := range ex.AssetNames() {
			sn := filepath.Base(n)
			sn = strings.TrimSuffix(sn, filepath.Ext(sn))
			byShortName[sn] = n
		}

		// Print a list of example names and exit
		if len(args) == 0 || args[0] == "list" {
			for sn := range byShortName {
				fmt.Println(sn)
			}
			return nil
		}

		// Retrieve and print the example
		name := args[0]
		key, ok := byShortName[name]
		if !ok {
			return fmt.Errorf("No example by the name of %s", name)
		}

		data, err := ex.Asset(key)
		if err != nil {
			return fmt.Errorf("No example by the name of %s", name)
		}

		print(string(data))
		return nil
	},
}
