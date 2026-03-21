import { EmptyState } from '@/components/status-state'
import { localeMessages } from '@/i18n/config'

export default function UnauthorizedPage() {
  const t = localeMessages.en.unauthorizedPage
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
