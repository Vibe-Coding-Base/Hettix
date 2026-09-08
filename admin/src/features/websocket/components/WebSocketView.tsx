import { Badge, Box, Tab, Tabs } from "@mui/material";
import { useState } from "react";

import WebSocketIntercept from "./WebSocketIntercept";
import WebSocketLogs from "./WebSocketLogs";

import { useInterceptedWebSocketMessagesQuery } from "lib/graphql/generated";

export default function WebSocketView(): JSX.Element {
  const [tab, setTab] = useState(0);
  const { data } = useInterceptedWebSocketMessagesQuery({ pollInterval: 1000 });
  const pending = data?.interceptedWebSocketMessages?.length ?? 0;

  return (
    <Box sx={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Tabs value={tab} onChange={(_, value) => setTab(value)} sx={{ mb: 2 }}>
        <Tab label="Traffic" />
        <Tab
          label={
            <Badge color="error" badgeContent={pending} sx={{ pr: pending ? 1.5 : 0 }}>
              Intercept
            </Badge>
          }
        />
      </Tabs>
      <Box sx={{ flex: 1, overflow: "auto" }} hidden={tab !== 0}>
        {tab === 0 && <WebSocketLogs />}
      </Box>
      <Box sx={{ flex: 1, overflow: "auto" }} hidden={tab !== 1}>
        {tab === 1 && <WebSocketIntercept />}
      </Box>
    </Box>
  );
}
