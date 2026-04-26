package handlers

import (
	"net/http"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

type leagueProvider interface {
	GetLeague(leagueID string) (provider.League, error)
}

func HandleGetLeague(p leagueProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leagueID := r.PathValue("leagueId")
		league, err := p.GetLeague(leagueID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encode(w, r, http.StatusOK, league)
	})
}
