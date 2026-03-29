import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Text, TextLink } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import type { Metadata } from 'next'

const notices = [
  { name: 'React', href: 'https://react.dev/', license: 'MIT' },
  { name: 'Next.js', href: 'https://nextjs.org/', license: 'MIT' },
  { name: 'Tailwind CSS', href: 'https://tailwindcss.com/', license: 'MIT' },
  { name: 'Headless UI', href: 'https://headlessui.com/', license: 'MIT' },
  { name: 'Heroicons (icon pack)', href: 'https://heroicons.com/', license: 'MIT' },
  { name: 'Inter (font)', href: 'https://rsms.me/inter/', license: 'SIL Open Font License 1.1' },
  { name: 'Vale', href: 'https://vale.sh/', license: 'MIT' },
  { name: 'LanguageTool', href: 'https://languagetool.org/', license: 'LGPL-2.1-or-later' },
  { name: 'spaCy', href: 'https://spacy.io/', license: 'MIT' },
  { name: 'TextDescriptives', href: 'https://hlasse.github.io/TextDescriptives/', license: 'Apache-2.0' },
  { name: 'Stanford CoreNLP (optional)', href: 'https://stanfordnlp.github.io/CoreNLP/', license: 'GPL-3.0-or-later' },
  { name: 'Ory Kratos', href: 'https://github.com/ory/kratos', license: 'Apache-2.0' },
  { name: 'SQLite', href: 'https://www.sqlite.org/', license: 'Public Domain' },
  {
    name: 'Tailwind Plus / Catalyst UI materials',
    href: 'https://tailwindcss.com/plus',
    license: 'Tailwind Plus commercial license',
  },
  { name: 'Writing Coach (this project)', href: 'https://github.com/tomasino/writing-coach', license: 'GPL-3.0-or-later' },
] as const

export const metadata: Metadata = {
  title: 'Third-Party Notices',
}

export default function ThirdPartyNoticesPage() {
  return (
    <div className="space-y-6">
      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Eyebrow>Legal</Eyebrow>
        <Heading className="mt-3">Third-Party Notices</Heading>
        <Text className="mt-3">
          This page provides a plain-language summary of key third-party tools, assets, and
          licenses used by this deployment.
        </Text>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>Summary list</Subheading>
        <ul className="mt-4 space-y-2 text-xs/6 text-zinc-700 dark:text-zinc-300">
          {notices.map((item) => (
            <li key={item.name}>
              <span className="font-semibold text-zinc-900 dark:text-zinc-100">{item.name}</span>{' '}
              <span className="text-zinc-500 dark:text-zinc-400">-</span>{' '}
              <TextLink href={item.href} target="_blank" rel="noreferrer">
                Project
              </TextLink>{' '}
              <span className="text-zinc-500 dark:text-zinc-400">-</span>{' '}
              <span className="text-zinc-600 dark:text-zinc-300">License: {item.license}</span>
            </li>
          ))}
        </ul>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>Repository legal files</Subheading>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>
            <TextLink href="https://github.com/tomasino/writing-coach/blob/main/LICENSE" target="_blank" rel="noreferrer">
              LICENSE
            </TextLink>{' '}
            (project GPL terms).
          </li>
          <li>
            <TextLink href="https://github.com/tomasino/writing-coach/blob/main/NOTICE.md" target="_blank" rel="noreferrer">
              NOTICE.md
            </TextLink>{' '}
            (mixed-license repository notice).
          </li>
          <li>
            <TextLink href="https://github.com/tomasino/writing-coach/blob/main/web/LICENSE.md" target="_blank" rel="noreferrer">
              web/LICENSE.md
            </TextLink>{' '}
            (Tailwind Plus terms used by this project).
          </li>
        </ul>
      </WorkspaceCard>
    </div>
  )
}

