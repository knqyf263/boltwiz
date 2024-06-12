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

		return server.StartServer(server.Options{
			DBPath:     dbPath,
			Port:       input.port,
			ProtoFiles: input.protoFiles,
			ProtoType:  input.protoType,
		})
	},
}

var input = new(struct {
	debug      bool
	local      bool
	port       int
	protoType  string
	protoFiles []string
})

func init() {
	// set global logger
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	rootCmd.Flags().BoolVarP(&input.local, "local", "l", false, "open the browser automatically")
	rootCmd.Flags().BoolVarP(&input.debug, "debug", "d", false, "debug mode")
	rootCmd.Flags().IntVarP(&input.port, "port", "p", 8090, "port to serve the server")
	rootCmd.Flags().StringVar(&input.protoType, "proto-type", "", "The full type name of the message within the input (e.g. acme.weather.v1.Units)")
	rootCmd.Flags().StringSliceVar(&input.protoFiles, "proto-files", nil, "Proto files")
}

func Execute() error {
	return rootCmd.Execute()
}
