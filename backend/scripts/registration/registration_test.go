package main

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

var (
	addr           = flag.String("addr", "localhost:8080", "the address of the active backend")
	clientSecret   = flag.String("client-secret", "secrets/client_secret_dev.json", "path to Google Client Secret JSON")
	googleAudience = flag.String("google-audience", "964146605436-cn068f6livloacebbi0itbhgvh9t5uns.apps.googleusercontent.com", "Google OAuth Client ID")
	callbackPort   = flag.Int("callback-port", 9999, "port for OAuth callback")
	nocache        = flag.Bool("nocache", false, "dummy flag to prevent failure when passed by user")
	verbose        = flag.Bool("v", false, "alias for verbose output")
)

func TestInteractiveRegistration(t *testing.T) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if *addr == "" {
		t.Fatal("--addr is required for interactive registration test")
	}

	// Resolve paths relative to execution root if needed (for Bazel)
	if _, err := os.Stat(*clientSecret); os.IsNotExist(err) {
		// Try going up 3 levels (backend/scripts/registration -> root)
		newPath := "../../../" + *clientSecret
		if _, err := os.Stat(newPath); err == nil {
			*clientSecret = newPath
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	t.Log("--- Interactive Registration Test ---")
	t.Log("This test will prompt you to log in via your browser.")

	// 1. Interactive Login
	idToken, err := testutil.InteractiveLogin(ctx, *clientSecret, *callbackPort)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	client := testutil.NewTestClient(*addr, "", nil)
	defer client.Close()
	
	authCtx := testutil.WithAuthToken(ctx, idToken)

	// 2. Ensure registered
	t.Log("Verifying user registration via GetMe...")
	user, err := client.UserClient.GetMe(authCtx, &pb.GetMeRequest{})
	if err != nil {
		t.Fatalf("GetMe failed: %v", err)
	}
	t.Logf("Successfully verified user: %s (ID: %s)", user.Name, user.Id)

	t.Log("✅ Registration/Login Verified Successful!")
}
