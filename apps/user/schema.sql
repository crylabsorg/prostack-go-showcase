CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    email text NOT NULL UNIQUE,
    first_name text,
    last_name text
);

CREATE TABLE user_outbox (
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    subject text NOT NULL,
    payload jsonb NOT NULL
);
