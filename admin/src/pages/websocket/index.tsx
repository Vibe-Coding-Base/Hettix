import { Layout, Page } from "features/Layout";
import WebSocketLogs from "features/websocket/components/WebSocketLogs";

function WebSocketPage(): JSX.Element {
  return (
    <Layout page={Page.WebSocket} title="WebSocket">
      <WebSocketLogs />
    </Layout>
  );
}

export default WebSocketPage;
