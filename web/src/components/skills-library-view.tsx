'use client'

import { useEffect, useMemo, useState } from 'react'
import { Badge } from '@/components/badge'
import { Eyebrow } from '@/components/eyebrow'
import { Heading } from '@/components/heading'
import { Link } from '@/components/link'
import { Text } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import { getSkillGraph } from '@/lib/api'
import { allSkillDetails, type SkillTier } from '@/lib/skill-details'
import { AppErrorState, LoadingState } from './status-state'

const TIERS: Array<{ tier: SkillTier; label: string; description: string }> = [
  {
    tier: 'core',
    label: 'Core skill families',
    description: 'These are the main skills the coach leans on most often.',
  },
  {
    tier: 'domain',
    label: 'Domain skill families',
    description: 'These add craft detail and help you shape stronger drafts.',
  },
  {
    tier: 'specialty',
    label: 'Specialty skill families',
    description: 'These are advanced style and theme skills used in specific paths.',
  },
]

export function SkillsLibraryView() {
  const [activeNames, setActiveNames] = useState<Set<string> | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        setLoading(true)
        const graph = await getSkillGraph()
        if (cancelled) {
          return
        }
        const names = new Set<string>()
        for (const node of graph.nodes ?? []) {
          const value = node.skill_name?.trim().toLowerCase()
          if (!value) {
            continue
          }
          names.add(value)
        }
        setActiveNames(names)
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load active skill families.')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [])

  const activeSkills = useMemo(() => {
    if (!activeNames) {
      return []
    }
    return allSkillDetails.filter((item) => activeNames.has(item.name.trim().toLowerCase()))
  }, [activeNames])

  if (loading) {
    return <LoadingState label="Loading skill family library..." />
  }

  if (error) {
    return <AppErrorState title="Skill family library unavailable" error={error} />
  }

  return (
    <div className="space-y-8">
      <WorkspaceCard>
        <Eyebrow>Skill families</Eyebrow>
        <Heading className="mt-3">Skill family detail pages</Heading>
        <Text className="mt-3">
          This page lists skill families. Skill objectives are listed separately in the{' '}
          <Link href="/tgos">Skill objective library</Link>.
        </Text>
      </WorkspaceCard>

      {TIERS.map((group) => {
        const skills = activeSkills.filter((item) => item.tier === group.tier)
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
                  <div className="font-semibold capitalize text-zinc-900 dark:text-white">{skill.name}</div>
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
