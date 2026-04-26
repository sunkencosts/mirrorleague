package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/sunkencosts/mirror-me/internal/provider"
)

func newTestServer(t *testing.T, sleeperHandler http.Handler) string {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/players/nfl", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]provider.Player{
			"111": {PlayerID: "111", FirstName: "Josh", LastName: "Allen"},
		})
	})
	mux.Handle("/", sleeperHandler)
	fakeSleeper := httptest.NewServer(mux)
	t.Cleanup(fakeSleeper.Close)

	port := freePort(t)
	getenv := func(key string) string {
		switch key {
		case "PORT":
			return port
		case "SLEEPER_BASE_URL":
			return fakeSleeper.URL
		case "DATABASE_URL":
			return "postgres://mirrorme:mirrorme@localhost:5433/mirrorme"
		}
		return ""
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go run(ctx, getenv, io.Discard, io.Discard)

	baseURL := "http://localhost:" + port
	if err := waitForReady(ctx, 5*time.Second, baseURL+"/healthz"); err != nil {
		t.Fatalf("server never became ready: %v", err)
	}
	return baseURL
}

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

func waitForReady(ctx context.Context, timeout time.Duration, endpoint string) error {
	client := http.Client{}
	startTime := time.Now()
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}

func TestGetRosters(t *testing.T) {
	baseURL := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/league/abc/rosters":
			json.NewEncoder(w).Encode([]map[string]any{
				{"roster_id": 1, "owner_id": "u1", "players": []string{"111"}, "starters": []string{"111"}},
			})
		case "/league/abc/users":
			json.NewEncoder(w).Encode([]map[string]any{
				{"user_id": "u1", "metadata": map[string]string{"team_name": "Test Team"}},
			})
		}
	}))

	resp, err := http.Get(baseURL + "/api/league/abc/rosters")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var rosters []provider.Roster
	if err := json.NewDecoder(resp.Body).Decode(&rosters); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(rosters) != 1 {
		t.Fatalf("expected 1 roster, got %d", len(rosters))
	}
	if rosters[0].TeamName != "Test Team" {
		t.Errorf("expected team name %q, got %q", "Test Team", rosters[0].TeamName)
	}
	if len(rosters[0].Players) != 1 || rosters[0].Players[0].PlayerID != "111" {
		t.Errorf("unexpected players: %+v", rosters[0].Players)
	}
}

func TestGetLeague(t *testing.T) {
	baseURL := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/league/abc" {
			json.NewEncoder(w).Encode(map[string]any{
				"name":   "Test League",
				"season": "2025",
			})
		}
	}))

	resp, err := http.Get(baseURL + "/api/league/abc")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var league provider.League
	if err := json.NewDecoder(resp.Body).Decode(&league); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if league.Name != "Test League" || league.Season != "2025" {
		t.Errorf("unexpected league: %+v", league)
	}
}

func TestHealthz(t *testing.T) {
	baseURL := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	resp, err := http.Get(baseURL + "/healthz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
