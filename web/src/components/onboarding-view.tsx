'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Checkbox, CheckboxField } from '@/components/checkbox'
import { Eyebrow } from '@/components/eyebrow'
import { Field, FieldGroup, Label } from '@/components/fieldset'
import { Heading, Subheading } from '@/components/heading'
import { Input } from '@/components/input'
import { PageHeader } from '@/components/page-header'
import { Select } from '@/components/select'
import { Text } from '@/components/text'
import { Textarea } from '@/components/textarea'
import { getOnboarding, getOnboardingOptions, saveOnboarding } from '@/lib/api'
import type { OnboardingOptions } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

const emptyOptions: OnboardingOptions = {
  writing_domains: [],
  assignment_formats: [],
  experience_levels: [],
  difficulty_levels: [],
  weaknesses: [],
  desired_outcomes: [],
}

export function OnboardingView() {
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

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const [onboarding, nextOptions] = await Promise.all([getOnboarding(), getOnboardingOptions()])
        if (cancelled) {
          return
        }
        setOptions(nextOptions)
        if (onboarding.profile) {
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
  }, [router])

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

  if (loading) {
    return <LoadingState label="Loading onboarding…" />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Track setup"
        title={existingProfile ? 'Change your track' : 'Set your starting path'}
        intro={
          existingProfile
            ? 'Update the writing profile that shapes your coaching track. Saving here refreshes the recommended path, active skills, and future assignment focus.'
            : 'Tell the coach what kind of writing you want to improve. This recommends a starting path into the writing skill map, including your first active skills and the regions most likely to matter first.'
        }
      />

      <WorkspaceCard>
        <CardHeader eyebrow="How it works" title="How the coaching loop works" />
        <div className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
          <p>You can focus on up to three skills at a time.</p>
          <p>Your assignment prompt and review are built around those active skills.</p>
          <p>When you show strong, consistent control, a skill can become mastered and stay in lighter maintenance checks going forward.</p>
        </div>
      </WorkspaceCard>

      {error ? <EmptyState title="Onboarding issue" body={error} /> : null}

      <WorkspaceCard>
        <form className="space-y-8" onSubmit={handleSubmit}>
          <FieldGroup>
            <Field>
              <Label>Primary writing domain</Label>
              <Select value={writingType} onChange={(event) => setWritingType(event.target.value)}>
                <option value="" disabled>
                  Choose a writing domain
                </option>
                {options.writing_domains.map((item) => (
                  <option key={item.value} value={item.value}>
                    {item.label}
                  </option>
                ))}
              </Select>
            </Field>
            <Field>
              <Label>Common assignment format</Label>
              <Select value={assignmentFormat} onChange={(event) => setAssignmentFormat(event.target.value)}>
                <option value="" disabled>
                  Choose an assignment format
                </option>
                {options.assignment_formats.map((item) => (
                  <option key={item.value} value={item.value}>
                    {item.label}
                  </option>
                ))}
              </Select>
            </Field>
            <Field>
              <Label>Experience level</Label>
              <Select value={experienceLevel} onChange={(event) => setExperienceLevel(event.target.value)}>
                <option value="" disabled>
                  Choose an experience level
                </option>
                {options.experience_levels.map((item) => (
                  <option key={item.value} value={item.value}>
                    {item.label}
                  </option>
                ))}
              </Select>
            </Field>
            <Field>
              <Label>Difficulty and intensity</Label>
              <Select value={difficultyIntensity} onChange={(event) => setDifficultyIntensity(event.target.value)}>
                <option value="" disabled>
                  Choose a pace
                </option>
                {options.difficulty_levels.map((item) => (
                  <option key={item.value} value={item.value}>
                    {item.label}
                  </option>
                ))}
              </Select>
            </Field>
          </FieldGroup>

          <FieldGroup>
            <Field>
              <Label>Target audience</Label>
              <Text className="mt-1 text-sm">Who should the writing feel written for?</Text>
              <Input
                value={targetAudience}
                onChange={(event) => setTargetAudience(event.target.value)}
                placeholder="Startup founders, hiring managers, general readers, fantasy fans…"
              />
            </Field>
            <Field>
              <Label>Typical subject matter</Label>
              <Text className="mt-1 text-sm">What kinds of situations, topics, or worlds should assignments draw from?</Text>
              <Input
                value={subjectMatter}
                onChange={(event) => setSubjectMatter(event.target.value)}
                placeholder="Developer tools, workplace conflict, family pressure, product launches…"
              />
            </Field>
          </FieldGroup>

          <Field>
            <Label>Tone target</Label>
            <Text className="mt-1 text-sm">How should the writing feel to a reader?</Text>
            <Input
              value={desiredTone}
              onChange={(event) => setDesiredTone(event.target.value)}
              placeholder="Weighty and restrained, clear and persuasive, analytical and direct…"
            />
          </Field>

          <Field>
            <Label>Writing goals</Label>
            <Textarea rows={6} value={writingGoals} onChange={(event) => setWritingGoals(event.target.value)} placeholder="Describe what you want this coaching track to help you become better at." />
          </Field>

          <div className="grid gap-8 lg:grid-cols-2">
            <div>
              <CardHeader eyebrow="Diagnosis" title="Biggest weaknesses" />
              <div className="mt-4 space-y-3">
                {options.weaknesses.map((item) => (
                  <CheckboxField key={item.value}>
                    <Checkbox checked={weaknesses.includes(item.value)} onChange={() => toggle(weaknesses, setWeaknesses, item.value)} />
                    <Label>{item.label}</Label>
                  </CheckboxField>
                ))}
              </div>
            </div>
            <div>
              <CardHeader eyebrow="Target state" title="Desired outcomes" />
              <div className="mt-4 space-y-3">
                {options.desired_outcomes.map((item) => (
                  <CheckboxField key={item.value}>
                    <Checkbox checked={outcomes.includes(item.value)} onChange={() => toggle(outcomes, setOutcomes, item.value)} />
                    <Label>{item.label}</Label>
                  </CheckboxField>
                ))}
              </div>
            </div>
          </div>

          <div className="flex justify-end">
            <Button type="submit" color="dark/zinc" disabled={saving}>
              {saving ? 'Preparing recommendations…' : existingProfile ? 'Update track' : 'Set starter path'}
            </Button>
          </div>
        </form>
      </WorkspaceCard>
    </div>
  )
}
