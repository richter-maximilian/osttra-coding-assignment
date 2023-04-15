CREATE TABLE IF NOT EXISTS messages (
    id text NOT NULL PRIMARY KEY CHECK (id <> ''),
    user_name text NOT NULL CHECK (user_name <> ''),
    content text NOT NULL,
    sent_at timestamptz NOT NULL,
    fetched_at timestamptz
);
