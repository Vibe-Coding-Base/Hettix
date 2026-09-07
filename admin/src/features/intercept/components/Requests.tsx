import { Box, Paper, Typography } from "@mui/material";
import { useNavigate, useSearchParams } from "react-router-dom";

import { useInterceptedRequests } from "lib/InterceptedRequestsContext";
import RequestsTable from "lib/components/RequestsTable";

function Requests(): JSX.Element {
  const interceptedRequests = useInterceptedRequests();

  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const activeId = searchParams.get("id") ?? undefined;

  const handleRowClick = (id: string) => {
    navigate(`/proxy/intercept?id=${id}`);
  };

  return (
    <Box>
      {interceptedRequests && interceptedRequests.length > 0 && (
        <RequestsTable requests={interceptedRequests} onRowClick={handleRowClick} activeRowId={activeId} />
      )}
      <Box sx={{ mt: 2, height: "100%" }}>
        {interceptedRequests?.length === 0 && (
          <Paper variant="centered">
            <Typography>No pending intercepted requests.</Typography>
          </Paper>
        )}
      </Box>
    </Box>
  );
}

export default Requests;
