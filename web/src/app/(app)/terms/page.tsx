import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Text, TextLink } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Terms',
}

export default function TermsPage() {
  return (
    <div className="space-y-6">
      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Eyebrow>Legal</Eyebrow>
        <Heading className="mt-3">Terms of Use</Heading>
        <Text className="mt-3">Effective date: March 29, 2026.</Text>
        <Text className="mt-2">
          These terms describe basic rules for using this Writing Coach deployment. They are
          written for clarity, not legal complexity.
        </Text>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>Use of the service</Subheading>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>Use the service lawfully and do not abuse, disrupt, or probe it.</li>
          <li>Do not attempt unauthorized access to accounts, systems, or data.</li>
          <li>You are responsible for content you submit and actions taken from your account.</li>
        </ul>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>Accounts and availability</Subheading>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>Account access may be suspended or removed for misuse or security risk.</li>
          <li>Service features may change over time.</li>
          <li>Availability is best effort and not guaranteed at all times.</li>
        </ul>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>AI and feedback</Subheading>
        <Text className="mt-3">
          Writing Coach uses deterministic checks and optional AI-assisted language generation.
          Outputs may be imperfect and should be reviewed by you before use.
        </Text>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>Open-source and third-party terms</Subheading>
        <Text className="mt-3">
          This project is mixed-license. Use of this deployment is also subject to third-party
          component terms.
        </Text>
        <ul className="mt-2 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>
            Project license: <TextLink href="https://github.com/tomasino/writing-coach">Writing Coach</TextLink>{' '}
            (GPL-3.0-or-later).
          </li>
          <li>
            Repository notices: <TextLink href="/about">About</TextLink> and{' '}
            <TextLink href="/third-party-notices">Third-Party Notices</TextLink>.
          </li>
        </ul>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>Warranty and liability</Subheading>
        <Text className="mt-3">
          The service is provided &quot;as is&quot; without warranties. To the fullest extent
          allowed by law, operators and contributors are not liable for indirect or consequential
          damages from use of the service.
        </Text>
      </WorkspaceCard>
    </div>
  )
}

