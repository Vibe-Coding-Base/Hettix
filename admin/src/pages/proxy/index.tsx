import ListIcon from "@mui/icons-material/List";
import { Button, Typography } from "@mui/material";
import { Link as RouterLink } from "react-router-dom";

import { Layout, Page } from "features/Layout";

function Index(): JSX.Element {
  return (
    <Layout page={Page.ProxySetup} title="Proxy setup">
      <Typography paragraph>Coming soon…</Typography>
      <Button
        component={RouterLink}
        to="/proxy/logs"
        variant="contained"
        color="primary"
        size="large"
        startIcon={<ListIcon />}
      >
        View logs
      </Button>
    </Layout>
  );
}

export default Index;
