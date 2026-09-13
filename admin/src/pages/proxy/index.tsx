import { Layout, Page } from "features/Layout";
import ProxySettings from "features/proxy/components/ProxySettings";

function Index(): JSX.Element {
  return (
    <Layout page={Page.ProxySetup} title="Proxy">
      <ProxySettings />
    </Layout>
  );
}

export default Index;
