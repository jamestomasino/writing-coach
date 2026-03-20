import { EmptyState } from '@/components/status-state'

export default function UnauthorizedPage() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl items-center px-6 py-16">
      <div className="w-full">
        <EmptyState
          title="Sign in required"
          body="This page is only available inside an authenticated coaching session."
          actions={[
            { href: '/login', label: 'Sign in' },
            { href: '/about', label: 'Back to home', outline: true },
          ]}
        />
      </div>
    </main>
  )
}
