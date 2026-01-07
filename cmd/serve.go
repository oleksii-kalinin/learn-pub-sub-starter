package cmd

import (
	"log"

	"github.com/oleksii-kalinin/learn-pub-sub-starter/cmd/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start game server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := server.Serve(amqpUrl); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
