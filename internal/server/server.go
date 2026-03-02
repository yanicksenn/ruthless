package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/yanicksenn/ruthless/internal/auth"
	"github.com/yanicksenn/ruthless/internal/domain"
	"github.com/yanicksenn/ruthless/internal/storage"
)

type Server struct {
	router chi.Router
	store  storage.Storage
	auth   auth.Authenticator
}

func New(store storage.Storage, authenticator auth.Authenticator) *Server {
	s := &Server{
		router: chi.NewRouter(),
		store:  store,
		auth:   authenticator,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Route("/cards", func(r chi.Router) {
			r.Get("/", s.handleListCards)
			r.Post("/", s.handleCreateCard)
		})

		r.Route("/sessions", func(r chi.Router) {
			r.Post("/", s.handleCreateSession)
			// Authentication required for most session interaction
			r.With(s.authMiddleware).Post("/{id}/join", s.handleJoinSession)
			r.With(s.authMiddleware).Get("/{id}", s.handleGetSession)
		})
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// -- Middleware

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		player, err := s.auth.Authenticate(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Simplified for now: just log or add to context in real app.
		_ = player
		next.ServeHTTP(w, r)
	})
}

// -- Handlers

func (s *Server) handleListCards(w http.ResponseWriter, r *http.Request) {
	cards, err := s.store.ListCards(r.Context())
	if err != nil {
		http.Error(w, "failed to list cards", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

type createCardRequest struct {
	Text string `json:"text"`
}

func (s *Server) handleCreateCard(w http.ResponseWriter, r *http.Request) {
	var req createCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	card, err := domain.NewCard(req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.CreateCard(r.Context(), card); err != nil {
		http.Error(w, "failed to save card", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	session := domain.NewSession()

	if err := s.store.CreateSession(r.Context(), session); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

// Simplistic join - in a real app would extract user from context
type joinSessionRequest struct {
	PlayerName string `json:"player_name"` // Simulating simple join
}

func (s *Server) handleJoinSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req joinSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	session, err := s.store.GetSession(r.Context(), id)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	session.Players = append(session.Players, domain.NewPlayer(req.PlayerName))

	if err := s.store.UpdateSession(r.Context(), session); err != nil {
		http.Error(w, "failed to update session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	session, err := s.store.GetSession(r.Context(), id)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}
