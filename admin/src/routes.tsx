import { RouteObject } from "react-router-dom";

import Assistant from "pages/assistant/index";
import Home from "pages/index";
import Projects from "pages/projects/index";
import Proxy from "pages/proxy/index";
import Intercept from "pages/proxy/intercept/index";
import ProxyLogs from "pages/proxy/logs/index";
import Scope from "pages/scope/index";
import Sender from "pages/sender/index";
import Settings from "pages/settings/index";

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
];
