import { EmptyState } from '@/components/status-state'
import { localeMessages } from '@/i18n/config'

export default function NotFound() {
  const t = localeMessages.en.notFoundPage
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
