import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import {
  Alert,
  Box,
  FormControlLabel,
  IconButton,
  Link,
  MenuItem,
  Snackbar,
  styled,
  Switch,
  TableCell,
  TableCellProps,
  ToggleButton,
  ToggleButtonGroup,
  Tooltip,
} from "@mui/material";
import { useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";

import Actions from "./Actions";
import LogDetail from "./LogDetail";
import Search from "./Search";

import RequestsTable from "lib/components/RequestsTable";
import SplitPane from "lib/components/SplitPane";
import useContextMenu from "lib/components/useContextMenu";
import { useCreateSenderRequestFromHttpRequestLogMutation, useHttpRequestLogsQuery } from "lib/graphql/generated";

const ActionsTableCell = styled(TableCell)<TableCellProps>(() => ({
  paddingTop: 0,
  paddingBottom: 0,
}));

const STATIC_EXTENSIONS = /\.(css|js|mjs|png|jpe?g|gif|svg|ico|webp|woff2?|ttf|eot|map)(\?|$)/i;

type LogEntry = { method: string; url: string; response?: { statusCode: number } | null };

function statusClass(code: number): string {
  return `${Math.floor(code / 100)}xx`;
}

function applyLogFilters<T extends LogEntry>(logs: readonly T[], statuses: string[], hideStatic: boolean): T[] {
  return logs.filter((log) => {
    if (hideStatic && STATIC_EXTENSIONS.test(log.url)) {
      return false;
    }
    if (statuses.length > 0) {
      const cls = log.response ? statusClass(log.response.statusCode) : "none";
      if (!statuses.includes(cls)) {
        return false;
      }
    }
    return true;
  });
}

export function RequestLogs(): JSX.Element {
  const navigate = useNavigate();
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

  const [copyToSenderId, setCopyToSenderId] = useState("");
  const [Menu, handleContextMenu, handleContextMenuClose] = useContextMenu();

  const handleCopyToSenderClick = () => {
    createSenderReqFromLog({
      variables: {
        id: copyToSenderId,
      },
    });
    handleContextMenuClose();
  };

  const [statusFilter, setStatusFilter] = useState<string[]>([]);
  const [hideStatic, setHideStatic] = useState(false);

  const filteredLogs = useMemo(
    () => applyLogFilters(data?.httpRequestLogs || [], statusFilter, hideStatic),
    [data?.httpRequestLogs, statusFilter, hideStatic]
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

  const handleRowContextClick = (e: React.MouseEvent, id: string) => {
    setCopyToSenderId(id);
    handleContextMenu(e);
  };

  const actionsCell = (id: string) => (
    <ActionsTableCell>
      <Tooltip title="Copy to Sender">
        <IconButton
          size="small"
          onClick={() => {
            setCopyToSenderId(id);
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
          <Search />
        </Box>
        <Box pt={0.5}>
          <Actions />
        </Box>
      </Box>
      <Box sx={{ display: "flex", alignItems: "center", gap: 2, px: 1, pb: 1, flexWrap: "wrap" }}>
        <ToggleButtonGroup size="small" value={statusFilter} onChange={(_, value: string[]) => setStatusFilter(value)}>
          <ToggleButton value="2xx" sx={{ color: "success.main" }}>
            2xx
          </ToggleButton>
          <ToggleButton value="3xx" sx={{ color: "info.main" }}>
            3xx
          </ToggleButton>
          <ToggleButton value="4xx" sx={{ color: "warning.main" }}>
            4xx
          </ToggleButton>
          <ToggleButton value="5xx" sx={{ color: "error.main" }}>
            5xx
          </ToggleButton>
        </ToggleButtonGroup>
        <FormControlLabel
          control={<Switch size="small" checked={hideStatic} onChange={(e) => setHideStatic(e.target.checked)} />}
          label="Hide static assets"
        />
      </Box>
      <Box sx={{ display: "flex", flex: "1 auto", position: "relative" }}>
        <SplitPane split="horizontal" size={"40%"}>
          <Box sx={{ width: "100%", height: "100%", pb: 2 }}>
            <Box sx={{ width: "100%", height: "100%", overflow: "scroll" }}>
              <Menu>
                <MenuItem onClick={handleCopyToSenderClick}>Copy request to Sender</MenuItem>
              </Menu>
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
