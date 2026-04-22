package sleeper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

func TestGetRosters_teamName(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/league/abc/rosters", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]roster{
			{RosterID: 1, OwnerID: "user1", Players: []string{"111"}, Starters: []string{"111"}},
		})
	})
	mux.HandleFunc("/league/abc/users", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]leagueUser{
			{UserID: "user1", Metadata: leagueUserMetadata{TeamName: "Mahomes Enjoyers"}},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cache := &PlayerCache{Players: map[string]provider.Player{"111": {PlayerID: "111"}}}
	c := New(srv.URL, cache)

	rosters, err := c.GetRosters("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rosters) != 1 {
		t.Fatalf("expected 1 roster, got %d", len(rosters))
	}
	if rosters[0].TeamName != "Mahomes Enjoyers" {
		t.Errorf("expected team name %q, got %q", "Mahomes Enjoyers", rosters[0].TeamName)
	}
}

func TestResolvePlayers(t *testing.T) {
	cache := &PlayerCache{
		Players: map[string]provider.Player{
			"111": {PlayerID: "111", FirstName: "Patrick", LastName: "Mahomes"},
			"222": {PlayerID: "222", FirstName: "Travis", LastName: "Kelce"},
		},
	}
	c := &Client{playerCache: cache}

	got := c.resolvePlayers([]string{"111", "222", "999"})

	if len(got) != 2 {
		t.Fatalf("expected 2 players, got %d", len(got))
	}
	if got[0].PlayerID != "111" {
		t.Errorf("expected player 111, got %s", got[0].PlayerID)
	}
}
