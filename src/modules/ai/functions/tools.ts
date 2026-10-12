import { Client } from "@modelcontextprotocol/sdk/client";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp";
import { InMemoryTransport } from "@modelcontextprotocol/sdk/inMemory";
import { mcpClient as docsClient } from "../../../core/docs";
import { createDiscourseServer, scrub } from "../mcp/discourse";

export type ToolDef = {
    type: "function";
    function: { name: string; description: string; parameters: object };
};

const GITHUB_ALLOW = new Set([
    "get_file_contents",
    "search_code",
    "search_repositories",
    "list_commits",
    "get_commit",
    "list_branches",
    "list_releases",
    "list_issues",
    "issue_read",
    "search_issues",
    "list_pull_requests",
    "pull_request_read",
    "search_pull_requests",
]);

const MAX_RESULT = 20000;
const RETRY_MS = 60_000;

const TIMEOUT = { timeout: 30_000 };

let defs: ToolDef[] = [];
const routes = new Map<string, { client: Client; name: string }>();
const connected = new Set<string>();
let lastTry = 0;
let initializing: Promise<void> | null = null;

const connectGithub = async () => {
    if (!process.env.GIT_TOKEN) throw new Error("GIT_TOKEN is not set");
    const client = new Client({
        name: "swiftlychan-mcp-github",
        version: "1.0.0",
    });
    await client.connect(
        new StreamableHTTPClientTransport(
            new URL("https://api.githubcopilot.com/mcp/"),
            {
                requestInit: {
                    headers: {
                        Authorization: `Bearer ${process.env.GIT_TOKEN}`,
                        "X-MCP-Readonly": "true",
                    },
                },
            },
        ),
        TIMEOUT,
    );
    return client;
};

const connectDiscourse = async () => {
    const [a, b] = InMemoryTransport.createLinkedPair();
    await createDiscourseServer().connect(b);
    const client = new Client({
        name: "swiftlychan-mcp-discourse",
        version: "1.0.0",
    });
    await client.connect(a);
    return client;
};

const sources: Record<
    string,
    { connect: () => Promise<Client>; allow?: Set<string> }
> = {
    docs: { connect: async () => docsClient }, // connected in src/index.ts
    forum: { connect: connectDiscourse },
    github: { connect: connectGithub, allow: GITHUB_ALLOW },
};

// forget a source so the next init reconnects it
const drop = (prefix: string) => {
    connected.delete(prefix);
    defs = defs.filter((d) => !d.function.name.startsWith(`${prefix}__`));
    for (const name of routes.keys())
        if (name.startsWith(`${prefix}__`)) routes.delete(name);
};

const load = async (prefix: string) => {
    const source = sources[prefix];
    let client: Client | undefined;
    try {
        client = await source.connect();
        const { tools } = await client.listTools(undefined, TIMEOUT);
        drop(prefix);
        for (const t of tools) {
            if (source.allow && !source.allow.has(t.name)) continue;
            const name = `${prefix}__${t.name}`;
            routes.set(name, { client, name: t.name });
            defs.push({
                type: "function",
                function: {
                    name,
                    description: t.description ?? "",
                    parameters: t.inputSchema,
                },
            });
        }
        connected.add(prefix);
    } catch (err) {
        console.error(`[AI] MCP "${prefix}" unavailable:`, scrub(String(err)));
        if (client && client !== docsClient) await client.close().catch(() => {});
    }
};

// ponytail: retry at most once a minute; no live re-listing when a server adds tools
const init = () => {
    if (initializing) return initializing;
    if (Date.now() - lastTry < RETRY_MS) return Promise.resolve();
    lastTry = Date.now();
    initializing = Promise.all(
        Object.keys(sources)
            .filter((p) => !connected.has(p))
            .map(load),
    )
        .then(() => {})
        .finally(() => (initializing = null));
    return initializing;
};

export const getTools = async () => {
    if (connected.size < Object.keys(sources).length) await init();
    return defs;
};

export const callTool = async (
    name: string,
    args: unknown,
): Promise<string> => {
    const route = routes.get(name);
    if (!route) return `Unknown tool ${name}`;
    try {
        const res = await route.client.callTool(
            {
                name: route.name,
                arguments: (args ?? {}) as Record<string, unknown>,
            },
            undefined,
            TIMEOUT,
        );
        const text = (res.content as { type: string; text?: string }[])
            .filter((c) => c.type == "text")
            .map((c) => c.text)
            .join("\n");
        return scrub(text).slice(0, MAX_RESULT) || "(empty result)";
    } catch (err) {
        drop(name.split("__")[0]); // session may have died; reconnect on next answer
        lastTry = 0;
        return scrub(`Tool error: ${String(err)}`);
    }
};
