package examples

import (
	"fmt"

	"github.com/ohsu-comp-bio/funnel/config"
	ex "github.com/ohsu-comp-bio/funnel/examples"
	"github.com/spf13/cobra"
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
		taskEx := ex.Examples()
		for n, v := range taskEx {
			byShortName[n] = v
		}

		confEx := config.Examples()
		for n, v := range confEx {
			byShortName[n] = v
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
		data, ok := byShortName[name]
		if !ok {
			return fmt.Errorf("no example by the name of %s", name)
		}

		fmt.Println(data)
		return nil
	},
}
