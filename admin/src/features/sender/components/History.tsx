import { Box, Paper, Typography } from "@mui/material";
import { useNavigate, useSearchParams } from "react-router-dom";

import RequestsTable from "lib/components/RequestsTable";
import { useGetSenderRequestsQuery } from "lib/graphql/generated";

function History(): JSX.Element {
  const { data, loading } = useGetSenderRequestsQuery({
    pollInterval: 1000,
  });

  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const activeId = searchParams.get("id") ?? undefined;

  const handleRowClick = (id: string) => {
    navigate(`/sender?id=${id}`);
  };

  return (
    <Box>
      {!loading && data?.senderRequests && data?.senderRequests.length > 0 && (
        <RequestsTable requests={data.senderRequests} onRowClick={handleRowClick} activeRowId={activeId} />
      )}
      <Box sx={{ mt: 2, height: "100%" }}>
        {!loading && data?.senderRequests.length === 0 && (
          <Paper variant="centered">
            <Typography>No requests created yet.</Typography>
          </Paper>
        )}
      </Box>
    </Box>
  );
}

export default History;
