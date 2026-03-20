import { EmptyState } from '@/components/status-state'

export default function NotFound() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl items-center px-6 py-16">
      <div className="w-full">
        <EmptyState
          title="Page not found"
          body="The page you requested does not exist, or it is not available in the current coaching context."
          actions={[
            { href: '/', label: 'Current assignment' },
            { href: '/assignments', label: 'Past assignments', outline: true },
            { href: '/progress', label: 'Track progress', outline: true },
            { href: '/about', label: 'About', outline: true },
          ]}
        />
      </div>
    </main>
  )
}
