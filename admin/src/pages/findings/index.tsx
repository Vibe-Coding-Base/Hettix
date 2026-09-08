import { Layout, Page } from "features/Layout";
import Findings from "features/findings/components/Findings";

function FindingsPage(): JSX.Element {
  return (
    <Layout page={Page.Findings} title="Findings">
      <Findings />
    </Layout>
  );
}

export default FindingsPage;
