package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

type RosterHandler struct {
	Provider provider.Provider
}

func (h *RosterHandler) GetRosters(w http.ResponseWriter, r *http.Request) {
	leagueID := r.PathValue("leagueId")
	rosters, err := h.Provider.GetRosters(leagueID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rosters)
}
