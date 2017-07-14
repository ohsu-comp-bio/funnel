package termdash

import (
	"fmt"
	ui "github.com/gizak/termui"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/compact"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/config"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/widgets"
	"github.com/spf13/cobra"
)

var (
	tesServer string
	cursor    *GridCursor
	cGrid     *compact.Grid
	header    *widgets.TermDashHeader
)

// Cmd represents the worker command
var Cmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Starts a Funnel dashboard in your terminal.",
	Run: func(cmd *cobra.Command, args []string) {
		termdash(tesServer)
	},
}

func init() {
	Cmd.Flags().StringVarP(&tesServer, "server", "S", "http://localhost:8000", "")
}

func termdash(tesHTTPServerAddress string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("An error occurred:\n", r)
		}
	}()

	// init global config
	config.Init()

	// override default colormap
	ui.ColorMap = colorMap

	if err := ui.Init(); err != nil {
		panic(fmt.Errorf("Error initializing termdash UI: %v", err))
	}
	defer Shutdown()

	// init grid, cursor, header
	cursor = NewGridCursor(tesHTTPServerAddress)
	cGrid = compact.NewGrid()
	header = widgets.NewTermDashHeader()

	for {
		exit := Display()
		if exit {
			return
		}
	}
}

func Shutdown() {
	ui.Close()
}
