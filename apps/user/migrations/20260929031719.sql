-- Modify "users" table
ALTER TABLE "public"."users" DROP COLUMN "name", ADD COLUMN "first_name" text NULL, ADD COLUMN "last_name" text NULL;
