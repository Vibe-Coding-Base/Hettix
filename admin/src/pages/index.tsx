import { Layout, Page } from "features/Layout";
import Home from "features/home/components/Home";

function Index(): JSX.Element {
  return (
    <Layout page={Page.Home} title="">
      <Home />
    </Layout>
  );
}

export default Index;
