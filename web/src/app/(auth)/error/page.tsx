import { EmptyState } from '@/components/status-state'
import { localeMessages } from '@/i18n/config'
import { Text } from '@/components/text'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: localeMessages.en.authErrorPage.metadataTitle,
}

export default function ErrorPage({
  searchParams,
}: {
  searchParams: { id?: string }
}) {
  const t = localeMessages.en.authErrorPage
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
