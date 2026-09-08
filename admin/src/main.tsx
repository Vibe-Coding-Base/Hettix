import { ApolloProvider } from "@apollo/client";
import { ThemeProvider } from "@mui/material";
import CssBaseline from "@mui/material/CssBaseline";
import React from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider, createBrowserRouter } from "react-router-dom";

import { routes } from "./routes";

import "lib/monaco";

import { ActiveProjectProvider } from "lib/ActiveProjectContext";
import { InterceptedRequestsProvider } from "lib/InterceptedRequestsContext";
import { useApollo } from "lib/graphql/useApollo";
import theme from "lib/mui/theme";

import "./styles.css";

const router = createBrowserRouter(routes);

function Root(): JSX.Element {
  const apolloClient = useApollo();

  return (
    <ApolloProvider client={apolloClient}>
      <ActiveProjectProvider>
        <InterceptedRequestsProvider>
          <ThemeProvider theme={theme}>
            <CssBaseline />
            <RouterProvider router={router} />
          </ThemeProvider>
        </InterceptedRequestsProvider>
      </ActiveProjectProvider>
    </ApolloProvider>
  );
}

const container = document.getElementById("root");
if (!container) {
  throw new Error("Root element not found");
}

createRoot(container).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>
);
