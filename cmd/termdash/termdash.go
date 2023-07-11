package termdash

import (
	"fmt"
	"os"

	ui "github.com/gizak/termui"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/compact"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/config"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/widgets"
	"github.com/spf13/cobra"
)

var (
	defaultTesServer = "http://localhost:8000"
	tesServer        string
	cursor           *GridCursor
	cGrid            *compact.Grid
	header           *widgets.TermDashHeader
)

// Cmd represents the worker command
var Cmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Start a Funnel dashboard in your terminal.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if tesServer == "" {
			if val := os.Getenv("FUNNEL_SERVER"); val != "" {
				tesServer = val
			} else {
				tesServer = defaultTesServer
			}
		}
	},
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return termdash(tesServer)
	},
}

func init() {
	Cmd.Flags().StringVarP(&tesServer, "server", "S", tesServer, fmt.Sprintf("(default \"%s\")", defaultTesServer))
}

func termdash(tesHTTPServerAddress string) error {
	// init global config
	config.Init()

	// override default colormap
	ui.ColorMap = colorMap

	if err := ui.Init(); err != nil {
		return fmt.Errorf("error initializing termdash UI: %v", err)
	}
	defer Shutdown()

	// init grid, cursor, header
	header = widgets.NewTermDashHeader()
	cGrid = compact.NewGrid()
	var err error
	cursor, err = NewGridCursor(tesHTTPServerAddress)
	if err != nil {
		return fmt.Errorf("error initializing the grid cursor: %v", err)
	}

	for {
		exit := Display()
		if exit {
			return nil
		}
	}
}

func Shutdown() {
	ui.Close()
}
