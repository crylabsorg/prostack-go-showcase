CREATE TABLE email_outbox (
    id uuid PRIMARY KEY REFERENCES users(id),
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
