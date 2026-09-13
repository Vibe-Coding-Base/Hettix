interface CurlRequest {
  method: string;
  url: string;
  headers?: { key: string; value: string }[];
  body?: string | null;
}

export type CurlPlatform = "unix" | "windows";

// Unix shells: single-quote and escape embedded single quotes.
function unixQuote(value: string): string {
  return `'${value.replace(/'/g, `'\\''`)}'`;
}

// Windows curl.exe (cmd): double-quote and backslash-escape embedded quotes.
function windowsQuote(value: string): string {
  return `"${value.replace(/"/g, `\\"`)}"`;
}

// toCurl renders a request as a copy-pasteable curl command. The Windows variant
// uses double quotes and `^` line continuations so it pastes into cmd; the Unix
// variant uses single quotes and `\` continuations for bash/zsh.
export function toCurl({ method, url, headers, body }: CurlRequest, platform: CurlPlatform = "unix"): string {
  const quote = platform === "windows" ? windowsQuote : unixQuote;
  const cont = platform === "windows" ? " ^\n  " : " \\\n  ";

  const parts = [`curl -i -X ${method} ${quote(url)}`];

  for (const { key, value } of headers ?? []) {
    if (key) {
      parts.push(`-H ${quote(`${key}: ${value}`)}`);
    }
  }

  if (body) {
    parts.push(`--data-raw ${quote(body)}`);
  }

  return parts.join(cont);
}
