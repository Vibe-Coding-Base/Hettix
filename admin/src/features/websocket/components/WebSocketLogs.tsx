import CallMadeIcon from "@mui/icons-material/CallMade";
import CallReceivedIcon from "@mui/icons-material/CallReceived";
import {
  Alert,
  Box,
  Chip,
  CircularProgress,
  List,
  ListItemButton,
  ListItemText,
  Paper,
  Typography,
} from "@mui/material";
import { useState } from "react";

import { WebSocketDirection, useWebSocketConnectionsQuery, useWebSocketMessagesQuery } from "lib/graphql/generated";

export default function WebSocketLogs(): JSX.Element {
  const { data, loading, error } = useWebSocketConnectionsQuery({ pollInterval: 2000 });
  const [selected, setSelected] = useState<string | null>(null);

  const connections = data?.webSocketConnections ?? [];

  if (error) {
    return <Alert severity="error">{error.message}</Alert>;
  }

  return (
    <Box sx={{ display: "flex", gap: 2, height: "calc(100vh - 120px)" }}>
      <Paper variant="outlined" sx={{ width: 360, flexShrink: 0, overflow: "auto" }}>
        {loading && connections.length === 0 ? (
          <Box sx={{ p: 2, display: "flex", justifyContent: "center" }}>
            <CircularProgress size={24} />
          </Box>
        ) : connections.length === 0 ? (
          <Typography color="text.secondary" sx={{ p: 2 }}>
            No WebSocket connections captured yet.
          </Typography>
        ) : (
          <List disablePadding>
            {connections.map((conn) => (
              <ListItemButton
                key={conn.id}
                selected={selected === conn.id}
                onClick={() => setSelected(conn.id)}
                divider
              >
                <ListItemText
                  primary={conn.host + conn.path}
                  secondary={
                    <>
                      {conn.messageCount} message{conn.messageCount === 1 ? "" : "s"}
                      {" · "}
                      {conn.closedAt ? "closed" : "open"}
                    </>
                  }
                  primaryTypographyProps={{ noWrap: true, title: conn.url }}
                />
              </ListItemButton>
            ))}
          </List>
        )}
      </Paper>

      <Box sx={{ flex: 1, overflow: "auto" }}>
        {selected ? (
          <MessageList connectionId={selected} />
        ) : (
          <Typography color="text.secondary" sx={{ p: 2 }}>
            Select a connection to inspect its messages.
          </Typography>
        )}
      </Box>
    </Box>
  );
}

function MessageList({ connectionId }: { connectionId: string }): JSX.Element {
  const { data, loading, error } = useWebSocketMessagesQuery({
    variables: { connectionId },
    pollInterval: 2000,
  });

  if (error) {
    return <Alert severity="error">{error.message}</Alert>;
  }

  const messages = data?.webSocketMessages ?? [];

  if (loading && messages.length === 0) {
    return (
      <Box sx={{ p: 2, display: "flex", justifyContent: "center" }}>
        <CircularProgress size={24} />
      </Box>
    );
  }

  if (messages.length === 0) {
    return (
      <Typography color="text.secondary" sx={{ p: 2 }}>
        No messages on this connection.
      </Typography>
    );
  }

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
      {messages.map((msg, i) => {
        const outbound = msg.direction === WebSocketDirection.ClientToServer;
        return (
          <Paper key={i} variant="outlined" sx={{ p: 1.5 }}>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 0.5 }}>
              <Chip
                size="small"
                color={outbound ? "primary" : "default"}
                icon={outbound ? <CallMadeIcon /> : <CallReceivedIcon />}
                label={outbound ? "Client → Server" : "Server → Client"}
              />
              <Typography variant="caption" color="text.secondary">
                opcode {msg.opcode} · {new Date(msg.timestamp).toLocaleTimeString()}
              </Typography>
            </Box>
            <Typography
              component="pre"
              variant="body2"
              sx={{ m: 0, whiteSpace: "pre-wrap", wordBreak: "break-all", fontFamily: "monospace" }}
            >
              {msg.payload}
            </Typography>
          </Paper>
        );
      })}
    </Box>
  );
}
