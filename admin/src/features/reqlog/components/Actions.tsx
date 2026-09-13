import AltRouteIcon from "@mui/icons-material/AltRoute";
import DeleteIcon from "@mui/icons-material/Delete";
import DownloadIcon from "@mui/icons-material/Download";
import { Alert, Badge, Button, IconButton, Menu, MenuItem, Tooltip } from "@mui/material";
import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";

import { useActiveProject } from "lib/ActiveProjectContext";
import { useInterceptedRequests } from "lib/InterceptedRequestsContext";
import { ConfirmationDialog, useConfirmationDialog } from "lib/components/ConfirmationDialog";
import { downloadText } from "lib/download";
import { HttpRequestLogsDocument, useClearHttpRequestLogMutation } from "lib/graphql/generated";

function Actions(): JSX.Element {
  const activeProject = useActiveProject();
  const interceptedRequests = useInterceptedRequests();
  const [clearHTTPRequestLog, clearLogsResult] = useClearHttpRequestLogMutation({
    refetchQueries: [{ query: HttpRequestLogsDocument }],
  });
  const clearHTTPConfirmationDialog = useConfirmationDialog();
  const [exportAnchor, setExportAnchor] = useState<null | HTMLElement>(null);
  const [exportError, setExportError] = useState("");

  // Fetch the export in-page and save it as a blob. A plain window.open would be
  // handed to the system browser by the desktop shell, which cannot reach the
  // in-process asset server.
  const download = async (format: "har" | "csv") => {
    setExportAnchor(null);
    setExportError("");
    try {
      const res = await fetch(`/api/export/${format}`);
      if (!res.ok) {
        throw new Error(await res.text());
      }
      const mime = format === "har" ? "application/json" : "text/csv";
      downloadText(`hettix-export.${format}`, await res.text(), mime);
    } catch (e) {
      setExportError(e instanceof Error ? e.message : String(e));
    }
  };

  return (
    <div>
      <ConfirmationDialog
        isOpen={clearHTTPConfirmationDialog.isOpen}
        onClose={clearHTTPConfirmationDialog.close}
        onConfirm={clearHTTPRequestLog}
      >
        All proxy logs are going to be removed. This action cannot be undone.
      </ConfirmationDialog>

      {clearLogsResult.error && (
        <Alert severity="error">Failed to clear HTTP logs: {clearLogsResult.error.message}</Alert>
      )}

      {exportError && <Alert severity="error">Failed to export: {exportError}</Alert>}

      {(activeProject?.settings.intercept.requestsEnabled || activeProject?.settings.intercept.responsesEnabled) && (
        <Button
          component={RouterLink}
          to="/proxy/intercept/?id="
          variant="contained"
          disabled={interceptedRequests === null || interceptedRequests.length === 0}
          color="primary"
          size="large"
          startIcon={
            <Badge color="error" badgeContent={interceptedRequests?.length || 0}>
              <AltRouteIcon />
            </Badge>
          }
          sx={{ mr: 1 }}
        >
          Review Intercepted…
        </Button>
      )}

      <Tooltip title="Export">
        <span>
          <IconButton disabled={!activeProject} onClick={(e) => setExportAnchor(e.currentTarget)}>
            <DownloadIcon />
          </IconButton>
        </span>
      </Tooltip>
      <Menu anchorEl={exportAnchor} open={Boolean(exportAnchor)} onClose={() => setExportAnchor(null)}>
        <MenuItem onClick={() => download("har")}>Export as HAR</MenuItem>
        <MenuItem onClick={() => download("csv")}>Export as CSV</MenuItem>
      </Menu>

      <Tooltip title="Clear all">
        <IconButton onClick={clearHTTPConfirmationDialog.open}>
          <DeleteIcon />
        </IconButton>
      </Tooltip>
    </div>
  );
}

export default Actions;
