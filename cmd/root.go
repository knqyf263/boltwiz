package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/knqyf263/boltwiz/server"
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Boltdb Server",
	Long:  `Start the boltdb browser server`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := args[0]
		if input.debug {
			slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
				Level: slog.LevelDebug,
			})))

		}
		if input.local {
			go func() {
				time.Sleep(2 * time.Second) // wait for server to start
				slog.Info("Opening browser....")
				err := browser.OpenURL(fmt.Sprintf("http://localhost:%d", input.port))

				if err != nil {
					slog.Error("Error while opening browser", slog.Any("err", err))
				}
			}()
		}

		server.StartServer(server.Options{
			DBPath: dbPath,
			Port:   input.port,
		})
		return nil
	},
}

var input = new(struct {
	debug bool
	local bool
	port  int
})

func init() {
	// set global logger
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	rootCmd.Flags().BoolVarP(&input.local, "local", "l", false, "open the browser automatically")
	rootCmd.Flags().BoolVarP(&input.debug, "debug", "d", false, "debug mode")
	rootCmd.Flags().IntVarP(&input.port, "port", "p", 8090, "port to serve the server")
}

func Execute() error {
	return rootCmd.Execute()
}
