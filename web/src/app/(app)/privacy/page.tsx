import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Text, TextLink } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Privacy',
}

export default function PrivacyPage() {
  return (
    <div className="space-y-6">
      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Eyebrow>Legal</Eyebrow>
        <Heading className="mt-3">Privacy Policy</Heading>
        <Text className="mt-3">Effective date: March 29, 2026.</Text>
        <Text className="mt-2">
          Writing Coach collects only the data needed to sign you in, run the app, and provide
          writing feedback. We do not run analytics or advertising trackers.
        </Text>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>What we collect</Subheading>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>Account data (for example email and authentication records).</li>
          <li>Content you submit in the app (assignments, drafts, revisions, reviews).</li>
          <li>Operational settings (for example selected language and configured providers).</li>
          <li>Service logs needed for reliability, abuse prevention, and troubleshooting.</li>
        </ul>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>Cookies</Subheading>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>Authentication/session cookies are used for sign-in and account security.</li>
          <li>A locale cookie is used to remember language preference.</li>
          <li>No analytics or advertising cookies are used by the app.</li>
        </ul>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>How data is used</Subheading>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>To create and manage your account.</li>
          <li>To run writing checks, generate feedback, and keep assignment history.</li>
          <li>To operate, secure, and improve reliability of the service.</li>
        </ul>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>Third-party processors</Subheading>
        <Text className="mt-3">
          Writing feedback may use configured model providers (for example OpenAI, Anthropic,
          Gemini, Groq, or xAI) depending on deployment settings and your choices in the app.
        </Text>
        <Text className="mt-2">
          Authentication is handled by Ory Kratos in this deployment model. See{' '}
          <TextLink href="/third-party-notices">Third-Party Notices</TextLink> for dependency and
          license details.
        </Text>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>Contact and requests</Subheading>
        <Text className="mt-3">
          For privacy questions, data access, or deletion requests, contact the deployment
          administrator or project maintainer for this instance.
        </Text>
      </WorkspaceCard>
    </div>
  )
}

