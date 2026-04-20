package sleeper

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

type roster struct {
	RosterID int      `json:"roster_id"`
	OwnerID  string   `json:"owner_id"`
	Players  []string `json:"players"`
	Starters []string `json:"starters"`
}

type Client struct {
	baseURL     string
	httpClient  *http.Client
	playerCache *PlayerCache
}

func New(baseURL string, cache *PlayerCache) *Client {
	return &Client{
		baseURL:     baseURL,
		httpClient:  &http.Client{},
		playerCache: cache,
	}
}

func (c *Client) resolvePlayers(ids []string) []provider.Player {
	var players []provider.Player
	for _, id := range ids {
		if player, ok := c.playerCache.Get(id); ok {
			players = append(players, player)
		}
	}
	return players
}

func (c *Client) GetRosters(leagueID string) ([]provider.Roster, error) {
	url := c.baseURL + "/league/" + leagueID + "/rosters"

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getting rosters for league %s: %w", leagueID, err)
	}
	defer resp.Body.Close()

	var rosters []roster
	if err := json.NewDecoder(resp.Body).Decode(&rosters); err != nil {
		return nil, fmt.Errorf("decoding rosters: %w", err)
	}

	var result []provider.Roster
	for _, r := range rosters {
		result = append(result, provider.Roster{
			RosterID: r.RosterID,
			OwnerID:  r.OwnerID,
			Players:  c.resolvePlayers(r.Players),
			Starters: c.resolvePlayers(r.Starters),
		})
	}

	return result, nil
}
