import DeleteIcon from "@mui/icons-material/Delete";
import DownloadIcon from "@mui/icons-material/Download";
import SearchIcon from "@mui/icons-material/Search";
import {
  Alert,
  Box,
  Button,
  Chip,
  IconButton,
  InputAdornment,
  MenuItem,
  Paper,
  TextField,
  Typography,
} from "@mui/material";
import { useMemo, useState } from "react";

import { downloadText, toCSV } from "lib/download";
import {
  FindingSeverity,
  FindingsDocument,
  useCreateFindingMutation,
  useDeleteFindingMutation,
  useFindingsQuery,
} from "lib/graphql/generated";

const SEVERITY_COLOR: Record<FindingSeverity, "default" | "info" | "success" | "warning" | "error"> = {
  [FindingSeverity.Info]: "default",
  [FindingSeverity.Low]: "info",
  [FindingSeverity.Medium]: "warning",
  [FindingSeverity.High]: "error",
  [FindingSeverity.Critical]: "error",
};

// Higher rank sorts first when ordering by severity.
const SEVERITY_RANK: Record<FindingSeverity, number> = {
  [FindingSeverity.Critical]: 4,
  [FindingSeverity.High]: 3,
  [FindingSeverity.Medium]: 2,
  [FindingSeverity.Low]: 1,
  [FindingSeverity.Info]: 0,
};

type SortBy = "severity" | "newest";

export default function Findings(): JSX.Element {
  return (
    <Box sx={{ maxWidth: 820 }}>
      <CreateFindingForm />
      <FindingList />
    </Box>
  );
}

function CreateFindingForm(): JSX.Element {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [severity, setSeverity] = useState<FindingSeverity>(FindingSeverity.Medium);
  const [requestLogID, setRequestLogID] = useState("");

  const [create, { loading, error }] = useCreateFindingMutation({
    refetchQueries: [{ query: FindingsDocument }],
    onCompleted() {
      setTitle("");
      setDescription("");
      setRequestLogID("");
    },
  });

  const onCreate = () => {
    create({
      variables: {
        input: {
          title,
          description: description || null,
          severity,
          requestLogID: requestLogID || null,
        },
      },
    });
  };

  return (
    <Paper variant="outlined" sx={{ p: 2, mb: 2 }}>
      <Typography variant="subtitle2" sx={{ mb: 1 }}>
        New finding
      </Typography>
      <Box sx={{ display: "flex", gap: 1 }}>
        <TextField
          size="small"
          label="Title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          sx={{ flex: 1 }}
        />
        <TextField
          size="small"
          select
          label="Severity"
          value={severity}
          onChange={(e) => setSeverity(e.target.value as FindingSeverity)}
          sx={{ width: 150 }}
        >
          {Object.values(FindingSeverity).map((s) => (
            <MenuItem key={s} value={s}>
              {s}
            </MenuItem>
          ))}
        </TextField>
      </Box>
      <TextField
        size="small"
        fullWidth
        multiline
        minRows={2}
        maxRows={8}
        label="Description"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
        sx={{ mt: 1 }}
      />
      <TextField
        size="small"
        fullWidth
        label="Request log id (optional)"
        value={requestLogID}
        onChange={(e) => setRequestLogID(e.target.value)}
        sx={{ mt: 1 }}
        InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
      />
      <Button variant="contained" sx={{ mt: 1 }} disabled={loading || !title} onClick={onCreate}>
        Add finding
      </Button>
      {error && (
        <Alert severity="error" sx={{ mt: 1 }}>
          {error.message}
        </Alert>
      )}
    </Paper>
  );
}

function FindingList(): JSX.Element {
  const { data, error } = useFindingsQuery({ pollInterval: 3000 });
  const [remove] = useDeleteFindingMutation({ refetchQueries: [{ query: FindingsDocument }] });

  const [search, setSearch] = useState("");
  const [sortBy, setSortBy] = useState<SortBy>("severity");

  const findings = useMemo(() => data?.findings ?? [], [data?.findings]);

  const visible = useMemo(() => {
    const query = search.trim().toLowerCase();
    const filtered = query
      ? findings.filter((f) =>
          [f.title, f.description ?? "", f.severity, f.requestLogID ?? ""].some((field) =>
            field.toLowerCase().includes(query)
          )
        )
      : findings;

    return [...filtered].sort((a, b) => {
      if (sortBy === "severity") {
        const diff = SEVERITY_RANK[b.severity] - SEVERITY_RANK[a.severity];
        if (diff !== 0) {
          return diff;
        }
      }
      // Newest first (ULIDs are time-ordered), and as the severity tiebreaker.
      return a.id < b.id ? 1 : a.id > b.id ? -1 : 0;
    });
  }, [findings, search, sortBy]);

  if (error) {
    return <Alert severity="error">{error.message}</Alert>;
  }

  if (findings.length === 0) {
    return (
      <Typography color="text.secondary">
        No findings yet. Record one above or let the assistant create them.
      </Typography>
    );
  }

  const exportCSV = () => {
    const rows = visible.map((f) => [f.timestamp, f.severity, f.title, f.description ?? "", f.requestLogID ?? ""]);
    downloadText(
      "findings.csv",
      toCSV(["Timestamp", "Severity", "Title", "Description", "Request log id"], rows),
      "text/csv"
    );
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
      <Box sx={{ display: "flex", gap: 1, alignItems: "center", flexWrap: "wrap" }}>
        <TextField
          size="small"
          placeholder="Search findings"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          sx={{ flex: 1, minWidth: 200 }}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon fontSize="small" />
              </InputAdornment>
            ),
          }}
        />
        <TextField
          size="small"
          select
          label="Sort by"
          value={sortBy}
          onChange={(e) => setSortBy(e.target.value as SortBy)}
          sx={{ width: 150 }}
        >
          <MenuItem value="severity">Severity</MenuItem>
          <MenuItem value="newest">Newest</MenuItem>
        </TextField>
        <Button size="small" startIcon={<DownloadIcon />} onClick={exportCSV}>
          Export CSV
        </Button>
      </Box>

      {visible.length === 0 ? (
        <Typography color="text.secondary" sx={{ py: 2 }}>
          No findings match &quot;{search}&quot;.
        </Typography>
      ) : (
        visible.map((f) => (
          <Paper key={f.id} variant="outlined" sx={{ p: 2 }}>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Chip size="small" color={SEVERITY_COLOR[f.severity]} label={f.severity.toLowerCase()} />
              <Typography sx={{ flex: 1, fontWeight: 500 }}>{f.title}</Typography>
              <IconButton size="small" aria-label="Delete finding" onClick={() => remove({ variables: { id: f.id } })}>
                <DeleteIcon fontSize="small" />
              </IconButton>
            </Box>
            {f.description && (
              <Box
                sx={{
                  mt: 1,
                  p: 1.5,
                  borderRadius: 1,
                  bgcolor: "action.hover",
                  fontFamily: "monospace",
                  fontSize: 13,
                  lineHeight: 1.5,
                  whiteSpace: "pre-wrap",
                  wordBreak: "break-word",
                }}
              >
                {f.description}
              </Box>
            )}
            {f.requestLogID && (
              <Typography
                variant="caption"
                color="text.secondary"
                sx={{ mt: 0.5, display: "block", fontFamily: "monospace" }}
              >
                request: {f.requestLogID}
              </Typography>
            )}
          </Paper>
        ))
      )}
    </Box>
  );
}
