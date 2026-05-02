package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sunkencosts/mirror-me/internal/provider"
)

type lineupStore interface {
	CreateLineup(ctx context.Context, userID, leagueID, source string, rosterID, weekNumber int, starters []string) (provider.Lineup, error)
	GetLineup(ctx context.Context, id string) (provider.Lineup, error)
	UpdateLineup(ctx context.Context, id string, starters []string) (provider.Lineup, error)
	ListLineups(ctx context.Context, userID, leagueID string, weekNumber int, rosterID *int) ([]provider.Lineup, error)
}

type lineupRosterProvider interface {
	GetRosters(ctx context.Context, leagueID string) ([]provider.Roster, error)
}

type createLineupRequest struct {
	UserID     string   `json:"user_id"`
	LeagueID   string   `json:"league_id"`
	Source     string   `json:"source"`
	RosterID   int      `json:"roster_id"`
	WeekNumber int      `json:"week_number"`
	Starters   []string `json:"starters"`
}

type updateLineupRequest struct {
	UserID   string   `json:"user_id"`
	Starters []string `json:"starters"`
}

func HandleCreateLineup(store lineupStore, p lineupRosterProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[createLineupRequest](r)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Source == "" {
			http.Error(w, "missing source", http.StatusBadRequest)
			return
		}

		if err := validateStarters(r.Context(), p, req.LeagueID, req.RosterID, req.Starters); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		lineup, err := store.CreateLineup(r.Context(), req.UserID, req.LeagueID, req.Source, req.RosterID, req.WeekNumber, req.Starters)
		if err != nil {
			http.Error(w, "failed to create lineup", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", "/api/lineups/"+lineup.ID)
		encode(w, r, http.StatusCreated, lineup)
	})
}

func HandleUpdateLineup(store lineupStore, p lineupRosterProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := uuid.Parse(id); err != nil {
			http.Error(w, "invalid lineup id", http.StatusBadRequest)
			return
		}

		req, err := decode[updateLineupRequest](r)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		existing, err := store.GetLineup(r.Context(), id)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "lineup not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "failed to get lineup", http.StatusInternalServerError)
			return
		}
		if existing.UserID != req.UserID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if err := validateStarters(r.Context(), p, existing.LeagueID, existing.RosterID, req.Starters); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		lineup, err := store.UpdateLineup(r.Context(), id, req.Starters)
		if err != nil {
			http.Error(w, "failed to update lineup", http.StatusInternalServerError)
			return
		}
		encode(w, r, http.StatusOK, lineup)
	})
}

func HandleListLineups(store lineupStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if _, err := uuid.Parse(userID); err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		leagueID := r.URL.Query().Get("league_id")
		if leagueID == "" {
			http.Error(w, "missing league_id", http.StatusBadRequest)
			return
		}
		weekNumber, err := strconv.Atoi(r.URL.Query().Get("week_number"))
		if err != nil {
			http.Error(w, "invalid week_number", http.StatusBadRequest)
			return
		}

		var rosterID *int
		if raw := r.URL.Query().Get("roster_id"); raw != "" {
			id, err := strconv.Atoi(raw)
			if err != nil {
				http.Error(w, "invalid roster_id", http.StatusBadRequest)
				return
			}
			rosterID = &id
		}

		lineups, err := store.ListLineups(r.Context(), userID, leagueID, weekNumber, rosterID)
		if err != nil {
			http.Error(w, "failed to list lineups", http.StatusInternalServerError)
			return
		}
		encode(w, r, http.StatusOK, lineups)
	})
}
func HandleGetLineupByID(store lineupStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := uuid.Parse(id); err != nil {
			http.Error(w, "invalid lineup id", http.StatusBadRequest)
			return
		}
		lineup, err := store.GetLineup(r.Context(), id)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "lineup not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "failed to get lineup", http.StatusInternalServerError)
			return
		}
		encode(w, r, http.StatusOK, lineup)
	})
}
func validateStarters(ctx context.Context, p lineupRosterProvider, leagueID string, rosterID int, starters []string) error {
	rosters, err := p.GetRosters(ctx, leagueID)
	if err != nil {
		return fmt.Errorf("fetching roster: %w", err)
	}

	var roster *provider.Roster
	for i := range rosters {
		if rosters[i].RosterID == rosterID {
			roster = &rosters[i]
			break
		}
	}
	if roster == nil {
		return fmt.Errorf("roster %d not found in league", rosterID)
	}

	playerSet := make(map[string]struct{}, len(roster.Players))
	for _, p := range roster.Players {
		playerSet[p.PlayerID] = struct{}{}
	}
	for _, id := range starters {
		if _, ok := playerSet[id]; !ok {
			return fmt.Errorf("player %s is not on this roster", id)
		}
	}
	return nil
}
