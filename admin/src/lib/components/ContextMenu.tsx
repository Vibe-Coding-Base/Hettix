import { Divider, ListItemIcon, ListItemText, Menu, MenuItem } from "@mui/material";
import { useState } from "react";
import type { ReactNode } from "react";

// A reusable right-click context menu, so any list or row can offer Burp-style
// actions ("Send to Sender", "Copy URL", …). Items are supplied per open() call
// so each row can present actions for its own data.
export interface ContextMenuItem {
  label: string;
  onClick: () => void;
  icon?: ReactNode;
  disabled?: boolean;
  divider?: boolean;
}

interface State {
  x: number;
  y: number;
  items: ContextMenuItem[];
}

export interface ContextMenu {
  // open shows the menu at the cursor with the given items; call it from an
  // element's onContextMenu handler.
  open: (event: React.MouseEvent, items: ContextMenuItem[]) => void;
  close: () => void;
  // menu is the element to render once anywhere in the component tree.
  menu: JSX.Element;
}

export function useContextMenu(): ContextMenu {
  const [state, setState] = useState<State | null>(null);

  const open = (event: React.MouseEvent, items: ContextMenuItem[]) => {
    if (items.length === 0) {
      return;
    }
    event.preventDefault();
    setState({ x: event.clientX, y: event.clientY, items });
  };

  const close = () => setState(null);

  const menu = (
    <Menu
      open={state !== null}
      onClose={close}
      anchorReference="anchorPosition"
      anchorPosition={state ? { top: state.y, left: state.x } : undefined}
    >
      {state?.items.flatMap((item, i) => {
        const entry = (
          <MenuItem
            key={i}
            disabled={item.disabled}
            onClick={() => {
              item.onClick();
              close();
            }}
          >
            {item.icon && <ListItemIcon>{item.icon}</ListItemIcon>}
            <ListItemText>{item.label}</ListItemText>
          </MenuItem>
        );
        return item.divider ? [<Divider key={`d${i}`} />, entry] : [entry];
      })}
    </Menu>
  );

  return { open, close, menu };
}
