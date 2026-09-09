import type { CSSProperties } from "react";

// Detection and helpers for running inside the Wails desktop shell. In the
// browser (headless server mode) isDesktop is false and the window helpers are
// no-ops, so the same admin build works in both environments.

interface WailsWindowRuntime {
  WindowMinimise: () => void;
  WindowToggleMaximise: () => void;
  WindowIsMaximised: () => Promise<boolean>;
  WindowReload: () => void;
  Quit: () => void;
}

function runtime(): WailsWindowRuntime | undefined {
  if (typeof window === "undefined") {
    return undefined;
  }
  return (window as unknown as { runtime?: WailsWindowRuntime }).runtime;
}

export const isDesktop = runtime() !== undefined;

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
