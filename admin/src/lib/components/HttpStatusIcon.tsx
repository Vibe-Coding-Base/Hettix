import FiberManualRecordIcon from "@mui/icons-material/FiberManualRecord";
import { SvgIconTypeMap } from "@mui/material";

interface Props {
  status: number;
}

export default function HttpStatusIcon({ status }: Props): JSX.Element {
  // Status colors are semantic and must not follow the (red) brand primary:
  // 2xx success/green, 3xx info/blue, 4xx warning/orange, 5xx error/red.
  let color: SvgIconTypeMap["props"]["color"] = "inherit";

  switch (Math.floor(status / 100)) {
    case 2:
      color = "success";
      break;
    case 3:
      color = "info";
      break;
    case 4:
      color = "warning";
      break;
    case 5:
      color = "error";
      break;
  }

  return <FiberManualRecordIcon sx={{ marginTop: "-.25rem", verticalAlign: "middle" }} color={color} />;
}
