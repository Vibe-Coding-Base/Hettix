// Some APIs prefix JSON with an anti-hijacking guard (Google's )]}' , or
// for(;;); / while(1);). Strip it before detecting or parsing JSON so the
// prettifier and highlighter still work.
const XSSI_GUARD = /^(\)\]\}'[,\s]*|for\s*\(;;\);|while\s*\(1\);|&&&START&&&)\s*/;

function stripGuard(body: string): string {
  return body.replace(XSSI_GUARD, "").trimStart();
}

// prettify pretty-prints JSON and XML/HTML bodies; other content is returned
// unchanged. Shared by the request and response body viewers.
export function prettify(body: string): string {
  const stripped = stripGuard(body);

  if (stripped[0] === "{" || stripped[0] === "[") {
    try {
      return JSON.stringify(JSON.parse(stripped), null, 2);
    } catch {
      return body;
    }
  }

  if (stripped[0] === "<") {
    let depth = 0;
    return stripped
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
  const c = stripGuard(body)[0];
  return c === "{" || c === "[" || c === "<";
}
