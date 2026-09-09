// prettify pretty-prints JSON and XML/HTML bodies; other content is returned
// unchanged. Shared by the request and response body viewers.
export function prettify(body: string): string {
  const trimmed = body.trimStart();

  if (trimmed[0] === "{" || trimmed[0] === "[") {
    try {
      return JSON.stringify(JSON.parse(body), null, 2);
    } catch {
      return body;
    }
  }

  if (trimmed[0] === "<") {
    let depth = 0;
    return trimmed
      .replace(/>\s*</g, "><")
      .replace(/</g, "\n<")
      .split("\n")
      .filter((line) => line.trim() !== "")
      .map((line) => {
        if (/^<\/.+/.test(line)) {
          depth = Math.max(depth - 1, 0);
        }
        const indented = "  ".repeat(depth) + line;
        if (/^<[^/!?][^>]*[^/]>$/.test(line)) {
          depth += 1;
        }
        return indented;
      })
      .join("\n");
  }

  return body;
}

// canPrettify reports whether prettify would change the body, so a Pretty
// control can be shown only when it's useful.
export function canPrettify(body?: string | null): boolean {
  if (!body) {
    return false;
  }
  const c = body.trimStart()[0];
  return c === "{" || c === "[" || c === "<";
}
