package cli

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	
	"github.com/yanicksenn/ruthless/backend/internal/auth"
	"github.com/yanicksenn/ruthless/backend/internal/config"
	"github.com/yanicksenn/ruthless/backend/internal/server"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"github.com/yanicksenn/ruthless/backend/internal/storage/memory"
	"github.com/yanicksenn/ruthless/backend/internal/storage/postgres"
	ruthlespb "github.com/yanicksenn/ruthless/api/v1"
)

var (
	storageFlag string
	authFlag    string
	seedFlag    string
	dbConnStr   string
	authSecret     string
	googleAudience string
	configPath     string
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func serverRun(cmd *cobra.Command, args []string) {
	// Priority: Flag > Env > Default
	storageFlag = getEnv("STORAGE", storageFlag)
	authFlag = getEnv("AUTH", authFlag)
	dbConnStr = getEnv("DB_CONN_STR", dbConnStr)
	authSecret = getEnv("AUTH_SECRET", authSecret)
	googleAudience = getEnv("GOOGLE_AUDIENCE", getEnv("GOOGLE_CLIENT_ID", googleAudience))

	log.Printf("Starting server with storage=%s auth=%s", storageFlag, authFlag)

	// Setup Config
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize config blocks if they are nil to avoid panics
	if cfg.Public == nil {
		cfg.Public = &ruthlespb.ConfigPublic{}
	}
	if cfg.Public.Registration == nil {
		cfg.Public.Registration = &ruthlespb.ConfigPublic_Registration{}
	}
	if cfg.Public.Limits == nil {
		cfg.Public.Limits = &ruthlespb.ConfigPublic_Limits{}
	}
	if cfg.Public.Game == nil {
		cfg.Public.Game = &ruthlespb.ConfigPublic_Game{}
	}
	if cfg.Private == nil {
		cfg.Private = &ruthlespb.ConfigPrivate{}
	}
	if cfg.Private.Registration == nil {
		cfg.Private.Registration = &ruthlespb.ConfigPrivate_Registration{}
	}

	// Setup Storage
	var store storage.Storage
	if storageFlag == "memory" {
		store = memory.New()
	} else if storageFlag == "postgres" {
		var err error
		store, err = postgres.New(dbConnStr)
		if err != nil {
			log.Fatalf("Failed to initialize postgres storage: %v", err)
		}
	} else {
		log.Fatalf("Unsupported storage type: %s", storageFlag)
	}

	// Setup Auth
	var authenticator auth.Authenticator
	var authHandler *auth.Handler

	if authFlag == "fake" {
		authenticator = auth.NewFakeAuthenticator()
	} else if authFlag == "google" {
		ctx := cmd.Context()
		
		// Ruthless Token logic
		tokenGen := auth.NewTokenGenerator(authSecret)
		authenticator = auth.NewTokenAuthenticator(tokenGen, store)

		// Google OAuth logic
		clientId := googleAudience
		if clientId == "" {
			clientId = cfg.Public.Registration.GoogleClientId
		}
		clientSecret := getEnv("GOOGLE_CLIENT_SECRET", cfg.Private.Registration.GoogleClientSecret)
		redirectURL := getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback")

		exchanger, err := auth.NewExchanger(
			ctx,
			clientId,
			clientSecret,
			redirectURL,
		)
		if err != nil {
			log.Fatalf("Failed to initialize Google OAuth: %v", err)
		}

		uiURL := getEnv("UI_URL", getEnv("FRONTEND_URL", "http://localhost:3000"))
		authHandler = auth.NewHandler(store, exchanger, tokenGen, uiURL)
	} else {
		log.Fatalf("Unsupported auth type: %s", authFlag)
	}

	srv := server.New(store, authenticator, cfg)

	// Handle Seeding
	if seedFlag != "" {
		if storageFlag != "memory" {
			log.Fatalf("Seeding is only supported with 'memory' storage")
		}
		log.Printf("Seeding data from %s", seedFlag)
		if err := server.LoadSeed(cmd.Context(), store, seedFlag); err != nil {
			log.Fatalf("Failed to seed data: %v", err)
		}
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			srv.UnaryLoggingInterceptor(),
			srv.UnaryAuthInterceptor(),
		),
	)

	srv.RegisterWithGRPC(grpcServer)

	wrappedGrpc := grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
	)

	mux := http.NewServeMux()
	
	// Register OAuth handlers if enabled
	if authHandler != nil {
		authHandler.RegisterHandlers(mux)
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
			return
		}

		// Fallback to standard HTTP handlers (OAuth)
		mux.ServeHTTP(w, r)
	})

	handler := corsHandler.Handler(h2c.NewHandler(httpHandler, &http2.Server{}))

	log.Println("Listening on :8080 (gRPC, gRPC-Web and HTTP Auth)")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the gRPC API server",
	Run:   serverRun,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&storageFlag, "storage", "memory", "storage engine (memory|postgres)")
	serverCmd.Flags().StringVar(&authFlag, "auth", "fake", "auth mechanism (fake|jwt)")
	serverCmd.Flags().StringVar(&seedFlag, "seed", "", "path to a JSON seed file (only works with --storage=memory)")
	serverCmd.Flags().StringVar(&dbConnStr, "db-conn-str", "", "PostgreSQL connection string")
	serverCmd.Flags().StringVar(&authSecret, "auth-secret", "dev-secret", "JWT shared secret for auth (only for 'jwt' type)")
	serverCmd.Flags().StringVar(&googleAudience, "google-audience", "", "Google OAuth Client ID (required for 'google' type)")
	serverCmd.Flags().StringVar(&configPath, "config", "backend/config.textproto", "path to textproto config file")
}
