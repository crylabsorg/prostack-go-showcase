CREATE TABLE user_outbox (
    id uuid PRIMARY KEY,
    subject text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO user_outbox (id, subject, payload, created_at)
SELECT id, 'user.created', payload, created_at FROM email_outbox;

DROP TABLE email_outbox;
