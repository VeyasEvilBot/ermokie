package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/rendering"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the overlay",
	Long:  `Serve the overlay`,
	Run: func(cmd *cobra.Command, args []string) {
		var port int
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		portFlag, getPortErr := cmd.Flags().GetInt("port")
		if getPortErr != nil {
			fmt.Printf("Error getting port: %v\n", getPortErr)
			os.Exit(1)
			port = cfg.Overlay.Port
		} else {
			port = portFlag
		}

		dir := filepath.Join(cfg.General.DataDir, "render")
		o := rendering.NewOverlayServer(port, dir)
		o.Serve()
	},
}

// add cmd that auto renders the active run on update

func init() {
	serveCmd.Flags().IntP("port", "p", 6767, "Port to serve the overlay on")
	serveCmd.Flags().StringP("dir", "d", "", "Directory to serve the overlay from")
	rootCmd.AddCommand(serveCmd)
}
