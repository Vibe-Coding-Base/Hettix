import { Box } from "@mui/material";

import EditRequest from "./EditRequest";
import History from "./History";

import SplitPane from "lib/components/SplitPane";

export default function Sender(): JSX.Element {
  return (
    <Box sx={{ height: "100%", position: "relative" }}>
      <SplitPane split="horizontal" size="30%">
        <Box sx={{ height: "100%", overflow: "auto" }}>
          <History />
        </Box>
        <Box sx={{ width: "100%", height: "100%", pt: "0.75rem" }}>
          <EditRequest />
        </Box>
      </SplitPane>
    </Box>
  );
}
