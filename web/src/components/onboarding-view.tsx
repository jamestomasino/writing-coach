'use client'

import { CardHeader } from '@/components/card-header'
import { Callout } from '@/components/callout'
import { PageHeader } from '@/components/page-header'
import { archiveTrack, getOnboarding, getOnboardingOptions, listTracks, saveOnboarding } from '@/lib/api'
import type { OnboardingOptions, OnboardingState, UserTrack } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { useTranslations } from 'next-intl'
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
  const t = useTranslations('onboardingView')
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
          setError(err instanceof Error ? err.message : t('loadError'))
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
  }, [mode, session, t])

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
      setError(err instanceof Error ? err.message : t('saveError'))
    } finally {
      setSaving(false)
    }
  }

  async function handleArchive() {
    const treeSlug = onboardingState?.context?.tree_slug
    if (!treeSlug) {
      return
    }
    const confirmed = window.confirm(t('archiveConfirm'))
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
      setError(err instanceof Error ? err.message : t('archiveError'))
    } finally {
      setArchiving(false)
    }
  }

  if (sessionLoading || loading) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError) {
    return <EmptyState title={t('issueTitle')} body={sessionError} />
  }

  const canArchive = mode === 'edit' && existingProfile && tracks.length > 1

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={setupFlow ? t('setupStepEyebrow') : t('setupEyebrow')}
        title={
          mode === 'create' ? t('createTitle') : existingProfile ? t('editTitle') : t('startingPointTitle')
        }
        intro={
          setupFlow
            ? t('setupIntro')
            : mode === 'create'
            ? t('createIntro')
            : existingProfile
              ? t('editIntro')
              : t('startingPointIntro')
        }
      />

      {setupFlow ? (
        <Callout
          tone="active"
          eyebrow={t('calloutEyebrow')}
          title={t('calloutTitle')}
          body={t('calloutBody')}
        >
          <ul className="space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            <li>{t('calloutBullet1')}</li>
            <li>{t('calloutBullet2')}</li>
            <li>{t('calloutBullet3')}</li>
          </ul>
        </Callout>
      ) : null}

      <WorkspaceCard>
        <CardHeader eyebrow={t('howItWorksEyebrow')} title={t('howItWorksTitle')} />
        <div className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
          <p>{t('howItWorksBody1')}</p>
          <p>{t('howItWorksBody2')}</p>
          <p>{t('howItWorksBody3')}</p>
        </div>
      </WorkspaceCard>

      {error ? <EmptyState title={t('issueTitle')} body={error} /> : null}

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
