import { Layout, Page } from "features/Layout";
import Assistant from "features/assistant/components/Assistant";

function AssistantPage(): JSX.Element {
  return (
    <Layout page={Page.Assistant} title="Assistant">
      <Assistant />
    </Layout>
  );
}

export default AssistantPage;
