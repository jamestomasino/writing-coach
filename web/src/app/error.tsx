'use client'

import { useEffect } from 'react'
import { Button } from '@/components/button'
import { Heading } from '@/components/heading'
import { Text } from '@/components/text'

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl items-center px-6 py-16">
      <div className="w-full rounded-2xl border border-dashed border-stone-300 bg-stone-50 p-10 text-center dark:border-white/10 dark:bg-white/5">
        <Heading level={2}>Something went wrong</Heading>
        <Text className="mx-auto mt-3 max-w-2xl">The app hit an unexpected error while rendering this page.</Text>
        <div className="mt-6 flex flex-wrap justify-center gap-3">
          <Button onClick={reset} outline>
            Try again
          </Button>
          <Button href="/" color="dark/zinc">
            Current assignment
          </Button>
          <Button href="/assignments" outline>
            Past assignments
          </Button>
          <Button href="/about" outline>
            About
          </Button>
        </div>
      </div>
    </main>
  )
}
