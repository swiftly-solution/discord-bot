import { Client } from "@modelcontextprotocol/sdk/client";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp";

export const mcpClient = new Client({
    name: "swiftlychan-mcp-docs",
    version: "1.0.0",
});

export const mcpTransport = new StreamableHTTPClientTransport(
    new URL(
        `https://swiftlys2.net/api/mcp?bypass_key=${process.env.CF_BYPASS_KEY}`,
    ),
);
