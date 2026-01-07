package cmd

import (
	"log"

	"github.com/oleksii-kalinin/learn-pub-sub-starter/cmd/client"
	"github.com/spf13/cobra"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start game client",
	Run: func(cmd *cobra.Command, args []string) {
		if err := client.Start(amqpUrl); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
