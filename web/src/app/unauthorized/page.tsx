import { EmptyState } from '@/components/status-state'

export default function UnauthorizedPage() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl items-center px-6 py-16">
      <div className="w-full">
        <EmptyState
          title="Sign in required"
          body="Sign in to open this page."
          actions={[
            { href: '/login', label: 'Sign in' },
            { href: '/about', label: 'Back to home', outline: true },
          ]}
        />
      </div>
    </main>
  )
}
