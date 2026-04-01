'use client'

import { Button } from '@/components/button'
import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Strong, Text, TextLink } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import { getSession } from '@/lib/api'
import {
  ArrowPathIcon,
  ChartBarSquareIcon,
  ChatBubbleLeftRightIcon,
  ClipboardDocumentListIcon,
  LinkIcon,
  MapIcon,
  Squares2X2Icon,
} from '@heroicons/react/20/solid'
import { useTranslations } from 'next-intl'
import { useEffect, useState } from 'react'

type SessionState = {
  checked: boolean
  authenticated: boolean
}

function sectionIdFromTitle(title: string): string {
  return title
    .normalize('NFKD')
    .replace(/\p{Mark}+/gu, '')
    .toLowerCase()
    .trim()
    .replace(/['’]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

const openSourceCheckTools = [
  {
    titleKey: 'ossHeuristicTitle',
    bodyKey: 'ossHeuristicBody',
    href: '',
    licenseKey: '',
  },
  {
    titleKey: 'ossValeTitle',
    bodyKey: 'ossValeBody',
    href: 'https://vale.sh/',
    licenseKey: 'ossValeLicense',
  },
  {
    titleKey: 'ossLanguageToolTitle',
    bodyKey: 'ossLanguageToolBody',
    href: 'https://languagetool.org/',
    licenseKey: 'ossLanguageToolLicense',
  },
  {
    titleKey: 'ossSpacyTitle',
    bodyKey: 'ossSpacyBody',
    href: 'https://spacy.io/',
    licenseKey: 'ossSpacyLicense',
  },
  {
    titleKey: 'ossTextDescriptivesTitle',
    bodyKey: 'ossTextDescriptivesBody',
    href: 'https://hlasse.github.io/TextDescriptives/',
    licenseKey: 'ossTextDescriptivesLicense',
  },
  {
    titleKey: 'ossCoreNLPTitle',
    bodyKey: 'ossCoreNLPBody',
    href: 'https://stanfordnlp.github.io/CoreNLP/',
    licenseKey: 'ossCoreNLPLicense',
  },
] as const

const openSourceCredits = [
  {
    titleKey: 'creditReactTitle',
    href: 'https://react.dev/',
    licenseKey: 'creditReactLicense',
  },
  {
    titleKey: 'creditNextTitle',
    href: 'https://nextjs.org/',
    licenseKey: 'creditNextLicense',
  },
  {
    titleKey: 'creditTailwindTitle',
    href: 'https://tailwindcss.com/',
    licenseKey: 'creditTailwindLicense',
  },
  {
    titleKey: 'creditHeadlessUITitle',
    href: 'https://headlessui.com/',
    licenseKey: 'creditHeadlessUILicense',
  },
  {
    titleKey: 'creditHeroiconsTitle',
    href: 'https://heroicons.com/',
    licenseKey: 'creditHeroiconsLicense',
  },
  {
    titleKey: 'creditInterTitle',
    href: 'https://rsms.me/inter/',
    licenseKey: 'creditInterLicense',
  },
  {
    titleKey: 'creditValeTitle',
    href: 'https://vale.sh/',
    licenseKey: 'creditValeLicense',
  },
  {
    titleKey: 'creditLanguageToolTitle',
    href: 'https://languagetool.org/',
    licenseKey: 'creditLanguageToolLicense',
  },
  {
    titleKey: 'creditSpacyTitle',
    href: 'https://spacy.io/',
    licenseKey: 'creditSpacyLicense',
  },
  {
    titleKey: 'creditTextDescriptivesTitle',
    href: 'https://hlasse.github.io/TextDescriptives/',
    licenseKey: 'creditTextDescriptivesLicense',
  },
  {
    titleKey: 'creditCoreNLPTitle',
    href: 'https://stanfordnlp.github.io/CoreNLP/',
    licenseKey: 'creditCoreNLPLicense',
  },
  {
    titleKey: 'creditKratosTitle',
    href: 'https://github.com/ory/kratos',
    licenseKey: 'creditKratosLicense',
  },
  {
    titleKey: 'creditSQLiteTitle',
    href: 'https://www.sqlite.org/',
    licenseKey: 'creditSQLiteLicense',
  },
  {
    titleKey: 'creditTailwindPlusTitle',
    href: 'https://tailwindcss.com/plus',
    licenseKey: 'creditTailwindPlusLicense',
  },
  {
    titleKey: 'creditWritingCoachTitle',
    href: 'https://github.com/tomasino/writing-coach',
    licenseKey: 'creditWritingCoachLicense',
  },
] as const

function AboutCard({
  id,
  eyebrow,
  title,
  body,
}: {
  id: string
  eyebrow: string
  title: string
  body: string
}) {
  return (
    <WorkspaceCard
      id={id}
      className="border-stone-200/80 bg-white/90 dark:border-white/10 dark:bg-zinc-900"
    >
      <Eyebrow>{eyebrow}</Eyebrow>
      <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{title}</Subheading>
      <Text className="mt-3">{body}</Text>
    </WorkspaceCard>
  )
}

export function AboutView() {
  const t = useTranslations('aboutView')
  const [session, setSession] = useState<SessionState>({ checked: false, authenticated: false })
  const heroTitle = t('heroTitle')
  const skillsTitle = t('skillsTitle')
  const whatYouGetTitle = t('whatYouGetTitle')
  const focusTitle = t('focusTitle')
  const feedbackTitle = t('feedbackTitle')
  const progressTitle = t('progressTitle')
  const differentTitle = t('differentTitle')
  const deterministicDetailTitle = t('deterministicDetailTitle')
  const ossToolsTitle = t('ossToolsTitle')
  const aiBoundaryTitle = t('aiBoundaryTitle')
  const openSourceCreditsTitle = t('openSourceCreditsTitle')

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
      <section
        id={sectionIdFromTitle(heroTitle)}
        className="overflow-hidden rounded-[2rem] border border-stone-200 bg-linear-to-br from-stone-100 via-white to-sky-50 p-8 shadow-sm ring-1 ring-black/3 dark:border-white/10 dark:from-zinc-900 dark:via-zinc-900 dark:to-sky-950/40 dark:ring-white/10"
      >
        <div className="max-w-3xl">
          <Eyebrow className="tracking-[0.22em]">{t('heroEyebrow')}</Eyebrow>
          <Heading className="mt-4 text-4xl/11 sm:text-3xl/10">{heroTitle}</Heading>
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
        <WorkspaceCard
          id={sectionIdFromTitle(skillsTitle)}
          className="border-stone-200/80 bg-stone-50 dark:border-white/10 dark:bg-white/5"
        >
          <Subheading className="text-xl/7 sm:text-lg/7">{skillsTitle}</Subheading>
          <div className="mt-4 space-y-4">
            <Text>{t('skillsBody1')}</Text>
            <Text>{t('skillsBody2')}</Text>
            <Text>{t('skillsBody3')}</Text>
          </div>
        </WorkspaceCard>

        <WorkspaceCard
          id={sectionIdFromTitle(whatYouGetTitle)}
          className="border-stone-200/80 bg-zinc-950 text-white dark:border-white/10"
        >
          <Subheading className="text-xl/7 text-white sm:text-lg/7">{whatYouGetTitle}</Subheading>
          <ul className="mt-4 space-y-3 text-sm/6 text-zinc-300">
            <li className="flex items-start gap-3">
              <span className="mt-0.5 inline-flex rounded-lg bg-amber-200/90 p-1.5 text-zinc-900">
                <ClipboardDocumentListIcon className="size-4" />
              </span>
              <span>{t('whatYouGet1')}</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="mt-0.5 inline-flex rounded-lg bg-amber-200/90 p-1.5 text-zinc-900">
                <ChatBubbleLeftRightIcon className="size-4" />
              </span>
              <span>{t('whatYouGet2')}</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="mt-0.5 inline-flex rounded-lg bg-amber-200/90 p-1.5 text-zinc-900">
                <ArrowPathIcon className="size-4" />
              </span>
              <span>{t('whatYouGet3')}</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="mt-0.5 inline-flex rounded-lg bg-amber-200/90 p-1.5 text-zinc-900">
                <ChartBarSquareIcon className="size-4" />
              </span>
              <span>{t('whatYouGet4')}</span>
            </li>
            <li className="flex items-start gap-3">
              <span className="mt-0.5 inline-flex rounded-lg bg-amber-200/90 p-1.5 text-zinc-900">
                <MapIcon className="size-4" />
              </span>
              <span>{t('whatYouGet5')}</span>
            </li>
          </ul>
        </WorkspaceCard>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <AboutCard
          id={sectionIdFromTitle(focusTitle)}
          eyebrow={t('focusEyebrow')}
          title={focusTitle}
          body={t('focusBody')}
        />
        <AboutCard
          id={sectionIdFromTitle(feedbackTitle)}
          eyebrow={t('feedbackEyebrow')}
          title={feedbackTitle}
          body={t('feedbackBody')}
        />
        <AboutCard
          id={sectionIdFromTitle(progressTitle)}
          eyebrow={t('progressEyebrow')}
          title={progressTitle}
          body={t('progressBody')}
        />
      </div>

      <WorkspaceCard
        id={sectionIdFromTitle(differentTitle)}
        className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900"
      >
        <Subheading className="text-xl/7 sm:text-lg/7">{differentTitle}</Subheading>
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="mb-3 inline-flex rounded-lg bg-amber-100 p-2 text-amber-700 dark:bg-zinc-800 dark:text-amber-300">
              <Squares2X2Icon className="size-5" />
            </div>
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('differentCard1Title')}</div>
            <Text className="mt-2">{t('differentCard1Body')}</Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="mb-3 inline-flex rounded-lg bg-amber-100 p-2 text-amber-700 dark:bg-zinc-800 dark:text-amber-300">
              <LinkIcon className="size-5" />
            </div>
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('differentCard2Title')}</div>
            <Text className="mt-2">{t('differentCard2Body')}</Text>
          </div>
        </div>
      </WorkspaceCard>

      <WorkspaceCard
        id={sectionIdFromTitle(deterministicDetailTitle)}
        className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900"
      >
        <Eyebrow>{t('deterministicDetailEyebrow')}</Eyebrow>
        <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{deterministicDetailTitle}</Subheading>
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

      <WorkspaceCard
        id={sectionIdFromTitle(ossToolsTitle)}
        className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900"
      >
        <Eyebrow>{t('ossToolsEyebrow')}</Eyebrow>
        <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{ossToolsTitle}</Subheading>
        <Text className="mt-3">{t('ossToolsIntro')}</Text>
        <div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {openSourceCheckTools.map((tool) => (
            <div
              key={tool.titleKey}
              className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5"
            >
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t(tool.titleKey)}</div>
              <Text className="mt-2">{t(tool.bodyKey)}</Text>
              {tool.href ? (
                <Text className="mt-3 text-sm">
                  <TextLink href={tool.href} target="_blank" rel="noreferrer">
                    {t('ossProjectLinkLabel')}
                  </TextLink>
                  <span className="ml-2 text-zinc-600 dark:text-zinc-300">
                    {t('licenseLabel', { license: t(tool.licenseKey) })}
                  </span>
                </Text>
              ) : null}
            </div>
          ))}
        </div>
      </WorkspaceCard>

      <WorkspaceCard
        id={sectionIdFromTitle(aiBoundaryTitle)}
        className="border-stone-200/80 bg-linear-to-br from-stone-50 via-white to-sky-50 dark:border-white/10 dark:from-zinc-900 dark:via-zinc-900 dark:to-sky-950/30"
      >
        <Eyebrow>{t('aiBoundaryEyebrow')}</Eyebrow>
        <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{aiBoundaryTitle}</Subheading>
        <Text className="mt-4 max-w-3xl">{t('aiBoundaryIntro')}</Text>

        <div className="mt-6">
          <div className="text-sm font-semibold uppercase tracking-wide text-zinc-700 dark:text-zinc-300">{t('promptFlowTitle')}</div>
          <div className="mt-3 grid gap-4 md:grid-cols-2">
            <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
              <div className="flex flex-wrap items-center gap-2">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('promptStep1Title')}</div>
                <span className="inline-flex rounded-full border border-zinc-300 bg-zinc-100 px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide text-zinc-700 dark:border-zinc-600 dark:bg-zinc-800 dark:text-zinc-200">
                  {t('stepTagNoAI')}
                </span>
              </div>
              <Text className="mt-2">{t('promptStep1Body')}</Text>
            </div>
            <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
              <div className="flex flex-wrap items-center gap-2">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('promptStep2Title')}</div>
                <span className="inline-flex rounded-full border border-sky-300 bg-sky-100 px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide text-sky-800 dark:border-sky-700 dark:bg-sky-900/40 dark:text-sky-200">
                  {t('stepTagAI')}
                </span>
              </div>
              <Text className="mt-2">{t('promptStep2Body')}</Text>
            </div>
          </div>
        </div>

        <div className="mt-6">
          <div className="text-sm font-semibold uppercase tracking-wide text-zinc-700 dark:text-zinc-300">{t('reviewFlowTitle')}</div>
          <div className="mt-3 grid gap-4 md:grid-cols-3">
            <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
              <div className="flex flex-wrap items-center gap-2">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('reviewStep1Title')}</div>
                <span className="inline-flex rounded-full border border-zinc-300 bg-zinc-100 px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide text-zinc-700 dark:border-zinc-600 dark:bg-zinc-800 dark:text-zinc-200">
                  {t('stepTagNoAI')}
                </span>
              </div>
              <Text className="mt-2">{t('reviewStep1Body')}</Text>
            </div>
            <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
              <div className="flex flex-wrap items-center gap-2">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('reviewStep2Title')}</div>
                <span className="inline-flex rounded-full border border-zinc-300 bg-zinc-100 px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide text-zinc-700 dark:border-zinc-600 dark:bg-zinc-800 dark:text-zinc-200">
                  {t('stepTagNoAI')}
                </span>
              </div>
              <Text className="mt-2">{t('reviewStep2Body')}</Text>
            </div>
            <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
              <div className="flex flex-wrap items-center gap-2">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('reviewStep3Title')}</div>
                <span className="inline-flex rounded-full border border-sky-300 bg-sky-100 px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide text-sky-800 dark:border-sky-700 dark:bg-sky-900/40 dark:text-sky-200">
                  {t('stepTagAI')}
                </span>
              </div>
              <Text className="mt-2">{t('reviewStep3Body')}</Text>
            </div>
          </div>
        </div>
      </WorkspaceCard>

      <WorkspaceCard
        id={sectionIdFromTitle(openSourceCreditsTitle)}
        className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900"
      >
        <Eyebrow>{t('openSourceCreditsEyebrow')}</Eyebrow>
        <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{openSourceCreditsTitle}</Subheading>
        <Text className="mt-3 text-sm/6">{t('openSourceCreditsIntro')}</Text>
        <ul className="mt-4 space-y-2 text-xs/6 text-zinc-700 dark:text-zinc-300">
          {openSourceCredits.map((tool) => (
            <li key={tool.titleKey}>
              <span className="font-semibold text-zinc-900 dark:text-zinc-100">{t(tool.titleKey)}</span>{' '}
              <span className="text-zinc-500 dark:text-zinc-400">-</span>{' '}
              <span>
                <TextLink href={tool.href} target="_blank" rel="noreferrer">
                  {t('ossProjectLinkLabel')}
                </TextLink>
              </span>{' '}
              <span className="text-zinc-500 dark:text-zinc-400">-</span>{' '}
              <span className="text-zinc-600 dark:text-zinc-300">
                  {t('licenseLabel', { license: t(tool.licenseKey) })}
              </span>
            </li>
          ))}
        </ul>
      </WorkspaceCard>

      <footer className="border-t border-stone-200/80 pt-6 dark:border-white/10">
        <div className="flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <TextLink href="/privacy">{t('legalPrivacy')}</TextLink>
          <TextLink href="/terms">{t('legalTerms')}</TextLink>
          <TextLink href="/third-party-notices">{t('legalThirdParty')}</TextLink>
        </div>
      </footer>
    </div>
  )
}
