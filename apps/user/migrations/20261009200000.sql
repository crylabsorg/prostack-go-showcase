ALTER TABLE user_outbox ADD COLUMN trace_context jsonb NOT NULL DEFAULT '{}'::jsonb;
