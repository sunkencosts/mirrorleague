package sleeper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

const rosterCacheTTL = 5 * time.Minute

type roster struct {
	RosterID int      `json:"roster_id"`
	OwnerID  string   `json:"owner_id"`
	Players  []string `json:"players"`
	Starters []string `json:"starters"`
	Reserve  []string `json:"reserve"`
	Taxi     []string `json:"taxi"`
}

type leagueUserMetadata struct {
	TeamName string `json:"team_name"`
}

type leagueUser struct {
	UserID   string             `json:"user_id"`
	Metadata leagueUserMetadata `json:"metadata"`
}

type rosterCacheEntry struct {
	rosters   []provider.Roster
	fetchedAt time.Time
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	players    playerLookup

	rosterMu    sync.RWMutex
	rosterCache map[string]rosterCacheEntry
}

type playerLookup interface {
	GetPlayersByIDs(ctx context.Context, ids []string) (map[string]provider.Player, error)
}

func New(baseURL string, players playerLookup) *Client {
	return &Client{
		baseURL:     baseURL,
		httpClient:  &http.Client{},
		players:     players,
		rosterCache: make(map[string]rosterCacheEntry),
	}
}

func (c *Client) resolvePlayers(playerMap map[string]provider.Player, ids []string) []provider.Player {
	players := []provider.Player{}
	for _, id := range ids {
		if player, ok := playerMap[id]; ok {
			player.ImageURL = fmt.Sprintf("https://sleepercdn.com/content/nfl/players/thumb/%s.jpg", player.PlayerID)
			players = append(players, player)
		}
	}
	return players
}

func (c *Client) getLeagueUsers(ctx context.Context, leagueID string) (map[string]leagueUser, error) {
	url := c.baseURL + "/league/" + leagueID + "/users"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating users request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
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
func (c *Client) GetLeague(ctx context.Context, leagueID string) (provider.League, error) {
	url := c.baseURL + "/league/" + leagueID
	var leagueSettings provider.League

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return provider.League{}, fmt.Errorf("creating league request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return provider.League{}, fmt.Errorf("getting league for leagueID %s: %w", leagueID, err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&leagueSettings); err != nil {
		return provider.League{}, fmt.Errorf("decoding league: %w", err)
	}
	return leagueSettings, nil
}

func (c *Client) GetRosters(ctx context.Context, leagueID string) ([]provider.Roster, error) {
	c.rosterMu.RLock()
	entry, ok := c.rosterCache[leagueID]
	c.rosterMu.RUnlock()
	if ok && time.Since(entry.fetchedAt) < rosterCacheTTL {
		return entry.rosters, nil
	}

	rosters, err := c.fetchRosters(ctx, leagueID)
	if err != nil {
		return nil, err
	}

	c.rosterMu.Lock()
	c.rosterCache[leagueID] = rosterCacheEntry{rosters: rosters, fetchedAt: time.Now()}
	c.rosterMu.Unlock()

	return rosters, nil
}

func (c *Client) fetchRosters(ctx context.Context, leagueID string) ([]provider.Roster, error) {
	var wg sync.WaitGroup
	var rawRosters []roster
	var usersByID map[string]leagueUser
	var rosterErr, userErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		url := c.baseURL + "/league/" + leagueID + "/rosters"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			rosterErr = fmt.Errorf("creating rosters request: %w", err)
			return
		}
		resp, err := c.httpClient.Do(req)
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
		usersByID, userErr = c.getLeagueUsers(ctx, leagueID)
	}()
	wg.Wait()

	if rosterErr != nil {
		return nil, rosterErr
	}
	if userErr != nil {
		return nil, userErr
	}

	seen := map[string]struct{}{}
	var allIDs []string
	for _, r := range rawRosters {
		for _, ids := range [][]string{r.Players, r.Starters, r.Reserve, r.Taxi} {
			for _, id := range ids {
				if _, ok := seen[id]; !ok {
					seen[id] = struct{}{}
					allIDs = append(allIDs, id)
				}
			}
		}
	}

	playerMap, err := c.players.GetPlayersByIDs(ctx, allIDs)
	if err != nil {
		return nil, fmt.Errorf("fetching players: %w", err)
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
			Players:  c.resolvePlayers(playerMap, r.Players),
			Starters: c.resolvePlayers(playerMap, r.Starters),
			Reserve:  c.resolvePlayers(playerMap, r.Reserve),
			Taxi:     c.resolvePlayers(playerMap, r.Taxi),
		})
	}

	return result, nil
}
