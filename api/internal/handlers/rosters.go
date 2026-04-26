package handlers

import (
	"net/http"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

type rosterProvider interface {
	GetRosters(leagueID string) ([]provider.Roster, error)
}

func HandleGetRosters(p rosterProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leagueID := r.PathValue("leagueId")
		rosters, err := p.GetRosters(leagueID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encode(w, r, http.StatusOK, rosters)
	})
}
