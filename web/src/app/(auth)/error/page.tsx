import { EmptyState } from '@/components/status-state'
import { cookies, headers } from 'next/headers'
import { getLocaleMessages, localeCookieName, resolveLocale } from '@/i18n/config'
import { Text } from '@/components/text'
import type { Metadata } from 'next'

export async function generateMetadata(): Promise<Metadata> {
  const cookieStore = await cookies()
  const headerStore = await headers()
  const locale = resolveLocale(cookieStore.get(localeCookieName)?.value, headerStore.get('accept-language'))
  const messages = await getLocaleMessages(locale)
  return {
    title: messages.authErrorPage.metadataTitle,
  }
}

export default async function ErrorPage({
  searchParams,
}: {
  searchParams: { id?: string }
}) {
  const cookieStore = await cookies()
  const headerStore = await headers()
  const locale = resolveLocale(cookieStore.get(localeCookieName)?.value, headerStore.get('accept-language'))
  const messages = await getLocaleMessages(locale)
  const t = messages.authErrorPage
  return (
    <div className="grid w-full max-w-lg grid-cols-1 gap-8">
      <EmptyState
        title={t.title}
        body={t.body}
        actions={[
          { href: '/login', label: t.signIn },
          { href: '/register', label: t.register, outline: true },
          { href: '/about', label: t.backToHome, outline: true },
        ]}
      />
      {searchParams.id ? (
        <Text className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 text-sm dark:border-white/10 dark:bg-white/5">
          {t.reference.replace('{id}', searchParams.id)}
        </Text>
      ) : null}
    </div>
  )
}
