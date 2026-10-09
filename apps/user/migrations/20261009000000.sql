-- Give outbox messages their own identity and retain acknowledged messages.
ALTER TABLE "public"."user_outbox"
  ADD COLUMN "user_id" uuid,
  ADD COLUMN "processed_at" timestamptz;

UPDATE "public"."user_outbox" SET "user_id" = "id", "id" = uuidv7();

ALTER TABLE "public"."user_outbox"
  ALTER COLUMN "user_id" SET NOT NULL,
  ADD CONSTRAINT "user_outbox_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id");

CREATE INDEX "user_outbox_pending_idx" ON "public"."user_outbox" ("id") WHERE "processed_at" IS NULL;

-- Pending events created with the old contract must use the new name fields.
UPDATE "public"."user_outbox"
SET "payload" = ("payload" - 'name') || jsonb_build_object('first_name', "payload"->>'name', 'last_name', '')
WHERE "subject" = 'user.created' AND "payload" ? 'name';
