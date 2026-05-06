package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	migrateURL := "pgx5://mirrorme:mirrorme@localhost:5433/mirrorme_test"
	mg, err := migrate.New("file://../../migrations", migrateURL)
	if err != nil {
		log.Fatalf("TestMain: create migrator: %v", err)
	}
	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("TestMain: migrate: %v", err)
	}

	pool, err := pgxpool.New(ctx, testDatabaseURL)
	if err != nil {
		log.Fatalf("TestMain: connect test db: %v", err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE lineups, players, league_bookmarks RESTART IDENTITY CASCADE"); err != nil {
		log.Fatalf("TestMain: truncate: %v", err)
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

	pool, err := pgxpool.New(context.Background(), testDatabaseURL)
	if err != nil {
		t.Fatalf("newTestServer: connect db: %v", err)
	}
	if _, err := pool.Exec(context.Background(), "TRUNCATE lineups, league_bookmarks"); err != nil {
		t.Fatalf("newTestServer: truncate lineups: %v", err)
	}
	pool.Close()

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

func lineupSleeperHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/league/test-league/rosters":
			json.NewEncoder(w).Encode([]map[string]any{
				{"roster_id": 1, "owner_id": "owner1", "players": []string{"111", "222", "333"}, "starters": []string{"111"}},
			})
		case "/league/test-league/users":
			json.NewEncoder(w).Encode([]map[string]any{
				{"user_id": "owner1", "metadata": map[string]string{"team_name": "Test Team"}},
			})
		}
	})
}

func createTestLineup(t *testing.T, baseURL string) provider.Lineup {
	t.Helper()
	body := `{"user_id":"00000000-0000-0000-0000-000000000001","source":"sleeper","league_id":"test-league","roster_id":1,"week_number":1,"starters":["111","222"]}`
	resp, err := http.Post(baseURL+"/api/lineups", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("createTestLineup: request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createTestLineup: expected 201, got %d", resp.StatusCode)
	}
	var lineup provider.Lineup
	if err := json.NewDecoder(resp.Body).Decode(&lineup); err != nil {
		t.Fatalf("createTestLineup: failed to decode: %v", err)
	}
	return lineup
}

func TestCreateLineup(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())

	body := `{"user_id":"00000000-0000-0000-0000-000000000001","source":"sleeper","league_id":"test-league","roster_id":1,"week_number":1,"starters":["111","222"]}`
	resp, err := http.Post(baseURL+"/api/lineups", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var lineup provider.Lineup
	if err := json.NewDecoder(resp.Body).Decode(&lineup); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if lineup.ID == "" {
		t.Error("expected a non-empty lineup ID")
	}
	if loc := resp.Header.Get("Location"); loc != "/api/lineups/"+lineup.ID {
		t.Errorf("expected Location header %q, got %q", "/api/lineups/"+lineup.ID, loc)
	}
	if lineup.UserID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("expected user_id %q, got %q", "00000000-0000-0000-0000-000000000001", lineup.UserID)
	}
	if lineup.LeagueID != "test-league" {
		t.Errorf("expected league_id %q, got %q", "test-league", lineup.LeagueID)
	}
	if lineup.RosterID != 1 {
		t.Errorf("expected roster_id 1, got %d", lineup.RosterID)
	}
	if lineup.WeekNumber != 1 {
		t.Errorf("expected week_number 1, got %d", lineup.WeekNumber)
	}
	if len(lineup.Starters) != 2 || lineup.Starters[0] != "111" || lineup.Starters[1] != "222" {
		t.Errorf("unexpected starters: %v", lineup.Starters)
	}
}
func TestCreateLineup_InvalidPlayer(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())

	body := `{"user_id":"00000000-0000-0000-0000-000000000001","source":"sleeper","league_id":"test-league","roster_id":1,"week_number":1,"starters":["999"]}`
	resp, err := http.Post(baseURL+"/api/lineups", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}
func TestGetLineup(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())
	created := createTestLineup(t, baseURL)

	resp, err := http.Get(baseURL + "/api/lineups/" + created.ID)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var lineup provider.Lineup
	if err := json.NewDecoder(resp.Body).Decode(&lineup); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if lineup.ID != created.ID {
		t.Errorf("expected id %q, got %q", created.ID, lineup.ID)
	}
}

func TestGetLineup_NotFound(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())

	resp, err := http.Get(baseURL + "/api/lineups/00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
func TestUpdateLineup(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())
	created := createTestLineup(t, baseURL)

	body := `{"user_id":"00000000-0000-0000-0000-000000000001","starters":["111","333"]}`
	req, _ := http.NewRequest(http.MethodPatch, baseURL+"/api/lineups/"+created.ID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var lineup provider.Lineup
	if err := json.NewDecoder(resp.Body).Decode(&lineup); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(lineup.Starters) != 2 || lineup.Starters[0] != "111" || lineup.Starters[1] != "333" {
		t.Errorf("unexpected starters: %v", lineup.Starters)
	}
}

func TestUpdateLineup_WrongUser(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())
	created := createTestLineup(t, baseURL)

	body := `{"user_id":"00000000-0000-0000-0000-000000000002","starters":["111"]}`
	req, _ := http.NewRequest(http.MethodPatch, baseURL+"/api/lineups/"+created.ID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestUpdateLineup_NotFound(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())

	body := `{"user_id":"00000000-0000-0000-0000-000000000001","starters":["111"]}`
	req, _ := http.NewRequest(http.MethodPatch, baseURL+"/api/lineups/00000000-0000-0000-0000-000000000000", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestListLineups_FilterByRoster(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())
	created := createTestLineup(t, baseURL)

	url := baseURL + "/api/lineups?user_id=" + created.UserID + "&league_id=test-league&week_number=1&roster_id=1"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var lineups []provider.Lineup
	if err := json.NewDecoder(resp.Body).Decode(&lineups); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(lineups) != 1 {
		t.Fatalf("expected 1 lineup, got %d", len(lineups))
	}
	if lineups[0].ID != created.ID {
		t.Errorf("expected id %q, got %q", created.ID, lineups[0].ID)
	}
}

func TestListLineups_All(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())
	created := createTestLineup(t, baseURL)

	url := baseURL + "/api/lineups?user_id=" + created.UserID + "&league_id=test-league&week_number=1"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var lineups []provider.Lineup
	if err := json.NewDecoder(resp.Body).Decode(&lineups); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(lineups) != 1 {
		t.Fatalf("expected 1 lineup, got %d", len(lineups))
	}
	if lineups[0].ID != created.ID {
		t.Errorf("expected id %q, got %q", created.ID, lineups[0].ID)
	}
}

func TestListLineups_Empty(t *testing.T) {
	baseURL := newTestServer(t, lineupSleeperHandler())

	url := baseURL + "/api/lineups?user_id=00000000-0000-0000-0000-000000000001&league_id=test-league&week_number=99"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var lineups []provider.Lineup
	if err := json.NewDecoder(resp.Body).Decode(&lineups); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(lineups) != 0 {
		t.Errorf("expected empty array, got %d lineups", len(lineups))
	}
}

func TestGetWeekMatchups(t *testing.T) {
	baseURL := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/league/abc/matchups/8":
			json.NewEncoder(w).Encode([]map[string]any{
				{"roster_id": 1, "matchup_id": 1, "players": []string{"111", "222"}, "starters": []string{"111"}, "points": 95.5, "custom_points": nil},
				{"roster_id": 2, "matchup_id": 1, "players": []string{"333"}, "starters": []string{"333"}, "points": 88.0, "custom_points": nil},
			})
		case "/league/abc/rosters":
			json.NewEncoder(w).Encode([]map[string]any{
				{"roster_id": 1, "owner_id": "u1", "players": []string{"111", "222"}, "starters": []string{"111"}},
				{"roster_id": 2, "owner_id": "u2", "players": []string{"333"}, "starters": []string{"333"}},
			})
		case "/league/abc/users":
			json.NewEncoder(w).Encode([]map[string]any{
				{"user_id": "u1", "metadata": map[string]string{"team_name": "Team One"}},
				{"user_id": "u2", "metadata": map[string]string{"team_name": "Team Two"}},
			})
		}
	}))

	resp, err := http.Get(baseURL + "/api/league/abc/week/8")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var matchups []provider.WeekMatchup
	if err := json.NewDecoder(resp.Body).Decode(&matchups); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(matchups) != 2 {
		t.Fatalf("expected 2 matchups, got %d", len(matchups))
	}
	if matchups[0].TeamName != "Team One" {
		t.Errorf("expected team name %q, got %q", "Team One", matchups[0].TeamName)
	}
	if len(matchups[0].Players) != 2 {
		t.Errorf("expected 2 players on roster 1, got %d", len(matchups[0].Players))
	}
	if matchups[0].Players[0].PlayerID != "111" {
		t.Errorf("expected player 111, got %s", matchups[0].Players[0].PlayerID)
	}
	if matchups[0].Points != 95.5 {
		t.Errorf("expected points 95.5, got %f", matchups[0].Points)
	}
	if matchups[0].CustomPoints != nil {
		t.Errorf("expected custom_points nil, got %v", matchups[0].CustomPoints)
	}
}

func TestGetWeekMatchups_InvalidWeek(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())

	resp, err := http.Get(baseURL + "/api/league/abc/week/notanumber")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetWeekMatchups_ZeroWeek(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())

	resp, err := http.Get(baseURL + "/api/league/abc/week/0")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func noopHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func saveTestUserLeague(t *testing.T, baseURL, userID, leagueID, label string) provider.UserLeague {
	t.Helper()
	body := fmt.Sprintf(`{"user_id":%q,"league_id":%q,"label":%q}`, userID, leagueID, label)
	resp, err := http.Post(baseURL+"/api/league-bookmarks", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("saveTestUserLeague: request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("saveTestUserLeague: expected 200, got %d", resp.StatusCode)
	}
	var ul provider.UserLeague
	if err := json.NewDecoder(resp.Body).Decode(&ul); err != nil {
		t.Fatalf("saveTestUserLeague: failed to decode: %v", err)
	}
	return ul
}

func TestSaveUserLeague(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())

	ul := saveTestUserLeague(t, baseURL, "user-a", "league-1", "My League")

	if ul.UserID != "user-a" {
		t.Errorf("expected user_id %q, got %q", "user-a", ul.UserID)
	}
	if ul.LeagueID != "league-1" {
		t.Errorf("expected league_id %q, got %q", "league-1", ul.LeagueID)
	}
	if ul.Label != "My League" {
		t.Errorf("expected label %q, got %q", "My League", ul.Label)
	}
}

func TestListUserLeagues(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())
	saveTestUserLeague(t, baseURL, "user-a", "league-1", "My League")

	resp, err := http.Get(baseURL + "/api/league-bookmarks?user_id=user-a")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var leagues []provider.UserLeague
	if err := json.NewDecoder(resp.Body).Decode(&leagues); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(leagues) != 1 {
		t.Fatalf("expected 1 league, got %d", len(leagues))
	}
	if leagues[0].LeagueID != "league-1" || leagues[0].Label != "My League" {
		t.Errorf("unexpected league: %+v", leagues[0])
	}
}

func TestListUserLeagues_Empty(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())

	resp, err := http.Get(baseURL + "/api/league-bookmarks?user_id=user-nobody")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var leagues []provider.UserLeague
	if err := json.NewDecoder(resp.Body).Decode(&leagues); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(leagues) != 0 {
		t.Errorf("expected empty array, got %d leagues", len(leagues))
	}
}

func TestListUserLeagues_Isolation(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())
	saveTestUserLeague(t, baseURL, "user-a", "league-1", "User A League")

	resp, err := http.Get(baseURL + "/api/league-bookmarks?user_id=user-b")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var leagues []provider.UserLeague
	if err := json.NewDecoder(resp.Body).Decode(&leagues); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(leagues) != 0 {
		t.Errorf("expected user-b to see no leagues, got %d", len(leagues))
	}
}

func TestSaveUserLeague_Upsert(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())
	saveTestUserLeague(t, baseURL, "user-a", "league-1", "First Label")
	ul := saveTestUserLeague(t, baseURL, "user-a", "league-1", "Updated Label")

	if ul.Label != "Updated Label" {
		t.Errorf("expected updated label, got %q", ul.Label)
	}

	resp, err := http.Get(baseURL + "/api/league-bookmarks?user_id=user-a")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var leagues []provider.UserLeague
	if err := json.NewDecoder(resp.Body).Decode(&leagues); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(leagues) != 1 {
		t.Errorf("expected 1 entry after upsert, got %d", len(leagues))
	}
	if leagues[0].Label != "Updated Label" {
		t.Errorf("expected %q, got %q", "Updated Label", leagues[0].Label)
	}
}

func TestUpdateUserLeague(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())
	saveTestUserLeague(t, baseURL, "user-a", "league-1", "Old Label")

	body := `{"user_id":"user-a","label":"New Label"}`
	req, _ := http.NewRequest(http.MethodPatch, baseURL+"/api/league-bookmarks/league-1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var ul provider.UserLeague
	if err := json.NewDecoder(resp.Body).Decode(&ul); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if ul.Label != "New Label" {
		t.Errorf("expected %q, got %q", "New Label", ul.Label)
	}
}

func TestUpdateUserLeague_NotFound(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())

	body := `{"user_id":"user-a","label":"Whatever"}`
	req, _ := http.NewRequest(http.MethodPatch, baseURL+"/api/league-bookmarks/nonexistent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteUserLeague(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())
	saveTestUserLeague(t, baseURL, "user-a", "league-1", "To Delete")

	req, _ := http.NewRequest(http.MethodDelete, baseURL+"/api/league-bookmarks/league-1?user_id=user-a", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	listResp, err := http.Get(baseURL + "/api/league-bookmarks?user_id=user-a")
	if err != nil {
		t.Fatalf("list request failed: %v", err)
	}
	defer listResp.Body.Close()

	var leagues []provider.UserLeague
	if err := json.NewDecoder(listResp.Body).Decode(&leagues); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(leagues) != 0 {
		t.Errorf("expected empty list after delete, got %d", len(leagues))
	}
}

func TestDeleteUserLeague_NotFound(t *testing.T) {
	baseURL := newTestServer(t, noopHandler())

	req, _ := http.NewRequest(http.MethodDelete, baseURL+"/api/league-bookmarks/nonexistent?user_id=user-a", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
