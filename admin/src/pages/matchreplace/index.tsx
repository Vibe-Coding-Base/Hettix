import { Layout, Page } from "features/Layout";
import MatchReplace from "features/matchreplace/components/MatchReplace";

function MatchReplacePage(): JSX.Element {
  return (
    <Layout page={Page.MatchReplace} title="Match & Replace">
      <MatchReplace />
    </Layout>
  );
}

export default MatchReplacePage;
