// Client-side quick filtering for the proxy log, driven by the filter popup.
// This mirrors Burp's proxy-history filter (request type, status class, MIME /
// file type, file-extension show/hide lists). The free-text bar above it is
// HTTPQL, which searches request/response bodies and headers server-side and is
// strictly more powerful than Burp's search term.

export interface LogFilters {
  hideNoResponse: boolean; // hide requests still awaiting a response
  onlyParameterized: boolean; // only requests carrying a query string
  statuses: string[]; // "2xx".."5xx"; empty = all
  types: string[]; // category ids; empty = all
  showExtensions: string; // whitelist of file extensions; empty = all
  hideExtensions: string; // blacklist of file extensions
}

export const emptyLogFilters: LogFilters = {
  hideNoResponse: false,
  onlyParameterized: false,
  statuses: [],
  types: [],
  showExtensions: "",
  hideExtensions: "",
};

export const STATUS_CLASSES = ["2xx", "3xx", "4xx", "5xx"] as const;

export const TYPE_CATEGORIES: { id: string; label: string }[] = [
  { id: "html", label: "HTML / page" },
  { id: "script", label: "Script" },
  { id: "css", label: "CSS" },
  { id: "image", label: "Images" },
  { id: "font", label: "Fonts" },
  { id: "data", label: "JSON / XML" },
  { id: "other", label: "Other" },
];

type LogEntry = { url: string; response?: { statusCode: number } | null };

function statusClass(code: number): string {
  return `${Math.floor(code / 100)}xx`;
}

function extension(url: string): string {
  const path = url.split("?")[0];
  return /\.([a-z0-9]+)$/i.exec(path)?.[1]?.toLowerCase() ?? "";
}

// parseExtensions splits a comma/space separated list into normalized, dot-less,
// lower-case extensions.
function parseExtensions(list: string): string[] {
  return list
    .split(/[\s,]+/)
    .map((e) => e.replace(/^\./, "").toLowerCase())
    .filter(Boolean);
}

// classify buckets a URL by its file extension; paths without a known extension
// are treated as HTML/pages (typical application endpoints).
export function classifyType(url: string): string {
  const ext = extension(url);
  if (!ext) {
    return "html";
  }
  if (["html", "htm", "xhtml"].includes(ext)) return "html";
  if (["js", "mjs", "jsx", "ts"].includes(ext)) return "script";
  if (ext === "css") return "css";
  if (["png", "jpg", "jpeg", "gif", "svg", "ico", "webp", "avif", "bmp"].includes(ext)) return "image";
  if (["woff", "woff2", "ttf", "eot", "otf"].includes(ext)) return "font";
  if (["json", "xml"].includes(ext)) return "data";
  return "other";
}

export function hasActiveLogFilters(filters: LogFilters): boolean {
  return (
    filters.hideNoResponse ||
    filters.onlyParameterized ||
    filters.statuses.length > 0 ||
    filters.types.length > 0 ||
    filters.showExtensions.trim() !== "" ||
    filters.hideExtensions.trim() !== ""
  );
}

export function applyLogFilters<T extends LogEntry>(logs: readonly T[], filters: LogFilters): T[] {
  const show = parseExtensions(filters.showExtensions);
  const hide = parseExtensions(filters.hideExtensions);

  return logs.filter((log) => {
    if (filters.hideNoResponse && !log.response) {
      return false;
    }
    if (filters.onlyParameterized && !log.url.includes("?")) {
      return false;
    }
    if (filters.statuses.length > 0) {
      const cls = log.response ? statusClass(log.response.statusCode) : "none";
      if (!filters.statuses.includes(cls)) {
        return false;
      }
    }
    if (filters.types.length > 0 && !filters.types.includes(classifyType(log.url))) {
      return false;
    }
    if (show.length > 0 || hide.length > 0) {
      const ext = extension(log.url);
      if (show.length > 0 && !show.includes(ext)) {
        return false;
      }
      if (hide.length > 0 && hide.includes(ext)) {
        return false;
      }
    }
    return true;
  });
}
