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
  TextField,
  Typography,
} from "@mui/material";
import { useState } from "react";

import { WebSocketDirection, useWebSocketConnectionsQuery, useWebSocketMessagesQuery } from "lib/graphql/generated";

export default function WebSocketLogs(): JSX.Element {
  const [search, setSearch] = useState("");
  const { data, loading, error } = useWebSocketConnectionsQuery({
    variables: { searchExpression: search || undefined },
    pollInterval: 2000,
  });
  const [selected, setSelected] = useState<string | null>(null);

  const connections = data?.webSocketConnections ?? [];

  return (
    <Box sx={{ display: "flex", gap: 2, height: "calc(100vh - 120px)" }}>
      <Paper variant="outlined" sx={{ width: 380, flexShrink: 0, display: "flex", flexDirection: "column" }}>
        <Box sx={{ p: 1 }}>
          <SearchField placeholder='Filter, e.g. ws.host cont "chat" or ws.payload cont "token"' onSearch={setSearch} />
        </Box>
        <Box sx={{ overflow: "auto", flex: 1 }}>
          {error ? (
            <Alert severity="error" sx={{ m: 1 }}>
              {error.message}
            </Alert>
          ) : loading && connections.length === 0 ? (
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
        </Box>
      </Paper>

      <Box sx={{ flex: 1, overflow: "hidden" }}>
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
  const [search, setSearch] = useState("");
  const { data, loading, error } = useWebSocketMessagesQuery({
    variables: { connectionId, searchExpression: search || undefined },
    pollInterval: 2000,
  });

  const messages = data?.webSocketMessages ?? [];

  return (
    <Box sx={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <Box sx={{ mb: 1 }}>
        <SearchField placeholder='Filter frames, e.g. ws.payload cont "error"' onSearch={setSearch} />
      </Box>
      <Box sx={{ overflow: "auto", flex: 1, display: "flex", flexDirection: "column", gap: 1 }}>
        {error ? (
          <Alert severity="error">{error.message}</Alert>
        ) : loading && messages.length === 0 ? (
          <Box sx={{ p: 2, display: "flex", justifyContent: "center" }}>
            <CircularProgress size={24} />
          </Box>
        ) : messages.length === 0 ? (
          <Typography color="text.secondary" sx={{ p: 2 }}>
            No messages on this connection.
          </Typography>
        ) : (
          messages.map((msg, i) => {
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
          })
        )}
      </Box>
    </Box>
  );
}

function SearchField({
  placeholder,
  onSearch,
}: {
  placeholder: string;
  onSearch: (value: string) => void;
}): JSX.Element {
  const [value, setValue] = useState("");

  return (
    <TextField
      size="small"
      fullWidth
      placeholder={placeholder}
      value={value}
      onChange={(e) => setValue(e.target.value)}
      onKeyDown={(e) => {
        if (e.key === "Enter") {
          onSearch(value.trim());
        }
      }}
      InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
    />
  );
}
