package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sunkencosts/mirror-me/internal/provider"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) UpsertPlayers(ctx context.Context, players []provider.Player) error {
	batch := &pgx.Batch{}
	for _, player := range players {
		if player.FantasyPositions == nil {
			player.FantasyPositions = []string{}
		}
		batch.Queue(`INSERT INTO players (player_id, first_name, last_name, team, active, fantasy_positions, number, age, rarity, updated_at)
					VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
					ON CONFLICT (player_id)
					DO UPDATE SET first_name=EXCLUDED.first_name, last_name=EXCLUDED.last_name, team=EXCLUDED.team, active=EXCLUDED.active,
      				fantasy_positions=EXCLUDED.fantasy_positions, number=EXCLUDED.number, age=EXCLUDED.age, rarity=EXCLUDED.rarity, updated_at=now()
					WHERE players.first_name IS DISTINCT FROM EXCLUDED.first_name
					   OR players.last_name IS DISTINCT FROM EXCLUDED.last_name
					   OR players.team IS DISTINCT FROM EXCLUDED.team
					   OR players.active IS DISTINCT FROM EXCLUDED.active
					   OR players.fantasy_positions IS DISTINCT FROM EXCLUDED.fantasy_positions
					   OR players.number IS DISTINCT FROM EXCLUDED.number
					   OR players.age IS DISTINCT FROM EXCLUDED.age
					   OR players.rarity IS DISTINCT FROM EXCLUDED.rarity`,
			player.PlayerID, player.FirstName, player.LastName, player.Team, player.Active, player.FantasyPositions, player.Number, player.Age, player.Rarity)
	}
	results := s.pool.SendBatch(ctx, batch)
	defer results.Close()

	for _, player := range players {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("upserting player %s: %w", player.PlayerID, err)
		}
	}
	return nil
}

func (s *Store) GetPlayersByIDs(ctx context.Context, ids []string) (map[string]provider.Player, error) {
	rows, err := s.pool.Query(ctx, "SELECT player_id, first_name, last_name, team, active, fantasy_positions, number, age, rarity FROM players WHERE player_id = ANY($1)", ids)

	if err != nil {
		return nil, fmt.Errorf("selecting players by id: %w", err)
	}
	defer rows.Close()

	result := map[string]provider.Player{}
	for rows.Next() {
		var p provider.Player
		if err := rows.Scan(&p.PlayerID, &p.FirstName, &p.LastName, &p.Team, &p.Active, &p.FantasyPositions, &p.Number, &p.Age, &p.Rarity); err != nil {
			return nil, fmt.Errorf("scanning player: %w", err)
		}
		result[p.PlayerID] = p
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating players: %w", err)
	}
	return result, nil
}

func (s *Store) CreateLineup(ctx context.Context, userID, leagueID string, rosterID, week int, starters []string) (provider.Lineup, error) {

	var l provider.Lineup
	err := s.pool.QueryRow(ctx, `
					INSERT INTO lineups (user_id, league_id, roster_id, week, starters)
					VALUES ($1, $2, $3, $4, $5)
					RETURNING id, user_id, league_id, roster_id, week, starters, created_at, updated_at
					`, userID, leagueID, rosterID, week, starters).Scan(
		&l.ID, &l.UserID, &l.LeagueID, &l.RosterID, &l.Week, &l.Starters, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return provider.Lineup{}, fmt.Errorf("creating lineup: %w", err)
	}
	return l, nil
}
func (s *Store) GetLineup(ctx context.Context, id string) (provider.Lineup, error) {
	var l provider.Lineup
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, league_id, roster_id, week, starters, created_at, updated_at
		FROM lineups WHERE id = $1
	`, id).Scan(
		&l.ID, &l.UserID, &l.LeagueID, &l.RosterID, &l.Week, &l.Starters, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return provider.Lineup{}, fmt.Errorf("getting lineup %s: %w", id, err)
	}
	return l, nil
}

func (s *Store) UpdateLineup(ctx context.Context, id string, starters []string) (provider.Lineup, error) {
	var l provider.Lineup
	err := s.pool.QueryRow(ctx, `
				UPDATE lineups 
				SET starters = $2, updated_at = now()
				WHERE id = $1
				RETURNING id, user_id, league_id, roster_id, week, starters, created_at, updated_at
							`,
		id, starters).Scan(&l.ID, &l.UserID, &l.LeagueID, &l.RosterID, &l.Week, &l.Starters, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return provider.Lineup{}, fmt.Errorf("updating lineup %s: %w", id, err)
	}
	return l, nil
}
