import { Server } from "@modelcontextprotocol/sdk/server";
import {
    CallToolRequestSchema,
    ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types";

const BASE = "https://forum.swiftlys2.net";

export const scrub = (text: string) => {
    const key = process.env.CF_BYPASS_KEY;
    return key
        ? text.replaceAll(key, "***").replaceAll(encodeURIComponent(key), "***")
        : text;
};

const get = async (
    path: string,
    params: Record<string, string> = {},
): Promise<any> => {
    const qs = new URLSearchParams(params);
    qs.set("bypass_key", process.env.CF_BYPASS_KEY ?? ""); // always last
    const res = await fetch(`${BASE}${path}?${qs}`, {
        headers: { Accept: "application/json" },
    });
    if (!res.ok)
        throw new Error(`Discourse answered ${res.status} for ${path}`);
    return res.json();
};

const link = (slug: string | undefined, topicId: number, postNumber?: number) =>
    `${BASE}/t/${slug || "-"}/${topicId}${postNumber ? `/${postNumber}` : ""}`;

const parse = (value: unknown) => {
    let u: URL;
    try {
        u = new URL(String(value));
    } catch {
        throw new Error("url must be a forum.swiftlys2.net topic or post url");
    }
    const m = /^\/t\/(?:[^/]*[^/\d][^/]*\/)?(\d+)(?:\/(\d+))?\/?$/.exec(
        u.pathname,
    );
    if (u.origin != BASE || !m)
        throw new Error("url must be a forum.swiftlys2.net topic or post url");
    return {
        topicId: Number(m[1]),
        postNumber: m[2] ? Number(m[2]) : undefined,
    };
};

const plain = (html: string) =>
    html
        .replace(/<[^>]+>/g, "")
        .replaceAll("&nbsp;", " ")
        .replaceAll("&lt;", "<")
        .replaceAll("&gt;", ">")
        .replaceAll("&quot;", '"')
        .replaceAll("&#39;", "'")
        .replaceAll("&amp;", "&");

const tool = (
    name: string,
    description: string,
    prop: string,
    type: string,
) => ({
    name,
    description,
    inputSchema: {
        type: "object" as const,
        properties: { [prop]: { type, description: prop } },
        required: [prop],
    },
});

export const createDiscourseServer = () => {
    const server = new Server(
        { name: "swiftlys2-discourse", version: "1.0.0" },
        { capabilities: { tools: {} } },
    );

    server.setRequestHandler(ListToolsRequestSchema, async () => ({
        tools: [
            tool(
                "search",
                "Search the SwiftlyS2 forum (Discourse). Returns matching topics and post snippets.",
                "query",
                "string",
            ),
            tool(
                "get_topic",
                "Read a forum topic (title and the first posts). Pass the topic or post url from a search result.",
                "url",
                "string",
            ),
            tool(
                "get_post",
                "Read one forum post (raw markdown). Pass the post url from a search result or topic.",
                "url",
                "string",
            ),
        ],
    }));

    server.setRequestHandler(CallToolRequestSchema, async (req) => {
        const args = (req.params.arguments ?? {}) as Record<string, unknown>;
        try {
            let result: unknown;
            if (req.params.name == "search") {
                const data = await get("/search.json", {
                    q: String(args.query ?? ""),
                });
                const slugs = new Map<number, string>(
                    (data.topics ?? []).map((t: any) => [t.id, t.slug]),
                );
                result = {
                    topics: (data.topics ?? []).slice(0, 10).map((t: any) => ({
                        title: t.title,
                        posts: t.posts_count,
                        url: link(t.slug, t.id),
                    })),
                    posts: (data.posts ?? []).slice(0, 10).map((p: any) => ({
                        user: p.username,
                        snippet: p.blurb,
                        url: link(
                            slugs.get(p.topic_id),
                            p.topic_id,
                            p.post_number,
                        ),
                    })),
                };
            } else if (req.params.name == "get_topic") {
                const { topicId } = parse(args.url);
                const data = await get(`/t/${topicId}.json`);
                result = {
                    title: data.title,
                    url: link(data.slug, topicId),
                    posts: (data.post_stream?.posts ?? []).map((p: any) => ({
                        number: p.post_number,
                        user: p.username,
                        text: plain(p.cooked ?? ""),
                        url: link(data.slug, topicId, p.post_number),
                    })),
                };
            } else if (req.params.name == "get_post") {
                const { topicId, postNumber } = parse(args.url);
                const data = await get(
                    `/posts/by_number/${topicId}/${postNumber ?? 1}.json`,
                );
                result = {
                    user: data.username,
                    url: link(data.topic_slug, topicId, data.post_number),
                    raw: data.raw,
                };
            } else {
                throw new Error(`Unknown tool ${req.params.name}`);
            }
            return {
                content: [{ type: "text", text: JSON.stringify(result) }],
            };
        } catch (err) {
            return {
                isError: true,
                content: [{ type: "text", text: scrub(String(err)) }],
            };
        }
    });

    return server;
};
