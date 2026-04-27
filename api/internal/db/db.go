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
      				fantasy_positions=EXCLUDED.fantasy_positions, number=EXCLUDED.number, age=EXCLUDED.age, rarity=EXCLUDED.rarity, updated_at=now() `,
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
