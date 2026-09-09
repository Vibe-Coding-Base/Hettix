import { Alert, Button, Snackbar } from "@mui/material";
import { useEffect, useState } from "react";

import { openExternal } from "lib/desktop";
import { checkForUpdate, dueForAutoCheck } from "lib/updateCheck";

// UpdateBanner runs a throttled check on startup and, when a newer release is
// available, shows a dismissible banner linking to it. Failures are silent.
export function UpdateBanner(): JSX.Element | null {
  const [update, setUpdate] = useState<{ latest: string; url: string } | null>(null);

  useEffect(() => {
    if (!dueForAutoCheck()) {
      return;
    }
    let active = true;
    checkForUpdate()
      .then((result) => {
        if (active && result.updateAvailable) {
          setUpdate({ latest: result.latest, url: result.url });
        }
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, []);

  if (!update) {
    return null;
  }

  return (
    <Snackbar
      open
      anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
      onClose={(_, reason) => reason !== "clickaway" && setUpdate(null)}
    >
      <Alert
        severity="info"
        onClose={() => setUpdate(null)}
        action={
          <Button color="inherit" size="small" onClick={() => openExternal(update.url)}>
            View
          </Button>
        }
      >
        Hettix {update.latest} is available.
      </Alert>
    </Snackbar>
  );
}
