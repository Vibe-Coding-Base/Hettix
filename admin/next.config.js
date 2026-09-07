// @ts-check

const { PHASE_DEVELOPMENT_SERVER } = require("next/constants");

/**
 * The admin UI ships as a static bundle that gets embedded into the `hetty`
 * binary, so production builds use `output: "export"`. Rewrites are not
 * supported in export mode, so the dev-only proxy to the GraphQL API is
 * applied for `next dev` only.
 *
 * @param {string} phase
 * @returns {import('next').NextConfig}
 */
module.exports = (phase) => {
  const isDev = phase === PHASE_DEVELOPMENT_SERVER;

  /** @type {import('next').NextConfig} */
  const nextConfig = {
    reactStrictMode: true,
    trailingSlash: true,
  };

  if (isDev) {
    return {
      ...nextConfig,
      async rewrites() {
        return [
          {
            source: "/api/:path/",
            destination: "http://localhost:8080/api/:path/",
          },
        ];
      },
    };
  }

  return {
    ...nextConfig,
    // Replaces the `next export` command, which was removed in Next.js 14.
    output: "export",
  };
};
