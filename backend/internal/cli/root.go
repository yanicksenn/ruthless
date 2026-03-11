package cli

import (
	"github.com/spf13/cobra"
)

var (
	grpcHost string
	token    string
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
	rootCmd.PersistentFlags().StringVar(&grpcHost, "host", "localhost:8080", "The host:port of the gRPC server")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Your auth token (fake auth uses player name)")
}
