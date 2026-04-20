package main

import (
	"log"
	"net/http"

	"github.com/sunkencosts/mirror-me/internal/handlers"
	"github.com/sunkencosts/mirror-me/internal/sleeper"
	"github.com/sunkencosts/mirror-me/pkg/config"
)

func main() {
	cfg := config.Load()
	cache := &sleeper.PlayerCache{}

	if err := cache.Load(cfg.SleeperBaseURL); err != nil {
		log.Fatal(err)
	}
	sleeperClient := sleeper.New(cfg.SleeperBaseURL, cache)
	h := &handlers.RosterHandler{Provider: sleeperClient}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /league/{leagueId}/rosters", h.GetRosters)
	log.Printf("server running on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
