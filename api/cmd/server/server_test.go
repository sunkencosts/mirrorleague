package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sunkencosts/mirror-me/internal/db"
	"github.com/sunkencosts/mirror-me/internal/provider"
)

const testDatabaseURL = "postgres://mirrorme:mirrorme@localhost:5433/mirrorme_test"

// testPlayers is the fixed reference dataset seeded once before all tests run.
// Use these player IDs in any Sleeper mock that needs player resolution.
var testPlayers = []provider.Player{
	{PlayerID: "111", FirstName: "Josh", LastName: "Allen", FantasyPositions: []string{"QB"}},
	{PlayerID: "222", FirstName: "Justin", LastName: "Jefferson", FantasyPositions: []string{"WR"}},
	{PlayerID: "333", FirstName: "Christian", LastName: "McCaffrey", FantasyPositions: []string{"RB"}},
	{PlayerID: "444", FirstName: "Travis", LastName: "Kelce", FantasyPositions: []string{"TE"}},
	{PlayerID: "555", FirstName: "Tyreek", LastName: "Hill", FantasyPositions: []string{"WR"}},
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDatabaseURL)
	if err != nil {
		log.Fatalf("TestMain: connect test db: %v", err)
	}
	if err := db.NewStore(pool).UpsertPlayers(ctx, testPlayers); err != nil {
		log.Fatalf("TestMain: seed players: %v", err)
	}
	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func newTestServer(t *testing.T, sleeperHandler http.Handler, extraEnv ...map[string]string) string {
	t.Helper()

	fakeSleeper := httptest.NewServer(sleeperHandler)
	t.Cleanup(fakeSleeper.Close)

	port := freePort(t)
	getenv := func(key string) string {
		for _, env := range extraEnv {
			if v, ok := env[key]; ok {
				return v
			}
		}
		switch key {
		case "PORT":
			return port
		case "SLEEPER_BASE_URL":
			return fakeSleeper.URL
		case "DATABASE_URL":
			return testDatabaseURL
		case "MIGRATIONS_URL":
			return "file://../../migrations"
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
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
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

func TestSyncPlayers(t *testing.T) {
	// Fake rankings CSV: 23 columns, two players with known rarities.
	// Column indices: 1=page_type, 5=pos, 22=merge_name.
	fakeRankings := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "c0,page_type,c2,c3,c4,pos,c6,c7,c8,c9,c10,c11,c12,c13,c14,c15,c16,c17,c18,c19,c20,c21,merge_name\n")
		fmt.Fprint(w, "x,dynasty-qb,x,x,x,QB,x,x,x,x,x,x,x,x,x,x,x,x,x,x,x,x,Josh Allen\n")
		fmt.Fprint(w, "x,dynasty-rb,x,x,x,RB,x,x,x,x,x,x,x,x,x,x,x,x,x,x,x,x,Christian McCaffrey\n")
	}))
	t.Cleanup(fakeRankings.Close)

	baseURL := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/players/nfl" {
			json.NewEncoder(w).Encode(map[string]provider.Player{
				"111": {PlayerID: "111", FirstName: "Josh", LastName: "Allen", FantasyPositions: []string{"QB"}},
				"333": {PlayerID: "333", FirstName: "Christian", LastName: "McCaffrey", FantasyPositions: []string{"RB"}},
			})
		}
	}), map[string]string{"RANKINGS_CSV_URL": fakeRankings.URL})

	resp, err := http.Post(baseURL+"/api/admin/sync-players", "", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["upserted"] != 2 {
		t.Errorf("expected 2 upserted, got %d", result["upserted"])
	}
}
