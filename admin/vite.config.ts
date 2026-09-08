import { fileURLToPath, URL } from "node:url";

import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// @mui/icons-material ships no "exports" map, so a deep import like
// "@mui/icons-material/Folder" resolves to its CommonJS build, whose default
// export Vite does not unwrap (you get { default: Icon } instead of the icon).
// Alias deep icon imports to the ESM build, which has a clean `export default`.
const iconsEsm = fileURLToPath(new URL("./node_modules/@mui/icons-material/esm/", import.meta.url));

// The admin UI is a static SPA embedded into the Hettix binary. In development
// it proxies the GraphQL API to a locally running Hettix instance on :8080.
export default defineConfig({
  plugins: [react()],
  define: {
    // Keep in sync with `version` in cmd/hettix/hetty.go.
    "import.meta.env.VITE_VERSION": JSON.stringify("0.1.0"),
  },
  resolve: {
    alias: [
      { find: /^@mui\/icons-material\/([^/]+)$/, replacement: `${iconsEsm}$1` },
      { find: "lib", replacement: fileURLToPath(new URL("./src/lib", import.meta.url)) },
      { find: "features", replacement: fileURLToPath(new URL("./src/features", import.meta.url)) },
      { find: "pages", replacement: fileURLToPath(new URL("./src/pages", import.meta.url)) },
    ],
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
