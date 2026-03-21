package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	googleoauth2 "golang.org/x/oauth2/google"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

var (
	ErrUnauthorized = errors.New("unauthorized request")
)

type Authenticator interface {
	// Authenticate validates a request (e.g., via token string) and returns the authenticated Player
	Authenticate(ctx context.Context, token string) (*pb.Player, error)
}

// --- Token Logic ---

type TokenGenerator struct {
	secret []byte
}

func NewTokenGenerator(secret string) *TokenGenerator {
	return &TokenGenerator{secret: []byte(secret)}
}

func (g *TokenGenerator) Generate(user *pb.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.Id,
		"name": user.Name,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(g.secret)
}

func (g *TokenGenerator) Validate(tokenString string) (*pb.Player, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	id, _ := claims["sub"].(string)
	name, _ := claims["name"].(string)

	if id == "" {
		return nil, fmt.Errorf("missing subject in token")
	}

	return &pb.Player{
		Id:   id,
		Name: name,
	}, nil
}

type TokenAuthenticator struct {
	generator *TokenGenerator
	store     storage.Storage
}

func NewTokenAuthenticator(generator *TokenGenerator, store storage.Storage) *TokenAuthenticator {
	return &TokenAuthenticator{generator: generator, store: store}
}

func (a *TokenAuthenticator) Authenticate(ctx context.Context, tokenString string) (*pb.Player, error) {
	player, err := a.generator.Validate(tokenString)
	if err != nil {
		return nil, ErrUnauthorized
	}
	
	isRevoked, err := a.store.IsTokenRevoked(ctx, tokenString)
	if err != nil {
		return nil, ErrUnauthorized
	}
	if isRevoked {
		return nil, ErrUnauthorized
	}

	return player, nil
}

// --- OAuth2 Logic ---

type ExchangeResult struct {
	ID    string
	Name  string
	Email string
}

type Exchanger struct {
	config   *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewExchanger(ctx context.Context, clientID, clientSecret, redirectURL string) (*Exchanger, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %v", err)
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     googleoauth2.Endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &Exchanger{
		config:   config,
		verifier: verifier,
	}, nil
}

func (e *Exchanger) AuthURL(state string) string {
	return e.config.AuthCodeURL(state)
}

func (e *Exchanger) Exchange(ctx context.Context, code string) (*ExchangeResult, error) {
	oauth2Token, err := e.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %v", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in response")
	}

	idToken, err := e.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %v", err)
	}

	var claims struct {
		Subject       string `json:"sub"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %v", err)
	}

	if !claims.EmailVerified {
		return nil, fmt.Errorf("email not verified")
	}

	return &ExchangeResult{
		ID:    claims.Subject,
		Name:  claims.Name,
		Email: claims.Email,
	}, nil
}

// --- HTTP Handler ---

type Handler struct {
	store     storage.Storage
	exchanger *Exchanger
	generator *TokenGenerator
	uiURL     string
	OnLogin   func(ctx context.Context, userId string)
}

func NewHandler(store storage.Storage, exchanger *Exchanger, generator *TokenGenerator, uiURL string, onLogin func(ctx context.Context, userId string)) *Handler {
	return &Handler{
		store:     store,
		exchanger: exchanger,
		generator: generator,
		uiURL:     uiURL,
		OnLogin:   onLogin,
	}
}

func (h *Handler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/auth/google", h.HandleLogin)
	mux.HandleFunc("/auth/google/callback", h.HandleCallback)
	mux.HandleFunc("/auth/logout", h.HandleLogout)
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := make([]byte, 16)
	rand.Read(state)
	stateStr := base64.URLEncoding.EncodeToString(state)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    stateStr,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   300,
	})

	http.Redirect(w, r, h.exchanger.AuthURL(stateStr), http.StatusTemporaryRedirect)
}

func (h *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value == "" || cookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	result, err := h.exchanger.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Exchange failed: %v", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	user, err := h.getOrCreateUser(r.Context(), result)
	if err != nil {
		log.Printf("User resolution failed: %v", err)
		http.Error(w, "Failed to resolve user", http.StatusInternalServerError)
		return
	}

	sessionToken, err := h.generator.Generate(user)
	if err != nil {
		log.Printf("Token generation failed: %v", err)
		http.Error(w, "Failed to generate session", http.StatusInternalServerError)
		return
	}

	if h.OnLogin != nil {
		h.OnLogin(r.Context(), user.Id)
	}

	target, err := url.Parse(h.uiURL)
	if err != nil {
		http.Error(w, "Invalid UI URL", http.StatusInternalServerError)
		return
	}
	target.Fragment = "token=" + sessionToken

	http.Redirect(w, r, target.String(), http.StatusTemporaryRedirect)
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization", http.StatusUnauthorized)
		return
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Validate the token just to extract the expiration time
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.generator.secret, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid claims", http.StatusUnauthorized)
		return
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		http.Error(w, "Missing expiration", http.StatusUnauthorized)
		return
	}
	expTime := time.Unix(int64(expFloat), 0)

	err = h.store.RevokeToken(r.Context(), tokenString, expTime)
	if err != nil {
		http.Error(w, "Failed to revoke token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getOrCreateUser(ctx context.Context, result *ExchangeResult) (*pb.User, error) {
	user, err := h.store.GetUser(ctx, result.ID)
	if err == nil {
		return user, nil
	}

	if err != storage.ErrNotFound {
		return nil, err
	}

	name := result.Name
	if name == "" {
		name = result.Email
	}

	newUser := &pb.User{
		Id:   result.ID,
		Name: name,
	}

	if err := h.store.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// --- Fake Authenticator (for development) ---

type FakeAuthenticator struct{}

func NewFakeAuthenticator() *FakeAuthenticator {
	return &FakeAuthenticator{}
}

func (a *FakeAuthenticator) Authenticate(ctx context.Context, token string) (*pb.Player, error) {
	if token == "" {
		return nil, ErrUnauthorized
	}
	return &pb.Player{
		Id:   token,
		Name: token,
	}, nil
}

var _ Authenticator = (*FakeAuthenticator)(nil)
