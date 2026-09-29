-- Create "user_outbox" table
CREATE TABLE "public"."user_outbox" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "subject" text NOT NULL,
  "payload" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "name" text NOT NULL,
  "email" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email")
);
