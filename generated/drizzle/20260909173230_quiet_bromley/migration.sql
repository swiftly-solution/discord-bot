CREATE TABLE "channel_roles" (
	"id" text PRIMARY KEY,
	"guild_id" text NOT NULL,
	"channel_id" text NOT NULL,
	"role" text NOT NULL,
	"created_at" timestamp NOT NULL
);
