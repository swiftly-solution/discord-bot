import { AttachmentBuilder, Message, ThreadChannel } from "discord.js";

const FENCE = "`".repeat(3);
const FENCE_LINE = /^\s*`{3}(.*)$/;
const INFO = /^[\w+#.-]{1,20}$/;

export const cleanMentions = (text: string, users: string[]) =>
    text
        .replace(/<@&\d+>/g, "")
        .replace(/<@!?(\d+)>/g, (_, id: string) =>
            users.includes(id) ? `<@${id}>` : "",
        )
        .replace(/@\s*(everyone|here)\b/gi, "$1");

export const linkTopics = (text: string) =>
    text.replace(
        /`?\btopic\s+#?(\d+)\b`?/gi,
        "https://forum.swiftlys2.net/t/-/$1",
    );

export const unwrapLinks = (text: string) =>
    text
        .replace(/`(https?:\/\/[^\s`]+)`/g, "$1")
        .replace(
            /`{3}[\w+#.-]*\n((?:[ \t]*https?:\/\/\S+[ \t]*\n?)+)`{3}/g,
            (_, urls: string) => urls.trim(),
        );

export const chunk = (text: string, max = 1900): string[] => {
    const out: string[] = [];
    let rest = text.trim();
    let fence: string | null = null;

    while (rest.length > 0) {
        const prefix = fence ? `${fence}\n` : "";
        const budget = max - prefix.length - FENCE.length - 1;

        let part: string;
        if (rest.length <= budget) {
            part = rest;
            rest = "";
        } else {
            let cut = rest.lastIndexOf("\n\n", budget);
            if (cut < budget / 2) cut = rest.lastIndexOf("\n", budget);
            if (cut < budget / 2) cut = rest.lastIndexOf(" ", budget);
            if (cut <= 0) cut = budget;
            part = rest.slice(0, cut);
            rest = rest.slice(cut).replace(/^(\n+| )/, "");
        }

        for (const line of part.split("\n")) {
            const info = FENCE_LINE.exec(line)?.[1].trim();
            if (info === undefined || info.includes(FENCE)) continue;
            fence = fence ? null : FENCE + (INFO.test(info) ? info : "");
        }
        out.push(prefix + part + (fence ? `\n${FENCE}` : ""));
    }
    return out;
};

export const prepare = (text: string, pingable: string[] = []) =>
    unwrapLinks(linkTopics(cleanMentions(text, pingable)));

const CODE_LIMIT = 1500;
const MAX_FILES = 10;
const EXT: Record<string, string> = {
    cs: "cs",
    csharp: "cs",
    csproj: "csproj",
    ts: "ts",
    typescript: "ts",
    js: "js",
    javascript: "js",
    py: "py",
    python: "py",
    cpp: "cpp",
    "c++": "cpp",
    h: "h",
    lua: "lua",
    json: "json",
    yml: "yml",
    yaml: "yaml",
    xml: "xml",
    toml: "toml",
    ini: "ini",
    cfg: "cfg",
    md: "md",
    markdown: "md",
    log: "log",
};

export const extractCode = (text: string) => {
    const files: { name: string; content: string }[] = [];
    const out = text.replace(
        /`{3}([\w+#.-]*)\n([\s\S]*?)\n?`{3}/g,
        (block, lang: string, code: string) => {
            if (code.length <= CODE_LIMIT || files.length >= MAX_FILES)
                return block;
            const name = `snippet-${files.length + 1}.${EXT[lang.toLowerCase()] ?? "txt"}`;
            files.push({ name, content: code });
            return `📎 **${name}** (${code.split("\n").length} lines) is attached below.`;
        },
    );
    return { text: out, files };
};

export const sendChunks = async (
    thread: ThreadChannel,
    text: string,
    replyTo?: Message,
    pingable: string[] = [],
) => {
    const { text: body, files } = extractCode(prepare(text, pingable));
    const parts = chunk(body);
    if (parts.length == 0) parts.push("I have nothing to say about that. 😶");

    for (const [i, content] of parts.entries()) {
        await thread.send({
            content,
            allowedMentions: { parse: [], users: pingable, roles: [] },
            reply:
                i == 0 && replyTo
                    ? { messageReference: replyTo.id }
                    : undefined,
            files:
                i == parts.length - 1 && files.length
                    ? files.map(
                          (f) =>
                              new AttachmentBuilder(Buffer.from(f.content), {
                                  name: f.name,
                              }),
                      )
                    : undefined,
        });
    }
};
