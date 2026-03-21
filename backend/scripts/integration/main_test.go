package main

import (
	"context"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
)

var (
	addr = flag.String("addr", "", "the address to connect to (if empty, starts in-process server)")
)

func TestIntegration(t *testing.T) {
	if !flag.Parsed() {
		flag.Parse()
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
		connectAddr, store, cleanup, err = testutil.StartTestServer(ctx, "")
		if err != nil {
			t.Fatalf("Failed to start test server: %v", err)
		}
		defer cleanup()
	}

	client := testutil.NewTestClient(connectAddr, "", store)
	defer client.Close()

	// Initialize Alice
	if store != nil {
		store.CreateUser(ctx, &pb.User{Id: "Alice", Name: "Alice"})
	}
	aliceCtx := client.GetAuthContext(ctx, "Alice")
	if _, err := client.UserClient.CompleteRegistration(aliceCtx, &pb.CompleteRegistrationRequest{Name: "Alice"}); err != nil {
		t.Fatalf("Failed to complete registration for Alice: %v", err)
	}

	// Initialize Bob
	if store != nil {
		store.CreateUser(ctx, &pb.User{Id: "Bob", Name: "Bob"})
	}
	bobCtx := client.GetAuthContext(ctx, "Bob")
	if _, err := client.UserClient.CompleteRegistration(bobCtx, &pb.CompleteRegistrationRequest{Name: "Bob"}); err != nil {
		t.Fatalf("Failed to complete registration for Bob: %v", err)
	}

	// Initialize Charlie
	if store != nil {
		store.CreateUser(ctx, &pb.User{Id: "Charlie", Name: "Charlie"})
	}
	charlieCtx := client.GetAuthContext(ctx, "Charlie")
	if _, err := client.UserClient.CompleteRegistration(charlieCtx, &pb.CompleteRegistrationRequest{Name: "Charlie"}); err != nil {
		t.Fatalf("Failed to complete registration for Charlie: %v", err)
	}

	// Initialize Dave
	if store != nil {
		store.CreateUser(ctx, &pb.User{Id: "Dave", Name: "Dave"})
	}
	daveCtx := client.GetAuthContext(ctx, "Dave")
	if _, err := client.UserClient.CompleteRegistration(daveCtx, &pb.CompleteRegistrationRequest{Name: "Dave"}); err != nil {
		t.Fatalf("Failed to complete registration for Dave: %v", err)
	}

	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	t.Run("AuthTests", func(t *testing.T) {
		runAuthTests(t, ctx, client, runID)
	})

	t.Run("CardTests", func(t *testing.T) {
		runCardTests(t, ctx, client, runID)
	})

	t.Run("DeckTests", func(t *testing.T) {
		runDeckTests(t, ctx, client, runID)
	})

	t.Run("GameTests", func(t *testing.T) {
		runGameTests(t, ctx, client, runID)
	})
	
	t.Run("SessionListTests", func(t *testing.T) {
		runSessionListTests(t, ctx, client, runID)
	})

	t.Run("AbandonmentTests", func(t *testing.T) {
		runAbandonmentTests(t, ctx, client, runID)
	})

	t.Run("VisibilityTests", func(t *testing.T) {
		runVisibilityTests(t, ctx, client, runID)
	})

	t.Run("FriendTests", func(t *testing.T) {
		runFriendTests(t, ctx, client, runID)
	})

	t.Run("NotificationTests", func(t *testing.T) {
		runNotificationTests(t, ctx, client, runID)
	})

	t.Run("SessionInvitationTests", func(t *testing.T) {
		runSessionInvitationTests(t, ctx, client, runID)
	})

	t.Log("✅ All Integration Tests Passed!")
}
