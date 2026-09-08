import { createContext, useCallback, useContext, useMemo, useState } from "react";
import type { ReactNode } from "react";

// Appearance holds the operator's editor font preferences, persisted per
// browser in localStorage. It drives the Monaco editors used to view request
// and response bodies.
export interface Appearance {
  fontFamily: string;
  fontSize: number;
}

export const FONT_FAMILIES: { label: string; value: string }[] = [
  { label: "JetBrains Mono", value: "'JetBrains Mono', monospace" },
  { label: "Space Mono", value: "'Space Mono', monospace" },
  { label: "Consolas", value: "Consolas, monospace" },
  { label: "Courier New", value: "'Courier New', monospace" },
  { label: "System monospace", value: "monospace" },
];

const DEFAULT: Appearance = { fontFamily: FONT_FAMILIES[0].value, fontSize: 13 };
const STORAGE_KEY = "hettix.appearance";

function load(): Appearance {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      return { ...DEFAULT, ...JSON.parse(raw) };
    }
  } catch {
    // ignore malformed or unavailable storage
  }
  return DEFAULT;
}

interface AppearanceContextValue extends Appearance {
  setAppearance: (next: Partial<Appearance>) => void;
}

const AppearanceContext = createContext<AppearanceContextValue>({
  ...DEFAULT,
  setAppearance: () => undefined,
});

export function AppearanceProvider({ children }: { children: ReactNode }): JSX.Element {
  const [appearance, setState] = useState<Appearance>(load);

  const setAppearance = useCallback((next: Partial<Appearance>) => {
    setState((prev) => {
      const merged = { ...prev, ...next };
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(merged));
      } catch {
        // ignore unavailable storage
      }
      return merged;
    });
  }, []);

  const value = useMemo(() => ({ ...appearance, setAppearance }), [appearance, setAppearance]);

  return <AppearanceContext.Provider value={value}>{children}</AppearanceContext.Provider>;
}

export function useAppearance(): AppearanceContextValue {
  return useContext(AppearanceContext);
}
