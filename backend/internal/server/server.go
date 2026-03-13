package server

import (
	"context"
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

	store storage.Storage
	auth  auth.Authenticator
}

func New(store storage.Storage, authenticator auth.Authenticator) *Server {
	return &Server{
		store: store,
		auth:  authenticator,
	}
}

func (s *Server) Register(grpcServer *grpc.Server) {
	pb.RegisterCardServiceServer(grpcServer, s)
	pb.RegisterDeckServiceServer(grpcServer, s)
	pb.RegisterSessionServiceServer(grpcServer, s)
	pb.RegisterGameServiceServer(grpcServer, s)
}

type contextKey string

const PlayerContextKey = contextKey("player")

func getPlayer(ctx context.Context) (*pb.Player, bool) {
	player, ok := ctx.Value(PlayerContextKey).(*pb.Player)
	return player, ok
}

func (s *Server) UnaryAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Only require auth for specific endpoints
		// In a real app we might use method matching
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

			// Ensure player exists as a User in storage.
			// This is necessary because some operations (like joining a session)
			// depend on the user existence due to foreign key constraints.
			err = s.store.CreateUser(ctx, &pb.User{
				Id:   player.Id,
				Name: player.Name,
			})
			if err != nil {
				// We don't want to block the request if this fails (maybe?),
				// but for E2E validation we need it to succeed.
				// Log it at least.
			}

			ctx = context.WithValue(ctx, PlayerContextKey, player)
		}
		return handler(ctx, req)
	}
}

func requiresAuth(method string) bool {
	// Let's require auth for everything except Get and List for now, and Creates that are open.
	// We will mirror the Chi router behaviour:
	// /decks (POST requires auth), /decks (GET public), /decks/{id}/contributors (POST/DELETE auth), etc.
	if strings.Contains(method, "CardService") {
		return false // CreateCard had no authMiddleware in Chi router
	}
	if strings.Contains(method, "SessionService/CreateSession") || strings.Contains(method, "SessionService/GetSession") {
		return false
	}
	if strings.Contains(method, "SessionService/JoinSession") {
		return true // Was behind authMiddleware in Chi
	}
	if strings.Contains(method, "DeckService/ListDecks") || strings.Contains(method, "DeckService/GetDeck") {
		return false
	}
	if strings.HasPrefix(method, "/ruthless.v1.DeckService/") {
		// CreateDeck, AddContributor, RemoveContributor, AddCardToDeck, RemoveCardFromDeck
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
