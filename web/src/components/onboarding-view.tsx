'use client'

import { CardHeader } from '@/components/card-header'
import { Callout } from '@/components/callout'
import { PageHeader } from '@/components/page-header'
import { archiveTrack, getOnboarding, getOnboardingOptions, listTracks, saveOnboarding } from '@/lib/api'
import type { OnboardingOptions, OnboardingState, UserTrack } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { usePathname, useRouter } from 'next/navigation'
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
  const pathname = usePathname()
  const { session, loading: sessionLoading, error: sessionError } = useRequiredAppSession(pathname)
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
  const [setupFlow, setSetupFlow] = useState(false)
  useEffect(() => {
    let cancelled = false
    async function load() {
      if (!session) {
        return
      }
      try {
        setSetupFlow(session.setup_step === 'needs_first_track')
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
  }, [mode, session])

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
      router.push(onboardingState?.onboarding_complete ? '/' : '/new-assignment')
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
      'Archive this practice path? Its history will be kept, but it will be removed from your active practice paths.'
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
      setError(err instanceof Error ? err.message : 'Could not archive practice path')
    } finally {
      setArchiving(false)
    }
  }

  if (sessionLoading || loading) {
    return <LoadingState label="Loading onboarding…" />
  }
  if (sessionError) {
    return <EmptyState title="Onboarding issue" body={sessionError} />
  }

  const canArchive = mode === 'edit' && existingProfile && tracks.length > 1

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={setupFlow ? 'Step 2 of 3 · Practice path setup' : 'Practice path setup'}
        title={
          mode === 'create' ? 'Create a new practice path' : existingProfile ? 'Edit practice path' : 'Set your starting point'
        }
        intro={
          setupFlow
            ? 'Tell us what kind of writing you want to practice first.'
            : mode === 'create'
            ? 'Start another practice path for a different kind of writing.'
            : existingProfile
              ? 'Update this practice path’s focus, tone, and goals.'
              : 'Tell us what kind of writing you want to improve first.'
        }
      />

      {setupFlow ? (
        <Callout
          tone="active"
          eyebrow="Onboarding"
          title="Next, create your first practice path"
          body="Your answers shape what you practice next."
        >
          <ul className="space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            <li>Describe the kind of writing you want to practice most often.</li>
            <li>Pick the audience, tone, and outcomes you want the coach to optimize for.</li>
            <li>Saving this step takes you straight to Step 3 of 3: your first assignment.</li>
          </ul>
        </Callout>
      ) : null}

      <WorkspaceCard>
        <CardHeader eyebrow="How it works" title="How practice paths work" />
        <div className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
          <p>You work on up to three skills at a time.</p>
          <p>Each assignment and each round of feedback stays focused on those skills.</p>
          <p>As you improve, older skills are checked more lightly and new ones open up.</p>
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
