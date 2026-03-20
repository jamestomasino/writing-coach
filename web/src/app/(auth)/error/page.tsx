import { EmptyState } from '@/components/status-state'
import { Text } from '@/components/text'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Account Error',
}

export default function ErrorPage({
  searchParams,
}: {
  searchParams: { id?: string }
}) {
  return (
    <div className="grid w-full max-w-lg grid-cols-1 gap-8">
      <EmptyState
        title="Account flow error"
        body="The account flow could not be completed. This can happen when a browser flow expires or a link is reused."
        actions={[
          { href: '/login', label: 'Sign in' },
          { href: '/register', label: 'Register', outline: true },
          { href: '/about', label: 'Back to home', outline: true },
        ]}
      />
      {searchParams.id ? (
        <Text className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 text-sm dark:border-white/10 dark:bg-white/5">
          Reference: {searchParams.id}
        </Text>
      ) : null}
    </div>
  )
}
