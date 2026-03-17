const apiBase = process.env.WRITING_COACH_API_INTERNAL_URL ?? 'http://app:8080'
const kratosPublicBase = process.env.WRITING_COACH_KRATOS_PUBLIC_INTERNAL_URL ?? 'http://kratos:4433'

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  async rewrites() {
    return [
      { source: '/api/:path*', destination: `${apiBase}/api/:path*` },
      { source: '/.ory/kratos/public/:path*', destination: `${kratosPublicBase}/:path*` },
    ]
  },
}

export default nextConfig
