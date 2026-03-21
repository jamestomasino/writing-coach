import { EmptyState } from '@/components/status-state'
import { cookies, headers } from 'next/headers'
import { getLocaleMessages, localeCookieName, resolveLocale } from '@/i18n/config'

export default async function NotFound() {
  const cookieStore = await cookies()
  const headerStore = await headers()
  const locale = resolveLocale(cookieStore.get(localeCookieName)?.value, headerStore.get('accept-language'))
  const messages = await getLocaleMessages(locale)
  const t = messages.notFoundPage
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl items-center px-6 py-16">
      <div className="w-full">
        <EmptyState
          title={t.title}
          body={t.body}
          actions={[
            { href: '/', label: t.currentAssignment },
            { href: '/assignments', label: t.pastAssignments, outline: true },
            { href: '/progress', label: t.progress, outline: true },
            { href: '/about', label: t.about, outline: true },
          ]}
        />
      </div>
    </main>
  )
}
