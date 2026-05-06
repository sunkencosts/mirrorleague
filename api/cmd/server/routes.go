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
	"github.com/sunkencosts/mirror-me/pkg/config"
)

type sleeperDeps interface {
	provider.Provider
	InvalidateRosters()
}

// Update api/api.md when adding or removing routes here.
func addRoutes(mux *http.ServeMux, sleeperClient sleeperDeps, store *db.Store, cfg config.Config) {
	mux.Handle("POST /api/league-bookmarks", handlers.HandleSaveUserLeague(store))
	mux.Handle("GET /api/league-bookmarks", handlers.HandleListUserLeagues(store))
	mux.Handle("PATCH /api/league-bookmarks/{leagueId}", handlers.HandleUpdateUserLeague(store))
	mux.Handle("DELETE /api/league-bookmarks/{leagueId}", handlers.HandleDeleteUserLeague(store))
	mux.Handle("POST /api/lineups", handlers.HandleCreateLineup(store, sleeperClient))
	mux.Handle("PATCH /api/lineups/{id}", handlers.HandleUpdateLineup(store, sleeperClient))
	mux.Handle("GET /api/lineups", handlers.HandleListLineups(store))
	mux.Handle("GET /api/lineups/{id}", handlers.HandleGetLineupByID(store))
	mux.Handle("GET /api/league/{leagueId}/rosters", handlers.HandleGetRosters(sleeperClient))
	mux.Handle("GET /api/league/{leagueId}", handlers.HandleGetLeague(sleeperClient))
	mux.Handle("POST /api/admin/sync-players", handlers.HandleSyncPlayers(store, sleeperClient, cfg.SleeperBaseURL, cfg.RankingsCSVURL))
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
