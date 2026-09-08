import DeleteIcon from "@mui/icons-material/Delete";
import { Alert, Box, Button, Chip, IconButton, MenuItem, Paper, TextField, Typography } from "@mui/material";
import { useState } from "react";

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

  if (error) {
    return <Alert severity="error">{error.message}</Alert>;
  }

  const findings = data?.findings ?? [];

  if (findings.length === 0) {
    return (
      <Typography color="text.secondary">
        No findings yet. Record one above or let the assistant create them.
      </Typography>
    );
  }

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
      {findings.map((f) => (
        <Paper key={f.id} variant="outlined" sx={{ p: 2 }}>
          <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
            <Chip size="small" color={SEVERITY_COLOR[f.severity]} label={f.severity.toLowerCase()} />
            <Typography sx={{ flex: 1, fontWeight: 500 }}>{f.title}</Typography>
            <IconButton size="small" aria-label="Delete finding" onClick={() => remove({ variables: { id: f.id } })}>
              <DeleteIcon fontSize="small" />
            </IconButton>
          </Box>
          {f.description && (
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, whiteSpace: "pre-wrap" }}>
              {f.description}
            </Typography>
          )}
          {f.requestLogID && (
            <Typography variant="caption" color="text.secondary" sx={{ fontFamily: "monospace" }}>
              request: {f.requestLogID}
            </Typography>
          )}
        </Paper>
      ))}
    </Box>
  );
}
