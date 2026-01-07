package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var amqpUrl string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "learn-pub-sub-starter",
	Short: "CLI to run server or client for a war game",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	defaultUrl := os.Getenv("AMQP_URL")
	if defaultUrl == "" {
		defaultUrl = "amqp://guest:guest@192.168.1.20:5672"
	}
	rootCmd.Flags().StringVar(&amqpUrl, "amqp-url", defaultUrl, "URL for AMQP broker")
}
