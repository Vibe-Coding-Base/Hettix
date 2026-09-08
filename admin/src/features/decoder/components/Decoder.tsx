import SwapVertIcon from "@mui/icons-material/SwapVert";
import { Alert, Box, Button, IconButton, Paper, TextField, Tooltip, Typography } from "@mui/material";
import { useState } from "react";

// Encoders/decoders. Each throws on invalid input so the UI can surface it.
const utf8 = new TextEncoder();
const utf8dec = new TextDecoder();

function base64Encode(s: string): string {
  const bytes = utf8.encode(s);
  let bin = "";
  bytes.forEach((b) => (bin += String.fromCharCode(b)));
  return btoa(bin);
}

function base64Decode(s: string): string {
  const bin = atob(s.trim());
  const bytes = Uint8Array.from(bin, (c) => c.charCodeAt(0));
  return utf8dec.decode(bytes);
}

function hexEncode(s: string): string {
  return Array.from(utf8.encode(s))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

function hexDecode(s: string): string {
  const clean = s.replace(/\s+/g, "");
  if (clean.length % 2 !== 0) {
    throw new Error("hex input must have an even number of digits");
  }
  const bytes = new Uint8Array(clean.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(clean.substr(i * 2, 2), 16);
  }
  return utf8dec.decode(bytes);
}

function htmlEncode(s: string): string {
  return s.replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`);
}

function htmlDecode(s: string): string {
  const el = document.createElement("textarea");
  el.innerHTML = s;
  return el.value;
}

type Op = { label: string; encode: (s: string) => string; decode: (s: string) => string };

const OPS: Op[] = [
  { label: "Base64", encode: base64Encode, decode: base64Decode },
  { label: "URL", encode: encodeURIComponent, decode: decodeURIComponent },
  { label: "HTML", encode: htmlEncode, decode: htmlDecode },
  { label: "Hex", encode: hexEncode, decode: hexDecode },
];

export default function Decoder(): JSX.Element {
  const [input, setInput] = useState("");
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
        {OPS.map((op) => (
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
