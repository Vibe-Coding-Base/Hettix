import UploadFileIcon from "@mui/icons-material/UploadFile";
import {
  Alert,
  Box,
  Button,
  Chip,
  LinearProgress,
  MenuItem,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { useState } from "react";
import { useSearchParams } from "react-router-dom";

import {
  HttpMethod,
  IntruderAttacksDocument,
  IntruderAttackStatus,
  useIntruderAttacksQuery,
  useIntruderResultsQuery,
  useStartIntruderAttackMutation,
} from "lib/graphql/generated";

export default function Intruder(): JSX.Element {
  const [selected, setSelected] = useState<string | null>(null);

  return (
    <Box sx={{ display: "flex", gap: 3, height: "calc(100vh - 110px)", p: 1 }}>
      <Box sx={{ width: 460, flexShrink: 0, overflow: "auto", pr: 1 }}>
        <AttackForm onStarted={setSelected} />
        <AttackList selected={selected} onSelect={setSelected} />
      </Box>
      <Box sx={{ flex: 1, overflow: "auto" }}>
        {selected ? (
          <Results attackId={selected} />
        ) : (
          <Paper variant="outlined" sx={{ p: 4, textAlign: "center", color: "text.secondary" }}>
            <Typography>Configure an attack on the left, or select a past attack to view its results.</Typography>
          </Paper>
        )}
      </Box>
    </Box>
  );
}

function methodFromParam(value: string | null): HttpMethod {
  const upper = (value ?? "").toUpperCase();
  return (Object.values(HttpMethod) as string[]).includes(upper) ? (upper as HttpMethod) : HttpMethod.Get;
}

function AttackForm({ onStarted }: { onStarted: (id: string) => void }): JSX.Element {
  const [searchParams] = useSearchParams();
  const [name, setName] = useState("");
  // Prefill from "Send to Intruder" (query params); the operator then marks the
  // insertion point with § and adds payloads.
  const [method, setMethod] = useState<HttpMethod>(methodFromParam(searchParams.get("method")));
  const [url, setUrl] = useState(searchParams.get("url") ?? "");
  const [body, setBody] = useState(searchParams.get("body") ?? "");
  const [payloadText, setPayloadText] = useState("");

  const [start, { loading, error }] = useStartIntruderAttackMutation({
    refetchQueries: [{ query: IntruderAttacksDocument }],
    onCompleted({ startIntruderAttack }) {
      onStarted(startIntruderAttack.id);
    },
  });

  const payloads = payloadText
    .split("\n")
    .map((p) => p.trim())
    .filter((p) => p.length > 0);

  const loadFromFile = (file: File) => {
    const reader = new FileReader();
    reader.onload = () => {
      const text = typeof reader.result === "string" ? reader.result : "";
      setPayloadText((prev) => (prev.trim() ? prev.replace(/\s*$/, "") + "\n" : "") + text);
    };
    reader.readAsText(file);
  };

  const canStart = url.includes("§") || body.includes("§");

  const onStart = () => {
    start({
      variables: {
        input: {
          name: name || "Attack",
          method,
          url,
          body: body || null,
          payloads,
        },
      },
    });
  };

  return (
    <Paper variant="outlined" sx={{ p: 2, mb: 2 }}>
      <Typography variant="subtitle2" sx={{ mb: 1 }}>
        New attack
      </Typography>
      <Typography variant="caption" color="text.secondary">
        Mark the insertion point with § in the URL or body. Each payload replaces every marker.
      </Typography>
      <Box sx={{ display: "flex", gap: 1, mt: 1 }}>
        <TextField size="small" label="Name" value={name} onChange={(e) => setName(e.target.value)} sx={{ flex: 1 }} />
        <TextField
          size="small"
          select
          label="Method"
          value={method}
          onChange={(e) => setMethod(e.target.value as HttpMethod)}
          sx={{ width: 110 }}
        >
          {Object.values(HttpMethod).map((m) => (
            <MenuItem key={m} value={m}>
              {m}
            </MenuItem>
          ))}
        </TextField>
      </Box>
      <TextField
        size="small"
        fullWidth
        label="URL"
        placeholder="https://target.com/search?q=§"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        sx={{ mt: 1 }}
        InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
      />
      <TextField
        size="small"
        fullWidth
        multiline
        minRows={2}
        maxRows={6}
        label="Body (optional)"
        value={body}
        onChange={(e) => setBody(e.target.value)}
        sx={{ mt: 1 }}
        InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
      />
      <TextField
        size="small"
        fullWidth
        multiline
        minRows={3}
        maxRows={10}
        label="Payloads (one per line)"
        value={payloadText}
        onChange={(e) => setPayloadText(e.target.value)}
        sx={{ mt: 1 }}
        InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
      />
      <Box sx={{ display: "flex", gap: 1, mt: 1, alignItems: "center" }}>
        <Button variant="contained" disabled={loading || !canStart || payloads.length === 0} onClick={onStart}>
          Start attack ({payloads.length})
        </Button>
        <Button component="label" size="small" startIcon={<UploadFileIcon />}>
          Load wordlist
          <input
            type="file"
            hidden
            accept=".txt,.lst,.list,text/plain"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) {
                loadFromFile(file);
              }
              e.target.value = "";
            }}
          />
        </Button>
        {payloadText && (
          <Button size="small" color="inherit" onClick={() => setPayloadText("")}>
            Clear
          </Button>
        )}
      </Box>
      {!canStart && (url || body) && (
        <Alert severity="info" sx={{ mt: 1 }}>
          Add a § marker to the URL or body to set an insertion point.
        </Alert>
      )}
      {error && (
        <Alert severity="error" sx={{ mt: 1 }}>
          {error.message}
        </Alert>
      )}
    </Paper>
  );
}

function AttackList({ selected, onSelect }: { selected: string | null; onSelect: (id: string) => void }): JSX.Element {
  const { data } = useIntruderAttacksQuery({ pollInterval: 2000 });
  const attacks = data?.intruderAttacks ?? [];

  if (attacks.length === 0) {
    return <></>;
  }

  return (
    <Paper variant="outlined">
      {attacks.map((attack) => (
        <Box
          key={attack.id}
          onClick={() => onSelect(attack.id)}
          sx={{
            p: 1.5,
            cursor: "pointer",
            borderBottom: 1,
            borderColor: "divider",
            bgcolor: selected === attack.id ? "action.selected" : undefined,
          }}
        >
          <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
            <Typography sx={{ flex: 1 }} noWrap>
              {attack.name}
            </Typography>
            <Chip
              size="small"
              color={attack.status === IntruderAttackStatus.Running ? "warning" : "success"}
              label={attack.status === IntruderAttackStatus.Running ? "running" : "done"}
            />
          </Box>
          <Typography variant="caption" color="text.secondary">
            {attack.completed}/{attack.total}
          </Typography>
        </Box>
      ))}
    </Paper>
  );
}

function Results({ attackId }: { attackId: string }): JSX.Element {
  const { data, loading, error } = useIntruderResultsQuery({ variables: { attackId }, pollInterval: 1000 });
  const results = data?.intruderResults ?? [];

  if (error) {
    return <Alert severity="error">{error.message}</Alert>;
  }

  if (loading && results.length === 0) {
    return (
      <Box sx={{ p: 2 }}>
        <LinearProgress />
      </Box>
    );
  }

  return (
    <Table size="small" stickyHeader>
      <TableHead>
        <TableRow>
          <TableCell>#</TableCell>
          <TableCell>Payload</TableCell>
          <TableCell align="right">Status</TableCell>
          <TableCell align="right">Length</TableCell>
          <TableCell align="right">Time (ms)</TableCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {results.map((r) => (
          <TableRow key={r.index} hover>
            <TableCell>{r.index + 1}</TableCell>
            <TableCell sx={{ fontFamily: "monospace", fontSize: 13, wordBreak: "break-all" }}>{r.payload}</TableCell>
            <TableCell align="right">
              {r.error ? (
                <Chip size="small" color="error" label="error" title={r.error} />
              ) : (
                <StatusChip code={r.statusCode} />
              )}
            </TableCell>
            <TableCell align="right">{r.length}</TableCell>
            <TableCell align="right">{r.durationMs}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function StatusChip({ code }: { code: number }): JSX.Element {
  const color = code >= 500 ? "error" : code >= 400 ? "warning" : code >= 300 ? "info" : "success";
  return <Chip size="small" color={color} label={code} />;
}
