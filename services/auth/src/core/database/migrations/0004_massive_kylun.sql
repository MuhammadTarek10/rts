ALTER TABLE "users" ADD COLUMN "email" text NOT NULL;--> statement-breakpoint
ALTER TABLE "users" DROP COLUMN "refresh_token_version";--> statement-breakpoint
ALTER TABLE "users" ADD CONSTRAINT "users_email_unique" UNIQUE("email");