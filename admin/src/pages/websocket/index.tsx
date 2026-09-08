import { Layout, Page } from "features/Layout";
import WebSocketView from "features/websocket/components/WebSocketView";

function WebSocketPage(): JSX.Element {
  return (
    <Layout page={Page.WebSocket} title="WebSocket">
      <WebSocketView />
    </Layout>
  );
}

export default WebSocketPage;
