import CloseIcon from "@mui/icons-material/Close";
import CropSquareIcon from "@mui/icons-material/CropSquare";
import FilterNoneIcon from "@mui/icons-material/FilterNone";
import MinimizeIcon from "@mui/icons-material/Minimize";
import { Box, IconButton } from "@mui/material";
import { useEffect, useState } from "react";

import { desktopWindow, noDragStyle } from "lib/desktop";

const buttonSx = {
  height: "100%",
  width: 46,
  borderRadius: 0,
  color: "inherit",
} as const;

// WindowControls renders Windows-style minimise, maximise/restore and close
// caption buttons for the frameless desktop window.
export function WindowControls(): JSX.Element {
  const [maximised, setMaximised] = useState(false);

  useEffect(() => {
    let active = true;
    const sync = () => {
      desktopWindow.isMaximised().then((m) => {
        if (active) {
          setMaximised(m);
        }
      });
    };
    sync();
    window.addEventListener("resize", sync);
    return () => {
      active = false;
      window.removeEventListener("resize", sync);
    };
  }, []);

  return (
    <Box sx={{ display: "flex", alignSelf: "stretch", ml: 1 }} style={noDragStyle}>
      <IconButton aria-label="Minimise" sx={buttonSx} onClick={() => desktopWindow.minimise()}>
        <MinimizeIcon sx={{ fontSize: 18, position: "relative", top: -4 }} />
      </IconButton>
      <IconButton
        aria-label={maximised ? "Restore" : "Maximise"}
        sx={buttonSx}
        onClick={() => desktopWindow.toggleMaximise()}
      >
        {maximised ? <FilterNoneIcon sx={{ fontSize: 15 }} /> : <CropSquareIcon sx={{ fontSize: 18 }} />}
      </IconButton>
      <IconButton
        aria-label="Close"
        sx={{ ...buttonSx, "&:hover": { backgroundColor: "#c42b1c", color: "#fff" } }}
        onClick={() => desktopWindow.quit()}
      >
        <CloseIcon sx={{ fontSize: 18 }} />
      </IconButton>
    </Box>
  );
}
