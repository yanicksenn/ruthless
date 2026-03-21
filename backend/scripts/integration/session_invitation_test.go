package main

import (
	"context"
	"fmt"
	"testing"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
	"google.golang.org/grpc/codes"
)

func runSessionInvitationTests(t *testing.T, ctx context.Context, client *testutil.TestClient, runID string) {
	aliceCtx := client.GetAuthContext(ctx, "Alice")
	bobCtx := client.GetAuthContext(ctx, "Bob")
    
	_, err := client.UserClient.GetMe(aliceCtx, &pb.GetMeRequest{})
	testutil.AssertSuccess(t, err, "GetMe Alice")
	bobProfile, err := client.UserClient.GetMe(bobCtx, &pb.GetMeRequest{})
	testutil.AssertSuccess(t, err, "GetMe Bob")

	// 1. Alice creates a session
	sessRes, err := client.SessionClient.CreateSession(aliceCtx, &pb.CreateSessionRequest{
		Name: "Alice's Game",
	})
	testutil.AssertSuccess(t, err, "CreateSession Alice")
	sessionID := sessRes.Id

	// 2. Alice invites Bob
	_, err = client.SessionInvitationClient.InviteFriendToSession(aliceCtx, &pb.InviteFriendToSessionRequest{
		SessionId: sessionID,
		FriendIdentifier: fmt.Sprintf("%s#%s", bobProfile.Name, bobProfile.Identifier),
	})
	// Should fail because they are not friends!
	testutil.AssertError(t, err, codes.PermissionDenied, "friends")

	// 2.1 Make them friends
	_, err = client.FriendClient.InviteFriend(aliceCtx, &pb.InviteFriendRequest{
		Identifier: fmt.Sprintf("%s#%s", bobProfile.Name, bobProfile.Identifier),
	})
	testutil.AssertSuccess(t, err, "InviteFriend")
    // Note: bob needs his invitation id
    invs, err := client.FriendClient.ListInvitations(bobCtx, &pb.ListInvitationsRequest{})
	testutil.AssertSuccess(t, err, "ListInvitations")
    if len(invs.Invitations) > 0 {
		_, err = client.FriendClient.RespondToInvitation(bobCtx, &pb.RespondToInvitationRequest{InvitationId: invs.Invitations[0].Id, Accept: true})
		testutil.AssertSuccess(t, err, "AcceptFriendRequest")
	}

	// 3. Retry Session Invitation
	_, err = client.SessionInvitationClient.InviteFriendToSession(aliceCtx, &pb.InviteFriendToSessionRequest{
		SessionId: sessionID,
		FriendIdentifier: fmt.Sprintf("%s#%s", bobProfile.Name, bobProfile.Identifier),
	})
	testutil.AssertSuccess(t, err, "InviteFriendToSession Alice -> Bob")

	// 4. Bob lists session invitations
	listRes, err := client.SessionInvitationClient.ListSessionInvitations(bobCtx, &pb.ListSessionInvitationsRequest{})
	testutil.AssertSuccess(t, err, "ListSessionInvitations Bob")
	if len(listRes.Invitations) == 0 {
		t.Fatal("Expected 1 session invitation for Bob, got 0")
	}
	inv := listRes.Invitations[0]

	// 5. Bob accepts
	acceptRes, err := client.SessionInvitationClient.RespondToSessionInvitation(bobCtx, &pb.RespondToSessionInvitationRequest{
		InvitationId: inv.Id,
		Accept:       true,
	})
	testutil.AssertSuccess(t, err, "AcceptSessionInvitation Bob")
	if acceptRes.SessionId != sessionID {
		t.Errorf("Expected joined session %s, got %s", sessionID, acceptRes.SessionId)
	}

	// 6. Verify Bob is in session
	getRes, err := client.SessionClient.GetSession(bobCtx, &pb.GetSessionRequest{Id: sessionID})
	testutil.AssertSuccess(t, err, "GetSession")
	found := false
	for _, pid := range getRes.PlayerIds {
		if pid == "Bob" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Bob not found in session players")
	}

	// 7. Verify ListFriends excludes Bob for Alice
	excludeSessionIDStr := sessionID
	friendsRes, err := client.FriendClient.ListFriends(aliceCtx, &pb.ListFriendsRequest{
		ExcludeFromSessionId: &excludeSessionIDStr,
	})
	testutil.AssertSuccess(t, err, "ListFriends with ExcludeFromSessionId Alice")
	for _, f := range friendsRes.Friends {
		if f.Id == "Bob" {
			t.Error("Bob still found in Alice's friends list despite exclude_from_session_id")
		}
	}
}
