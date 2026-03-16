package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	grpcHost  string
	token     string
	tokenFile string
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
}

// AddTokenFlags adds the --token and --token-file flags to the given command.
func AddTokenFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&token, "token", "", "Your auth token (fake auth uses player name)")
	cmd.PersistentFlags().StringVar(&tokenFile, "token-file", "", "Path to a file containing your auth token")
}

// ResolveToken returns the token from the --token flag or reads it from the --token-file flag.
// It returns an error if both are provided.
func ResolveToken(cmd *cobra.Command) (string, error) {
	t, _ := cmd.Flags().GetString("token")
	tf, _ := cmd.Flags().GetString("token-file")

	if t != "" && tf != "" {
		return "", fmt.Errorf("either --token or --token-file can be used, but not both")
	}

	if tf != "" {
		content, err := os.ReadFile(tf)
		if err != nil {
			return "", fmt.Errorf("failed to read token file: %v", err)
		}
		return strings.TrimSpace(string(content)), nil
	}

	return t, nil
}
