import { Alert, Box, Button, Paper, Snackbar, TextField, Typography } from "@mui/material";
import { useEffect, useState } from "react";

import { useProxySettingsQuery, useSetProxyPortMutation } from "lib/graphql/generated";

export default function ProxySettings(): JSX.Element {
  const { data } = useProxySettingsQuery();
  const [setProxyPort, { loading: saving, error }] = useSetProxyPortMutation();
  const [port, setPort] = useState("");
  const [saved, setSaved] = useState(false);

  const currentPort = data?.proxySettings.port ?? 0;

  useEffect(() => {
    if (currentPort > 0) {
      setPort(String(currentPort));
    }
  }, [currentPort]);

  const handleApply = (e: React.FormEvent) => {
    e.preventDefault();
    const value = parseInt(port, 10);
    if (!Number.isInteger(value) || value < 1 || value > 65535) {
      return;
    }
    setProxyPort({ variables: { port: value } })
      .then(() => setSaved(true))
      .catch(() => undefined);
  };

  const invalid = (() => {
    const v = parseInt(port, 10);
    return !Number.isInteger(v) || v < 1 || v > 65535;
  })();

  return (
    <Box sx={{ maxWidth: 640 }}>
      <Typography color="text.secondary" sx={{ mb: 2 }}>
        The MITM proxy listens on this port. Point your browser&apos;s HTTP proxy here, or use the launched pentest
        browser, which is wired to it automatically. Changing the port applies immediately.
      </Typography>

      <Paper variant="outlined" sx={{ p: 2 }}>
        <Box component="form" onSubmit={handleApply} sx={{ display: "flex", alignItems: "flex-start", gap: 2 }}>
          <TextField
            label="Proxy port"
            value={port}
            onChange={(e) => setPort(e.target.value.replace(/[^0-9]/g, ""))}
            size="small"
            error={port !== "" && invalid}
            helperText={port !== "" && invalid ? "Enter a port between 1 and 65535" : `Current: ${currentPort}`}
            sx={{ width: 200 }}
          />
          <Button
            type="submit"
            variant="contained"
            disabled={saving || invalid || parseInt(port, 10) === currentPort}
            sx={{ mt: 0.5 }}
          >
            Apply
          </Button>
        </Box>
      </Paper>

      {error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {error.message}
        </Alert>
      )}

      <Snackbar
        open={saved}
        autoHideDuration={4000}
        onClose={() => setSaved(false)}
        anchorOrigin={{ horizontal: "center", vertical: "bottom" }}
      >
        <Alert severity="success" onClose={() => setSaved(false)}>
          Proxy now listening on port {currentPort}.
        </Alert>
      </Snackbar>
    </Box>
  );
}
