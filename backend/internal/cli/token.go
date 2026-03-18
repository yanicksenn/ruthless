package cli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/yanicksenn/ruthless/api/v1"
)
var (
	clientSecretFile string
	callbackPort     int
	saveTo           string
)

func init() {
	tokenCmd.PersistentFlags().StringVar(&clientSecretFile, "client-secret", "secrets/client_secret_dev.json", "path to Google Client Secret JSON")
	tokenCmd.PersistentFlags().IntVar(&callbackPort, "callback-port", 9999, "local port to listen on for the OAuth callback")
	tokenCmd.PersistentFlags().StringVar(&saveTo, "save-to", "", "optional path to save the received ID Token")
	
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

		if saveTo != "" {
			if err := os.MkdirAll(filepath.Dir(saveTo), 0755); err != nil {
				log.Fatalf("Unable to create directories for token file: %v", err)
			}
			if err := os.WriteFile(saveTo, []byte(idToken), 0600); err != nil {
				log.Fatalf("Unable to write token to file: %v", err)
			}
		}
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register the current user",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewUserServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token, err := ResolveToken(cmd)
		if err != nil {
			log.Fatalf("Token error: %v", err)
		}
		if token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		}

	user, err := client.GetMe(ctx, &pb.GetMeRequest{})
	if err != nil {
		log.Fatalf("Retrieving user info failed: %v", err)
	}

	fmt.Printf("User: %s (ID: %s)\n", user.Name, user.Id)
	if user.PendingCompletion {
		fmt.Println("Status: PENDING REGISTRATION COMPLETION (Run 'cah user complete' if implemented, or use the UI)")
	} else {
		fmt.Printf("Status: ACTIVE (Alias: %s)\n", user.Identifier)
	}
	},
}

func init() {
	tokenCmd.AddCommand(loginCmd)
	tokenCmd.AddCommand(registerCmd)

	AddTokenFlags(tokenCmd)
}
