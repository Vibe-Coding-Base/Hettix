import { TabContext, TabList, TabPanel } from "@mui/lab";
import { Box, Button, Tab, ToggleButton } from "@mui/material";
import { useState } from "react";

import { KeyValuePairTable, KeyValuePair, KeyValuePairTableProps } from "./KeyValuePair";

import Editor from "lib/components/Editor";
import { canPrettify, prettify } from "lib/prettify";

enum TabValue {
  QueryParams = "queryParams",
  Headers = "headers",
  Body = "body",
}

interface RequestTabsProps {
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
    queryParams,
    onQueryParamChange,
    onQueryParamDelete,
    headers,
    onHeaderChange,
    onHeaderDelete,
    body,
    onBodyChange,
  } = props;
  const [tabValue, setTabValue] = useState(TabValue.QueryParams);
  const [pretty, setPretty] = useState(false);
  const readOnly = onBodyChange === undefined;

  const tabSx = {
    textTransform: "none",
  };

  const queryParamsLength = onQueryParamChange ? queryParams.length - 1 : queryParams.length;
  const headersLength = onHeaderChange ? headers.length - 1 : headers.length;
  const showPretty = tabValue === TabValue.Body && canPrettify(body);

  return (
    <Box sx={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <TabContext value={tabValue}>
        <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 1, display: "flex", alignItems: "center" }}>
          <TabList onChange={(_, value) => setTabValue(value)} sx={{ flex: 1 }}>
            <Tab
              value={TabValue.QueryParams}
              label={"Query Params" + (queryParamsLength ? ` (${queryParamsLength})` : "")}
              sx={tabSx}
            />
            <Tab value={TabValue.Headers} label={"Headers" + (headersLength ? ` (${headersLength})` : "")} sx={tabSx} />
            <Tab
              value={TabValue.Body}
              label={"Body" + (body?.length ? ` (${body.length} byte` + (body.length > 1 ? "s" : "") + ")" : "")}
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
        <Box flex="1 auto" overflow="scroll" height="100%">
          <TabPanel value={TabValue.QueryParams} sx={{ p: 0, height: "100%" }}>
            <Box>
              <KeyValuePairTable items={queryParams} onChange={onQueryParamChange} onDelete={onQueryParamDelete} />
            </Box>
          </TabPanel>
          <TabPanel value={TabValue.Headers} sx={{ p: 0, height: "100%" }}>
            <Box>
              <KeyValuePairTable items={headers} onChange={onHeaderChange} onDelete={onHeaderDelete} />
            </Box>
          </TabPanel>
          <TabPanel value={TabValue.Body} sx={{ p: 0, height: "100%" }}>
            <Editor
              content={readOnly && pretty && body ? prettify(body) : body || ""}
              onChange={(value) => {
                onBodyChange && onBodyChange(value || "");
              }}
              monacoOptions={{ readOnly }}
              contentType={headers.find(({ key }) => key.toLowerCase() === "content-type")?.value}
            />
          </TabPanel>
        </Box>
      </TabContext>
    </Box>
  );
}

export default RequestTabs;
