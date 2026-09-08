import { Layout, Page } from "features/Layout";
import Decoder from "features/decoder/components/Decoder";

function DecoderPage(): JSX.Element {
  return (
    <Layout page={Page.Decoder} title="Decoder">
      <Decoder />
    </Layout>
  );
}

export default DecoderPage;
