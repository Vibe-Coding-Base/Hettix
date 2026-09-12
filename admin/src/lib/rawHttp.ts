import type { KeyValuePair } from "lib/components/KeyValuePair";

// Reconstructs raw HTTP messages (request line / status line + headers + body)
// from the structured fields, so the request and response views can show the
// full raw message the way Burp does.

function headerLines(headers: KeyValuePair[]): string[] {
  return headers.filter((h) => h.key).map((h) => `${h.key}: ${h.value}`);
}

function assemble(startLine: string, headers: KeyValuePair[], body?: string | null): string {
  const raw = [startLine, ...headerLines(headers)].join("\r\n") + "\r\n\r\n";
  return body ? raw + body : raw;
}

export function rawRequest(input: {
  method?: string;
  url?: string;
  proto?: string | null;
  headers: KeyValuePair[];
  body?: string | null;
}): string {
  let target = input.url ?? "";
  try {
    const u = new URL(input.url ?? "");
    target = u.pathname + u.search;
  } catch {
    // Not an absolute URL; use it verbatim as the request target.
  }
  const startLine = `${input.method ?? "GET"} ${target || "/"} ${input.proto || "HTTP/1.1"}`;
  return assemble(startLine, input.headers, input.body);
}

export function rawResponse(input: {
  proto?: string | null;
  statusCode?: number | null;
  statusReason?: string | null;
  headers: KeyValuePair[];
  body?: string | null;
}): string {
  const startLine = [input.proto || "HTTP/1.1", input.statusCode ?? "", input.statusReason ?? ""]
    .filter((part) => part !== "" && part !== null && part !== undefined)
    .join(" ");
  return assemble(startLine, input.headers, input.body);
}
