import FolderOpenIcon from "@mui/icons-material/FolderOpen";
import TravelExploreIcon from "@mui/icons-material/TravelExplore";
import { Alert, Box, Button, List, ListItemButton, ListItemText, Paper, Snackbar, Typography } from "@mui/material";
import { useState } from "react";
import { Link as RouterLink } from "react-router-dom";

import NewProject from "features/projects/components/NewProject";
import useOpenProjectMutation from "features/projects/hooks/useOpenProjectMutation";
import { useActiveProject } from "lib/ActiveProjectContext";
import { useLaunchBrowserMutation, useProjectsQuery } from "lib/graphql/generated";

export default function Home(): JSX.Element {
  const activeProject = useActiveProject();
  const [launchBrowser, { loading: launching }] = useLaunchBrowserMutation();
  const [error, setError] = useState<string | null>(null);

  const onLaunch = () => launchBrowser().catch((e) => setError(e.message));

  const highlight = { color: "primary.main" };

  return (
    <Box p={4} sx={{ maxWidth: 900 }}>
      <Snackbar open={error !== null} autoHideDuration={6000} onClose={() => setError(null)}>
        <Alert severity="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      </Snackbar>

      <Typography variant="h3" sx={{ mb: 1 }}>
        <Box component="span" sx={highlight}>
          Hettix://
        </Box>
      </Typography>
      <Typography variant="h6" color="text.secondary" sx={{ mb: 4 }}>
        An HTTP toolkit for security research.
      </Typography>

      {activeProject ? (
        <Paper variant="outlined" sx={{ p: 3, mb: 3 }}>
          <Typography sx={{ mb: 2 }}>
            Project{" "}
            <Box component="span" sx={{ ...highlight, fontWeight: 600 }}>
              {activeProject.name}
            </Box>{" "}
            is open. Point your browser at the Hettix proxy, or open the built-in browser, and start testing.
          </Typography>
          <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap" }}>
            <Button variant="contained" startIcon={<TravelExploreIcon />} onClick={onLaunch} disabled={launching}>
              Open browser
            </Button>
            <Button variant="outlined" component={RouterLink} to="/proxy/logs">
              Go to proxy logs
            </Button>
          </Box>
        </Paper>
      ) : (
        <NoProjectOnboarding onLaunch={onLaunch} launching={launching} />
      )}
    </Box>
  );
}

function NoProjectOnboarding({ onLaunch, launching }: { onLaunch: () => void; launching: boolean }): JSX.Element {
  const { data } = useProjectsQuery({ fetchPolicy: "network-only" });
  const [openProject, { loading: opening }] = useOpenProjectMutation();
  const projects = data?.projects ?? [];

  return (
    <Box sx={{ display: "flex", gap: 3, flexWrap: "wrap" }}>
      <Paper variant="outlined" sx={{ p: 3, flex: "1 1 340px" }}>
        <Typography variant="h6" sx={{ mb: 1 }}>
          Get started
        </Typography>
        <Typography color="text.secondary" sx={{ mb: 3 }}>
          A project holds the traffic, scope, findings and settings for one session. Create one to begin — everything
          else unlocks once a project is open.
        </Typography>
        <NewProject />
      </Paper>

      {projects.length > 0 && (
        <Paper variant="outlined" sx={{ p: 3, flex: "1 1 300px" }}>
          <Typography variant="h6" sx={{ mb: 1 }}>
            Open a project
          </Typography>
          <List dense>
            {projects.map((project) => (
              <ListItemButton
                key={project.id}
                disabled={opening}
                onClick={() => openProject({ variables: { id: project.id } })}
              >
                <FolderOpenIcon sx={{ mr: 1.5, color: "text.secondary" }} fontSize="small" />
                <ListItemText primary={project.name} />
              </ListItemButton>
            ))}
          </List>
          <Button sx={{ mt: 2 }} startIcon={<TravelExploreIcon />} onClick={onLaunch} disabled={launching}>
            Open browser
          </Button>
        </Paper>
      )}
    </Box>
  );
}
