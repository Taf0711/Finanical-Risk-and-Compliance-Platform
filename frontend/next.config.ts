import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
    NEXT_PUBLIC_WS_URL: process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080',
  },
  images: {
    domains: ['localhost'],
  },
  experimental: {
    optimizePackageImports: ['recharts', 'framer-motion'],
  },
};

export default nextConfig;
