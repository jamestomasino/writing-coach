import { Badge } from '@/components/badge'
import { Eyebrow } from '@/components/eyebrow'
import { Heading } from '@/components/heading'
import { Link } from '@/components/link'
import { Text } from '@/components/text'
import { allSkillDetails, type SkillTier } from '@/lib/skill-details'
import { WorkspaceCard } from '@/components/workspace-card'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Skills',
}

const TIERS: Array<{ tier: SkillTier; label: string; description: string }> = [
  {
    tier: 'core',
    label: 'Core skills',
    description: 'These are the main skills the coach leans on most often.',
  },
  {
    tier: 'domain',
    label: 'Domain skills',
    description: 'These add craft detail and help you shape stronger drafts.',
  },
  {
    tier: 'specialty',
    label: 'Specialty skills',
    description: 'These are advanced style and theme skills used in specific paths.',
  },
]

export default function SkillsPage() {
  return (
    <div className="space-y-8">
      <WorkspaceCard>
        <Eyebrow>Skill library</Eyebrow>
        <Heading className="mt-3">Skill detail pages</Heading>
        <Text className="mt-3">
          Each skill has a plain-language guide with examples and revision moves. Open any skill to see what it means,
          how to spot it, and how to improve it.
        </Text>
        <Text className="mt-2">Total skills: {allSkillDetails.length}</Text>
      </WorkspaceCard>

      {TIERS.map((group) => {
        const skills = allSkillDetails.filter((item) => item.tier === group.tier)
        if (skills.length === 0) {
          return null
        }

        return (
          <WorkspaceCard key={group.tier}>
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <Heading level={2}>{group.label}</Heading>
                <Text className="mt-2">{group.description}</Text>
              </div>
              <Badge color="zinc">{skills.length}</Badge>
            </div>
            <div className="mt-5 grid gap-4 md:grid-cols-2">
              {skills.map((skill) => (
                <Link
                  key={skill.slug}
                  href={`/skills/${skill.slug}`}
                  className="rounded-2xl border border-stone-200 bg-stone-50 p-4 data-hover:border-stone-300 data-hover:bg-white dark:border-white/10 dark:bg-white/5 dark:data-hover:border-white/20"
                >
                  <div className="font-semibold text-zinc-900 capitalize dark:text-white">{skill.name}</div>
                  <Text className="mt-2 text-sm">{skill.oneLine}</Text>
                </Link>
              ))}
            </div>
          </WorkspaceCard>
        )
      })}
    </div>
  )
}
