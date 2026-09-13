import AccountTreeIcon from "@mui/icons-material/AccountTree";
import AltRouteIcon from "@mui/icons-material/AltRoute";
import AutoAwesomeIcon from "@mui/icons-material/AutoAwesome";
import BugReportIcon from "@mui/icons-material/BugReport";
import ChevronLeftIcon from "@mui/icons-material/ChevronLeft";
import ChevronRightIcon from "@mui/icons-material/ChevronRight";
import ExtensionIcon from "@mui/icons-material/Extension";
import FindReplaceIcon from "@mui/icons-material/FindReplace";
import FolderIcon from "@mui/icons-material/Folder";
import FormatListBulletedIcon from "@mui/icons-material/FormatListBulleted";
import GpsFixedIcon from "@mui/icons-material/GpsFixed";
import HomeIcon from "@mui/icons-material/Home";
import HubIcon from "@mui/icons-material/Hub";
import LocationSearchingIcon from "@mui/icons-material/LocationSearching";
import MenuIcon from "@mui/icons-material/Menu";
import SendIcon from "@mui/icons-material/Send";
import SettingsIcon from "@mui/icons-material/Settings";
import SettingsEthernetIcon from "@mui/icons-material/SettingsEthernet";
import SwapHorizIcon from "@mui/icons-material/SwapHoriz";
import TransformIcon from "@mui/icons-material/Transform";
import {
  Theme,
  useTheme,
  Toolbar,
  IconButton,
  Typography,
  Divider,
  List,
  Tooltip,
  styled,
  CSSObject,
  Box,
  ListItemText,
  Badge,
} from "@mui/material";
import MuiAppBar, { AppBarProps as MuiAppBarProps } from "@mui/material/AppBar";
import MuiDrawer from "@mui/material/Drawer";
import MuiListItemButton, { ListItemButtonProps } from "@mui/material/ListItemButton";
import MuiListItemIcon, { ListItemIconProps } from "@mui/material/ListItemIcon";
import React, { useState } from "react";
import { Link as RouterLink } from "react-router-dom";

import { useActiveProject } from "lib/ActiveProjectContext";
import { useInterceptedRequests } from "lib/InterceptedRequestsContext";
import { AppMenuBar } from "lib/components/AppMenuBar";
import { UpdateBanner } from "lib/components/UpdateBanner";
import { WindowControls } from "lib/components/WindowControls";
import { dragStyle, isDesktop, isFramelessDesktop, noDragStyle } from "lib/desktop";

export enum Page {
  Home,
  GetStarted,
  Intercept,
  Projects,
  ProxySetup,
  ProxyLogs,
  Sender,
  Scope,
  Settings,
  Assistant,
  MatchReplace,
  WebSocket,
  Intruder,
  Sitemap,
  Findings,
  Workflows,
  Decoder,
  Plugins,
}

interface NavItem {
  page: Page;
  to: string;
  label: string;
  icon: JSX.Element;
  // project is true for pages that need an active project (disabled otherwise).
  project?: boolean;
}

const drawerWidth = 240;

// hideScrollbar keeps the nav scrollable without painting a scrollbar, which
// looked out of place on the dark drawer.
const hideScrollbar: CSSObject = {
  overflowY: "auto",
  scrollbarWidth: "none",
  "&::-webkit-scrollbar": { display: "none" },
};

const openedMixin = (theme: Theme): CSSObject => ({
  width: drawerWidth,
  transition: theme.transitions.create("width", {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.enteringScreen,
  }),
  overflowX: "hidden",
  ...hideScrollbar,
});

const closedMixin = (theme: Theme): CSSObject => ({
  transition: theme.transitions.create("width", {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  overflowX: "hidden",
  width: 56,
  ...hideScrollbar,
});

const DrawerHeader = styled("div")(({ theme }) => ({
  display: "flex",
  alignItems: "center",
  justifyContent: "flex-start",
  padding: theme.spacing(0, 1),
  // necessary for content to be below app bar
  ...theme.mixins.toolbar,
}));

interface AppBarProps extends MuiAppBarProps {
  open?: boolean;
}

const AppBar = styled(MuiAppBar, {
  shouldForwardProp: (prop) => prop !== "open",
})<AppBarProps>(({ theme, open }) => ({
  backgroundColor: theme.palette.secondary.dark,
  zIndex: theme.zIndex.drawer + 1,
  transition: theme.transitions.create(["width", "margin"], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  ...(open && {
    marginLeft: drawerWidth,
    width: `calc(100% - ${drawerWidth}px)`,
    transition: theme.transitions.create(["width", "margin"], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  }),
}));

const Drawer = styled(MuiDrawer, { shouldForwardProp: (prop) => prop !== "open" })(({ theme, open }) => ({
  width: drawerWidth,
  flexShrink: 0,
  whiteSpace: "nowrap",
  boxSizing: "border-box",
  ...(open && {
    ...openedMixin(theme),
    "& .MuiDrawer-paper": openedMixin(theme),
  }),
  ...(!open && {
    ...closedMixin(theme),
    "& .MuiDrawer-paper": closedMixin(theme),
  }),
}));

const ListItemButton = styled(MuiListItemButton)<ListItemButtonProps & { component?: React.ElementType; to?: string }>(
  ({ theme }) => ({
    [theme.breakpoints.up("sm")]: {
      px: 1,
    },
    "&.MuiListItemButton-root": {
      // The nav items render as router links (anchors); without an explicit
      // color the text inherits the browser's default link color (purple once
      // visited), which is illegible on the dark drawer.
      color: theme.palette.text.primary,
      "& .MuiListItemText-root": {
        color: theme.palette.text.primary,
      },
      "&.Mui-selected": {
        backgroundColor: theme.palette.primary.main,
        "& .MuiListItemIcon-root": {
          color: theme.palette.secondary.dark,
        },
        "& .MuiListItemText-root": {
          color: theme.palette.secondary.dark,
        },
      },
    },
  })
);

const ListItemIcon = styled(MuiListItemIcon)<ListItemIconProps>(() => ({
  minWidth: 42,
}));

interface Props {
  children: React.ReactNode;
  title: string;
  page: Page;
}

export function Layout({ title, page, children }: Props): JSX.Element {
  const activeProject = useActiveProject();
  const interceptedRequests = useInterceptedRequests();
  const theme = useTheme();
  const [open, setOpen] = useState(true);

  const handleDrawerOpen = () => {
    setOpen(true);
  };

  const handleDrawerClose = () => {
    setOpen(false);
  };

  // Grouped by pentest workflow: observe traffic, then attack, then analyze,
  // then manage. Groups are separated by dividers in the drawer.
  const navGroups: NavItem[][] = [
    [{ page: Page.Home, to: "/", label: "Home", icon: <HomeIcon /> }],
    [
      { page: Page.Scope, to: "/scope", label: "Scope", icon: <LocationSearchingIcon />, project: true },
      { page: Page.ProxyLogs, to: "/proxy/logs", label: "Proxy logs", icon: <FormatListBulletedIcon />, project: true },
      {
        page: Page.Intercept,
        to: "/proxy/intercept",
        label: "Intercept",
        icon: (
          <Badge color="error" badgeContent={interceptedRequests?.length || 0}>
            <AltRouteIcon />
          </Badge>
        ),
        project: true,
      },
      { page: Page.Sitemap, to: "/sitemap", label: "Sitemap", icon: <AccountTreeIcon />, project: true },
      { page: Page.WebSocket, to: "/websocket", label: "WebSocket", icon: <SwapHorizIcon />, project: true },
    ],
    [
      { page: Page.Sender, to: "/sender", label: "Sender", icon: <SendIcon />, project: true },
      { page: Page.Intruder, to: "/intruder", label: "Intruder", icon: <GpsFixedIcon />, project: true },
      {
        page: Page.MatchReplace,
        to: "/match-replace",
        label: "Match & Replace",
        icon: <FindReplaceIcon />,
        project: true,
      },
      { page: Page.Workflows, to: "/workflows", label: "Workflows", icon: <HubIcon />, project: true },
      { page: Page.Decoder, to: "/decoder", label: "Decoder", icon: <TransformIcon /> },
    ],
    [
      { page: Page.Findings, to: "/findings", label: "Findings", icon: <BugReportIcon />, project: true },
      { page: Page.Assistant, to: "/assistant", label: "Assistant", icon: <AutoAwesomeIcon />, project: true },
    ],
    [
      { page: Page.Projects, to: "/projects", label: "Projects", icon: <FolderIcon /> },
      { page: Page.ProxySetup, to: "/proxy", label: "Proxy", icon: <SettingsEthernetIcon /> },
      { page: Page.Plugins, to: "/plugins", label: "Plugins", icon: <ExtensionIcon /> },
      { page: Page.Settings, to: "/settings", label: "Settings", icon: <SettingsIcon /> },
    ],
  ];

  const SiteTitle = styled("span")({
    ...(title !== "" && {
      color: theme.palette.primary.main,
      marginRight: 4,
    }),
  });

  return (
    <Box sx={{ display: "flex", height: "100%" }}>
      <AppBar position="fixed" open={open}>
        <Toolbar
          disableGutters={isFramelessDesktop}
          style={isFramelessDesktop ? dragStyle : undefined}
          sx={isFramelessDesktop ? { pl: 1.5 } : undefined}
        >
          <IconButton
            color="inherit"
            aria-label="Open drawer"
            onClick={handleDrawerOpen}
            edge="start"
            style={isFramelessDesktop ? noDragStyle : undefined}
            sx={{
              mr: isDesktop ? 1 : 5,
              ...(open && { display: "none" }),
            }}
          >
            <MenuIcon />
          </IconButton>
          {isDesktop && <AppMenuBar />}
          <Box
            sx={{
              display: "flex",
              justifyContent: "space-around",
              alignItems: "center",
              width: "100%",
              ...(isDesktop && { ml: 3 }),
            }}
          >
            <Typography variant="h5" noWrap sx={{ width: "100%" }}>
              <SiteTitle>Hettix://</SiteTitle>
              {title}
            </Typography>
            <Box sx={{ flexShrink: 0, pt: 0.75 }}>v{import.meta.env.VITE_VERSION || "0.0"}</Box>
          </Box>
          {isFramelessDesktop && <WindowControls />}
        </Toolbar>
      </AppBar>
      <Drawer variant="permanent" open={open}>
        <DrawerHeader>
          <IconButton onClick={handleDrawerClose}>
            {theme.direction === "rtl" ? <ChevronRightIcon /> : <ChevronLeftIcon />}
          </IconButton>
        </DrawerHeader>
        <Divider />
        {navGroups.map((group, groupIndex) => (
          <List sx={{ p: 0 }} key={groupIndex}>
            {groupIndex > 0 && <Divider />}
            {group.map((item) => {
              const disabled = item.project === true && !activeProject;

              return (
                <ListItemButton
                  component={RouterLink}
                  to={item.to}
                  key={item.to}
                  disabled={disabled}
                  selected={page === item.page}
                >
                  <Tooltip title={item.label} placement="right">
                    <ListItemIcon>{item.icon}</ListItemIcon>
                  </Tooltip>
                  <ListItemText primary={item.label} />
                </ListItemButton>
              );
            })}
          </List>
        ))}
      </Drawer>
      <Box component="main" sx={{ flexGrow: 1, mx: 3, mt: 11 }}>
        {children}
      </Box>
      <UpdateBanner />
    </Box>
  );
}
