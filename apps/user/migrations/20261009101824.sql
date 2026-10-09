-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "email" text NOT NULL,
  "first_name" text NULL,
  "last_name" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email")
);
-- Create "user_outbox" table
CREATE TABLE "public"."user_outbox" (
  "id" uuid NOT NULL DEFAULT uuidv7(),
  "user_id" uuid NOT NULL,
  "subject" text NOT NULL,
  "payload" jsonb NOT NULL,
  "processed_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "user_outbox_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "user_outbox_pending_idx" to table: "user_outbox"
CREATE INDEX "user_outbox_pending_idx" ON "public"."user_outbox" ("id") WHERE (processed_at IS NULL);
