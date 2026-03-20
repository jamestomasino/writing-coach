'use client'

import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { archiveTrack, getOnboarding, getOnboardingOptions, listTracks, saveOnboarding } from '@/lib/api'
import type { OnboardingOptions, OnboardingState, UserTrack } from '@/lib/types'
import { useRouter } from 'next/navigation'
import { FormEvent, useEffect, useState } from 'react'
import { OnboardingTrackForm } from './onboarding-track-form'
import { EmptyState, LoadingState } from './status-state'
import { TrackManagementCard } from './track-management-card'
import { WorkspaceCard } from './workspace-card'

const emptyOptions: OnboardingOptions = {
  writing_domains: [],
  assignment_formats: [],
  experience_levels: [],
  difficulty_levels: [],
  weaknesses: [],
  desired_outcomes: [],
}

export function OnboardingView({ mode = 'edit' }: { mode?: 'create' | 'edit' }) {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [existingProfile, setExistingProfile] = useState(false)
  const [options, setOptions] = useState<OnboardingOptions>(emptyOptions)
  const [writingType, setWritingType] = useState('')
  const [assignmentFormat, setAssignmentFormat] = useState('')
  const [targetAudience, setTargetAudience] = useState('')
  const [subjectMatter, setSubjectMatter] = useState('')
  const [experienceLevel, setExperienceLevel] = useState('')
  const [desiredTone, setDesiredTone] = useState('')
  const [difficultyIntensity, setDifficultyIntensity] = useState('')
  const [writingGoals, setWritingGoals] = useState('')
  const [weaknesses, setWeaknesses] = useState<string[]>([])
  const [outcomes, setOutcomes] = useState<string[]>([])
  const [saving, setSaving] = useState(false)
  const [archiving, setArchiving] = useState(false)
  const [onboardingState, setOnboardingState] = useState<OnboardingState | null>(null)
  const [tracks, setTracks] = useState<UserTrack[]>([])
  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const [onboarding, nextOptions, trackList] = await Promise.all([
          getOnboarding(),
          getOnboardingOptions(),
          listTracks(),
        ])
        if (cancelled) {
          return
        }
        setOnboardingState(onboarding)
        setOptions(nextOptions)
        setTracks(trackList)
        if (mode === 'edit' && onboarding.profile) {
          setWritingType(onboarding.profile.writing_type)
          setAssignmentFormat(onboarding.profile.assignment_format)
          setTargetAudience(onboarding.profile.target_audience)
          setSubjectMatter(onboarding.profile.subject_matter)
          setExperienceLevel(onboarding.profile.experience_level)
          setDesiredTone(onboarding.profile.desired_tone)
          setDifficultyIntensity(onboarding.profile.difficulty_intensity)
          setWritingGoals(onboarding.profile.writing_goals)
          setWeaknesses(onboarding.profile.biggest_weaknesses)
          setOutcomes(onboarding.profile.desired_outcomes)
          setExistingProfile(true)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load onboarding')
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
  }, [mode, router])

  function toggle(list: string[], setList: (items: string[]) => void, value: string) {
    if (list.includes(value)) {
      setList(list.filter((item) => item !== value))
      return
    }
    setList([...list, value])
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    try {
      setSaving(true)
      setError(null)
      await saveOnboarding({
        mode,
        writing_type: writingType,
        assignment_format: assignmentFormat,
        target_audience: targetAudience,
        subject_matter: subjectMatter,
        experience_level: experienceLevel,
        desired_tone: desiredTone,
        biggest_weaknesses: weaknesses,
        desired_outcomes: outcomes,
        difficulty_intensity: difficultyIntensity,
        writing_goals: writingGoals,
      })
      router.push('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save onboarding')
    } finally {
      setSaving(false)
    }
  }

  async function handleArchive() {
    const treeSlug = onboardingState?.context?.tree_slug
    if (!treeSlug) {
      return
    }
    const confirmed = window.confirm(
      'Archive this track? Its history will be kept, but it will be removed from the active track list.'
    )
    if (!confirmed) {
      return
    }
    try {
      setArchiving(true)
      setError(null)
      await archiveTrack(treeSlug)
      router.push('/')
      router.refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not archive track')
    } finally {
      setArchiving(false)
    }
  }

  if (loading) {
    return <LoadingState label="Loading onboarding…" />
  }

  const canArchive = mode === 'edit' && existingProfile && tracks.length > 1

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Track setup"
        title={mode === 'create' ? 'Create a new track' : existingProfile ? 'Edit track' : 'Set your starting path'}
        intro={
          mode === 'create'
            ? 'Create an additional writing track with its own skill map, progress, and assignment history.'
            : existingProfile
              ? 'Update the writing profile that shapes your coaching track. Saving here refreshes the recommended path, active skills, and future assignment focus.'
              : 'Tell the coach what kind of writing you want to improve. This recommends a starting path into the writing skill map, including your first active skills and the regions most likely to matter first.'
        }
      />

      <WorkspaceCard>
        <CardHeader eyebrow="How it works" title="How the coaching loop works" />
        <div className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
          <p>You can focus on up to three skills at a time.</p>
          <p>Your assignment prompt and review are built around those active skills.</p>
          <p>
            When you show strong, consistent control, a skill can become mastered and stay in lighter maintenance checks
            going forward.
          </p>
        </div>
      </WorkspaceCard>

      {error ? <EmptyState title="Onboarding issue" body={error} /> : null}

      <OnboardingTrackForm
        mode={mode}
        options={options}
        existingProfile={existingProfile}
        writingType={writingType}
        assignmentFormat={assignmentFormat}
        targetAudience={targetAudience}
        subjectMatter={subjectMatter}
        experienceLevel={experienceLevel}
        desiredTone={desiredTone}
        difficultyIntensity={difficultyIntensity}
        writingGoals={writingGoals}
        weaknesses={weaknesses}
        outcomes={outcomes}
        saving={saving}
        onWritingTypeChange={setWritingType}
        onAssignmentFormatChange={setAssignmentFormat}
        onTargetAudienceChange={setTargetAudience}
        onSubjectMatterChange={setSubjectMatter}
        onExperienceLevelChange={setExperienceLevel}
        onDesiredToneChange={setDesiredTone}
        onDifficultyIntensityChange={setDifficultyIntensity}
        onWritingGoalsChange={setWritingGoals}
        onWeaknessToggle={(value) => toggle(weaknesses, setWeaknesses, value)}
        onOutcomeToggle={(value) => toggle(outcomes, setOutcomes, value)}
        onSubmit={handleSubmit}
      />

      {canArchive ? <TrackManagementCard archiving={archiving} onArchive={handleArchive} /> : null}
    </div>
  )
}
