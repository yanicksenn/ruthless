package main

import (
	"context"
	"fmt"
	"testing"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
	"google.golang.org/grpc/codes"
)

func runFriendTests(t *testing.T, ctx context.Context, client *testutil.TestClient, runID string) {
	aliceCtx := client.GetAuthContext(ctx, "Alice")
	bobCtx := client.GetAuthContext(ctx, "Bob")

	// Get Alice's and Bob's profiles to get their identifiers
	_, err := client.UserClient.GetMe(aliceCtx, &pb.GetMeRequest{})
	testutil.AssertSuccess(t, err, "GetMe Alice")
	
	bobProfile, err := client.UserClient.GetMe(bobCtx, &pb.GetMeRequest{})
	testutil.AssertSuccess(t, err, "GetMe Bob")

	// 1. Alice invites Bob
	_, err = client.FriendClient.InviteFriend(aliceCtx, &pb.InviteFriendRequest{
		Identifier: fmt.Sprintf("%s#%s", bobProfile.Name, bobProfile.Identifier),
	})
	testutil.AssertSuccess(t, err, "InviteFriend Alice -> Bob")

	// 2. Alice invites Bob again (should fail)
	_, err = client.FriendClient.InviteFriend(aliceCtx, &pb.InviteFriendRequest{
		Identifier: fmt.Sprintf("%s#%s", bobProfile.Name, bobProfile.Identifier),
	})
	testutil.AssertError(t, err, codes.AlreadyExists, "invitation already sent")

	// 3. Bob lists invitations
	invRes, err := client.FriendClient.ListInvitations(bobCtx, &pb.ListInvitationsRequest{})
	testutil.AssertSuccess(t, err, "ListInvitations Bob")
	if len(invRes.Invitations) == 0 {
		t.Fatal("Expected 1 invitation for Bob, got 0")
	}
	inv := invRes.Invitations[0]
	if inv.Sender.Id != "Alice" {
		t.Errorf("Expected sender Alice, got %s", inv.Sender.Id)
	}

	// 4. Bob accepts invitation
	_, err = client.FriendClient.RespondToInvitation(bobCtx, &pb.RespondToInvitationRequest{
		InvitationId: inv.Id,
		Accept:       true,
	})
	testutil.AssertSuccess(t, err, "RespondToInvitation Bob (Accept)")

	// 5. Alice lists friends
	friendsRes, err := client.FriendClient.ListFriends(aliceCtx, &pb.ListFriendsRequest{})
	testutil.AssertSuccess(t, err, "ListFriends Alice")
	found := false
	for _, f := range friendsRes.Friends {
		if f.Id == "Bob" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Bob not found in Alice's friends list")
	}

	// 6. Bob lists friends
	friendsRes, err = client.FriendClient.ListFriends(bobCtx, &pb.ListFriendsRequest{})
	testutil.AssertSuccess(t, err, "ListFriends Bob")
	found = false
	for _, f := range friendsRes.Friends {
		if f.Id == "Alice" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Alice not found in Bob's friends list")
	}

	// 7. Alice removes Bob
	_, err = client.FriendClient.RemoveFriend(aliceCtx, &pb.RemoveFriendRequest{
		FriendId: "Bob",
	})
	testutil.AssertSuccess(t, err, "RemoveFriend Alice -> Bob")

	// 8. Bob lists friends (should be empty)
	friendsRes, err = client.FriendClient.ListFriends(bobCtx, &pb.ListFriendsRequest{})
	testutil.AssertSuccess(t, err, "ListFriends Bob (After removal)")
	if len(friendsRes.Friends) > 0 {
		t.Errorf("Expected 0 friends for Bob, got %d", len(friendsRes.Friends))
	}
}
