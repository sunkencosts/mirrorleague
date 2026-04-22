package sleeper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

const playerCacheFile = "players.json"
const cacheTTL = 24 * time.Hour

type PlayerCache struct {
	Players map[string]provider.Player
}

func (p *PlayerCache) Load(baseURL string) error {
	if info, err := os.Stat(playerCacheFile); err == nil {
		if time.Since(info.ModTime()) < cacheTTL {
			data, err := os.ReadFile(playerCacheFile)
			if err != nil {
				return fmt.Errorf("reading player cache: %w", err)
			}
			return json.Unmarshal(data, &p.Players)
		}
	}

	resp, err := http.Get(baseURL + "/players/nfl")
	if err != nil {
		return fmt.Errorf("fetching Players: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if err := os.WriteFile(playerCacheFile, data, 0644); err != nil {
		return fmt.Errorf("writing player cache: %w", err)
	}

	return json.Unmarshal(data, &p.Players)
}

func (p *PlayerCache) Get(playerID string) (provider.Player, bool) {

	player, ok := p.Players[playerID]
	return player, ok
}
