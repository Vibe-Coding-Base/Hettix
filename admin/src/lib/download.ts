// downloadText saves text content to a file the viewer can keep. It builds a
// Blob and clicks a temporary anchor, which works in the browser and in the
// desktop WebView2 runtime (unlike server routes opened with window.open, which
// the desktop shell would hand to the system browser).
export function downloadText(filename: string, content: string, mime = "text/plain;charset=utf-8"): void {
  const blob = new Blob([content], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

// csvCell quotes a value for CSV, doubling embedded quotes.
export function csvCell(value: string | number | null | undefined): string {
  const s = String(value ?? "");
  return /[",\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s;
}

// toCSV builds a CSV document from a header row and data rows.
export function toCSV(header: string[], rows: (string | number | null | undefined)[][]): string {
  return [header, ...rows].map((row) => row.map(csvCell).join(",")).join("\r\n");
}
