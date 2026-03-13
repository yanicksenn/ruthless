package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)
var (
	clientSecretFile string
	callbackPort     int
	tokenFile        string
)

func init() {
	tokenCmd.PersistentFlags().StringVar(&clientSecretFile, "client-secret", "secrets/client_secret_dev.json", "path to Google Client Secret JSON")
	tokenCmd.PersistentFlags().IntVar(&callbackPort, "callback-port", 9999, "local port to listen on for the OAuth callback")
	tokenCmd.PersistentFlags().StringVar(&tokenFile, "token-file", "", "optional path to save the received ID Token")
	rootCmd.AddCommand(tokenCmd)
}

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Auth token utilities",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Interactive Google Login to get an ID Token",
	Run: func(cmd *cobra.Command, args []string) {
		path := clientSecretFile
		if workspace := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); workspace != "" {
			if !filepath.IsAbs(path) {
				path = filepath.Join(workspace, path)
			}
		}

		b, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Unable to read client secret file %q: %v", path, err)
		}

		config, err := google.ConfigFromJSON(b, "openid", "profile", "email")
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}

		// Use a local server for the callback. 
		// If the authorized redirect is just "http://localhost", we use that.
		// Google's OAuth2 implementation for Web clients is strict on port 80 
		// if no port is specified in the authorized list.
		config.RedirectURL = "http://localhost"
		if callbackPort != 80 && callbackPort != 0 {
			config.RedirectURL = fmt.Sprintf("http://localhost:%d", callbackPort)
		}

		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		fmt.Printf("Go to the following link in your browser then type the "+
			"authorization code: \n%v\n", authURL)

		// Start a temporary server to receive the code
		codeChan := make(chan string)
		server := &http.Server{Addr: fmt.Sprintf(":%d", callbackPort)}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			if code != "" {
				fmt.Fprintf(w, "Login successful! You can close this window.")
				codeChan <- code
			} else {
				fmt.Fprintf(w, "Login failed!")
			}
		})

		go func() {
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("HTTP server ListenAndServe: %v", err)
			}
		}()

		authCode := <-codeChan
		server.Shutdown(context.Background())

		tok, err := config.Exchange(context.Background(), authCode)
		if err != nil {
			log.Fatalf("Unable to retrieve token from web: %v", err)
		}

		idToken, ok := tok.Extra("id_token").(string)
		if !ok {
			log.Fatalf("No id_token found in response")
		}

		fmt.Printf("\nID Token:\n%s\n", idToken)

		if tokenFile != "" {
			if err := os.MkdirAll(filepath.Dir(tokenFile), 0755); err != nil {
				log.Fatalf("Unable to create directories for token file: %v", err)
			}
			if err := os.WriteFile(tokenFile, []byte(idToken), 0600); err != nil {
				log.Fatalf("Unable to write token to file: %v", err)
			}
			fmt.Printf("Token saved to: %s\n", tokenFile)
		}
	},
}

func init() {
	tokenCmd.AddCommand(loginCmd)
}
