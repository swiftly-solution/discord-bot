export type ParsedCall = { name: string; args: Record<string, unknown> };

const MARK = /<[^<>\n]*DSML[^<>\n]*>/;

export const hasDsml = (text: string) => /DSML/.test(text);

export const parseDsml = (text: string): ParsedCall[] =>
    text
        .split(/invoke\s+name="/)
        .slice(1)
        .map((part) => {
            const name = part.slice(0, part.indexOf('"'));
            const args: Record<string, unknown> = {};
            for (const m of part.matchAll(
                /parameter\s+name="([^"]+)"([^>]*)>([\s\S]*?)<[^<>]*\/[^<>]*parameter>/g,
            )) {
                const raw = m[3].trim();
                if (/string="false"/.test(m[2])) {
                    try {
                        args[m[1]] = JSON.parse(raw);
                        continue;
                    } catch {}
                }
                args[m[1]] = raw;
            }
            return { name, args };
        })
        .filter((c) => c.name);

export const stripDsml = (text: string) => {
    const i = text.search(MARK);
    return (i < 0 ? text : text.slice(0, i)).trim();
};
