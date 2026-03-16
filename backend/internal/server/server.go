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

type contextKey string

const PlayerContextKey = contextKey("player")

func getPlayer(ctx context.Context) (*pb.Player, bool) {
	player, ok := ctx.Value(PlayerContextKey).(*pb.Player)
	return player, ok
}

func (s *Server) UnaryAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if requiresAuth(info.FullMethod) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
			}

			authHeaders := md.Get("authorization")
			if len(authHeaders) == 0 {
				return nil, status.Errorf(codes.Unauthenticated, "missing Authorization header")
			}

			token := strings.TrimPrefix(authHeaders[0], "Bearer ")
			player, err := s.auth.Authenticate(ctx, token)
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "invalid token")
			}

			// Verify user exists in storage (unless registering)
			if !strings.HasSuffix(info.FullMethod, "UserService/Register") {
				_, err := s.store.GetUser(ctx, player.Id)
				if err != nil {
					return nil, status.Errorf(codes.PermissionDenied, "user not registered: %v", err)
				}
			}

			ctx = context.WithValue(ctx, PlayerContextKey, player)
		}
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
	// Let's require auth for everything except Get and List for now, and Creates that are open.
	// We will mirror the Chi router behaviour:
	// /decks (POST requires auth), /decks (GET public), /decks/{id}/contributors (POST/DELETE auth), etc.
	if strings.Contains(method, "CardService/CreateCard") || strings.Contains(method, "CardService/DeleteCard") {
		return true
	}
	if strings.HasPrefix(method, "/ruthless.v1.SessionService/") {
		if strings.Contains(method, "GetSession") || strings.Contains(method, "ListSessions") {
			return false
		}
		return true
	}
	if strings.HasPrefix(method, "/ruthless.v1.DeckService/") {
		if strings.Contains(method, "ListDecks") || strings.Contains(method, "GetDeck") {
			return false
		}
		return true
	}
	if strings.HasPrefix(method, "/ruthless.v1.UserService/") {
		return true
	}
	if strings.HasPrefix(method, "/ruthless.v1.GameService/") {
		if strings.Contains(method, "GetGame") || strings.Contains(method, "CreateGame") {
			return false
		}
		return true
	}
	return false
}
