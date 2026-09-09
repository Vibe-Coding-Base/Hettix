import SwapVertIcon from "@mui/icons-material/SwapVert";
import { Alert, Box, Button, IconButton, Paper, TextField, Tooltip, Typography } from "@mui/material";
import { useState } from "react";
import { useSearchParams } from "react-router-dom";

import { CODECS } from "lib/codec";

export default function Decoder(): JSX.Element {
  const [searchParams] = useSearchParams();
  const [input, setInput] = useState(() => searchParams.get("input") ?? "");
  const [output, setOutput] = useState("");
  const [error, setError] = useState<string | null>(null);

  const apply = (fn: (s: string) => string) => {
    try {
      setOutput(fn(input));
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      setOutput("");
    }
  };

  const monospace = { sx: { fontFamily: "monospace", fontSize: 13 } };

  return (
    <Box sx={{ maxWidth: 900 }}>
      <Typography color="text.secondary" sx={{ mb: 2 }}>
        Transform data between encodings. Apply an operation to the input, then swap the result up to chain transforms.
      </Typography>

      <Paper variant="outlined" sx={{ p: 2, mb: 1 }}>
        <TextField
          label="Input"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          fullWidth
          multiline
          minRows={4}
          maxRows={12}
          InputProps={monospace}
        />
      </Paper>

      <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1, mb: 1, alignItems: "center" }}>
        {CODECS.map((op) => (
          <Box key={op.label} sx={{ display: "flex", gap: 0.5 }}>
            <Button size="small" variant="outlined" onClick={() => apply(op.encode)}>
              {op.label} encode
            </Button>
            <Button size="small" variant="outlined" onClick={() => apply(op.decode)}>
              {op.label} decode
            </Button>
          </Box>
        ))}
        <Tooltip title="Move output to input">
          <IconButton
            onClick={() => {
              setInput(output);
              setOutput("");
            }}
          >
            <SwapVertIcon />
          </IconButton>
        </Tooltip>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 1 }}>
          {error}
        </Alert>
      )}

      <Paper variant="outlined" sx={{ p: 2 }}>
        <TextField
          label="Output"
          value={output}
          fullWidth
          multiline
          minRows={4}
          maxRows={12}
          InputProps={{ ...monospace, readOnly: true }}
        />
      </Paper>
    </Box>
  );
}
