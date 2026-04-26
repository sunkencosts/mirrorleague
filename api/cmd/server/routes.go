package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sunkencosts/mirror-me/internal/db"
	"github.com/sunkencosts/mirror-me/internal/handlers"
	"github.com/sunkencosts/mirror-me/internal/provider"
)

func addRoutes(mux *http.ServeMux, sleeperClient provider.Provider, store *db.Store) {
	mux.Handle("GET /api/league/{leagueId}/rosters", handlers.HandleGetRosters(sleeperClient))
	mux.Handle("GET /api/league/{leagueId}", handlers.HandleGetLeague(sleeperClient))
	mux.HandleFunc("GET /healthz", handleHealthz(store))
	mux.Handle("/", spaHandler("web/dist"))
}

func handleHealthz(store *db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := store.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func spaHandler(distDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(distDir, filepath.Clean("/"+r.URL.Path))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
			return
		}
		http.FileServer(http.Dir(distDir)).ServeHTTP(w, r)
	}
}
