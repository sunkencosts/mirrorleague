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
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sunkencosts/mirror-me/internal/db"
	"github.com/sunkencosts/mirror-me/internal/provider"
	"github.com/sunkencosts/mirror-me/internal/rankings"
	"github.com/sunkencosts/mirror-me/internal/sleeper"
	"github.com/sunkencosts/mirror-me/pkg/config"
)

func run(ctx context.Context, getenv func(string) string, stdout, stderr io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg := config.Load(getenv)
	playerCache := &sleeper.PlayerCache{}
	rankingsCache := &rankings.Cache{}
	sleeperClient := sleeper.New(cfg.SleeperBaseURL, playerCache, rankingsCache)

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

	srv := NewServer(sleeperClient, playerCache, rankingsCache, cfg, store)

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

func NewServer(sleeperClient provider.Provider, cache *sleeper.PlayerCache, rankingsCache *rankings.Cache, cfg config.Config, store *db.Store) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, sleeperClient, store)

	var (
		once                   sync.Once
		playerErr, rankingsErr error
	)

	core := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() {
			var wg sync.WaitGroup
			wg.Add(2)
			go func() { defer wg.Done(); playerErr = cache.Load(cfg.SleeperBaseURL) }()
			go func() { defer wg.Done(); rankingsErr = rankingsCache.Load(cfg.RankingsCSVURL) }()
			wg.Wait()
		})
		if playerErr != nil || rankingsErr != nil {
			http.Error(w, "failed to load data", http.StatusServiceUnavailable)
			return
		}
		mux.ServeHTTP(w, r)
	})

	var handler http.Handler = core
	handler = corsMiddleware(handler)
	return handler
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
