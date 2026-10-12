import { eq } from "drizzle-orm";
import { db } from "../../../core/db/drizzle";
import { channelRoles } from "../../../core/db/schema/schema";

export const aiChannels: string[] = [];

export default async () => {
    const rows = await db
        .select()
        .from(channelRoles)
        .where(eq(channelRoles.role, "forum"));

    aiChannels.splice(0, aiChannels.length, ...rows.map((r) => r.channelId));
};
