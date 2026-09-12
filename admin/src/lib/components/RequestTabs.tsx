import { TabContext, TabList, TabPanel } from "@mui/lab";
import { Box, Button, Tab, ToggleButton } from "@mui/material";
import { useState } from "react";

import { KeyValuePairTable, KeyValuePair, KeyValuePairTableProps } from "./KeyValuePair";

import Editor from "lib/components/Editor";
import { canPrettify, prettify } from "lib/prettify";
import { rawRequest } from "lib/rawHttp";

enum TabValue {
  Raw = "raw",
  Body = "body",
  Headers = "headers",
  QueryParams = "queryParams",
}

interface RequestTabsProps {
  method?: string;
  url?: string;
  proto?: string | null;
  queryParams: KeyValuePair[];
  headers: KeyValuePair[];
  onQueryParamChange?: KeyValuePairTableProps["onChange"];
  onQueryParamDelete?: KeyValuePairTableProps["onDelete"];
  onHeaderChange?: KeyValuePairTableProps["onChange"];
  onHeaderDelete?: KeyValuePairTableProps["onDelete"];
  body?: string | null;
  onBodyChange?: (value: string) => void;
}

function RequestTabs(props: RequestTabsProps): JSX.Element {
  const {
    method,
    url,
    proto,
    queryParams,
    onQueryParamChange,
    onQueryParamDelete,
    headers,
    onHeaderChange,
    onHeaderDelete,
    body,
    onBodyChange,
  } = props;
  const readOnly = onBodyChange === undefined;
  // Viewers land on the full raw message (Burp-style); editors land on the
  // editable body.
  const [tabValue, setTabValue] = useState(readOnly ? TabValue.Raw : TabValue.Body);
  const [pretty, setPretty] = useState(false);

  const tabSx = { textTransform: "none" };

  const queryParamsLength = onQueryParamChange ? queryParams.length - 1 : queryParams.length;
  const headersLength = onHeaderChange ? headers.length - 1 : headers.length;
  const showPretty = tabValue === TabValue.Body && canPrettify(body);
  const contentType = headers.find(({ key }) => key.toLowerCase() === "content-type")?.value;

  return (
    <Box sx={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <TabContext value={tabValue}>
        <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 1, display: "flex", alignItems: "center" }}>
          <TabList onChange={(_, value) => setTabValue(value)} sx={{ flex: 1, minHeight: 40 }}>
            <Tab value={TabValue.Raw} label="Raw" sx={tabSx} />
            <Tab value={TabValue.Body} label={"Body" + (body?.length ? ` (${body.length})` : "")} sx={tabSx} />
            <Tab value={TabValue.Headers} label={"Headers" + (headersLength ? ` (${headersLength})` : "")} sx={tabSx} />
            <Tab
              value={TabValue.QueryParams}
              label={"Query" + (queryParamsLength ? ` (${queryParamsLength})` : "")}
              sx={tabSx}
            />
          </TabList>
          {showPretty &&
            (readOnly ? (
              <ToggleButton
                value="pretty"
                size="small"
                selected={pretty}
                onChange={() => setPretty((p) => !p)}
                sx={{ mr: 1, textTransform: "none", py: 0.2 }}
              >
                Pretty
              </ToggleButton>
            ) : (
              <Button size="small" sx={{ mr: 1 }} onClick={() => onBodyChange && body && onBodyChange(prettify(body))}>
                Format
              </Button>
            ))}
        </Box>
        <Box flex="1 auto" overflow="hidden" height="100%">
          <TabPanel value={TabValue.Raw} sx={{ p: 0, height: "100%" }}>
            <Editor content={rawRequest({ method, url, proto, headers, body })} language="plaintext" />
          </TabPanel>
          <TabPanel value={TabValue.Body} sx={{ p: 0, height: "100%" }}>
            <Editor
              content={readOnly && pretty && body ? prettify(body) : body || ""}
              onChange={(value) => {
                onBodyChange && onBodyChange(value || "");
              }}
              monacoOptions={{ readOnly }}
              contentType={contentType}
            />
          </TabPanel>
          <TabPanel value={TabValue.Headers} sx={{ p: 0, height: "100%", overflow: "auto" }}>
            <KeyValuePairTable items={headers} onChange={onHeaderChange} onDelete={onHeaderDelete} />
          </TabPanel>
          <TabPanel value={TabValue.QueryParams} sx={{ p: 0, height: "100%", overflow: "auto" }}>
            <KeyValuePairTable items={queryParams} onChange={onQueryParamChange} onDelete={onQueryParamDelete} />
          </TabPanel>
        </Box>
      </TabContext>
    </Box>
  );
}

export default RequestTabs;
