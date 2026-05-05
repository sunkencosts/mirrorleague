package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/sunkencosts/mirror-me/internal/provider"
)

type userLeagueStore interface {
	SaveUserLeague(ctx context.Context, userID, leagueID, label string) (provider.UserLeague, error)
	ListUserLeagues(ctx context.Context, userID string) ([]provider.UserLeague, error)
	UpdateUserLeague(ctx context.Context, userID, leagueID, label string) (provider.UserLeague, error)
	DeleteUserLeague(ctx context.Context, userID, leagueID string) error
}

type saveUserLeagueRequest struct {
	UserID   string `json:"user_id"`
	LeagueID string `json:"league_id"`
	Label    string `json:"label"`
}

type updateUserLeagueRequest struct {
	UserID string `json:"user_id"`
	Label  string `json:"label"`
}

func HandleSaveUserLeague(store userLeagueStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[saveUserLeagueRequest](r)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.UserID == "" || req.LeagueID == "" {
			http.Error(w, "missing user_id or league_id", http.StatusBadRequest)
			return
		}
		ul, err := store.SaveUserLeague(r.Context(), req.UserID, req.LeagueID, req.Label)
		if err != nil {
			http.Error(w, "failed to save bookmark", http.StatusInternalServerError)
			return
		}
		encode(w, r, http.StatusOK, ul)
	})
}

func HandleListUserLeagues(store userLeagueStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "missing user_id", http.StatusBadRequest)
			return
		}
		leagues, err := store.ListUserLeagues(r.Context(), userID)
		if err != nil {
			http.Error(w, "failed to list bookmarks", http.StatusInternalServerError)
			return
		}
		encode(w, r, http.StatusOK, leagues)
	})
}

func HandleUpdateUserLeague(store userLeagueStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leagueID := r.PathValue("leagueId")
		req, err := decode[updateUserLeagueRequest](r)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.UserID == "" {
			http.Error(w, "missing user_id", http.StatusBadRequest)
			return
		}
		ul, err := store.UpdateUserLeague(r.Context(), req.UserID, leagueID, req.Label)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "bookmark not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to update bookmark", http.StatusInternalServerError)
			return
		}
		encode(w, r, http.StatusOK, ul)
	})
}

func HandleDeleteUserLeague(store userLeagueStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leagueID := r.PathValue("leagueId")
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "missing user_id", http.StatusBadRequest)
			return
		}
		if err := store.DeleteUserLeague(r.Context(), userID, leagueID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "bookmark not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to delete bookmark", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
