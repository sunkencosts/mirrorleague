package sleeper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

type roster struct {
	RosterID int      `json:"roster_id"`
	OwnerID  string   `json:"owner_id"`
	Players  []string `json:"players"`
	Starters []string `json:"starters"`
}

type leagueUserMetadata struct {
	TeamName string `json:"team_name"`
}

type leagueUser struct {
	UserID   string             `json:"user_id"`
	Metadata leagueUserMetadata `json:"metadata"`
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
			player.ImageURL = fmt.Sprintf("https://sleepercdn.com/content/nfl/players/thumb/%s.jpg", player.PlayerID)
			players = append(players, player)
		}
	}
	return players
}

func (c *Client) getLeagueUsers(leagueID string) (map[string]leagueUser, error) {
	url := c.baseURL + "/league/" + leagueID + "/users"
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getting users for league %s: %w", leagueID, err)
	}
	defer resp.Body.Close()

	var users []leagueUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("decoding league users: %w", err)
	}

	byID := make(map[string]leagueUser, len(users))
	for _, u := range users {
		byID[u.UserID] = u
	}
	return byID, nil
}
func (c *Client) GetLeague(leagueID string) (provider.League, error) {
	url := c.baseURL + "/league/" + leagueID
	var leagueSettings provider.League

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return provider.League{}, fmt.Errorf("getting league for leagueID %s", leagueID)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&leagueSettings); err != nil {
		return provider.League{}, fmt.Errorf("getting decoding league: %w", err)
	}
	return leagueSettings, nil
}

func (c *Client) GetRosters(leagueID string) ([]provider.Roster, error) {
	var wg sync.WaitGroup
	var rawRosters []roster
	var usersByID map[string]leagueUser
	var rosterErr, userErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		url := c.baseURL + "/league/" + leagueID + "/rosters"
		resp, err := c.httpClient.Get(url)
		if err != nil {
			rosterErr = fmt.Errorf("getting rosters for league %s: %w", leagueID, err)
			return
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&rawRosters); err != nil {
			rosterErr = fmt.Errorf("decoding rosters: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		usersByID, userErr = c.getLeagueUsers(leagueID)
	}()
	wg.Wait()

	if rosterErr != nil {
		return nil, rosterErr
	}
	if userErr != nil {
		return nil, userErr
	}

	var result []provider.Roster
	for _, r := range rawRosters {
		teamName := ""
		if u, ok := usersByID[r.OwnerID]; ok {
			teamName = u.Metadata.TeamName
		}
		result = append(result, provider.Roster{
			RosterID: r.RosterID,
			OwnerID:  r.OwnerID,
			TeamName: teamName,
			Players:  c.resolvePlayers(r.Players),
			Starters: c.resolvePlayers(r.Starters),
		})
	}

	return result, nil
}
