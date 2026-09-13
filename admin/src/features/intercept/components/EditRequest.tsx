import { ApolloCache } from "@apollo/client";
import CancelIcon from "@mui/icons-material/Cancel";
import DownloadIcon from "@mui/icons-material/Download";
import SendIcon from "@mui/icons-material/Send";
import SettingsIcon from "@mui/icons-material/Settings";
import { Alert, Box, Button, CircularProgress, IconButton, Tooltip, Typography } from "@mui/material";
import React, { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";

import { useInterceptedRequests } from "lib/InterceptedRequestsContext";
import Link from "lib/components/Link";
import RequestTabs from "lib/components/RequestTabs";
import ResponseStatus from "lib/components/ResponseStatus";
import ResponseTabs from "lib/components/ResponseTabs";
import { HttpMethod, HttpProtocol } from "lib/graphql/generated";
import {
  HttpRequest,
  useCancelRequestMutation,
  useCancelResponseMutation,
  useGetInterceptedRequestQuery,
  useModifyRequestMutation,
  useModifyResponseMutation,
} from "lib/graphql/generated";
import { parseRawRequest, parseRawResponse, rawRequest, rawResponse } from "lib/rawHttp";

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
  const interceptedRequests = useInterceptedRequests();

  useEffect(() => {
    // If nothing is selected and requests are pending, jump to the first one.
    if (!searchParams.get("id") && interceptedRequests?.length) {
      navigate(`/proxy/intercept?id=${interceptedRequests[0].id}`, { replace: true });
    }
  }, [searchParams, navigate, interceptedRequests]);

  const reqId = searchParams.get("id") ?? undefined;

  const [rawReq, setRawReq] = useState("");
  const [rawRes, setRawRes] = useState("");
  const [scheme, setScheme] = useState("https");

  const getReqResult = useGetInterceptedRequestQuery({
    variables: { id: reqId as string },
    skip: reqId === undefined,
    onCompleted: ({ interceptedRequest }) => {
      if (!interceptedRequest) {
        return;
      }
      setRawReq(
        rawRequest({
          method: interceptedRequest.method,
          url: interceptedRequest.url,
          proto: interceptedRequest.proto,
          headers: interceptedRequest.headers ?? [],
          body: interceptedRequest.body,
        })
      );
      setScheme(schemeOf(interceptedRequest.url));

      if (interceptedRequest.response) {
        setRawRes(
          rawResponse({
            proto: interceptedRequest.response.proto,
            statusCode: interceptedRequest.response.statusCode,
            statusReason: interceptedRequest.response.statusReason,
            headers: interceptedRequest.response.headers ?? [],
            body: interceptedRequest.response.body,
          })
        );
      }
    },
  });
  const interceptedReq =
    reqId && !getReqResult?.data?.interceptedRequest?.response ? getReqResult?.data?.interceptedRequest : undefined;
  const interceptedRes = reqId ? getReqResult?.data?.interceptedRequest?.response : undefined;

  const [modifyRequest, modifyReqResult] = useModifyRequestMutation();
  const [cancelRequest, cancelReqResult] = useCancelRequestMutation();
  const [modifyResponse, modifyResResult] = useModifyResponseMutation();
  const [cancelResponse, cancelResResult] = useCancelResponseMutation();

  const dropFromCache = (cache: ApolloCache<unknown>, id: string) => {
    cache.modify<{ interceptedRequests: HttpRequest[] }>({
      fields: {
        interceptedRequests(existing, { readField }) {
          return existing.filter((ref) => id !== readField("id", ref));
        },
      },
    });
  };

  const onActionCompleted = () => {
    setRawReq("");
    setRawRes("");
    navigate(`/proxy/intercept`, { replace: true });
  };

  const handleFormSubmit: React.FormEventHandler = (e) => {
    e.preventDefault();

    if (interceptedReq) {
      const parsed = parseRawRequest(rawReq, scheme);
      modifyRequest({
        variables: {
          request: {
            id: interceptedReq.id,
            url: parsed.url,
            method: parsed.method.toUpperCase() as HttpMethod,
            proto: HttpProtocol.Http20,
            headers: parsed.headers.filter((kv) => kv.key !== ""),
            body: parsed.body || undefined,
          },
        },
        update: (cache) => dropFromCache(cache, interceptedReq.id),
        onCompleted: onActionCompleted,
      });
    }

    if (interceptedRes) {
      const parsed = parseRawResponse(rawRes);
      modifyResponse({
        variables: {
          response: {
            requestID: interceptedRes.id,
            proto: interceptedRes.proto, // proto is kept; status/headers/body are editable
            statusCode: parsed.statusCode || interceptedRes.statusCode,
            statusReason: parsed.statusReason,
            headers: parsed.headers.filter((kv) => kv.key !== ""),
            body: parsed.body || undefined,
          },
        },
        update: (cache) => dropFromCache(cache, interceptedRes.id),
        onCompleted: onActionCompleted,
      });
    }
  };

  const handleReqCancelClick = () => {
    if (!interceptedReq) {
      return;
    }
    cancelRequest({
      variables: { id: interceptedReq.id },
      update: (cache) => dropFromCache(cache, interceptedReq.id),
      onCompleted: onActionCompleted,
    });
  };

  const handleResCancelClick = () => {
    if (!interceptedRes) {
      return;
    }
    cancelResponse({
      variables: { requestID: interceptedRes.id },
      update: (cache) => dropFromCache(cache, interceptedRes.id),
      onCompleted: onActionCompleted,
    });
  };

  return (
    <Box display="flex" flexDirection="column" height="100%" gap={2}>
      <Box component="form" autoComplete="off" onSubmit={handleFormSubmit}>
        <Box sx={{ display: "flex", gap: 1, flexWrap: "wrap", alignItems: "center" }}>
          {!interceptedRes && (
            <>
              <Button
                variant="contained"
                disableElevation
                type="submit"
                disabled={!interceptedReq || modifyReqResult.loading || cancelReqResult.loading}
                startIcon={modifyReqResult.loading ? <CircularProgress size={22} /> : <SendIcon />}
              >
                Send
              </Button>
              <Button
                variant="contained"
                color="error"
                disableElevation
                onClick={handleReqCancelClick}
                disabled={!interceptedReq || modifyReqResult.loading || cancelReqResult.loading}
                startIcon={cancelReqResult.loading ? <CircularProgress size={22} /> : <CancelIcon />}
              >
                Cancel
              </Button>
            </>
          )}
          {interceptedRes && (
            <>
              <Button
                variant="contained"
                disableElevation
                type="submit"
                disabled={modifyResResult.loading || cancelResResult.loading}
                endIcon={modifyResResult.loading ? <CircularProgress size={22} /> : <DownloadIcon />}
              >
                Receive
              </Button>
              <Button
                variant="contained"
                color="error"
                disableElevation
                onClick={handleResCancelClick}
                disabled={modifyResResult.loading || cancelResResult.loading}
                endIcon={cancelResResult.loading ? <CircularProgress size={22} /> : <CancelIcon />}
              >
                Cancel
              </Button>
            </>
          )}
          <Box sx={{ flex: "1 auto" }} />
          <Tooltip title="Intercept settings">
            <IconButton LinkComponent={Link} href="/settings#intercept">
              <SettingsIcon />
            </IconButton>
          </Tooltip>
        </Box>
        {modifyReqResult.error && (
          <Alert severity="error" sx={{ mt: 1 }}>
            {modifyReqResult.error.message}
          </Alert>
        )}
        {cancelReqResult.error && (
          <Alert severity="error" sx={{ mt: 1 }}>
            {cancelReqResult.error.message}
          </Alert>
        )}
      </Box>

      <Box flex="1 auto" overflow="hidden">
        {interceptedReq && (
          <Box sx={{ height: "100%", pb: 2 }}>
            <Typography variant="overline" color="textSecondary" sx={{ position: "absolute", right: 0, mt: 1.2 }}>
              Request
            </Typography>
            <RequestTabs headers={[]} raw={rawReq} onRawChange={setRawReq} />
          </Box>
        )}
        {interceptedRes && (
          <Box sx={{ height: "100%", pb: 2 }}>
            <Box sx={{ position: "absolute", right: 0, mt: 1.4 }}>
              <Box sx={{ float: "right", mt: 0.2 }}>
                <ResponseStatus
                  proto={interceptedRes.proto}
                  statusCode={interceptedRes.statusCode}
                  statusReason={interceptedRes.statusReason}
                />
              </Box>
            </Box>
            <ResponseTabs
              headers={(interceptedRes.headers ?? []).map(({ key, value }) => ({ key, value }))}
              body={interceptedRes.body}
              hasResponse
              raw={rawRes}
              onRawChange={setRawRes}
            />
          </Box>
        )}
      </Box>
    </Box>
  );
}

export default EditRequest;
