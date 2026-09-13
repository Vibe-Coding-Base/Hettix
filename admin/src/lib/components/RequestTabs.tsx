import { Box, ToggleButton } from "@mui/material";
import { useState } from "react";

import { KeyValuePair } from "./KeyValuePair";

import Editor from "lib/components/Editor";
import { canPrettify } from "lib/prettify";
import { rawRequest } from "lib/rawHttp";

interface RequestTabsProps {
  // Read-only rendering builds the raw message from these structured fields.
  method?: string;
  url?: string;
  proto?: string | null;
  headers: KeyValuePair[];
  body?: string | null;
  // Editable mode: the parent owns the raw text and parses it on send.
  raw?: string;
  onRawChange?: (raw: string) => void;
}

// RequestTabs shows a request as a single raw HTTP message (Burp-style). When
// editable, the parent's raw text is edited directly; when read-only, a Pretty
// toggle pretty-prints the body.
function RequestTabs({ method, url, proto, headers, body, raw, onRawChange }: RequestTabsProps): JSX.Element {
  const editable = onRawChange !== undefined;
  const [pretty, setPretty] = useState(false);

  const content = editable ? raw ?? "" : rawRequest({ method, url, proto, headers, body, pretty });

  return (
    <Box sx={{ display: "flex", flexDirection: "column", height: "100%" }}>
      {!editable && canPrettify(body) && (
        <Box sx={{ display: "flex", justifyContent: "flex-end", pb: 0.5 }}>
          <ToggleButton
            value="pretty"
            size="small"
            selected={pretty}
            onChange={() => setPretty((p) => !p)}
            sx={{ textTransform: "none", py: 0.2 }}
          >
            Pretty
          </ToggleButton>
        </Box>
      )}
      <Box sx={{ flex: "1 auto", overflow: "hidden" }}>
        <Editor
          content={content}
          language="http"
          monacoOptions={{ readOnly: !editable }}
          onChange={onRawChange ? (value) => onRawChange(value ?? "") : undefined}
        />
      </Box>
    </Box>
  );
}

export default RequestTabs;
