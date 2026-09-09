import { Box } from "@mui/material";

import ResponseStatus from "lib/components/ResponseStatus";
import ResponseTabs from "lib/components/ResponseTabs";
import { HttpResponseLog } from "lib/graphql/generated";

interface ResponseProps {
  response?: HttpResponseLog | null;
}

function Response({ response }: ResponseProps): JSX.Element {
  return (
    <Box sx={{ height: "100%", display: "flex", flexDirection: "column", minHeight: 0 }}>
      {response && (
        <Box sx={{ flexShrink: 0, px: 1, pb: 0.5, overflowX: "auto" }}>
          <ResponseStatus
            proto={response.proto}
            statusCode={response.statusCode}
            statusReason={response.statusReason}
          />
        </Box>
      )}
      <Box sx={{ flex: 1, minHeight: 0 }}>
        <ResponseTabs
          body={response?.body}
          headers={response?.headers || []}
          hasResponse={response !== undefined && response !== null}
        />
      </Box>
    </Box>
  );
}

export default Response;
