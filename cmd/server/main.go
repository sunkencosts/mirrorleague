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
	sleeperClient := sleeper.New(cfg.SleeperBaseURL)
	h := &handlers.RosterHandler{Provider: sleeperClient}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /league/{leagueId}/rosters", h.GetRosters)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
