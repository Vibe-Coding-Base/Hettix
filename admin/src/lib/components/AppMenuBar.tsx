import { Box, Button, Dialog, DialogContent, Divider, Link, Menu, MenuItem, Typography } from "@mui/material";
import { useState } from "react";
import type { MouseEvent } from "react";

import { desktopWindow, noDragStyle, openExternal } from "lib/desktop";
import { useLaunchBrowserMutation } from "lib/graphql/generated";
import { checkForUpdate } from "lib/updateCheck";

const VERSION = import.meta.env.VITE_VERSION || "0.0";

interface MenuAction {
  label?: string;
  divider?: boolean;
  onClick?: () => void;
}

interface MenuGroup {
  label: string;
  items: MenuAction[];
}

// AppMenuBar renders the desktop window's File/View/Help menu strip. It only
// appears in the desktop shell; window actions use the Wails runtime and the
// browser launch reuses the same GraphQL mutation as the Home screen.
type UpdateState =
  | { status: "checking" }
  | { status: "result"; current: string; latest: string; updateAvailable: boolean; url: string }
  | { status: "error"; message: string };

export function AppMenuBar(): JSX.Element {
  const [anchor, setAnchor] = useState<{ el: HTMLElement; group: number } | null>(null);
  const [aboutOpen, setAboutOpen] = useState(false);
  const [update, setUpdate] = useState<UpdateState | null>(null);
  const [launchBrowser] = useLaunchBrowserMutation();

  const runUpdateCheck = () => {
    setUpdate({ status: "checking" });
    checkForUpdate()
      .then((r) => setUpdate({ status: "result", ...r }))
      .catch((e) => setUpdate({ status: "error", message: e instanceof Error ? e.message : String(e) }));
  };

  const groups: MenuGroup[] = [
    {
      label: "File",
      items: [
        {
          label: "Launch Pentest Browser",
          onClick: () => {
            launchBrowser().catch(() => undefined);
          },
        },
        { divider: true },
        { label: "Exit", onClick: () => desktopWindow.quit() },
      ],
    },
    {
      label: "View",
      items: [{ label: "Reload", onClick: () => desktopWindow.reload() }],
    },
    {
      label: "Help",
      items: [
        { label: "Check for updates…", onClick: runUpdateCheck },
        { divider: true },
        { label: "About Hettix", onClick: () => setAboutOpen(true) },
      ],
    },
  ];

  const openGroup = (event: MouseEvent<HTMLElement>, group: number) => {
    setAnchor({ el: event.currentTarget, group });
  };
  const close = () => setAnchor(null);

  return (
    <Box sx={{ display: "flex", alignItems: "center", ml: 1 }} style={noDragStyle}>
      {groups.map((group, index) => (
        <Button
          key={group.label}
          size="small"
          color="inherit"
          onClick={(event) => openGroup(event, index)}
          sx={{ minWidth: 0, px: 1.25, textTransform: "none", fontWeight: 400, opacity: 0.9 }}
        >
          {group.label}
        </Button>
      ))}

      <Menu
        anchorEl={anchor?.el}
        open={anchor !== null}
        onClose={close}
        anchorOrigin={{ vertical: "bottom", horizontal: "left" }}
        transformOrigin={{ vertical: "top", horizontal: "left" }}
      >
        {anchor !== null &&
          groups[anchor.group].items.flatMap((item, index) =>
            item.divider
              ? [<Divider key={`divider-${index}`} />]
              : [
                  <MenuItem
                    key={index}
                    dense
                    onClick={() => {
                      item.onClick?.();
                      close();
                    }}
                  >
                    {item.label}
                  </MenuItem>,
                ]
          )}
      </Menu>

      <AboutDialog open={aboutOpen} onClose={() => setAboutOpen(false)} />
      <UpdateDialog state={update} onClose={() => setUpdate(null)} />
    </Box>
  );
}

function UpdateDialog({ state, onClose }: { state: UpdateState | null; onClose: () => void }): JSX.Element {
  return (
    <Dialog open={state !== null} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogContent sx={{ py: 3 }}>
        {state?.status === "checking" && <Typography>Checking for updates…</Typography>}
        {state?.status === "error" && (
          <Typography color="error">Could not check for updates: {state.message}</Typography>
        )}
        {state?.status === "result" &&
          (state.updateAvailable ? (
            <Typography>
              Hettix {state.latest} is available (you have {state.current}).{" "}
              <Link component="button" type="button" onClick={() => openExternal(state.url)}>
                Open the release page
              </Link>
              .
            </Typography>
          ) : (
            <Typography>You are on the latest version ({state.current}).</Typography>
          ))}
      </DialogContent>
    </Dialog>
  );
}

function AboutDialog({ open, onClose }: { open: boolean; onClose: () => void }): JSX.Element {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogContent sx={{ textAlign: "center", py: 4 }}>
        <Typography variant="h4" sx={{ fontFamily: "'JetBrains Mono', monospace", color: "primary.main" }}>
          Hettix
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
          Version {VERSION}
        </Typography>
        <Typography variant="body2" sx={{ mt: 2 }}>
          An HTTP toolkit for security research — proxy, intercept, fuzzing and an agentic AI assistant.
        </Typography>
        <Divider sx={{ my: 2 }} />
        <Typography variant="caption" color="text.secondary">
          © {new Date().getFullYear()} Hettix
        </Typography>
      </DialogContent>
    </Dialog>
  );
}
