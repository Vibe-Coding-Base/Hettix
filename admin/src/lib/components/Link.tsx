import MuiLink, { LinkProps as MuiLinkProps } from "@mui/material/Link";
import clsx from "clsx";
import * as React from "react";
import { Link as RouterLink, useLocation } from "react-router-dom";

export type LinkProps = {
  activeClassName?: string;
  href: string;
} & Omit<MuiLinkProps, "href">;

// A router-aware link that renders an MUI-styled anchor for internal routes and
// a plain MUI link for external ones. It sets `activeClassName` when the current
// location matches the target path.
const Link = React.forwardRef<HTMLAnchorElement, LinkProps>(function Link(props, ref) {
  const { activeClassName = "active", className: classNameProps, href, role: _role, ...other } = props;

  const location = useLocation();
  const pathname = href.split("?")[0];
  const className = clsx(classNameProps, {
    [activeClassName]: location.pathname === pathname && activeClassName,
  });

  const isExternal = href.indexOf("http") === 0 || href.indexOf("mailto:") === 0;

  if (isExternal) {
    return <MuiLink className={className} href={href} ref={ref} {...other} />;
  }

  return <MuiLink component={RouterLink} to={href} className={className} ref={ref} {...other} />;
});

export default Link;
