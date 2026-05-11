package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sunkencosts/mirror-me/internal/db"
	"github.com/sunkencosts/mirror-me/internal/googleauth"
	"github.com/sunkencosts/mirror-me/internal/sleeper"
	"github.com/sunkencosts/mirror-me/pkg/config"
)

func run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg := config.Load(getenv)

	if cfg.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET must be set")
	}

	dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("creating connection pool: %w", err)
	}
	defer dbpool.Close()

	store := db.NewStore(dbpool)
	if err := store.Ping(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}
	log.Println("database connected")
	sleeperClient := sleeper.New(cfg.SleeperBaseURL, store, cfg.CurrentWeek)
	googleClient := googleauth.New(googleauth.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		AuthURL:      cfg.GoogleAuthURL,
		TokenURL:     cfg.GoogleTokenURL,
		UserInfoURL:  cfg.GoogleUserInfoURL,
	})

	migrateURL := strings.Replace(cfg.DatabaseURL, "postgresql://", "pgx5://", 1)
	migrateURL = strings.Replace(migrateURL, "postgres://", "pgx5://", 1)
	m, err := migrate.New(cfg.MigrationsURL, migrateURL)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}
	srv := NewServer(sleeperClient, cfg, store, googleClient)

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	httpServer := &http.Server{Handler: srv}

	go func() {
		fmt.Fprintf(stdout, "listening on %s\n", listener.Addr())
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}

func NewServer(sleeperClient sleeperDeps, cfg config.Config, store *db.Store, googleClient *googleauth.Client) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, sleeperClient, store, cfg, googleClient)

	var handler http.Handler = mux
	handler = corsMiddleware(cfg.FrontendURL)(handler)
	return handler
}

func corsMiddleware(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Origin") != "" {
				w.Header().Set("Access-Control-Allow-Origin", frontendURL)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
