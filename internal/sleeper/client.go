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
	baseURL    string
	httpClient *http.Client
}

func New(baseUrl string) *Client {
	return &Client{
		baseURL:    baseUrl,
		httpClient: &http.Client{},
	}
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
	for _, roster := range rosters {
		result = append(result, provider.Roster{
			RosterID: roster.RosterID,
			OwnerID:  roster.OwnerID,
			Players:  roster.Players,
			Starters: roster.Starters,
		})

	}

	return result, nil

}
