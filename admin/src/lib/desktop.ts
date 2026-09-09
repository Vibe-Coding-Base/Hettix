import type { CSSProperties } from "react";

// Detection and helpers for running inside the Wails desktop shell. In the
// browser (headless server mode) isDesktop is false and the window helpers are
// no-ops, so the same admin build works in both environments.

interface WailsWindowRuntime {
  WindowMinimise: () => void;
  WindowToggleMaximise: () => void;
  WindowIsMaximised: () => Promise<boolean>;
  WindowReload: () => void;
  BrowserOpenURL: (url: string) => void;
  Quit: () => void;
}

function runtime(): WailsWindowRuntime | undefined {
  if (typeof window === "undefined") {
    return undefined;
  }
  return (window as unknown as { runtime?: WailsWindowRuntime }).runtime;
}

export const isDesktop = runtime() !== undefined;

// The window is frameless (custom title bar and controls) only on Windows; the
// webview's user agent reliably identifies the host OS.
export const isFramelessDesktop = isDesktop && typeof navigator !== "undefined" && /windows/i.test(navigator.userAgent);

// openExternal opens a URL in the user's real browser rather than navigating the
// app's webview to it. In browser mode it falls back to a new tab.
export function openExternal(url: string): void {
  const rt = runtime();
  if (rt) {
    rt.BrowserOpenURL(url);
  } else {
    window.open(url, "_blank", "noopener,noreferrer");
  }
}

export const desktopWindow = {
  minimise(): void {
    runtime()?.WindowMinimise();
  },
  toggleMaximise(): void {
    runtime()?.WindowToggleMaximise();
  },
  async isMaximised(): Promise<boolean> {
    return (await runtime()?.WindowIsMaximised()) ?? false;
  },
  reload(): void {
    runtime()?.WindowReload();
  },
  quit(): void {
    runtime()?.Quit();
  },
};

// Wails treats an element as a window-drag handle when its computed
// --wails-draggable is "drag"; children opt back out with "no-drag". The value
// inherits, so a draggable bar only needs "no-drag" on its interactive parts.
export const dragStyle = { "--wails-draggable": "drag" } as unknown as CSSProperties;
export const noDragStyle = { "--wails-draggable": "no-drag" } as unknown as CSSProperties;
