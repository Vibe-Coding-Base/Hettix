import { Alert, Box, Button, Chip, FormControlLabel, Paper, Switch, TextField, Typography } from "@mui/material";
import { useEffect, useState } from "react";

import {
  InterceptedWebSocketMessagesDocument,
  WebSocketDirection,
  useDropWebSocketMessageMutation,
  useForwardWebSocketMessageMutation,
  useInterceptedWebSocketMessagesQuery,
  useModifyWebSocketMessageMutation,
  useUpdateWebSocketInterceptSettingsMutation,
  useWebSocketInterceptSettingsQuery,
} from "lib/graphql/generated";

export default function WebSocketIntercept(): JSX.Element {
  return (
    <Box sx={{ maxWidth: 820 }}>
      <InterceptSettings />
      <PendingFrames />
    </Box>
  );
}

function InterceptSettings(): JSX.Element {
  const { data } = useWebSocketInterceptSettingsQuery();
  const [enabled, setEnabled] = useState(false);
  const [filter, setFilter] = useState("");
  const [update, { error }] = useUpdateWebSocketInterceptSettingsMutation();

  useEffect(() => {
    if (data?.webSocketInterceptSettings) {
      setEnabled(data.webSocketInterceptSettings.enabled);
      setFilter(data.webSocketInterceptSettings.filter ?? "");
    }
  }, [data]);

  const save = (nextEnabled: boolean, nextFilter: string) => {
    update({ variables: { input: { enabled: nextEnabled, filter: nextFilter || null } } });
  };

  return (
    <Paper variant="outlined" sx={{ p: 2, mb: 2 }}>
      <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
        <FormControlLabel
          control={
            <Switch
              checked={enabled}
              onChange={(e) => {
                setEnabled(e.target.checked);
                save(e.target.checked, filter);
              }}
            />
          }
          label="Intercept WebSocket messages"
        />
        <TextField
          size="small"
          fullWidth
          label="Filter (HTTPQL, optional)"
          placeholder='e.g. ws.payload cont "token"'
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              save(enabled, filter.trim());
            }
          }}
          InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
        />
      </Box>
      {error && (
        <Alert severity="error" sx={{ mt: 1 }}>
          {error.message}
        </Alert>
      )}
    </Paper>
  );
}

function PendingFrames(): JSX.Element {
  const { data, error } = useInterceptedWebSocketMessagesQuery({ pollInterval: 1000 });
  const frames = data?.interceptedWebSocketMessages ?? [];

  if (error) {
    return <Alert severity="error">{error.message}</Alert>;
  }

  if (frames.length === 0) {
    return (
      <Typography color="text.secondary">
        No messages are being held. Enable interception to hold matching frames.
      </Typography>
    );
  }

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
      {frames.map((frame) => (
        <FrameCard key={frame.id} frame={frame} />
      ))}
    </Box>
  );
}

type Frame = {
  id: string;
  direction: WebSocketDirection;
  opcode: number;
  payload: string;
};

function FrameCard({ frame }: { frame: Frame }): JSX.Element {
  const [payload, setPayload] = useState(frame.payload);
  const refetch = [{ query: InterceptedWebSocketMessagesDocument }];
  const [modify, { loading: modifying }] = useModifyWebSocketMessageMutation({ refetchQueries: refetch });
  const [forward, { loading: forwarding }] = useForwardWebSocketMessageMutation({ refetchQueries: refetch });
  const [drop, { loading: dropping }] = useDropWebSocketMessageMutation({ refetchQueries: refetch });

  const busy = modifying || forwarding || dropping;
  const outbound = frame.direction === WebSocketDirection.ClientToServer;
  const edited = payload !== frame.payload;

  return (
    <Paper variant="outlined" sx={{ p: 2 }}>
      <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 1 }}>
        <Chip
          size="small"
          color={outbound ? "primary" : "default"}
          label={outbound ? "Client → Server" : "Server → Client"}
        />
        <Typography variant="caption" color="text.secondary">
          opcode {frame.opcode}
        </Typography>
      </Box>
      <TextField
        multiline
        fullWidth
        minRows={2}
        maxRows={12}
        value={payload}
        onChange={(e) => setPayload(e.target.value)}
        InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
      />
      <Box sx={{ display: "flex", gap: 1, mt: 1 }}>
        <Button
          variant="contained"
          disabled={busy}
          onClick={() =>
            edited
              ? modify({ variables: { input: { id: frame.id, payload } } })
              : forward({ variables: { id: frame.id } })
          }
        >
          {edited ? "Forward edited" : "Forward"}
        </Button>
        <Button color="error" disabled={busy} onClick={() => drop({ variables: { id: frame.id } })}>
          Drop
        </Button>
      </Box>
    </Paper>
  );
}
