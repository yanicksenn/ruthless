package testutil

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/auth"
	"github.com/yanicksenn/ruthless/backend/internal/server"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"github.com/yanicksenn/ruthless/backend/internal/storage/memory"
)

type TestClient struct {
	conn             *grpc.ClientConn
	CardClient       pb.CardServiceClient
	DeckClient       pb.DeckServiceClient
	SessionClient    pb.SessionServiceClient
	SessionInvitationClient pb.SessionInvitationServiceClient
	GameClient       pb.GameServiceClient
	UserClient       pb.UserServiceClient
	FriendClient     pb.FriendServiceClient
	NotificationClient pb.NotificationServiceClient
	AuthSecret       string
	UserTokenSources map[string]oauth2.TokenSource
	Store            storage.Storage
}

func NewTestClient(addr string, authSecret string, store storage.Storage) *TestClient {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return &TestClient{
		conn:          conn,
		CardClient:    pb.NewCardServiceClient(conn),
		DeckClient:    pb.NewDeckServiceClient(conn),
		SessionClient: pb.NewSessionServiceClient(conn),
		SessionInvitationClient: pb.NewSessionInvitationServiceClient(conn),
		GameClient:    pb.NewGameServiceClient(conn),
		UserClient:    pb.NewUserServiceClient(conn),
		FriendClient:  pb.NewFriendServiceClient(conn),
		NotificationClient: pb.NewNotificationServiceClient(conn),
		AuthSecret:    authSecret,
		UserTokenSources: make(map[string]oauth2.TokenSource),
		Store:         store,
	}
}

func (c *TestClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func StartTestServer(ctx context.Context, googleAudience string) (string, storage.Storage, func(), error) {
	store := memory.New()
	
	authenticator := auth.NewFakeAuthenticator()

	srv := server.New(store, authenticator, &pb.Config{})
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to listen: %v", err)
	}
	addr := listener.Addr().String()

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(srv.UnaryAuthInterceptor()),
	)
	srv.RegisterWithGRPC(grpcServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	cleanup := func() {
		grpcServer.Stop()
		listener.Close()
	}

	return addr, store, cleanup, nil
}


func (c *TestClient) GetAuthContextForUser(ctx context.Context, name string) (context.Context, error) {
	ts, ok := c.UserTokenSources[name]
	if !ok {
		return nil, fmt.Errorf("token source for user %q not initialized", name)
	}
	tok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token for %q: %v", name, err)
	}
	idToken, err := GetIDToken(tok)
	if err != nil {
		return nil, err
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+idToken), nil
}

func SeedUser(ctx context.Context, store storage.Storage, id, name string) error {
	return store.CreateUser(ctx, &pb.User{Id: id, Name: name})
}

func (c *TestClient) EnsureUserRegistered(ctx context.Context, name string, store storage.Storage) (string, error) {
	ts, ok := c.UserTokenSources[name]
	if !ok {
		return "", fmt.Errorf("token source for user %q not initialized", name)
	}

	tok, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get %s token: %v", name, err)
	}

	idToken, err := GetIDToken(tok)
	if err != nil {
		return "", err
	}

	sub := GetSub(idToken)
	if sub == "" {
		return "", fmt.Errorf("failed to extract sub from %s token", name)
	}

	if store != nil {
		// In-process server: seed directly for speed
		if err := SeedUser(ctx, store, sub, name); err != nil {
			return "", err
		}
		return sub, nil
	}

	// External server: use API
	authCtx := WithAuthToken(ctx, idToken)
	_, err = c.UserClient.GetMe(authCtx, &pb.GetMeRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to ensure %s is registered: %v", name, err)
	}

	return sub, nil
}

func (c *TestClient) GetAuthContext(ctx context.Context, name string) context.Context {
	token := c.signToken(name)
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

func WithAuthToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

func (c *TestClient) signToken(name string) string {
	if c.AuthSecret == "" {
		return name
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  name,
		"name": name,
	})
	signed, err := t.SignedString([]byte(c.AuthSecret))
	if err != nil {
		log.Fatalf("could not sign token for %s: %v", name, err)
	}
	return signed
}


func AssertError(t testing.TB, err error, code codes.Code, msgContains string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error %v, but got success", code)
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("Expected gRPC status error, got: %v", err)
	}
	if st.Code() != code {
		t.Fatalf("Expected error code %v, got %v (msg: %s)", code, st.Code(), st.Message())
	}
	if msgContains != "" && !strings.Contains(strings.ToLower(st.Message()), strings.ToLower(msgContains)) {
		t.Fatalf("Expected error message to contain %q, but got %q", msgContains, st.Message())
	}
}

func AssertSuccess(t testing.TB, err error, action string) {
	t.Helper()
	if err != nil {
		t.Fatalf("Action %q failed unexpectedly: %v", action, err)
	}
}

func CountBlanks(text string) int {
	return strings.Count(text, "___")
}
