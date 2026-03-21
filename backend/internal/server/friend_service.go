package server

import (
	"context"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) InviteFriend(ctx context.Context, req *pb.InviteFriendRequest) (*pb.InviteFriendResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.Identifier == "" {
		return nil, status.Error(codes.InvalidArgument, "identifier is required")
	}

	name, ident, err := parseIdentifier(req.Identifier)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identifier format: %v", err)
	}

	targetUser, err := s.store.GetUserByNameAndIdentifier(ctx, name, ident)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "user with identifier %s not found", req.Identifier)
		}
		return nil, status.Errorf(codes.Internal, "failed to lookup user: %v", err)
	}

	if targetUser.Id == player.Id {
		return nil, status.Error(codes.InvalidArgument, "you cannot invite yourself")
	}

	err = s.store.CreateInvitation(ctx, player.Id, targetUser.Id)
	if err != nil {
		if err == storage.ErrAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "invitation already sent")
		}
		return nil, status.Errorf(codes.Internal, "failed to send invitation: %v", err)
	}

	err = s.store.IncrementNotificationCounter(ctx, targetUser.Id, pb.NotificationType_NOTIFICATION_TYPE_FRIENDS_PENDING_INVITATIONS)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to increment notification counter: %v", err)
	}

	return &pb.InviteFriendResponse{}, nil
}

func (s *Server) RespondToInvitation(ctx context.Context, req *pb.RespondToInvitationRequest) (*pb.RespondToInvitationResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	invitation, err := s.store.GetInvitation(ctx, req.InvitationId)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, status.Error(codes.NotFound, "invitation not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get invitation: %v", err)
	}

	if invitation.Receiver.Id != player.Id {
		return nil, status.Error(codes.PermissionDenied, "this invitation was not sent to you")
	}

	if req.Accept {
		err = s.store.CreateFriendship(ctx, invitation.Sender.Id, invitation.Receiver.Id)
		if err != nil && err != storage.ErrAlreadyExists {
			return nil, status.Errorf(codes.Internal, "failed to create friendship: %v", err)
		}
	}

	err = s.store.DeleteInvitation(ctx, req.InvitationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete invitation: %v", err)
	}

	return &pb.RespondToInvitationResponse{}, nil
}

func (s *Server) ListFriends(ctx context.Context, req *pb.ListFriendsRequest) (*pb.ListFriendsResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	excludeSessionID := req.GetExcludeFromSessionId()
	excludeDeckID := req.GetExcludeFromDeckId()
	filter := req.GetFilter()
	friends, totalCount, err := s.store.ListFriends(ctx, player.Id, excludeSessionID, excludeDeckID, filter, req.PageSize, req.PageNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list friends: %v", err)
	}

	return &pb.ListFriendsResponse{
		Friends:    friends,
		TotalCount: totalCount,
	}, nil
}

func (s *Server) ListInvitations(ctx context.Context, req *pb.ListInvitationsRequest) (*pb.ListInvitationsResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	invitations, totalCount, err := s.store.ListInvitations(ctx, player.Id, req.PageSize, req.PageNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list invitations: %v", err)
	}

	return &pb.ListInvitationsResponse{
		Invitations: invitations,
		TotalCount:  totalCount,
	}, nil
}

func (s *Server) RemoveFriend(ctx context.Context, req *pb.RemoveFriendRequest) (*pb.RemoveFriendResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.FriendId == "" {
		return nil, status.Error(codes.InvalidArgument, "friend_id is required")
	}

	err := s.store.DeleteFriendship(ctx, player.Id, req.FriendId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove friend: %v", err)
	}

	return &pb.RemoveFriendResponse{}, nil
}
