import '@/styles/tailwind.css'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: {
    template: '%s - Writing Coach',
    default: 'Writing Coach',
  },
  description: 'Assignment-first writing coaching built on a progressive writing skill map.',
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
      <body>{children}</body>
    </html>
  )
}
