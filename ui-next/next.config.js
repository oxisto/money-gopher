const createNextIntlPlugin = require("next-intl/plugin");

const withNextIntl = createNextIntlPlugin();

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

module.exports = withNextIntl(nextConfig);
