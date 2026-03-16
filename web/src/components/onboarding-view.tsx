'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/button'
import { Checkbox, CheckboxField } from '@/components/checkbox'
import { Field, FieldGroup, Label } from '@/components/fieldset'
import { Heading, Subheading } from '@/components/heading'
import { Input } from '@/components/input'
import { Select } from '@/components/select'
import { Text } from '@/components/text'
import { Textarea } from '@/components/textarea'
import { getOnboarding, saveOnboarding } from '@/lib/api'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

const weaknessOptions = [
  'word choice',
  'sentence variety',
  'sentence economy',
  'paragraph control',
  'narrative clarity',
  'scene architecture',
  'symbolic control',
  'tone calibration',
  'evidence integration',
]

const outcomeOptions = [
  'publish stronger fiction',
  'write clearer essays',
  'improve professional communication',
  'develop a distinctive voice',
  'build revision discipline',
  'write thought leadership with authority',
]

export function OnboardingView() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [writingType, setWritingType] = useState('fiction')
  const [experienceLevel, setExperienceLevel] = useState('intermediate')
  const [desiredTone, setDesiredTone] = useState('')
  const [difficultyIntensity, setDifficultyIntensity] = useState('steady')
  const [writingGoals, setWritingGoals] = useState('')
  const [weaknesses, setWeaknesses] = useState<string[]>([])
  const [outcomes, setOutcomes] = useState<string[]>([])
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const onboarding = await getOnboarding()
        if (cancelled) {
          return
        }
        if (onboarding.profile) {
          setWritingType(onboarding.profile.writing_type)
          setExperienceLevel(onboarding.profile.experience_level)
          setDesiredTone(onboarding.profile.desired_tone)
          setDifficultyIntensity(onboarding.profile.difficulty_intensity)
          setWritingGoals(onboarding.profile.writing_goals)
          setWeaknesses(onboarding.profile.biggest_weaknesses)
          setOutcomes(onboarding.profile.desired_outcomes)
        }
        if (onboarding.onboarding_complete) {
          router.replace('/')
          return
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
      <header>
        <Heading>Skill tree creator</Heading>
        <Text className="mt-2 max-w-3xl">
          Tell the coach what kind of writing you want to improve. This creates your active track, seeds the first TGOs, and determines what unlocks next.
        </Text>
      </header>

      {error ? <EmptyState title="Onboarding issue" body={error} /> : null}

      <WorkspaceCard>
        <form className="space-y-8" onSubmit={handleSubmit}>
          <FieldGroup>
            <Field>
              <Label>Writing type</Label>
              <Select value={writingType} onChange={(event) => setWritingType(event.target.value)}>
                <option value="fiction">Fiction</option>
                <option value="thought leadership">Thought leadership</option>
                <option value="professional">Professional writing</option>
                <option value="other">Other</option>
              </Select>
            </Field>
            <Field>
              <Label>Experience level</Label>
              <Select value={experienceLevel} onChange={(event) => setExperienceLevel(event.target.value)}>
                <option value="beginner">Beginner</option>
                <option value="intermediate">Intermediate</option>
                <option value="advanced">Advanced</option>
              </Select>
            </Field>
            <Field>
              <Label>Difficulty and intensity</Label>
              <Select value={difficultyIntensity} onChange={(event) => setDifficultyIntensity(event.target.value)}>
                <option value="steady">Steady</option>
                <option value="ambitious">Ambitious</option>
                <option value="gentle">Gentle</option>
              </Select>
            </Field>
          </FieldGroup>

          <Field>
            <Label>Desired tone or style</Label>
            <Input value={desiredTone} onChange={(event) => setDesiredTone(event.target.value)} placeholder="Mythic and grave, practical and concise, analytical and decisive…" />
          </Field>

          <Field>
            <Label>Writing goals</Label>
            <Textarea rows={6} value={writingGoals} onChange={(event) => setWritingGoals(event.target.value)} placeholder="Describe what you want this coaching track to help you become better at." />
          </Field>

          <div className="grid gap-8 lg:grid-cols-2">
            <div>
              <Subheading>Biggest weaknesses</Subheading>
              <div className="mt-4 space-y-3">
                {weaknessOptions.map((item) => (
                  <CheckboxField key={item}>
                    <Checkbox checked={weaknesses.includes(item)} onChange={() => toggle(weaknesses, setWeaknesses, item)} />
                    <Label>{item}</Label>
                  </CheckboxField>
                ))}
              </div>
            </div>
            <div>
              <Subheading>Desired outcomes</Subheading>
              <div className="mt-4 space-y-3">
                {outcomeOptions.map((item) => (
                  <CheckboxField key={item}>
                    <Checkbox checked={outcomes.includes(item)} onChange={() => toggle(outcomes, setOutcomes, item)} />
                    <Label>{item}</Label>
                  </CheckboxField>
                ))}
              </div>
            </div>
          </div>

          <div className="flex justify-end">
            <Button type="submit" color="dark/zinc" disabled={saving}>
              {saving ? 'Generating track…' : 'Generate skill tree'}
            </Button>
          </div>
        </form>
      </WorkspaceCard>
    </div>
  )
}
