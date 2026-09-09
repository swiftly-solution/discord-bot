import { pgTable, text, timestamp } from "drizzle-orm/pg-core";

export const channelRoles = pgTable("channel_roles", {
    id: text("id").primaryKey(),
    guildId: text("guild_id").notNull(),
    channelId: text("channel_id").notNull(),
    role: text("role").notNull(),
    createdAt: timestamp("created_at").notNull(),
});

export type ChannelRole = typeof channelRoles.$inferSelect;
