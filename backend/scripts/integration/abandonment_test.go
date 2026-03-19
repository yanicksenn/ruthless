package main

import (
	"context"
	"testing"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

func runAbandonmentTests(t *testing.T, ctx context.Context, c *testutil.TestClient, runID string) {
	aliceCtx := c.GetAuthContext(ctx, "Alice")
	bobCtx := c.GetAuthContext(ctx, "Bob")

	t.Log("\n--- Abandonment Suite ---")

	// 1. Create a session (Alice is owner)
	t.Log("  [RUN] Alice creates session...")
	session, err := c.SessionClient.CreateSession(aliceCtx, &pb.CreateSessionRequest{
		Name: "Abandonment Session",
	})
	testutil.AssertSuccess(t, err, "CreateSession")

	// 2. Bob joins
	t.Log("  [RUN] Bob joins session...")
	_, err = c.SessionClient.JoinSession(bobCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "Bob Join")

	// Verify Bob is in
	session, _ = c.SessionClient.GetSession(aliceCtx, &pb.GetSessionRequest{Id: session.Id})
	if len(session.PlayerIds) != 2 {
		t.Errorf("Expected 2 players, got %d", len(session.PlayerIds))
	}

	// 3. Alice (owner) leaves while state is WAITING
	t.Log("  [RUN] Alice (owner) leaves session while WAITING...")
	_, err = c.SessionClient.LeaveSession(aliceCtx, &pb.LeaveSessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "Alice Leave")

	// 4. Verify Game state is ABANDONED
	t.Log("  [RUN] Verifying game state is ABANDONED...")
	game, err := c.GameClient.GetGameBySession(aliceCtx, &pb.GetGameBySessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "GetGameBySession")
	if game.State != pb.GameState_GAME_STATE_ABANDONED {
		t.Errorf("Expected game state ABANDONED, got %v", game.State)
	}

	// 5. Verify Bob is removed from session
	t.Log("  [RUN] Verifying Bob was also removed from session participation...")
	session, _ = c.SessionClient.GetSession(bobCtx, &pb.GetSessionRequest{Id: session.Id})
	if len(session.PlayerIds) != 0 {
		t.Errorf("Expected 0 players (nobody left in session), got %d: %v", len(session.PlayerIds), session.PlayerIds)
	}
}
