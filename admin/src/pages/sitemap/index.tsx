import { Layout, Page } from "features/Layout";
import Sitemap from "features/sitemap/components/Sitemap";

function SitemapPage(): JSX.Element {
  return (
    <Layout page={Page.Sitemap} title="Sitemap">
      <Sitemap />
    </Layout>
  );
}

export default SitemapPage;
