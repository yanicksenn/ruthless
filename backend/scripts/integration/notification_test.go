package main

import (
	"context"
	"fmt"
	"testing"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

func runNotificationTests(t *testing.T, ctx context.Context, client *testutil.TestClient, runID string) {
	charlieCtx := client.GetAuthContext(ctx, "Charlie")
	daveCtx := client.GetAuthContext(ctx, "Dave")

	// Get Charlie's and Dave's profiles to get their identifiers
	_, err := client.UserClient.GetMe(charlieCtx, &pb.GetMeRequest{})
	testutil.AssertSuccess(t, err, "GetMe Charlie")
	
	daveProfile, err := client.UserClient.GetMe(daveCtx, &pb.GetMeRequest{})
	testutil.AssertSuccess(t, err, "GetMe Dave")

	// 1. Initial state: Dave should have 0 notifications
	notifRes, err := client.NotificationClient.GetNotifications(daveCtx, &pb.GetNotificationsRequest{})
	testutil.AssertSuccess(t, err, "GetNotifications Dave (Initial)")
	if len(notifRes.Notifications) > 0 {
		t.Fatalf("Expected 0 notifications for Dave, got %d", len(notifRes.Notifications))
	}

	// 2. Charlie invites Dave
	_, err = client.FriendClient.InviteFriend(charlieCtx, &pb.InviteFriendRequest{
		Identifier: fmt.Sprintf("%s#%s", daveProfile.Name, daveProfile.Identifier),
	})
	testutil.AssertSuccess(t, err, "InviteFriend Charlie -> Dave")

	// 3. Dave should now have 1 notification of type FRIENDS_PENDING_INVITATIONS
	notifRes, err = client.NotificationClient.GetNotifications(daveCtx, &pb.GetNotificationsRequest{})
	testutil.AssertSuccess(t, err, "GetNotifications Dave (After Invite)")
	if len(notifRes.Notifications) != 1 {
		t.Fatalf("Expected 1 notification for Dave, got %d", len(notifRes.Notifications))
	}
	if notifRes.Notifications[0].Type != pb.NotificationType_NOTIFICATION_TYPE_FRIENDS_PENDING_INVITATIONS {
		t.Errorf("Expected notification type FRIENDS_PENDING_INVITATIONS, got %v", notifRes.Notifications[0].Type)
	}
	if notifRes.Notifications[0].Count != 1 {
		t.Errorf("Expected notification count 1, got %d", notifRes.Notifications[0].Count)
	}

	// 4. Reset counter
	_, err = client.NotificationClient.ResetNotificationCounter(daveCtx, &pb.ResetNotificationCounterRequest{
		Type: pb.NotificationType_NOTIFICATION_TYPE_FRIENDS_PENDING_INVITATIONS,
	})
	testutil.AssertSuccess(t, err, "ResetNotificationCounter Dave")

	// 5. Verify counter is reset
	notifRes, err = client.NotificationClient.GetNotifications(daveCtx, &pb.GetNotificationsRequest{})
	testutil.AssertSuccess(t, err, "GetNotifications Dave (After Reset)")
	if len(notifRes.Notifications) > 0 {
		t.Fatalf("Expected 0 notifications for Dave, got %d", len(notifRes.Notifications))
	}
}
