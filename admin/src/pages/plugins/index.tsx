import { Layout, Page } from "features/Layout";
import Plugins from "features/plugins/components/Plugins";

function PluginsPage(): JSX.Element {
  return (
    <Layout page={Page.Plugins} title="Plugins">
      <Plugins />
    </Layout>
  );
}

export default PluginsPage;
