import '@/styles/tailwind.css'
import type { Metadata } from 'next'
import { AppProviders } from '@/components/app-providers'
import { defaultLocale, localeMessages } from '@/i18n/config'

export const metadata: Metadata = {
  title: {
    template: '%s - Writing Coach',
    default: 'Writing Coach',
  },
  description: localeMessages.en.metadata.appDescription,
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
      lang={defaultLocale}
      className="bg-stone-100 text-zinc-950 antialiased dark:bg-zinc-950 dark:text-white"
    >
      <head>
        <link rel="preconnect" href="https://rsms.me/" />
        <link rel="stylesheet" href="https://rsms.me/inter/inter.css" />
      </head>
      <body>
        <AppProviders locale={defaultLocale} messages={localeMessages[defaultLocale]}>
          {children}
        </AppProviders>
      </body>
    </html>
  )
}
