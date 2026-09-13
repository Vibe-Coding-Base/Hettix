import { RouteObject } from "react-router-dom";

import Assistant from "pages/assistant/index";
import Decoder from "pages/decoder/index";
import Findings from "pages/findings/index";
import Home from "pages/index";
import Intruder from "pages/intruder/index";
import MatchReplace from "pages/matchreplace/index";
import Plugins from "pages/plugins/index";
import Projects from "pages/projects/index";
import Proxy from "pages/proxy/index";
import Intercept from "pages/proxy/intercept/index";
import ProxyLogs from "pages/proxy/logs/index";
import Scope from "pages/scope/index";
import Sender from "pages/sender/index";
import Settings from "pages/settings/index";
import Sitemap from "pages/sitemap/index";
import WebSocket from "pages/websocket/index";
import Workflows from "pages/workflows/index";

export const routes: RouteObject[] = [
  { path: "/", element: <Home /> },
  { path: "/projects", element: <Projects /> },
  { path: "/proxy", element: <Proxy /> },
  { path: "/proxy/intercept", element: <Intercept /> },
  { path: "/proxy/logs", element: <ProxyLogs /> },
  { path: "/scope", element: <Scope /> },
  { path: "/sender", element: <Sender /> },
  { path: "/settings", element: <Settings /> },
  { path: "/assistant", element: <Assistant /> },
  { path: "/match-replace", element: <MatchReplace /> },
  { path: "/websocket", element: <WebSocket /> },
  { path: "/intruder", element: <Intruder /> },
  { path: "/sitemap", element: <Sitemap /> },
  { path: "/findings", element: <Findings /> },
  { path: "/workflows", element: <Workflows /> },
  { path: "/decoder", element: <Decoder /> },
  { path: "/plugins", element: <Plugins /> },
];
