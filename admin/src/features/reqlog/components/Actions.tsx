import AltRouteIcon from "@mui/icons-material/AltRoute";
import DeleteIcon from "@mui/icons-material/Delete";
import DownloadIcon from "@mui/icons-material/Download";
import { Alert, Badge, Button, IconButton, Menu, MenuItem, Tooltip } from "@mui/material";
import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";

import { useActiveProject } from "lib/ActiveProjectContext";
import { useInterceptedRequests } from "lib/InterceptedRequestsContext";
import { ConfirmationDialog, useConfirmationDialog } from "lib/components/ConfirmationDialog";
import { HttpRequestLogsDocument, useClearHttpRequestLogMutation } from "lib/graphql/generated";

function Actions(): JSX.Element {
  const activeProject = useActiveProject();
  const interceptedRequests = useInterceptedRequests();
  const [clearHTTPRequestLog, clearLogsResult] = useClearHttpRequestLogMutation({
    refetchQueries: [{ query: HttpRequestLogsDocument }],
  });
  const clearHTTPConfirmationDialog = useConfirmationDialog();
  const [exportAnchor, setExportAnchor] = useState<null | HTMLElement>(null);

  const download = (format: "har" | "csv") => {
    window.open(`/api/export/${format}`, "_blank");
    setExportAnchor(null);
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
