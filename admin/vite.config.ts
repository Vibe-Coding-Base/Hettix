import { fileURLToPath, URL } from "node:url";

import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// The admin UI is a static SPA embedded into the Hettix binary. In development
// it proxies the GraphQL API to a locally running Hettix instance on :8080.
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      lib: fileURLToPath(new URL("./src/lib", import.meta.url)),
      features: fileURLToPath(new URL("./src/features", import.meta.url)),
      pages: fileURLToPath(new URL("./src/pages", import.meta.url)),
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    port: 3000,
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
});
