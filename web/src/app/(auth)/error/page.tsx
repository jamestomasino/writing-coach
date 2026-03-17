import { Heading } from '@/components/heading'
import { Text, TextLink } from '@/components/text'
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
      <div>
        <Heading>Account flow error</Heading>
        <Text className="mt-3">
          The account flow could not be completed. This can happen when a browser flow expires or a link is reused.
        </Text>
      </div>
      {searchParams.id ? (
        <Text className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 text-sm dark:border-white/10 dark:bg-white/5">
          Reference: {searchParams.id}
        </Text>
      ) : null}
      <Text>
        Try again from <TextLink href="/login">sign in</TextLink> or <TextLink href="/register">registration</TextLink>.
      </Text>
    </div>
  )
}
