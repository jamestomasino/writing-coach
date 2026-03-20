import '@/styles/tailwind.css'
import type { Metadata } from 'next'
import { AppProviders } from '@/components/app-providers'

export const metadata: Metadata = {
  title: {
    template: '%s - Writing Coach',
    default: 'Writing Coach',
  },
  description:
    'Writing Coach provides a skill-based coaching system for writers, combining focused objectives, progressive skill tracking, targeted prompts, and detailed review feedback.',
  icons: {
    icon: [
      { url: '/favicon-writing-coach.svg', type: 'image/svg+xml' },
      { url: '/favicon.ico' },
    ],
    shortcut: '/favicon-writing-coach.svg',
    apple: '/favicon-writing-coach.svg',
  },
}

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html
      lang="en"
      className="bg-stone-100 text-zinc-950 antialiased dark:bg-zinc-950 dark:text-white"
    >
      <head>
        <link rel="preconnect" href="https://rsms.me/" />
        <link rel="stylesheet" href="https://rsms.me/inter/inter.css" />
      </head>
      <body>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  )
}
