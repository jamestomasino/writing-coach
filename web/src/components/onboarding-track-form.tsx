'use client'

import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Checkbox, CheckboxField } from '@/components/checkbox'
import { Field, FieldGroup, Label } from '@/components/fieldset'
import { Input } from '@/components/input'
import { Select } from '@/components/select'
import { Text } from '@/components/text'
import { Textarea } from '@/components/textarea'
import type { OnboardingOptions } from '@/lib/types'
import { useTranslations } from 'next-intl'
import type { FormEvent } from 'react'
import { WorkspaceCard } from './workspace-card'

type Props = {
  mode: 'create' | 'edit'
  options: OnboardingOptions
  existingProfile: boolean
  writingType: string
  assignmentFormat: string
  targetAudience: string
  subjectMatter: string
  experienceLevel: string
  desiredTone: string
  difficultyIntensity: string
  writingGoals: string
  weaknesses: string[]
  outcomes: string[]
  saving: boolean
  onWritingTypeChange: (value: string) => void
  onAssignmentFormatChange: (value: string) => void
  onTargetAudienceChange: (value: string) => void
  onSubjectMatterChange: (value: string) => void
  onExperienceLevelChange: (value: string) => void
  onDesiredToneChange: (value: string) => void
  onDifficultyIntensityChange: (value: string) => void
  onWritingGoalsChange: (value: string) => void
  onWeaknessToggle: (value: string) => void
  onOutcomeToggle: (value: string) => void
  onSubmit: (event: FormEvent<HTMLFormElement>) => void
}

export function OnboardingTrackForm({
  mode,
  options,
  existingProfile,
  writingType,
  assignmentFormat,
  targetAudience,
  subjectMatter,
  experienceLevel,
  desiredTone,
  difficultyIntensity,
  writingGoals,
  weaknesses,
  outcomes,
  saving,
  onWritingTypeChange,
  onAssignmentFormatChange,
  onTargetAudienceChange,
  onSubjectMatterChange,
  onExperienceLevelChange,
  onDesiredToneChange,
  onDifficultyIntensityChange,
  onWritingGoalsChange,
  onWeaknessToggle,
  onOutcomeToggle,
  onSubmit,
}: Props) {
  const t = useTranslations('onboardingTrackForm')
  return (
    <WorkspaceCard>
      <form className="space-y-8" onSubmit={onSubmit} data-testid="track-form">
        <FieldGroup>
          <Field>
            <Label>{t('writingTypeLabel')}</Label>
            <Select value={writingType} onChange={(event) => onWritingTypeChange(event.target.value)}>
              <option value="" disabled>
                {t('writingTypePlaceholder')}
              </option>
              {options.writing_domains.map((item) => (
                <option key={item.value} value={item.value}>
                  {item.label}
                </option>
              ))}
            </Select>
          </Field>
          <Field>
            <Label>{t('assignmentFormatLabel')}</Label>
            <Select value={assignmentFormat} onChange={(event) => onAssignmentFormatChange(event.target.value)}>
              <option value="" disabled>
                {t('assignmentFormatPlaceholder')}
              </option>
              {options.assignment_formats.map((item) => (
                <option key={item.value} value={item.value}>
                  {item.label}
                </option>
              ))}
            </Select>
          </Field>
          <Field>
            <Label>{t('experienceLevelLabel')}</Label>
            <Select value={experienceLevel} onChange={(event) => onExperienceLevelChange(event.target.value)}>
              <option value="" disabled>
                {t('experienceLevelPlaceholder')}
              </option>
              {options.experience_levels.map((item) => (
                <option key={item.value} value={item.value}>
                  {item.label}
                </option>
              ))}
            </Select>
          </Field>
          <Field>
            <Label>{t('difficultyLabel')}</Label>
            <Select value={difficultyIntensity} onChange={(event) => onDifficultyIntensityChange(event.target.value)}>
              <option value="" disabled>
                {t('difficultyPlaceholder')}
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
            <Label>{t('targetAudienceLabel')}</Label>
            <Text className="mt-1 text-sm">{t('targetAudienceHelp')}</Text>
            <Input
              value={targetAudience}
              onChange={(event) => onTargetAudienceChange(event.target.value)}
              placeholder={t('targetAudiencePlaceholder')}
            />
          </Field>
          <Field>
            <Label>{t('subjectMatterLabel')}</Label>
            <Text className="mt-1 text-sm">{t('subjectMatterHelp')}</Text>
            <Input
              value={subjectMatter}
              onChange={(event) => onSubjectMatterChange(event.target.value)}
              placeholder={t('subjectMatterPlaceholder')}
            />
          </Field>
        </FieldGroup>

        <Field>
          <Label>{t('toneLabel')}</Label>
          <Text className="mt-1 text-sm">{t('toneHelp')}</Text>
          <Input
            value={desiredTone}
            onChange={(event) => onDesiredToneChange(event.target.value)}
            placeholder={t('tonePlaceholder')}
          />
        </Field>

        <Field>
          <Label>{t('goalsLabel')}</Label>
          <Textarea
            rows={6}
            value={writingGoals}
            onChange={(event) => onWritingGoalsChange(event.target.value)}
            placeholder={t('goalsPlaceholder')}
          />
        </Field>

        <div className="grid gap-8 lg:grid-cols-2">
          <div>
            <CardHeader eyebrow={t('rightNowEyebrow')} title={t('rightNowTitle')} />
            <div className="mt-4 space-y-3">
              {options.weaknesses.map((item) => (
                <CheckboxField key={item.value}>
                  <Checkbox checked={weaknesses.includes(item.value)} onChange={() => onWeaknessToggle(item.value)} />
                  <Label>{item.label}</Label>
                </CheckboxField>
              ))}
            </div>
          </div>
          <div>
            <CardHeader eyebrow={t('goalEyebrow')} title={t('goalTitle')} />
            <div className="mt-4 space-y-3">
              {options.desired_outcomes.map((item) => (
                <CheckboxField key={item.value}>
                  <Checkbox checked={outcomes.includes(item.value)} onChange={() => onOutcomeToggle(item.value)} />
                  <Label>{item.label}</Label>
                </CheckboxField>
              ))}
            </div>
          </div>
        </div>

        <div className="flex justify-end">
          <Button type="submit" color="dark/zinc" disabled={saving} data-testid="save-track-button">
            {saving
              ? t('saving')
              : mode === 'create'
                ? t('create')
                : existingProfile
                  ? t('save')
                  : t('create')}
          </Button>
        </div>
      </form>
    </WorkspaceCard>
  )
}
