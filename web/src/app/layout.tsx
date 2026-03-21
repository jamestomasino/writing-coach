import '@/styles/tailwind.css'
import type { Metadata } from 'next'
import { cookies, headers } from 'next/headers'
import { AppProviders } from '@/components/app-providers'
import { getLocaleMessages, localeCookieName, resolveLocale } from '@/i18n/config'

export async function generateMetadata(): Promise<Metadata> {
  const cookieStore = await cookies()
  const headerStore = await headers()
  const locale = resolveLocale(cookieStore.get(localeCookieName)?.value, headerStore.get('accept-language'))
  const messages = await getLocaleMessages(locale)

  return {
    title: {
      template: '%s - Writing Coach',
      default: 'Writing Coach',
    },
    description: messages.metadata.appDescription,
    icons: {
      icon: [
        { url: '/favicon-writing-coach.svg', type: 'image/svg+xml' },
        { url: '/favicon.ico' },
      ],
      shortcut: '/favicon-writing-coach.svg',
      apple: '/favicon-writing-coach.svg',
    },
  }
}

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  const cookieStore = await cookies()
  const headerStore = await headers()
  const locale = resolveLocale(cookieStore.get(localeCookieName)?.value, headerStore.get('accept-language'))
  const messages = await getLocaleMessages(locale)

  return (
    <html lang={locale} className="bg-stone-100 text-zinc-950 antialiased dark:bg-zinc-950 dark:text-white">
      <head>
        <link rel="preconnect" href="https://rsms.me/" />
        <link rel="stylesheet" href="https://rsms.me/inter/inter.css" />
      </head>
      <body>
        <AppProviders locale={locale} messages={messages}>
          {children}
        </AppProviders>
      </body>
    </html>
  )
}
