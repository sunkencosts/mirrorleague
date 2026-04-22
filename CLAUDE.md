# mirror-me

## Project Overview
A fantasy football "mirror" app where users can import a celebrity or public figure's Sleeper fantasy football team and set their own lineup using the same roster — then compare scores to prove they're a better fantasy manager.

**Core concept:** Same players, same pool, different brain making the decisions.

---

## Tech Stack
- **Backend:** Go (primary focus)
- **Database:** (Postgres) — added in Step 2
- **Cache:** In-memory to start, Redis (Upstash) later
- **External APIs:** Sleeper API (free, no auth), Tank01 NFL API via RapidAPI (player stats, added in Step 3)

---

## Current Status
- [ ] Step 1 — Mirror a league (in progress)
  - [x] `GET /league/:leagueId/rosters` — returns rosters with resolved player names, team name, and player image URLs
  - [ ] `GET /user/:username` — returns user object + their leagues
  - [ ] `GET /league/:leagueId/week/:week` — returns each team's lineup for that week
- [ ] Step 2 — Reaction layer (votes, predictions)
- [ ] Step 3 — User accounts + lineup picks
- [ ] Step 4 — Scoring and winner declaration

---

## Architecture

```
/cmd
  /server
    main.go              ← HTTP server entry point

/internal
  /sleeper
    client.go            ← all Sleeper API calls (GetRosters, getLeagueUsers)
    cache.go             ← player map caching (once/day, 5MB, saved to players.json)
  /handlers
    rosters.go           ← HTTP handler for roster endpoint
    rosters_test.go      ← handler tests
  /provider
    provider.go          ← shared interfaces and models (Player, Roster, Provider)

/pkg
  /config
    config.go            ← env vars and configuration
```

---

## Sleeper API — Step 1 Endpoints

All calls are read-only, no auth token required.
Base URL: `https://api.sleeper.app/v1`

| Step | Endpoint | Purpose |
|------|----------|---------|
| 1 | `GET /user/<username>` | Resolve username → user_id |
| 2 | `GET /user/<user_id>/leagues/nfl/2025` | Get all leagues for user |
| 3 | `GET /league/<league_id>/rosters` | Get rosters (player IDs) |
| 4 | `GET /league/<league_id>/users` | Get owners of each roster |
| 5 | `GET /players/nfl` | Full player ID → name map (cache, once/day) |
| 6 | `GET /league/<league_id>/matchups/<week>` | Get lineups + scores for a week |

**Rate limit:** Stay under 1,000 API calls/minute or risk IP block.
**Player map:** 5MB response — save to memory/disk, do not call repeatedly.

---

## Core Go Patterns to Follow

### Sleeper Client Interface
```go
type SleeperClient interface {
    GetUser(username string) (*User, error)
    GetLeagues(userID string) ([]League, error)
    GetRosters(leagueID string) ([]Roster, error)
    GetUsers(leagueID string) ([]LeagueUser, error)
    GetMatchup(leagueID string, week int) ([]Matchup, error)
    GetPlayers() (map[string]Player, error)
}
```

### Key Models (internal/provider/provider.go)
```go
type Player struct {
    PlayerID         string   `json:"player_id"`
    FirstName        string   `json:"first_name"`
    LastName         string   `json:"last_name"`
    Number           int      `json:"number"`
    Age              int      `json:"age"`
    Team             string   `json:"team"`
    Active           bool     `json:"active"`
    FantasyPositions []string `json:"fantasy_positions"`
    ImageURL         string   `json:"image_url"`  // constructed from sleepercdn.com
}

type Roster struct {
    RosterID int      `json:"roster_id"`
    OwnerID  string   `json:"owner_id"`
    TeamName string   `json:"team_name"`  // from /league/:id/users metadata
    Players  []Player `json:"players"`
    Starters []Player `json:"starters"`
}
```

### Player Image URLs
Sleeper doesn't return image URLs directly. Construct them from the CDN:
```
https://sleepercdn.com/content/nfl/players/thumb/<player_id>.jpg
```

### Parallel API Calls Pattern
```go
var wg sync.WaitGroup
var rosters []Roster
var users []LeagueUser
var rosterErr, userErr error

wg.Add(2)
go func() {
    defer wg.Done()
    rosters, rosterErr = client.GetRosters(leagueID)
}()
go func() {
    defer wg.Done()
    users, userErr = client.GetUsers(leagueID)
}()
wg.Wait()
```

### Error Wrapping Pattern
```go
user, err := client.GetUser(username)
if err != nil {
    return nil, fmt.Errorf("fetching user %s: %w", username, err)
}
```

---

## API Endpoints to Build (Step 1)

```
GET /user/:username          → returns user object + their leagues
GET /league/:leagueId        → returns league info + all rosters with owner names
GET /league/:leagueId/week/:week → returns each team's lineup for that week
```

**Done when:** `curl localhost:8080/user/someSleeperUsername` returns clean JSON with their leagues and roster.

---

## Key Constraints and Notes

- **Sleeper is read-only** — you cannot modify any league data through the API
- **No official terms of service** for the Sleeper API — be a good citizen, cache aggressively, attribute Sleeper in the UI
- **Player stats are NOT in Sleeper API** — Sleeper licenses that from Genius Sports (official NFL data partner). Tank01 on RapidAPI is the indie developer solution for Step 3.
- **Scoring settings live in the League object** — pull them alongside the roster so you can apply correct PPR/half-PPR/standard scoring later
- **Sleeper username vs user_id** — usernames can change, always store and reference by user_id

---

## Future Steps (Not Building Yet)

**Step 2 — Reaction layer**
- Users vote on trades, sit/start decisions
- Anonymous sessions first, no auth required yet
- Store reactions in Supabase

**Step 3 — User accounts + lineup picks**
- Supabase Auth
- Users set their own lineup from the celebrity's roster before Sunday 1pm ET
- Picks locked at kickoff

**Step 4 — Scoring + winner declaration**
- Integrate Tank01 NFL API for individual player stats
- Apply league scoring settings to both lineups
- Generate shareable result card

**Future — Celebrity challenge layer**
- Celebrity shares their Sleeper username publicly
- Fans mirror their team
- Season-long leaderboard of who manages the roster best
- AI-generated weekly roast of bad decisions (Claude API, Haiku model)

---

## Environment Variables
```
PORT=8080
SLEEPER_BASE_URL=https://api.sleeper.app/v1
# Later:
# DATABASE_URL=
# DATABASE_KEY=
# RAPIDAPI_KEY=       (Tank01)
# ANTHROPIC_API_KEY=  (Claude AI features)
```
