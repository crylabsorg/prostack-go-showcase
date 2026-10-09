CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    email text NOT NULL UNIQUE,
    first_name text,
    last_name text
);

CREATE TABLE user_outbox (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id uuid NOT NULL REFERENCES users(id),
    subject text NOT NULL,
    payload jsonb NOT NULL,
    processed_at timestamptz
);

CREATE INDEX user_outbox_pending_idx ON user_outbox (id) WHERE processed_at IS NULL;
