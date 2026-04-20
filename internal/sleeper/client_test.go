package sleeper

import (
	"testing"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

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
