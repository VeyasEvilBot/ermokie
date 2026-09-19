package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"codeberg.org/veya/ermokie/pkg/config"
	"codeberg.org/veya/ermokie/pkg/rendering"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the browser overlay",
	Long:  "Serve the generated overlay on localhost for an OBS browser source.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return fmt.Errorf("read port: %w", err)
		}
		if port == 0 {
			port = cfg.Overlay.Port
		}
		dir, err := cmd.Flags().GetString("dir")
		if err != nil {
			return fmt.Errorf("read render directory: %w", err)
		}
		if dir == "" {
			dir = filepath.Join(cfg.General.DataDir, "render")
		}

		cmd.Printf("Overlay: http://127.0.0.1:%d/overlay\n", port)
		server := rendering.NewOverlayServer(port, dir)
		if err := server.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve overlay: %w", err)
		}
		return nil
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 0, "Port to serve the overlay on (defaults to config)")
	serveCmd.Flags().StringP("dir", "d", "", "Directory containing output.html and output.css")
	rootCmd.AddCommand(serveCmd)
}
