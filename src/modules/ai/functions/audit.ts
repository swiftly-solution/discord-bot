import { API, MODEL } from "./deepseek";

const SITE = new Set([
    "swiftlys2.net",
    "www.swiftlys2.net",
    "docs.swiftlys2.net",
    "forum.swiftlys2.net",
]);

export type LinkReport = { total: number; verified: number; missing: string[] };
export type Judge = { supported: number; unbacked: string[] };
export type Report = {
    links: LinkReport;
    tools: string[];
    judge: Judge | null;
};
type Call = (messages: { role: string; content: string }[]) => Promise<any>;

const URL_RE = /https?:\/\/[^\s<>()[\]"'`]+/g;
const TOPIC =
    /forum\.swiftlys2\.net\/t\/(?:[^/\s"]*[^/\s"\d][^/\s"]*\/)?(\d+)/g;
const norm = (s: string) =>
    s.toLowerCase().replace(/#.*$/, "").replace(/\/+$/, "");

export const checkLinks = (text: string, evidence: string[]): LinkReport => {
    const urls = [
        ...new Set(
            (text.match(URL_RE) ?? []).map((u) =>
                u.replace(/[.,;:!?*_]+$/, ""),
            ),
        ),
    ];
    const blob = evidence.join("\n").toLowerCase();
    const topics = new Set([...blob.matchAll(TOPIC)].map((m) => m[1]));
    const gh = evidence
        .filter((e) => e.startsWith("github__"))
        .join("\n")
        .toLowerCase();

    const backed = (url: string) => {
        let u: URL;
        try {
            u = new URL(url);
        } catch {
            return false;
        }
        const host = u.hostname.toLowerCase();
        const path = u.pathname.toLowerCase().replace(/\/+$/, "");
        if (blob.includes(norm(url))) return true;
        if (host == "forum.swiftlys2.net") {
            const m = /^\/t\/(?:[^/]*[^/\d][^/]*\/)?(\d+)/.exec(path);
            if (m && topics.has(m[1])) return true;
        }
        if (host == "github.com") {
            const [owner, repo] = path.split("/").filter(Boolean);
            return !!owner && !!repo && gh.includes(owner) && gh.includes(repo);
        }
        if (SITE.has(host)) return path == "" || blob.includes(path);
        return false;
    };

    const missing = urls.filter((u) => !backed(u));
    return {
        total: urls.length,
        verified: urls.length - missing.length,
        missing,
    };
};

export const parseVerdicts = (raw: string): Judge | null => {
    try {
        const json = JSON.parse(/\{[\s\S]*\}/.exec(raw)?.[0] ?? "");
        if (!Array.isArray(json.claims)) return null;
        const claims = json.claims as { claim?: string; verdict?: string }[];
        return {
            supported: claims.filter((c) => c.verdict == "supported").length,
            unbacked: claims
                .filter((c) => c.verdict == "unsupported")
                .map((c) => String(c.claim ?? "")),
        };
    } catch {
        return null;
    }
};

export const JUDGE = `You are a strict fact checker. You get an EVIDENCE block (tool results and messages from the thread) and an ANSWER written by an assistant. List the concrete factual claims in the ANSWER (APIs, function or class names, versions, commands, links, behaviours). For each claim give a verdict: "supported" if the EVIDENCE confirms it, "unsupported" if the EVIDENCE does not confirm it, "opinion" for jokes, greetings, advice or opinions. Reply with JSON only: {"claims":[{"claim":"short text","verdict":"supported|unsupported|opinion"}]}`;

const deepseek: Call = async (messages) => {
    const res = await fetch(API, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${process.env.DEEPSEEK_AI_API_KEY}`,
        },
        signal: AbortSignal.timeout(30_000),
        body: JSON.stringify({
            model: MODEL,
            messages,
            temperature: 0,
            response_format: { type: "json_object" },
        }),
    });
    if (!res.ok) throw new Error(`DeepSeek judge ${res.status}`);
    return res.json();
};

export const judgeAnswer = async (
    answer: string,
    evidence: string[],
    call: Call = deepseek,
): Promise<Judge | null> => {
    if (answer.length < 50) return null;
    try {
        const res = await call([
            { role: "system", content: JUDGE },
            {
                role: "user",
                content: `EVIDENCE:\n${evidence.join("\n---\n").slice(0, 60_000)}\n\nANSWER:\n${answer}`,
            },
        ]);
        return parseVerdicts(res.choices[0].message.content ?? "");
    } catch (err) {
        console.error("[AI] Judge failed:", err);
        return null;
    }
};

export const isAudit = (line: string) => line.startsWith("-# ");

const short = (s: string, n = 80) => (s.length > n ? `${s.slice(0, n)}…` : s);

export const footer = ({ links, tools, judge }: Report): string | null => {
    const claims = judge ? judge.supported + judge.unbacked.length : 0;
    const total = links.total + claims;
    const verified = links.verified + (judge?.supported ?? 0);
    if (total == 0 && tools.length == 0) return null;

    const lines: string[] = [];
    if (total == 0) {
        lines.push(`-# 🛠️ Sources: ${tools.join(", ")}`);
    } else {
        const pct = Math.round((100 * verified) / total);
        const mark = pct >= 80 ? "🟢" : pct >= 50 ? "🟡" : "🔴";
        const detail = [
            links.total ? `links ${links.verified}/${links.total}` : "",
            claims ? `claims ${judge!.supported}/${claims}` : "",
            tools.length ? `tools: ${tools.join(", ")}` : "no tools used",
        ].filter(Boolean);
        lines.push(
            `-# ${mark} ${pct}% verified (${verified}/${total}) · ${detail.join(" · ")}`,
        );
    }
    if (links.missing.length)
        lines.push(
            `-# ⚠️ not found in sources: ${links.missing
                .slice(0, 3)
                .map((u) => `<${short(u, 120)}>`)
                .join(" ")}`,
        );
    if (judge?.unbacked.length)
        lines.push(
            `-# ⚠️ unverified: ${judge.unbacked
                .slice(0, 3)
                .map((c) => short(c))
                .join("; ")}`,
        );
    return lines.join("\n");
};

export const audit = async (
    text: string,
    evidence: string[],
    tools: string[],
    call?: Call,
) =>
    footer({
        links: checkLinks(text, evidence),
        tools,
        judge:
            tools.length == 0 ? await judgeAnswer(text, evidence, call) : null,
    });
