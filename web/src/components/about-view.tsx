'use client'

import { Button } from '@/components/button'
import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Strong, Text } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import { getSession } from '@/lib/api'
import { useTranslations } from 'next-intl'
import { useEffect, useState } from 'react'

type SessionState = {
  checked: boolean
  authenticated: boolean
}

function AboutCard({ eyebrow, title, body }: { eyebrow: string; title: string; body: string }) {
  return (
    <WorkspaceCard className="border-stone-200/80 bg-white/90 dark:border-white/10 dark:bg-zinc-900">
      <Eyebrow>{eyebrow}</Eyebrow>
      <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{title}</Subheading>
      <Text className="mt-3">{body}</Text>
    </WorkspaceCard>
  )
}

export function AboutView() {
  const t = useTranslations('aboutView')
  const [session, setSession] = useState<SessionState>({ checked: false, authenticated: false })

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const next = await getSession()
        if (!cancelled) {
          setSession({ checked: true, authenticated: next.authenticated })
        }
      } catch {
        if (!cancelled) {
          setSession({ checked: true, authenticated: false })
        }
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="space-y-8">
      <section className="overflow-hidden rounded-[2rem] border border-stone-200 bg-linear-to-br from-stone-100 via-white to-sky-50 p-8 shadow-sm ring-1 ring-black/3 dark:border-white/10 dark:from-zinc-900 dark:via-zinc-900 dark:to-sky-950/40 dark:ring-white/10">
        <div className="max-w-3xl">
          <Eyebrow className="tracking-[0.22em]">{t('heroEyebrow')}</Eyebrow>
          <Heading className="mt-4 text-4xl/11 sm:text-3xl/10">{t('heroTitle')}</Heading>
          <Text className="mt-4 text-lg/8 text-zinc-600 dark:text-zinc-300">
            {t('heroBody')}
          </Text>
          <div className="mt-6 flex flex-wrap gap-3">
            {session.authenticated ? (
              <>
                <Button href="/" color="dark/zinc">
                  {t('openCurrentAssignment')}
                </Button>
                <Button href="/progress" outline>
                  {t('viewProgress')}
                </Button>
              </>
            ) : (
              <>
                <Button href="/register" color="dark/zinc">
                  {t('register')}
                </Button>
                <Button href="/login" outline>
                  {t('signIn')}
                </Button>
              </>
            )}
          </div>
        </div>
      </section>

      <div className="grid gap-6 xl:grid-cols-[1.3fr_0.9fr]">
        <WorkspaceCard className="border-stone-200/80 bg-stone-50 dark:border-white/10 dark:bg-white/5">
          <Subheading className="text-xl/7 sm:text-lg/7">{t('skillsTitle')}</Subheading>
          <div className="mt-4 space-y-4">
            <Text>{t('skillsBody1')}</Text>
            <Text>{t('skillsBody2')}</Text>
            <Text>{t('skillsBody3')}</Text>
          </div>
        </WorkspaceCard>

        <WorkspaceCard className="border-stone-200/80 bg-zinc-950 text-white dark:border-white/10">
          <Subheading className="text-xl/7 text-white sm:text-lg/7">{t('whatYouGetTitle')}</Subheading>
          <ul className="mt-4 space-y-3 text-sm/6 text-zinc-300">
            <li>{t('whatYouGet1')}</li>
            <li>{t('whatYouGet2')}</li>
            <li>{t('whatYouGet3')}</li>
            <li>{t('whatYouGet4')}</li>
            <li>{t('whatYouGet5')}</li>
          </ul>
        </WorkspaceCard>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <AboutCard
          eyebrow={t('focusEyebrow')}
          title={t('focusTitle')}
          body={t('focusBody')}
        />
        <AboutCard
          eyebrow={t('feedbackEyebrow')}
          title={t('feedbackTitle')}
          body={t('feedbackBody')}
        />
        <AboutCard
          eyebrow={t('progressEyebrow')}
          title={t('progressTitle')}
          body={t('progressBody')}
        />
      </div>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading className="text-xl/7 sm:text-lg/7">{t('differentTitle')}</Subheading>
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('differentCard1Title')}</div>
            <Text className="mt-2">{t('differentCard1Body')}</Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('differentCard2Title')}</div>
            <Text className="mt-2">{t('differentCard2Body')}</Text>
          </div>
        </div>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-linear-to-br from-stone-50 via-white to-sky-50 dark:border-white/10 dark:from-zinc-900 dark:via-zinc-900 dark:to-sky-950/30">
        <Eyebrow>{t('howItWorksEyebrow')}</Eyebrow>
        <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{t('howItWorksTitle')}</Subheading>
        <div className="mt-4 max-w-3xl space-y-4">
          <Text>{t('howItWorksBody1')}</Text>
          <Text>{t('howItWorksBody2')}</Text>
          <Text>{t('howItWorksBody3')}</Text>
          <Text>{t('howItWorksBody4')}</Text>
        </div>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Eyebrow>{t('deterministicDetailEyebrow')}</Eyebrow>
        <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{t('deterministicDetailTitle')}</Subheading>
        <Text className="mt-3">{t('deterministicDetailIntro')}</Text>
        <div className="mt-6 grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('checksCardTitle')}</div>
            <Text className="mt-2">{t('checksCardBody')}</Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('styleCardTitle')}</div>
            <Text className="mt-2">{t('styleCardBody')}</Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('skillCardTitle')}</div>
            <Text className="mt-2">{t('skillCardBody')}</Text>
          </div>
        </div>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-linear-to-br from-stone-50 via-white to-sky-50 dark:border-white/10 dark:from-zinc-900 dark:via-zinc-900 dark:to-sky-950/30">
        <Eyebrow>{t('aiBoundaryEyebrow')}</Eyebrow>
        <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{t('aiBoundaryTitle')}</Subheading>
        <div className="mt-4 max-w-3xl space-y-4">
          <Text>{t('aiBoundaryBody1')}</Text>
          <Text>{t('aiBoundaryBody2')}</Text>
          <Text>{t('aiBoundaryBody3')}</Text>
        </div>

        <div className="mt-6 grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('step1Title')}</div>
            <Text className="mt-2">{t('step1Body')}</Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('step2Title')}</div>
            <Text className="mt-2">{t('step2Body')}</Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('step3Title')}</div>
            <Text className="mt-2">{t('step3Body')}</Text>
          </div>
        </div>
      </WorkspaceCard>
    </div>
  )
}
