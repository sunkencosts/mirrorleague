CREATE TABLE players(
    player_id text PRIMARY KEY,
    first_name text NOT NULL DEFAULT '',
    last_name text NOT NULL DEFAULT '',
    team text NOT NULL DEFAULT '',
    active boolean NOT NULL DEFAULT FALSE,
    fantasy_positions text[] NOT NULL DEFAULT '{}',
    number int NOT NULL DEFAULT 0,
    age int NOT NULL DEFAULT 0,
    rarity text NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT now())
