import { EmptyState } from '@/components/status-state'
import { cookies, headers } from 'next/headers'
import { getLocaleMessages, localeCookieName, resolveLocale } from '@/i18n/config'

export default async function UnauthorizedPage() {
  const cookieStore = await cookies()
  const headerStore = await headers()
  const locale = resolveLocale(cookieStore.get(localeCookieName)?.value, headerStore.get('accept-language'))
  const messages = await getLocaleMessages(locale)
  const t = messages.unauthorizedPage
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl items-center px-6 py-16">
      <div className="w-full">
        <EmptyState
          title={t.title}
          body={t.body}
          actions={[
            { href: '/login', label: t.signIn },
            { href: '/about', label: t.backToHome, outline: true },
          ]}
        />
      </div>
    </main>
  )
}
