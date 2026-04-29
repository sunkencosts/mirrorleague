CREATE TABLE lineups(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id text NOT NULL,
    league_id text NOT NULL,
    roster_id int NOT NULL,
    week int NOT NULL,
    starters text[] NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

