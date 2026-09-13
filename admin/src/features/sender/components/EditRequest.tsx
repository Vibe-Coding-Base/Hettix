import AddIcon from "@mui/icons-material/Add";
import { Alert, Box, Button, Fab, ToggleButton, ToggleButtonGroup, Tooltip, Typography, useTheme } from "@mui/material";
import React, { useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";

import RequestTabs from "lib/components/RequestTabs";
import Response from "lib/components/Response";
import SplitPane from "lib/components/SplitPane";
import { HttpMethod, HttpProto, httpProtoMap } from "lib/components/UrlBar";
import {
  GetSenderRequestQuery,
  useCreateOrUpdateSenderRequestMutation,
  useGetSenderRequestQuery,
  useSendRequestMutation,
} from "lib/graphql/generated";
import { parseRawRequest, rawRequest } from "lib/rawHttp";

const newRaw = "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n";

function schemeOf(url: string): string {
  try {
    return new URL(url).protocol.replace(":", "") || "https";
  } catch {
    return "https";
  }
}

function EditRequest(): JSX.Element {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const reqId = searchParams.get("id") ?? undefined;

  const theme = useTheme();

  const [raw, setRaw] = useState(newRaw);
  const [scheme, setScheme] = useState("https");
  const [response, setResponse] = useState<NonNullable<GetSenderRequestQuery["senderRequest"]>["response"]>(null);

  const getReqResult = useGetSenderRequestQuery({
    variables: { id: reqId as string },
    skip: reqId === undefined,
    onCompleted: ({ senderRequest }) => {
      if (!senderRequest) {
        return;
      }
      setRaw(
        rawRequest({
          method: senderRequest.method,
          url: senderRequest.url,
          headers: senderRequest.headers ?? [],
          body: senderRequest.body,
        })
      );
      setScheme(schemeOf(senderRequest.url));
      setResponse(senderRequest.response);
    },
  });

  const [createOrUpdateRequest, createResult] = useCreateOrUpdateSenderRequestMutation();
  const [sendRequest, sendResult] = useSendRequestMutation();

  const createOrUpdateRequestAndSend = () => {
    const senderReq = getReqResult?.data?.senderRequest;
    const parsed = parseRawRequest(raw, scheme);
    createOrUpdateRequest({
      variables: {
        request: {
          // Update the existing request if it was cloned from a log and not sent yet.
          ...(senderReq && senderReq.sourceRequestLogID && !senderReq.response && { id: senderReq.id }),
          url: parsed.url,
          method: parsed.method.toUpperCase() as HttpMethod,
          proto: httpProtoMap.get(HttpProto.Http20),
          headers: parsed.headers.filter((kv) => kv.key !== ""),
          body: parsed.body || undefined,
        },
      },
      onCompleted: ({ createOrUpdateSenderRequest }) => {
        sendRequestAndPushRoute(createOrUpdateSenderRequest.id);
      },
    });
  };

  const sendRequestAndPushRoute = (id: string) => {
    sendRequest({
      errorPolicy: "all",
      onCompleted: () => navigate(`/sender?id=${id}`),
      variables: { id },
    });
  };

  const handleFormSubmit: React.FormEventHandler = (e) => {
    e.preventDefault();
    createOrUpdateRequestAndSend();
  };

  const handleNewRequest = () => {
    setRaw(newRaw);
    setScheme("https");
    setResponse(null);
    navigate(`/sender`);
  };

  return (
    <Box display="flex" flexDirection="column" height="100%" gap={2}>
      <Box sx={{ position: "absolute", bottom: theme.spacing(2), right: theme.spacing(2) }}>
        <Tooltip title="New request">
          <Fab color="primary" onClick={handleNewRequest}>
            <AddIcon />
          </Fab>
        </Tooltip>
      </Box>
      <Box component="form" autoComplete="off" onSubmit={handleFormSubmit}>
        <Box sx={{ display: "flex", gap: 1, alignItems: "center" }}>
          <ToggleButtonGroup exclusive size="small" value={scheme} onChange={(_, value) => value && setScheme(value)}>
            <ToggleButton value="https" sx={{ textTransform: "none", py: 0.4 }}>
              https
            </ToggleButton>
            <ToggleButton value="http" sx={{ textTransform: "none", py: 0.4 }}>
              http
            </ToggleButton>
          </ToggleButtonGroup>
          <Box sx={{ flex: "1 auto" }} />
          <Button
            variant="contained"
            disableElevation
            sx={{ width: "8rem" }}
            type="submit"
            disabled={createResult.loading || sendResult.loading}
          >
            Send
          </Button>
        </Box>
        {createResult.error && (
          <Alert severity="error" sx={{ mt: 1 }}>
            {createResult.error.message}
          </Alert>
        )}
        {sendResult.error && (
          <Alert severity="error" sx={{ mt: 1 }}>
            {sendResult.error.message}
          </Alert>
        )}
      </Box>

      <Box flex="1 auto" position="relative">
        <SplitPane split="vertical" size={"50%"}>
          <Box sx={{ height: "100%", mr: 2, pb: 2, position: "relative" }}>
            <Typography variant="overline" color="textSecondary" sx={{ position: "absolute", right: 0, mt: 1.2 }}>
              Request
            </Typography>
            <RequestTabs headers={[]} raw={raw} onRawChange={setRaw} />
          </Box>
          <Box sx={{ height: "100%", position: "relative", ml: 2, pb: 2 }}>
            <Response response={response} />
          </Box>
        </SplitPane>
      </Box>
    </Box>
  );
}

export default EditRequest;
