package server

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/auth"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedCardServiceServer
	pb.UnimplementedDeckServiceServer
	pb.UnimplementedSessionServiceServer
	pb.UnimplementedGameServiceServer
	pb.UnimplementedUserServiceServer

	store  storage.Storage
	auth   auth.Authenticator
	config *pb.Config
}

func New(store storage.Storage, authenticator auth.Authenticator, config *pb.Config) *Server {
	if config.Public == nil {
		config.Public = &pb.ConfigPublic{}
	}
	if config.Public.Registration == nil {
		config.Public.Registration = &pb.ConfigPublic_Registration{}
	}
	if config.Public.Limits == nil {
		config.Public.Limits = &pb.ConfigPublic_Limits{}
	}
	if config.Public.Game == nil {
		config.Public.Game = &pb.ConfigPublic_Game{}
	}
	if config.Private == nil {
		config.Private = &pb.ConfigPrivate{}
	}
	if config.Private.Registration == nil {
		config.Private.Registration = &pb.ConfigPrivate_Registration{}
	}

	// Detect development mode based on authenticator type
	if _, ok := authenticator.(*auth.FakeAuthenticator); ok {
		config.Public.IsDevelopment = true
	}


	return &Server{
		store:  store,
		auth:   authenticator,
		config: config,
	}
}

func (s *Server) RegisterWithGRPC(grpcServer *grpc.Server) {
	pb.RegisterCardServiceServer(grpcServer, s)
	pb.RegisterDeckServiceServer(grpcServer, s)
	pb.RegisterSessionServiceServer(grpcServer, s)
	pb.RegisterGameServiceServer(grpcServer, s)
	pb.RegisterUserServiceServer(grpcServer, s)
}

type PlayerKey struct{}

func getPlayer(ctx context.Context) (*pb.Player, bool) {
	player, ok := ctx.Value(PlayerKey{}).(*pb.Player)
	return player, ok
}

func (s *Server) UnaryAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !requiresAuth(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing authorization header")
		}

		token := strings.TrimPrefix(values[0], "Bearer ")
		player, err := s.auth.Authenticate(ctx, token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Check if user exists in our storage
		user, err := s.store.GetUser(ctx, player.Id)
		if err != nil {
			if err == storage.ErrNotFound {
				// This shouldn't happen with our new flow, but if it does, the user record is missing.
				return nil, status.Errorf(codes.Unauthenticated, "user profile not found")
			}
			return nil, status.Errorf(codes.Internal, "failed to fetch user: %v", err)
		}

		// If user is pending completion, block all requests except GetMe and CompleteRegistration
		if user.Identifier == "" {
			if !isAllowedWhilePending(info.FullMethod) {
				return nil, status.Errorf(codes.FailedPrecondition, "registration completion required")
			}
		}

		ctx = context.WithValue(ctx, PlayerKey{}, player)
		return handler(ctx, req)
	}
}

func (s *Server) UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			st, _ := status.FromError(err)
			log.Printf(" [RPC Error] Method: %s | Code: %s | Message: %s", info.FullMethod, st.Code(), st.Message())
		}
		return resp, err
	}
}

func requiresAuth(method string) bool {
	if strings.HasPrefix(method, "/ruthless.v1.CardService/") {
		if strings.Contains(method, "GetConfig") {
			return false
		}
		return true
	}
	if strings.HasPrefix(method, "/ruthless.v1.DeckService/") {
		if strings.Contains(method, "GetDeck") {
			return false
		}
		return true
	}
	if strings.HasPrefix(method, "/ruthless.v1.SessionService/") {
		if strings.Contains(method, "GetSession") {
			return false
		}
		return true
	}
	if strings.HasPrefix(method, "/ruthless.v1.GameService/") {
		if strings.Contains(method, "GetGame") || strings.Contains(method, "GetGameBySession") {
			return false
		}
		return true
	}
	if strings.HasPrefix(method, "/ruthless.v1.UserService/") {
		return true
	}

	return false
}

func isAllowedWhilePending(method string) bool {
	allowed := []string{
		"UserService/CompleteRegistration",
		"UserService/GetMe",
		"CardService/GetConfig",
	}
	for _, m := range allowed {
		if strings.HasSuffix(method, m) {
			return true
		}
	}
	return false
}
func (s *Server) GetConfig(ctx context.Context, _ *emptypb.Empty) (*pb.ConfigPublic, error) {
	return s.config.Public, nil
}
