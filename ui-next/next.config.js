/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    ppr: true,
  },
  rewrites: () => {
    return [
      {
        source: "/",
        destination: "/dashboard",
      },
    ];
  },
};

module.exports = nextConfig;
