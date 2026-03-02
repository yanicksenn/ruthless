package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Join a game session and play",
}

var playStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new game session",
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/api/v1/sessions", apiURL)
		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			log.Fatalf("Failed to start session: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusCreated {
			log.Fatalf("Error: [%d] %s", resp.StatusCode, string(respBody))
		}

		fmt.Printf("Session started:\n%s\n", string(respBody))
	},
}

var playJoinCmd = &cobra.Command{
	Use:   "join [session_id]",
	Short: "Join an existing game session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		playerName, _ := cmd.Flags().GetString("name")
		if playerName == "" {
			log.Fatal("Player name is required")
		}

		token, _ := cmd.Flags().GetString("token")
		if token == "" {
			// For testing with fake auth, we just use the name as token
			token = playerName
		}

		payload := map[string]string{"player_name": playerName}
		body, _ := json.Marshal(payload)

		url := fmt.Sprintf("%s/api/v1/sessions/%s/join", apiURL, sessionID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("Failed to join session: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Error: [%d] %s", resp.StatusCode, string(respBody))
		}

		fmt.Printf("Successfully joined session:\n%s\n", string(respBody))
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
	playCmd.AddCommand(playStartCmd)
	playCmd.AddCommand(playJoinCmd)

	playJoinCmd.Flags().String("name", "", "Your player name")
	playJoinCmd.Flags().String("token", "", "Auth token (optional, uses name for fake auth if omitted)")
	playJoinCmd.MarkFlagRequired("name")
}
