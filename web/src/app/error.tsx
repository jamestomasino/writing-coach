'use client'

import { useEffect } from 'react'
import { Button } from '@/components/button'
import { Heading } from '@/components/heading'
import { Text } from '@/components/text'
import { useTranslations } from 'next-intl'

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  const t = useTranslations('globalErrorPage')
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl items-center px-6 py-16">
      <div className="w-full rounded-2xl border border-dashed border-stone-300 bg-stone-50 p-10 text-center dark:border-white/10 dark:bg-white/5">
        <Heading level={2}>{t('title')}</Heading>
        <Text className="mx-auto mt-3 max-w-2xl">{t('body')}</Text>
        <div className="mt-6 flex flex-wrap justify-center gap-3">
          <Button onClick={reset} outline>
            {t('tryAgain')}
          </Button>
          <Button href="/" color="dark/zinc">
            {t('currentAssignment')}
          </Button>
          <Button href="/assignments" outline>
            {t('pastAssignments')}
          </Button>
          <Button href="/about" outline>
            {t('about')}
          </Button>
        </div>
      </div>
    </main>
  )
}
