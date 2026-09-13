import { TabContext, TabList, TabPanel } from "@mui/lab";
import { Box, Paper, Tab, ToggleButton, Typography } from "@mui/material";
import { useState } from "react";

import { KeyValuePair } from "./KeyValuePair";

import Editor from "lib/components/Editor";
import { canPrettify } from "lib/prettify";
import { rawResponse } from "lib/rawHttp";

interface ResponseTabsProps {
  // Read-only rendering builds the raw message from these structured fields.
  proto?: string | null;
  statusCode?: number | null;
  statusReason?: string | null;
  headers: KeyValuePair[];
  body?: string | null;
  hasResponse: boolean;
  // Editable mode: the parent owns the raw text and parses it on forward.
  raw?: string;
  onRawChange?: (raw: string) => void;
}

enum TabValue {
  Raw = "raw",
  Render = "render",
}

function isHTML(contentType?: string, body?: string | null): boolean {
  if (contentType?.toLowerCase().includes("html")) {
    return true;
  }
  const trimmed = body?.trimStart() ?? "";
  return /^<!doctype html|^<html[\s>]/i.test(trimmed);
}

const reqNotSent = (
  <Paper variant="centered">
    <Typography>Response not received yet.</Typography>
  </Paper>
);

// ResponseTabs shows a response as a single raw HTTP message plus, for HTML, a
// Render tab that displays it as a browser would. When editable, the raw text is
// edited directly (used by intercept).
function ResponseTabs({
  proto,
  statusCode,
  statusReason,
  headers,
  body,
  hasResponse,
  raw,
  onRawChange,
}: ResponseTabsProps): JSX.Element {
  const editable = onRawChange !== undefined;
  const [tabValue, setTabValue] = useState(TabValue.Raw);
  const [pretty, setPretty] = useState(false);

  const contentType = headers.find((header) => header.key.toLowerCase() === "content-type")?.value;
  const canRender = isHTML(contentType, body);
  const content = editable ? raw ?? "" : rawResponse({ proto, statusCode, statusReason, headers, body, pretty });

  return (
    <Box height="100%" sx={{ display: "flex", flexDirection: "column" }}>
      <TabContext value={tabValue}>
        <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 1, display: "flex", alignItems: "center" }}>
          <TabList onChange={(_, value) => setTabValue(value)} sx={{ flex: 1, minHeight: 40 }}>
            <Tab value={TabValue.Raw} label="Raw" sx={{ textTransform: "none" }} />
            {canRender && <Tab value={TabValue.Render} label="Render" sx={{ textTransform: "none" }} />}
          </TabList>
          {!editable && tabValue === TabValue.Raw && canPrettify(body) && (
            <ToggleButton
              value="pretty"
              size="small"
              selected={pretty}
              onChange={() => setPretty((p) => !p)}
              sx={{ mr: 1, textTransform: "none", py: 0.2 }}
            >
              Pretty
            </ToggleButton>
          )}
        </Box>
        <Box flex="1 auto" overflow="hidden">
          <TabPanel value={TabValue.Raw} sx={{ p: 0, height: "100%" }}>
            {hasResponse ? (
              <Editor
                content={content}
                language="http"
                monacoOptions={{ readOnly: !editable }}
                onChange={onRawChange ? (value) => onRawChange(value ?? "") : undefined}
              />
            ) : (
              reqNotSent
            )}
          </TabPanel>
          {canRender && (
            <TabPanel value={TabValue.Render} sx={{ p: 0, height: "100%" }}>
              {/* Sandboxed with no allow-* tokens: the response's scripts never
                  run, so rendering an attacker-controlled page is safe. */}
              <iframe
                title="Rendered response"
                srcDoc={body || ""}
                sandbox=""
                style={{ width: "100%", height: "100%", border: "none", background: "#fff" }}
              />
            </TabPanel>
          )}
        </Box>
      </TabContext>
    </Box>
  );
}

export default ResponseTabs;
