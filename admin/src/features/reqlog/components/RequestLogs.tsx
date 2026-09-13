import { useApolloClient } from "@apollo/client";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import { Alert, Box, IconButton, Link, Snackbar, styled, TableCell, TableCellProps, Tooltip } from "@mui/material";
import { useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";

import Actions from "./Actions";
import LogDetail from "./LogDetail";
import Search from "./Search";

import { useContextMenu } from "lib/components/ContextMenu";
import RequestsTable from "lib/components/RequestsTable";
import SplitPane from "lib/components/SplitPane";
import { downloadText } from "lib/download";
import {
  HttpRequestLogDocument,
  HttpRequestLogQuery,
  useCreateSenderRequestFromHttpRequestLogMutation,
  useHttpRequestLogsQuery,
} from "lib/graphql/generated";
import { applyLogFilters, emptyLogFilters, LogFilters } from "lib/logFilters";
import { rawRequest, rawResponse } from "lib/rawHttp";
import { toCurl, CurlPlatform } from "lib/toCurl";

const ActionsTableCell = styled(TableCell)<TableCellProps>(() => ({
  paddingTop: 0,
  paddingBottom: 0,
}));

export function RequestLogs(): JSX.Element {
  const navigate = useNavigate();
  const client = useApolloClient();
  const [searchParams] = useSearchParams();
  const id = searchParams.get("id") ?? undefined;
  const { data } = useHttpRequestLogsQuery({
    pollInterval: 1000,
  });

  const [createSenderReqFromLog] = useCreateSenderRequestFromHttpRequestLogMutation({
    onCompleted({ createSenderRequestFromHttpRequestLog }) {
      const { id } = createSenderRequestFromHttpRequestLog;
      setNewSenderReqId(id);
      setCopiedReqNotifOpen(true);
    },
  });

  const ctxMenu = useContextMenu();

  const [filters, setFilters] = useState<LogFilters>(emptyLogFilters);

  const filteredLogs = useMemo(
    () => applyLogFilters(data?.httpRequestLogs || [], filters),
    [data?.httpRequestLogs, filters]
  );

  const [newSenderReqId, setNewSenderReqId] = useState("");
  const [copiedReqNotifOpen, setCopiedReqNotifOpen] = useState(false);
  const handleCloseCopiedNotif = (_: Event | React.SyntheticEvent, reason?: string) => {
    if (reason === "clickaway") {
      return;
    }
    setCopiedReqNotifOpen(false);
  };

  const handleRowClick = (id: string) => {
    navigate(`/proxy/logs?id=${id}`);
  };

  const fetchLog = async (id: string) => {
    const { data } = await client.query<HttpRequestLogQuery>({ query: HttpRequestLogDocument, variables: { id } });
    return data.httpRequestLog ?? undefined;
  };

  const copyAsCurl = async (id: string, platform: CurlPlatform) => {
    const log = await fetchLog(id);
    if (!log) {
      return;
    }
    navigator.clipboard?.writeText(
      toCurl({ method: log.method, url: log.url, headers: log.headers ?? [], body: log.body }, platform)
    );
  };

  const downloadLog = async (id: string, includeResponse: boolean) => {
    const log = await fetchLog(id);
    if (!log) {
      return;
    }
    let content = rawRequest({
      method: log.method,
      url: log.url,
      proto: log.proto,
      headers: log.headers ?? [],
      body: log.body,
    });
    if (includeResponse && log.response) {
      content +=
        "\n\n" +
        rawResponse({
          proto: log.response.proto,
          statusCode: log.response.statusCode,
          statusReason: log.response.statusReason,
          headers: log.response.headers ?? [],
          body: log.response.body,
        });
    }
    downloadText(`request-${id}.txt`, content);
  };

  const handleRowContextClick = (e: React.MouseEvent, id: string) => {
    const log = filteredLogs.find((l) => l.id === id);
    if (!log) {
      return;
    }

    ctxMenu.open(e, [
      { label: "Send to Sender", onClick: () => createSenderReqFromLog({ variables: { id } }) },
      {
        label: "Send to Intruder",
        onClick: () => navigate(`/intruder?url=${encodeURIComponent(log.url)}&method=${log.method}`),
      },
      { label: "Copy URL", onClick: () => navigator.clipboard?.writeText(log.url), divider: true },
      { label: "Copy as curl (bash)", onClick: () => copyAsCurl(id, "unix") },
      { label: "Copy as curl (Windows)", onClick: () => copyAsCurl(id, "windows") },
      { label: "Save request", onClick: () => downloadLog(id, false), divider: true },
      { label: "Save request + response", onClick: () => downloadLog(id, true) },
    ]);
  };

  const actionsCell = (id: string) => (
    <ActionsTableCell>
      <Tooltip title="Copy to Sender">
        <IconButton
          size="small"
          onClick={() => {
            createSenderReqFromLog({
              variables: {
                id,
              },
            });
          }}
        >
          <ContentCopyIcon fontSize="small" />
        </IconButton>
      </Tooltip>
    </ActionsTableCell>
  );

  return (
    <Box display="flex" flexDirection="column" height="100%">
      <Box display="flex">
        <Box flex="1 auto">
          <Search filters={filters} onFiltersChange={setFilters} />
        </Box>
        <Box pt={0.5}>
          <Actions />
        </Box>
      </Box>
      <Box sx={{ display: "flex", flex: "1 auto", position: "relative" }}>
        <SplitPane split="horizontal" size={"40%"}>
          <Box sx={{ width: "100%", height: "100%", pb: 2 }}>
            <Box sx={{ width: "100%", height: "100%", overflow: "scroll" }}>
              {ctxMenu.menu}
              <Snackbar
                open={copiedReqNotifOpen}
                autoHideDuration={3000}
                onClose={handleCloseCopiedNotif}
                anchorOrigin={{ horizontal: "center", vertical: "bottom" }}
              >
                <Alert onClose={handleCloseCopiedNotif} severity="info">
                  Request was copied. <Link href={`/sender?id=${newSenderReqId}`}>Edit in Sender.</Link>
                </Alert>
              </Snackbar>
              <RequestsTable
                requests={filteredLogs}
                activeRowId={id}
                actionsCell={actionsCell}
                onRowClick={handleRowClick}
                onContextMenu={handleRowContextClick}
              />
            </Box>
          </Box>
          <LogDetail id={id} />
        </SplitPane>
      </Box>
    </Box>
  );
}
