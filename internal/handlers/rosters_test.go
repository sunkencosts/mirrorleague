package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

type mockProvider struct {
	rosters []provider.Roster
	err     error
}

func (m *mockProvider) GetRosters(leagueID string) ([]provider.Roster, error) {
	return m.rosters, m.err
}

func TestGetRosters_success(t *testing.T) {
	rosters := []provider.Roster{{RosterID: 1, Players: []provider.Player{{PlayerID: "111"}}}}
	h := &RosterHandler{Provider: &mockProvider{rosters: rosters}}
	req := httptest.NewRequest("GET", "/league/test/rosters", nil)
	req.SetPathValue("leagueId", "test")
	w := httptest.NewRecorder()

	h.GetRosters(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Result().StatusCode)
	}

	var got []provider.Roster
	json.NewDecoder(w.Result().Body).Decode(&got)
	if len(got) != 1 || got[0].Players[0].PlayerID != "111" {
		t.Errorf("unexpected response: %+v", got)
	}
}

func TestGetRosters_providerError(t *testing.T) {
	h := &RosterHandler{Provider: &mockProvider{err: errors.New("down")}}
	req := httptest.NewRequest("GET", "/league/test/rosters", nil)
	req.SetPathValue("leagueId", "test")
	w := httptest.NewRecorder()

	h.GetRosters(w, req)

	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Result().StatusCode)
	}
}
