package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

var (
	addr           = flag.String("addr", "", "the address to connect to (if empty, starts in-process server)")
	clientSecret   = flag.String("client-secret", "", "path to Google Client Secret JSON")
	aliceRefresh   = flag.String("alice-refresh", "", "path to Alice's refresh token")
	bobRefresh     = flag.String("bob-refresh", "", "path to Bob's refresh token")
	googleAudience = flag.String("google-audience", "964146605436-cn068f6livloacebbi0itbhgvh9t5uns.apps.googleusercontent.com", "Google OAuth Client ID")
)

func TestIntegration(t *testing.T) {
	if !flag.Parsed() {
		flag.Parse()
	}

	// Fallback for local dev if flags are empty
	if *clientSecret == "" {
		*clientSecret = "secrets/client_secret_dev.json"
	}
	if *aliceRefresh == "" {
		*aliceRefresh = "secrets/ruthless.alice.sec"
	}
	if *bobRefresh == "" {
		*bobRefresh = "secrets/ruthless.bob.sec"
	}

	// Resolve paths relative to execution root if needed (for Bazel)
	for _, p := range []*string{clientSecret, aliceRefresh, bobRefresh} {
		if _, err := os.Stat(*p); os.IsNotExist(err) {
			// Try going up 3 levels (backend/scripts/integration -> root)
			newPath := "../../../" + *p
			if _, err := os.Stat(newPath); err == nil {
				*p = newPath
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var connectAddr string
	var cleanup func()
	var store storage.Storage

	if *addr != "" {
		connectAddr = *addr
	} else {
		var err error
		connectAddr, store, cleanup, err = testutil.StartTestServer(ctx, *googleAudience)
		if err != nil {
			t.Fatalf("Failed to start test server: %v", err)
		}
		defer cleanup()
	}

	client := testutil.NewTestClient(connectAddr, "")
	defer client.Close()

	// Initialize identities
	if *aliceRefresh != "" {
		ts, err := testutil.TokenSourceFromRefresh(ctx, *clientSecret, *aliceRefresh)
		if err != nil {
			t.Fatalf("Failed to create Alice token source: %v", err)
		}
		client.UserTokenSources["Alice"] = ts
		if _, err := client.EnsureUserRegistered(ctx, "Alice", store); err != nil {
			t.Fatalf("Failed to ensure Alice is registered: %v", err)
		}
	}

	if *bobRefresh != "" {
		ts, err := testutil.TokenSourceFromRefresh(ctx, *clientSecret, *bobRefresh)
		if err != nil {
			t.Fatalf("Failed to create Bob token source: %v", err)
		}
		client.UserTokenSources["Bob"] = ts
		if _, err := client.EnsureUserRegistered(ctx, "Bob", store); err != nil {
			t.Logf("Warning: Failed to ensure Bob is registered: %v", err)
		}
	}

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	t.Run("AuthTests", func(t *testing.T) {
		runAuthTests(t, ctx, client, runID)
	})

	t.Run("DeckTests", func(t *testing.T) {
		runDeckTests(t, ctx, client, runID)
	})

	t.Run("GameTests", func(t *testing.T) {
		runGameTests(t, ctx, client, runID)
	})

	t.Log("✅ All Integration Tests Passed!")
}
