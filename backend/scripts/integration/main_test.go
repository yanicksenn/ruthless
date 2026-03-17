package main

import (
	"context"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
	pb "github.com/yanicksenn/ruthless/api/v1"
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

	if *addr != "" {
		connectAddr = *addr
	} else {
		var err error
		connectAddr, _, cleanup, err = testutil.StartTestServer(ctx, "")
		if err != nil {
			t.Fatalf("Failed to start test server: %v", err)
		}
		defer cleanup()
	}

	client := testutil.NewTestClient(connectAddr, "")
	defer client.Close()

	// Initialize Alice
	aliceCtx := client.GetAuthContext(ctx, "Alice")
	if _, err := client.UserClient.Register(aliceCtx, &pb.RegisterRequest{}); err != nil {
		t.Fatalf("Failed to register Alice: %v", err)
	}
	if _, err := client.UserClient.CompleteRegistration(aliceCtx, &pb.CompleteRegistrationRequest{Name: "Alice"}); err != nil {
		t.Fatalf("Failed to complete registration for Alice: %v", err)
	}

	// Initialize Bob
	bobCtx := client.GetAuthContext(ctx, "Bob")
	if _, err := client.UserClient.Register(bobCtx, &pb.RegisterRequest{}); err != nil {
		t.Fatalf("Failed to register Bob: %v", err)
	}
	if _, err := client.UserClient.CompleteRegistration(bobCtx, &pb.CompleteRegistrationRequest{Name: "Bob"}); err != nil {
		t.Fatalf("Failed to complete registration for Bob: %v", err)
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

	t.Log("✅ All Integration Tests Passed!")
}
