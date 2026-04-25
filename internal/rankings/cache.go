package rankings

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const cacheTTL = 24 * time.Hour

const (
	colPageType  = 1
	colPos       = 5
	colMergeName = 22
)

var dynastyPageTypes = map[string]bool{
	"dynasty-qb": true, "dynasty-rb": true,
	"dynasty-wr": true, "dynasty-te": true,
}

type Cache struct {
	players   map[string]string
	cacheFile string // settable in tests; defaults to "rankings.json"
}

func (c *Cache) file() string {
	if c.cacheFile != "" {
		return c.cacheFile
	}
	return "rankings.json"
}

func cacheKey(name, pos string) string {
	return normalize(name) + "|" + pos
}

func normalize(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, ".", "")
	for _, suffix := range []string{" iv", " iii", " ii", " jr", " sr"} {
		name = strings.TrimSuffix(name, suffix)
	}
	return strings.TrimSpace(name)
}

func (c *Cache) Load(csvURL string) error {
	if info, err := os.Stat(c.file()); err == nil {
		if time.Since(info.ModTime()) < cacheTTL {
			data, err := os.ReadFile(c.file())
			if err != nil {
				return fmt.Errorf("reading rankings cache: %w", err)
			}
			return json.Unmarshal(data, &c.players)
		}
	}

	resp, err := http.Get(csvURL)
	if err != nil {
		return fmt.Errorf("fetching rankings CSV: %w", err)
	}
	defer resp.Body.Close()

	byPos := map[string][]string{}
	r := csv.NewReader(resp.Body)
	if _, err := r.Read(); err != nil {
		return fmt.Errorf("reading rankings CSV header: %w", err)
	}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("parsing rankings CSV: %w", err)
		}
		if !dynastyPageTypes[row[colPageType]] {
			continue
		}
		pos := row[colPos]
		byPos[pos] = append(byPos[pos], row[colMergeName])
	}

	c.players = make(map[string]string)
	for pos, names := range byPos {
		total := len(names)
		for rank, name := range names {
			pct := float64(rank+1) / float64(total)
			c.players[cacheKey(name, pos)] = rarityFromPct(pct)
		}
	}

	data, err := json.Marshal(c.players)
	if err != nil {
		return fmt.Errorf("marshaling rankings: %w", err)
	}
	if err := os.WriteFile(c.file(), data, 0644); err != nil {
		return fmt.Errorf("writing rankings cache: %w", err)
	}
	return nil
}

func rarityFromPct(pct float64) string {
	switch {
	case pct <= 0.05:
		return "orange"
	case pct <= 0.20:
		return "purple"
	case pct <= 0.45:
		return "blue"
	case pct <= 0.75:
		return "green"
	default:
		return "grey"
	}
}

func (c *Cache) Get(fullName, pos string) (string, bool) {
	rarity, ok := c.players[cacheKey(fullName, pos)]
	return rarity, ok
}
