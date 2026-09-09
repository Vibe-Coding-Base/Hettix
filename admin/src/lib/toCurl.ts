interface CurlRequest {
  method: string;
  url: string;
  headers?: { key: string; value: string }[];
  body?: string | null;
}

function shellQuote(value: string): string {
  return `'${value.replace(/'/g, `'\\''`)}'`;
}

// toCurl renders a request as a copy-pasteable curl command.
export function toCurl({ method, url, headers, body }: CurlRequest): string {
  const parts = [`curl -i -X ${method} ${shellQuote(url)}`];

  for (const { key, value } of headers ?? []) {
    if (key) {
      parts.push(`-H ${shellQuote(`${key}: ${value}`)}`);
    }
  }

  if (body) {
    parts.push(`--data-raw ${shellQuote(body)}`);
  }

  return parts.join(" \\\n  ");
}
