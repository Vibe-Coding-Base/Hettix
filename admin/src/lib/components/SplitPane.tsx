import { alpha, styled } from "@mui/material/styles";
import { Allotment } from "allotment";
import React, { Children } from "react";

import "allotment/dist/style.css";

// `react-split-pane` is unmaintained and only supports React 16, so the split
// view is built on `allotment` instead. This component keeps the original
// `react-split-pane` API (a `split` direction, a `size` for the first pane, and
// two children) so call sites did not have to change.
export interface SplitPaneProps {
  split: "horizontal" | "vertical";
  size?: string | number;
  children: React.ReactNode;
}

// `react-split-pane` positioned itself absolutely within the nearest positioned
// ancestor. Call sites rely on that to fill their container, so it's preserved.
const StyledAllotment = styled(Allotment)(({ theme }) => ({
  position: "absolute",
  inset: 0,
  "--focus-border": theme.palette.primary.main,
  "--separator-border": alpha(theme.palette.grey[400], 0.15),
}));

export default function SplitPane({ split, size, children }: SplitPaneProps): JSX.Element {
  // `split="horizontal"` stacks the panes top/bottom, which is Allotment's
  // `vertical` orientation.
  const vertical = split === "horizontal";

  // Call sites may render a falsy second child (e.g. a response that isn't
  // there yet), so only real children become panes.
  const panes = Children.toArray(children);

  return (
    <StyledAllotment vertical={vertical}>
      {panes.map((child, index) => (
        <Allotment.Pane key={index} preferredSize={index === 0 ? size : undefined}>
          {child}
        </Allotment.Pane>
      ))}
    </StyledAllotment>
  );
}
