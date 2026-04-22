package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sunkencosts/mirror-me/internal/handlers"
	"github.com/sunkencosts/mirror-me/internal/sleeper"
	"github.com/sunkencosts/mirror-me/pkg/config"
)

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

func main() {
	cfg := config.Load()
	cache := &sleeper.PlayerCache{}

	if err := cache.Load(cfg.SleeperBaseURL); err != nil {
		log.Fatal(err)
	}
	sleeperClient := sleeper.New(cfg.SleeperBaseURL, cache)
	h := &handlers.RosterHandler{Provider: sleeperClient}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/league/{leagueId}/rosters", h.GetRosters)
	mux.HandleFunc("/", spaHandler("web/dist"))
	log.Printf("server running on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, corsMiddleware(mux)))
}
