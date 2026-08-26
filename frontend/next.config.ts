import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    const internalApi = process.env.INTERNAL_API_URL || "http://caelus-api:8080";
    return [
      {
        source: "/api/:path*",
        destination: `${internalApi}/api/:path*`,
      },
      {
        source: "/install.sh",
        destination: `${internalApi}/install.sh`,
      },
      {
        source: "/agent-bin",
        destination: `${internalApi}/agent-bin`,
      },
      {
        source: "/health",
        destination: `${internalApi}/health`,
      },
    ];
  },
};

export default nextConfig;
