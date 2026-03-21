package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
)

func (s *Server) GetNotifications(ctx context.Context, req *pb.GetNotificationsRequest) (*pb.GetNotificationsResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	notifications, err := s.store.GetNotifications(ctx, player.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get notifications: %v", err)
	}

	return &pb.GetNotificationsResponse{
		Notifications: notifications,
	}, nil
}

func (s *Server) ResetNotificationCounter(ctx context.Context, req *pb.ResetNotificationCounterRequest) (*pb.ResetNotificationCounterResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if req.Type == pb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "notification type must be specified")
	}

	err := s.store.ResetNotificationCounter(ctx, player.Id, req.Type)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset notification counter: %v", err)
	}

	return &pb.ResetNotificationCounterResponse{}, nil
}
