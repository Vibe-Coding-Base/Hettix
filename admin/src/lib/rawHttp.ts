import type { KeyValuePair } from "lib/components/KeyValuePair";
import { prettify } from "lib/prettify";

// Reconstructs raw HTTP messages (request line / status line + headers + body)
// from the structured fields, so the request and response views can show the
// full raw message the way Burp does.

function headerLines(headers: KeyValuePair[]): string[] {
  return headers.filter((h) => h.key).map((h) => `${h.key}: ${h.value}`);
}

function assemble(startLine: string, headers: KeyValuePair[], body?: string | null, pretty?: boolean): string {
  const raw = [startLine, ...headerLines(headers)].join("\r\n") + "\r\n\r\n";
  const finalBody = body && pretty ? prettify(body) : body;
  return finalBody ? raw + finalBody : raw;
}

function requestTarget(url?: string): string {
  try {
    const u = new URL(url ?? "");
    return u.pathname + u.search;
  } catch {
    return url || "";
  }
}

export function rawRequest(input: {
  method?: string;
  url?: string;
  proto?: string | null;
  headers: KeyValuePair[];
  body?: string | null;
  pretty?: boolean;
}): string {
  const startLine = `${input.method ?? "GET"} ${requestTarget(input.url) || "/"} ${input.proto || "HTTP/1.1"}`;
  return assemble(startLine, input.headers, input.body, input.pretty);
}

// The response proto arrives as the GraphQL enum (HTTP10/HTTP11/HTTP20); show it
// in wire form.
const PROTO_DISPLAY: Record<string, string> = { HTTP10: "HTTP/1.0", HTTP11: "HTTP/1.1", HTTP20: "HTTP/2" };

export function rawResponse(input: {
  proto?: string | null;
  statusCode?: number | null;
  statusReason?: string | null;
  headers: KeyValuePair[];
  body?: string | null;
  pretty?: boolean;
}): string {
  const proto = (input.proto && PROTO_DISPLAY[input.proto]) || input.proto || "HTTP/1.1";
  const startLine = [proto, input.statusCode ?? "", input.statusReason ?? ""]
    .filter((part) => part !== "" && part !== null && part !== undefined)
    .join(" ");
  return assemble(startLine, input.headers, input.body, input.pretty);
}

export interface ParsedRawRequest {
  method: string;
  url: string;
  headers: KeyValuePair[];
  body: string;
}

// parseRawRequest turns an edited raw request back into structured fields for
// sending. The request target may be absolute or origin-form; for origin-form
// the URL is built from the Host header and the fallback scheme.
export function parseRawRequest(raw: string, fallbackScheme = "https"): ParsedRawRequest {
  const sep = raw.search(/\r?\n\r?\n/);
  const head = sep >= 0 ? raw.slice(0, sep) : raw;
  const body = sep >= 0 ? raw.slice(sep).replace(/^\r?\n\r?\n/, "") : "";

  const lines = head.split(/\r?\n/);
  const requestLine = lines.shift() ?? "";
  const [method = "GET", target = "/"] = requestLine.trim().split(/\s+/);

  const headers: KeyValuePair[] = [];
  for (const line of lines) {
    const idx = line.indexOf(":");
    if (idx > 0) {
      headers.push({ key: line.slice(0, idx).trim(), value: line.slice(idx + 1).trim() });
    }
  }

  let url = target;
  if (!/^https?:\/\//i.test(target)) {
    const host = headers.find((h) => h.key.toLowerCase() === "host")?.value ?? "";
    url = `${fallbackScheme}://${host}${target.startsWith("/") ? target : `/${target}`}`;
  }

  return { method, url, headers, body };
}

export interface ParsedRawResponse {
  statusCode: number;
  statusReason: string;
  headers: KeyValuePair[];
  body: string;
}

// parseRawResponse turns an edited raw response back into structured fields.
// The proto is left to the caller (it's rarely edited and needs enum mapping).
export function parseRawResponse(raw: string): ParsedRawResponse {
  const sep = raw.search(/\r?\n\r?\n/);
  const head = sep >= 0 ? raw.slice(0, sep) : raw;
  const body = sep >= 0 ? raw.slice(sep).replace(/^\r?\n\r?\n/, "") : "";

  const lines = head.split(/\r?\n/);
  const statusLine = (lines.shift() ?? "").trim();
  const match = /^\S+\s+(\d{3})\s*(.*)$/.exec(statusLine);

  const headers: KeyValuePair[] = [];
  for (const line of lines) {
    const idx = line.indexOf(":");
    if (idx > 0) {
      headers.push({ key: line.slice(0, idx).trim(), value: line.slice(idx + 1).trim() });
    }
  }

  return {
    statusCode: match ? parseInt(match[1], 10) : 0,
    statusReason: match ? match[2].trim() : "",
    headers,
    body,
  };
}
