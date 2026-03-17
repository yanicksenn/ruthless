package main

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

func runAuthTests(t *testing.T, ctx context.Context, c *testutil.TestClient, runID string) {
	t.Log("\n--- Auth & Authentication Suite ---")

	// 1. SUCCESS: GetMe Alice (Fake Token)
	aliceCtx := c.GetAuthContext(ctx, "Alice")
	t.Log("  [RUN] GetMe Alice (Fake Token)...")
	_, err := c.UserClient.GetMe(aliceCtx, &pb.GetMeRequest{})
	testutil.AssertSuccess(t, err, "GetMe Alice")

	// 2. SUCCESS: GetMe Charlie (Fake Token) -> Pending -> Complete
	charlieName := "AuthCharlie_" + runID
	charlieCtx := c.GetAuthContext(ctx, charlieName)
	_, err = c.UserClient.Register(charlieCtx, &pb.RegisterRequest{})
	testutil.AssertSuccess(t, err, "Register Charlie (Fake)")

	// GetMe should show pending
	me, err := c.UserClient.GetMe(charlieCtx, &pb.GetMeRequest{})
	testutil.AssertSuccess(t, err, "GetMe Charlie (Pending)")
	if !me.PendingCompletion {
		t.Errorf("Expected Charlie to be pending completion")
	}

	// Another RPC should be blocked
	_, err = c.CardClient.CreateCard(charlieCtx, &pb.CreateCardRequest{Text: "Illegal Card"})
	testutil.AssertError(t, err, codes.FailedPrecondition, "registration completion required")

	// Complete Registration
	_, err = c.UserClient.CompleteRegistration(charlieCtx, &pb.CompleteRegistrationRequest{Name: charlieName})
	testutil.AssertSuccess(t, err, "CompleteRegistration Charlie")

	// GetMe should now be complete
	me, err = c.UserClient.GetMe(charlieCtx, &pb.GetMeRequest{})
	testutil.AssertSuccess(t, err, "GetMe Charlie (Complete)")
	if me.PendingCompletion {
		t.Errorf("Expected Charlie to NOT be pending completion")
	}

	// 3. SUCCESS: Register Bob (Fake Token) -> Complete
	bobName := "AuthBob_" + runID
	bobCtx := c.GetAuthContext(ctx, bobName)
	_, err = c.UserClient.Register(bobCtx, &pb.RegisterRequest{})
	testutil.AssertSuccess(t, err, "Register Bob")
	_, err = c.UserClient.CompleteRegistration(bobCtx, &pb.CompleteRegistrationRequest{Name: bobName})
	testutil.AssertSuccess(t, err, "CompleteRegistration Bob")

	// 4. FAILURE: Unauthenticated GetMe
	_, err = c.UserClient.GetMe(ctx, &pb.GetMeRequest{})
	testutil.AssertError(t, err, codes.Unauthenticated, "missing Authorization")
}
