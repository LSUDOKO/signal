import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // output standalone is intentionally omitted — Vercel manages its own build output
  images: {
    unoptimized: true,
  },
  typescript: {
    ignoreBuildErrors: true,
  },
};

export default nextConfig;

export default nextConfig;
