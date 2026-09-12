import { TabContext, TabList, TabPanel } from "@mui/lab";
import { Box, Paper, Tab, ToggleButton, Typography } from "@mui/material";
import { useState } from "react";

import { KeyValuePairTable, KeyValuePair, KeyValuePairTableProps } from "./KeyValuePair";

import Editor from "lib/components/Editor";
import { canPrettify, prettify } from "lib/prettify";
import { rawResponse } from "lib/rawHttp";

interface ResponseTabsProps {
  proto?: string | null;
  statusCode?: number | null;
  statusReason?: string | null;
  headers: KeyValuePair[];
  onHeaderChange?: KeyValuePairTableProps["onChange"];
  onHeaderDelete?: KeyValuePairTableProps["onDelete"];
  body?: string | null;
  onBodyChange?: (value: string) => void;
  hasResponse: boolean;
}

enum TabValue {
  Raw = "raw",
  Body = "body",
  Render = "render",
  Headers = "headers",
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

function ResponseTabs(props: ResponseTabsProps): JSX.Element {
  const { proto, statusCode, statusReason, headers, onHeaderChange, onHeaderDelete, body, onBodyChange, hasResponse } =
    props;
  const [tabValue, setTabValue] = useState(TabValue.Raw);
  const [pretty, setPretty] = useState(false);
  const readOnly = onBodyChange === undefined;

  const contentType = headers.find((header) => header.key.toLowerCase() === "content-type")?.value;
  const canRender = isHTML(contentType, body);

  const tabSx = { textTransform: "none" };
  const headersLength = onHeaderChange ? headers.length - 1 : headers.length;

  return (
    <Box height="100%" sx={{ display: "flex", flexDirection: "column" }}>
      <TabContext value={tabValue}>
        <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 1, display: "flex", alignItems: "center" }}>
          <TabList onChange={(_, value) => setTabValue(value)} sx={{ flex: 1, minHeight: 40 }}>
            <Tab value={TabValue.Raw} label="Raw" sx={tabSx} />
            <Tab value={TabValue.Body} label={"Body" + (body?.length ? ` (${body.length})` : "")} sx={tabSx} />
            {canRender && <Tab value={TabValue.Render} label="Render" sx={tabSx} />}
            <Tab value={TabValue.Headers} label={"Headers" + (headersLength ? ` (${headersLength})` : "")} sx={tabSx} />
          </TabList>
          {readOnly && tabValue === TabValue.Body && canPrettify(body) && (
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
              <Editor content={rawResponse({ proto, statusCode, statusReason, headers, body })} language="plaintext" />
            ) : (
              reqNotSent
            )}
          </TabPanel>
          <TabPanel value={TabValue.Body} sx={{ p: 0, height: "100%" }}>
            {hasResponse ? (
              <Editor
                content={pretty && body ? prettify(body) : body || ""}
                onChange={(value) => {
                  onBodyChange && onBodyChange(value || "");
                }}
                monacoOptions={{ readOnly }}
                contentType={contentType}
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
          <TabPanel value={TabValue.Headers} sx={{ p: 0, height: "100%", overflow: "auto" }}>
            {hasResponse && <KeyValuePairTable items={headers} onChange={onHeaderChange} onDelete={onHeaderDelete} />}
            {!hasResponse && reqNotSent}
          </TabPanel>
        </Box>
      </TabContext>
    </Box>
  );
}

export default ResponseTabs;
