import { Badge } from '@/components/badge'
import { Eyebrow } from '@/components/eyebrow'
import { SkillFamilyObjectives } from '@/components/skill-family-objectives'
import { Heading, Subheading } from '@/components/heading'
import { Link } from '@/components/link'
import { Text } from '@/components/text'
import { allSkillDetails, getSkillDetailBySlug } from '@/lib/skill-details'
import { WorkspaceCard } from '@/components/workspace-card'
import type { Metadata } from 'next'
import { notFound } from 'next/navigation'

type Params = {
  slug: string
}

const tierLabel: Record<string, string> = {
  core: 'Core',
  domain: 'Domain',
  specialty: 'Specialty',
}

export function generateStaticParams() {
  return allSkillDetails.map((skill) => ({ slug: skill.slug }))
}

export async function generateMetadata({ params }: { params: Promise<Params> }): Promise<Metadata> {
  const { slug } = await params
  const detail = getSkillDetailBySlug(slug)
  if (!detail) {
    return { title: 'Skill not found' }
  }
  return {
    title: `${detail.name} skill`,
  }
}

export default async function SkillDetailPage({ params }: { params: Promise<Params> }) {
  const { slug } = await params
  const detail = getSkillDetailBySlug(slug)

  if (!detail) {
    notFound()
  }

  return (
    <div className="space-y-8">
      <WorkspaceCard>
        <Eyebrow>Skill detail</Eyebrow>
        <Heading className="mt-3 capitalize">{detail.name}</Heading>
        <div className="mt-3 flex flex-wrap items-center gap-2">
          <Badge color="zinc">{tierLabel[detail.tier] ?? detail.tier}</Badge>
          <Badge color="blue">Teaching guide</Badge>
        </div>
        <Text className="mt-4">{detail.oneLine}</Text>
      </WorkspaceCard>

      <WorkspaceCard>
        <Subheading>What this means</Subheading>
        <Text className="mt-3">{detail.whatItMeans}</Text>
        <Subheading className="mt-6">Why it matters</Subheading>
        <Text className="mt-3">{detail.whyItMatters}</Text>
      </WorkspaceCard>

      <WorkspaceCard>
        <Subheading>What to look for</Subheading>
        <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
          {detail.lookFor.map((item) => (
            <li key={item}>• {item}</li>
          ))}
        </ul>
      </WorkspaceCard>

      <WorkspaceCard>
        <Subheading>Examples</Subheading>
        <div className="mt-4 grid gap-4 lg:grid-cols-2">
          <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-400/30 dark:bg-emerald-500/10">
            <div className="text-sm font-semibold text-emerald-900 dark:text-emerald-200">Stronger example</div>
            <Text className="mt-2 text-sm text-emerald-900 dark:text-emerald-100">{detail.strongExample}</Text>
          </div>
          <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-400/30 dark:bg-amber-500/10">
            <div className="text-sm font-semibold text-amber-900 dark:text-amber-200">Needs work example</div>
            <Text className="mt-2 text-sm text-amber-900 dark:text-amber-100">{detail.weakExample}</Text>
          </div>
        </div>
      </WorkspaceCard>

      <WorkspaceCard>
        <Subheading>Revision moves</Subheading>
        <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
          {detail.revisionMoves.map((item) => (
            <li key={item}>• {item}</li>
          ))}
        </ul>
        <div className="mt-4 rounded-xl border border-stone-200 bg-stone-50 p-4 text-sm text-zinc-700 dark:border-white/10 dark:bg-white/5 dark:text-zinc-300">
          <span className="font-semibold text-zinc-900 dark:text-white">Coach tip: </span>
          {detail.coachTip}
        </div>
      </WorkspaceCard>

      <WorkspaceCard>
        <Subheading>Skill objectives in this family</Subheading>
        <Text className="mt-3 text-sm">
          These are the objective-level checkpoints that currently map to this skill family in active trees.
        </Text>
        <SkillFamilyObjectives familyName={detail.name} />
      </WorkspaceCard>

      <WorkspaceCard>
        <Subheading>Content sources</Subheading>
        <Text className="mt-3">These guides are generated from app curriculum data and internal pedagogy documentation.</Text>
        <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
          {detail.contentSources.map((source) => (
            <li key={source}>• {source}</li>
          ))}
        </ul>
        <Text className="mt-4 text-sm">
          Want the full list? <Link href="/skills">Go back to the skill library.</Link>
        </Text>
      </WorkspaceCard>
    </div>
  )
}
