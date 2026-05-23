# Plan: Game-Lock Enforcement

## Goal

Prevent users from changing their lineup starters once a player's game has kicked off. A starter whose game has already started should be locked — `PATCH /api/lineups/{id}` and `POST /api/lineups` should both reject with `423 Locked` if any proposed starter is on a team whose kickoff has passed.

## What Was Attempted (and Why It Was Deleted)

A `checkGameLocks` function was added to `HandleCreateLineup` and `HandleUpdateLineup`. It called a `GetNFLSchedule` method on the Sleeper client that hit `/schedule/nfl/regular/{season}/{week}` — **this endpoint does not exist on the Sleeper API**. Because `checkGameLocks` was written to fail open (return nil on any error, to avoid blocking users), the entire feature was silently non-functional. No tests caught this because `lineupSleeperHandler` returned an empty 200 for unknown paths, which also caused a silent pass.

## What Needs to Be Figured Out Before Implementing

### 1. Where does the kickoff schedule actually come from?

Sleeper does not expose a schedule endpoint. Options:
- **Sleeper matchup `starting_slot` / `custom_points` fields** — Sleeper matchup data may include game start timestamps on individual player entries. Investigate the raw JSON from `GET /league/{leagueId}/matchups/{week}`.
- **Tank01 NFL API (RapidAPI)** — Already planned for Step 3. Has game schedules. Could be pulled in early just for kickoff times.
- **`nfl-data-py` / `nfl_data_py` open dataset** — Free, cached weekly schedule data derived from the official NFL API.
- **ESPN schedule API** — Undocumented but widely used: `site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard?week={week}&seasontype=2&dates={year}`. Returns game times.
- **Official NFL API** — Requires a feed subscription, not practical.

**Action needed:** Pick a source, verify the endpoint exists and returns game times with team abbreviations that match Sleeper's team codes.

### 2. Do Sleeper team codes match the source's team codes?

The lock check maps `player.Team` (a Sleeper team abbreviation, e.g. `"BUF"`) to a kickoff time. Whatever schedule source is used must use the same abbreviations, or a mapping layer is needed.

### 3. Fail-open vs fail-closed policy

The deleted code failed open (schedule unavailable → no locking). Decide:
- **Fail open** (current choice): user-friendly, but the lock feature degrades silently. Should at minimum log a warning or return a response header so operators know locking was skipped.
- **Fail closed**: reject the request if schedule is unavailable. Safer for integrity, worse UX.

### 4. Test coverage gap

When implementing, `lineupSleeperHandler` (or whatever fake handler is used in tests) **must** serve the schedule endpoint path. The old implementation was never caught because the test handler returned an empty 200 for unknown paths, which caused a silent pass through `checkGameLocks`. Add a schedule route to the shared test handler, or assert in tests that the schedule endpoint is actually called.

## Intended Handler Flow (once schedule source is known)

```
POST /api/lineups  or  PATCH /api/lineups/{id}
  1. Auth check
  2. Decode + validate request
  3. validateStarters (players on roster for that week)
  4. checkGameLocks (any starter's game already started?)
       - fetch schedule for week (with caching)
       - look up each starter's team → kickoff time
       - if time.Now().After(kickoff) → return errGameStarted
  5. store.CreateLineup / store.UpdateLineup
```

## Caching Note

Schedule data for a given week changes only before kickoffs. A 30-minute TTL is reasonable during game day; a longer TTL (e.g. 6 hours) is fine for future weeks. The cache implementation in `sleeper.Client` was structurally correct — just pointed at a non-existent URL.

## Related Files

- `api/internal/handlers/lineup.go` — `HandleCreateLineup`, `HandleUpdateLineup`, `validateStarters`
- `api/internal/sleeper/client.go` — where `GetNFLSchedule` (or equivalent) will live
- `api/cmd/server/routes.go` — `sleeperDeps` interface needs the new method
- `api/cmd/server/server_test.go` — `lineupSleeperHandler` needs a schedule route
- `api/pkg/config/config.go` — `CurrentSeason` field (already added, can keep)
