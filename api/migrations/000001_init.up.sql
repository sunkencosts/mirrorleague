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
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE lineups(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id text NOT NULL,
    league_id text NOT NULL,
    roster_id int NOT NULL,
    week_number int NOT NULL,
    source text NOT NULL,
    starters text[] NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT lineups_key UNIQUE (user_id, league_id, roster_id, week_number, source)
);

CREATE TABLE league_bookmarks(
    user_id     text NOT NULL,
    league_id   text NOT NULL,
    label       text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, league_id)
);
