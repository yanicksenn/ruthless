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

var cardsCmd = &cobra.Command{
	Use:   "cards",
	Short: "Manage cards",
}

var cardsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new card",
	Run: func(cmd *cobra.Command, args []string) {
		text, _ := cmd.Flags().GetString("text")
		if text == "" {
			log.Fatal("Card text is required")
		}

		payload := map[string]string{"text": text}
		body, _ := json.Marshal(payload)

		url := fmt.Sprintf("%s/api/v1/cards", apiURL)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Fatalf("Failed to create card: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusCreated {
			log.Fatalf("Error: [%d] %s", resp.StatusCode, string(respBody))
		}

		fmt.Printf("Success! Card created:\n%s\n", string(respBody))
	},
}

var cardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cards",
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("%s/api/v1/cards", apiURL)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to list cards: %v", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Error: [%d] %s", resp.StatusCode, string(respBody))
		}

		fmt.Printf("Cards:\n%s\n", string(respBody))
	},
}

func init() {
	rootCmd.AddCommand(cardsCmd)
	cardsCmd.AddCommand(cardsCreateCmd)
	cardsCmd.AddCommand(cardsListCmd)

	cardsCreateCmd.Flags().String("text", "", "The text of the card, containing at least one '___'")
	cardsCreateCmd.MarkFlagRequired("text")
}
