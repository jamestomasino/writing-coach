const apiBase = process.env.WRITING_COACH_API_INTERNAL_URL ?? 'http://app:8080'
const kratosPublicBase = process.env.WRITING_COACH_KRATOS_PUBLIC_INTERNAL_URL ?? 'http://kratos:4433'
const kratosUIBase = process.env.WRITING_COACH_KRATOS_UI_INTERNAL_URL ?? 'http://kratos-selfservice-ui:3000'

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  async rewrites() {
    return [
      { source: '/api/:path*', destination: `${apiBase}/api/:path*` },
      { source: '/.ory/kratos/public/:path*', destination: `${kratosPublicBase}/:path*` },
      { source: '/.ory/kratos/ui/:path*', destination: `${kratosUIBase}/:path*` },
    ]
  },
}

export default nextConfig
