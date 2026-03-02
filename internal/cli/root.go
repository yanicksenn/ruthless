package cli

import (
	"github.com/spf13/cobra"
)

var (
	apiURL string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cah",
	Short: "Cards Against Humanity clone CLI",
	Long:  `A command line interface to manage and play the Ruthless (CAH Clone) server.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURL, "url", "http://localhost:8080", "The base URL of the CAH server")
}
