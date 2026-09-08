import FolderIcon from "@mui/icons-material/Folder";
import TravelExploreIcon from "@mui/icons-material/TravelExplore";
import { Alert, Box, Button, Snackbar, Typography } from "@mui/material";
import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";

import { Layout, Page } from "features/Layout";
import { useLaunchBrowserMutation } from "lib/graphql/generated";

function Index(): JSX.Element {
  const highlightSx = { color: "primary.main" };
  const [launchBrowser, { loading }] = useLaunchBrowserMutation();
  const [error, setError] = useState<string | null>(null);

  const onLaunch = () => {
    launchBrowser().catch((e) => setError(e.message));
  };

  return (
    <Layout page={Page.Home} title="">
      <Snackbar open={error !== null} autoHideDuration={6000} onClose={() => setError(null)}>
        <Alert severity="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      </Snackbar>

      <Box p={4}>
        <Box mb={4} width="60%">
          <Typography variant="h2">
            <Box component="span" sx={highlightSx}>
              Hettix://
            </Box>
            <br />
            The simple HTTP toolkit for security research.
          </Typography>
        </Box>

        <Typography
          paragraph
          sx={{
            fontSize: "1.6rem",
            width: "60%",
            lineHeight: 2,
            mb: 5,
          }}
        >
          Welcome to{" "}
          <Box component="span" sx={highlightSx}>
            Hettix
          </Box>
          . Get started by creating a project.
        </Typography>

        <Button
          component={RouterLink}
          to="/projects"
          sx={{ mr: 2 }}
          variant="contained"
          color="primary"
          size="large"
          startIcon={<FolderIcon />}
        >
          Manage projects
        </Button>
        <Button
          variant="outlined"
          color="primary"
          size="large"
          onClick={onLaunch}
          disabled={loading}
          startIcon={<TravelExploreIcon />}
        >
          Open browser
        </Button>
      </Box>
    </Layout>
  );
}

export default Index;
