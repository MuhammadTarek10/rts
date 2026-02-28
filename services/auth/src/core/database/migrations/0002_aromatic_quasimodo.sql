DO $$ BEGIN
 CREATE TYPE "public"."user_roles" AS ENUM('user', 'admin');
EXCEPTION
WHEN duplicate_object THEN null;
END $$;

ALTER TABLE "users" ADD COLUMN "role" "user_roles" DEFAULT 'user' NOT NULL;