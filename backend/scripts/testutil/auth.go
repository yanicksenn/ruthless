package testutil

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/golang-jwt/jwt/v5"
)

// TokenSourceFromRefresh creates an oauth2.TokenSource using a client secret file and a refresh token file.
func TokenSourceFromRefresh(ctx context.Context, clientSecretPath, refreshTokenPath string) (oauth2.TokenSource, error) {
	b, err := os.ReadFile(clientSecretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, "openid", "profile", "email")
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	refreshData, err := os.ReadFile(refreshTokenPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read refresh token file: %v", err)
	}

	token := &oauth2.Token{
		RefreshToken: string(refreshData),
	}

	return config.TokenSource(ctx, token), nil
}

// GetIDToken extracts the id_token from an oauth2.Token.
func GetIDToken(tok *oauth2.Token) (string, error) {
	idToken, ok := tok.Extra("id_token").(string)
	if !ok {
		return "", fmt.Errorf("no id_token found in response")
	}
	return idToken, nil
}

// InteractiveLogin performs an interactive OAuth flow and returns the ID token.
func InteractiveLogin(ctx context.Context, clientSecretPath string, port int) (string, error) {
	b, err := os.ReadFile(clientSecretPath)
	if err != nil {
		return "", fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, "openid", "profile", "email")
	if err != nil {
		return "", fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	config.RedirectURL = "http://localhost"
	if port != 80 && port != 0 {
		config.RedirectURL = fmt.Sprintf("http://localhost:%d", port)
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then login: \n%v\n", authURL)

	codeChan := make(chan string)
	errChan := make(chan error)

	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "Login successful! You can close this window.")
			codeChan <- code
		} else {
			fmt.Fprintf(w, "Login failed!")
			errChan <- fmt.Errorf("no code in callback")
		}
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case code := <-codeChan:
		server.Shutdown(ctx)
		tok, err := config.Exchange(ctx, code)
		if err != nil {
			return "", fmt.Errorf("unable to retrieve token from web: %v", err)
		}
		return GetIDToken(tok)
	case err := <-errChan:
		server.Shutdown(ctx)
		return "", err
	case <-ctx.Done():
		server.Shutdown(ctx)
		return "", ctx.Err()
	}
}

// GetSub extracts the subject (user ID) from a JWT ID token.
func GetSub(token string) string {
	tok, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return ""
	}
	if claims, ok := tok.Claims.(jwt.MapClaims); ok {
		if sub, ok := claims["sub"].(string); ok {
			return sub
		}
	}
	return ""
}
