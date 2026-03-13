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

	client := testutil.NewTestClient(*addr, "")
	defer client.Close()
	
	authCtx := testutil.WithAuthToken(ctx, idToken)

	// 2. Register
	t.Log("Registering user...")
	user, err := client.UserClient.Register(authCtx, &pb.RegisterRequest{})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}
	t.Logf("Successfully registered: %s (ID: %s)", user.Name, user.Id)

	// 3. Verify
	t.Log("Verifying registration via GetMe...")
	me, err := client.UserClient.GetMe(authCtx, &pb.GetMeRequest{})
	if err != nil {
		t.Fatalf("GetMe failed: %v", err)
	}
	t.Logf("GetMe verified: Hello, %s!", me.Name)

	t.Log("✅ Interactive Registration Successful!")
}
