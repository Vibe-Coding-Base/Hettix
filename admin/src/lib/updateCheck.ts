// Update check against the project's GitHub Releases. It runs client-side (in
// the desktop webview or the browser) so it works the same in both, with no
// backend involvement.

const REPO = "Vibe-Coding-Base/Hettix";
const CURRENT = import.meta.env.VITE_VERSION || "0.0.0";

export interface UpdateResult {
  current: string;
  latest: string;
  updateAvailable: boolean;
  url: string;
}

function parseVersion(v: string): number[] {
  return v
    .replace(/^v/, "")
    .split("-")[0]
    .split(".")
    .map((n) => parseInt(n, 10) || 0);
}

// isNewer reports whether latest is a strictly higher version than current.
function isNewer(latest: string, current: string): boolean {
  const a = parseVersion(latest);
  const b = parseVersion(current);
  for (let i = 0; i < Math.max(a.length, b.length); i++) {
    const x = a[i] ?? 0;
    const y = b[i] ?? 0;
    if (x !== y) {
      return x > y;
    }
  }
  return false;
}

// checkForUpdate queries the latest GitHub release and compares it to the
// running version. It throws on network or API errors.
export async function checkForUpdate(): Promise<UpdateResult> {
  const res = await fetch(`https://api.github.com/repos/${REPO}/releases/latest`, {
    headers: { Accept: "application/vnd.github+json" },
  });
  if (!res.ok) {
    throw new Error(`GitHub API returned ${res.status}`);
  }
  const data = (await res.json()) as { tag_name?: string; html_url?: string };
  const latest = (data.tag_name ?? "").replace(/^v/, "");
  return {
    current: CURRENT,
    latest,
    updateAvailable: latest !== "" && isNewer(latest, CURRENT),
    url: data.html_url ?? `https://github.com/${REPO}/releases/latest`,
  };
}

const LAST_CHECK_KEY = "hettix:lastUpdateCheck";
const DAY_MS = 24 * 60 * 60 * 1000;

// dueForAutoCheck throttles the passive startup check to at most once a day,
// tolerating unavailable or unwritable storage.
export function dueForAutoCheck(): boolean {
  try {
    const last = Number(localStorage.getItem(LAST_CHECK_KEY) ?? 0);
    if (Date.now() - last < DAY_MS) {
      return false;
    }
    localStorage.setItem(LAST_CHECK_KEY, String(Date.now()));
  } catch {
    // No storage: check this session but don't block on remembering it.
  }
  return true;
}
