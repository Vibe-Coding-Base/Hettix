import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  MenuItem,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";

import { CODECS } from "lib/codec";

interface Props {
  open: boolean;
  initialText: string;
  onClose: () => void;
  // onApply hands back the chosen codec label and the decoded value when the
  // user confirms.
  onApply: (label: string, value: string) => void;
}

// DecoderDialog lets the operator pick a codec, edit the input, and preview the
// decoded output live before applying it to the selection.
export function DecoderDialog({ open, initialText, onClose, onApply }: Props): JSX.Element {
  const [codecLabel, setCodecLabel] = useState(CODECS[0].label);
  const [input, setInput] = useState(initialText);

  useEffect(() => {
    if (open) {
      setInput(initialText);
    }
  }, [open, initialText]);

  const codec = CODECS.find((c) => c.label === codecLabel) ?? CODECS[0];
  const result = useMemo(() => {
    try {
      return { value: codec.decode(input), error: false };
    } catch (e) {
      return { value: e instanceof Error ? e.message : String(e), error: true };
    }
  }, [codec, input]);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ pb: 1 }}>Decode</DialogTitle>
      <DialogContent>
        <TextField
          select
          label="Type"
          value={codecLabel}
          onChange={(e) => setCodecLabel(e.target.value)}
          size="small"
          sx={{ mb: 2, minWidth: 160 }}
        >
          {CODECS.map((c) => (
            <MenuItem key={c.label} value={c.label}>
              {c.label}
            </MenuItem>
          ))}
        </TextField>
        <TextField
          label="Input"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          fullWidth
          multiline
          minRows={3}
          maxRows={8}
          InputProps={{ sx: { fontFamily: "monospace", fontSize: 13 } }}
          sx={{ mb: 2 }}
        />
        <Box sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", mb: 0.5 }}>
          <Typography variant="caption" color={result.error ? "error" : "text.secondary"}>
            {result.error ? "Cannot decode" : "Output"}
          </Typography>
          {!result.error && (
            <Tooltip title="Copy">
              <IconButton size="small" onClick={() => navigator.clipboard?.writeText(result.value)}>
                <ContentCopyIcon sx={{ fontSize: 14 }} />
              </IconButton>
            </Tooltip>
          )}
        </Box>
        <Box
          sx={{
            fontFamily: "monospace",
            fontSize: 13,
            whiteSpace: "pre-wrap",
            wordBreak: "break-all",
            bgcolor: "action.hover",
            borderRadius: 1,
            p: 1.5,
            minHeight: 72,
            maxHeight: 240,
            overflow: "auto",
            color: result.error ? "error.main" : "text.primary",
          }}
        >
          {result.value}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button
          variant="contained"
          disabled={result.error}
          onClick={() => {
            onApply(codec.label, result.value);
            onClose();
          }}
        >
          OK
        </Button>
      </DialogActions>
    </Dialog>
  );
}
