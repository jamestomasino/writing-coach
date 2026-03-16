'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { ArrowUpTrayIcon, SparklesIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { Heading, Subheading } from '@/components/heading'
import { Strong, Text } from '@/components/text'
import { Textarea } from '@/components/textarea'
import { createRevisionAssignment, getDashboard, getExercises, getReviews, getSession, getSubmissions, reviewSubmission, submitDraft } from '@/lib/api'
import type { Dashboard, Exercise, Review, Submission } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

type WorkspaceState = {
  dashboard: Dashboard
  exercise?: Exercise
  submission?: Submission
  review?: Review
}

function countWords(value: string) {
  return value.trim() === '' ? 0 : value.trim().split(/\s+/).length
}

export function CurrentAssignmentView() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sessionRequired, setSessionRequired] = useState(false)
  const [workspace, setWorkspace] = useState<WorkspaceState | null>(null)
  const [draft, setDraft] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        setLoading(true)
        const session = await getSession()
        if (!session.authenticated) {
          if (!cancelled) {
            setSessionRequired(true)
          }
          return
        }
        if (!session.onboarding_complete) {
          router.replace('/onboarding')
          return
        }

        const dashboard = await getDashboard()
        const exercises = await getExercises(1)
        const exercise = exercises[0]
        let submission: Submission | undefined
        let review: Review | undefined
        if (exercise) {
          const submissions = await getSubmissions(exercise.id, 1)
          submission = submissions[0]
          if (submission) {
            const reviews = await getReviews(submission.id, 1)
            review = reviews[0]
          }
        }

        if (!cancelled) {
          setWorkspace({ dashboard, exercise, submission, review })
          setDraft(submission?.content ?? '')
          setSessionRequired(false)
          setError(null)
        }
      } catch (err) {
        if (!cancelled) {
          const message = err instanceof Error ? err.message : 'Failed to load current assignment'
          if (message.toLowerCase().includes('unauthorized')) {
            setSessionRequired(true)
          } else {
            setError(message)
          }
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

  const wordCount = useMemo(() => countWords(draft), [draft])

  async function handleFile(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (!file) {
      return
    }
    setDraft(await file.text())
  }

  async function handleReview() {
    if (!workspace?.exercise || draft.trim() === '') {
      return
    }
    try {
      setSubmitting(true)
      setError(null)
      const submission = await submitDraft({
        exerciseId: workspace.exercise.id,
        content: draft,
        parentSubmissionId: workspace.submission?.id,
      })
      const review = await reviewSubmission(submission.id)
      router.push(`/reviews/${review.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Review failed')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleRevisionPrompt() {
    if (!workspace?.submission) {
      return
    }
    try {
      setSubmitting(true)
      const exercise = await createRevisionAssignment(workspace.submission.id)
      router.push(`/?revisionExercise=${exercise.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create revision prompt')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return <LoadingState />
  }
  if (sessionRequired) {
    return (
      <EmptyState
        title="Sign in to continue"
        body="Authentication runs through the Kratos browser flow. Once you sign in, this page becomes your current assignment workspace."
        actionHref="/login"
        actionLabel="Open sign in"
      />
    )
  }
  if (error) {
    return <EmptyState title="Workspace unavailable" body={error} actionHref="/" actionLabel="Try again" />
  }
  if (!workspace) {
    return <LoadingState />
  }

  const { dashboard, exercise, review } = workspace

  if (!exercise) {
    return (
      <div className="space-y-8">
        <header>
          <Heading>Current assignment</Heading>
          <Text className="mt-2">You do not have an active assignment yet. Choose 3 unlocked TGOs and generate the next prompt.</Text>
        </header>
        <WorkspaceCard>
          <Subheading>Active TGOs</Subheading>
          <div className="mt-4 flex flex-wrap gap-2">
            {dashboard.active_tgos.map((tgo) => (
              <Badge key={tgo.code} color="blue">
                {tgo.title}
              </Badge>
            ))}
          </div>
          <div className="mt-6">
            <Button href="/new-assignment" color="dark/zinc">
              Start new assignment
            </Button>
          </div>
        </WorkspaceCard>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <Heading>{exercise.title}</Heading>
          <Text className="mt-2 max-w-3xl">{exercise.brief}</Text>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button href="/new-assignment" outline>
            New assignment
          </Button>
          {review ? (
            <Button onClick={handleRevisionPrompt} color="dark/zinc" disabled={submitting}>
              Revise from latest review
            </Button>
          ) : null}
        </div>
      </header>

      <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
        <WorkspaceCard>
          <div className="flex items-center justify-between gap-4">
            <Subheading>Prompt</Subheading>
            <Badge color="zinc">{exercise.generation_kind}</Badge>
          </div>
          <div className="mt-4 space-y-5">
            <div>
              <Strong>Constraints</Strong>
              <ul className="mt-2 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                {exercise.constraints.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
            <div>
              <Strong>Success criteria</Strong>
              <ul className="mt-2 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                {exercise.success_criteria.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
          </div>
        </WorkspaceCard>

        <WorkspaceCard>
          <Subheading>Active TGOs</Subheading>
          <Text className="mt-2">These are the three objectives the review will measure most heavily on this assignment.</Text>
          <div className="mt-4 space-y-3">
            {dashboard.active_tgos.map((tgo) => (
              <div key={tgo.code} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                <div className="flex items-center justify-between gap-4">
                  <Strong>{tgo.title}</Strong>
                  <Badge color="cyan">{tgo.stage}</Badge>
                </div>
                <Text className="mt-2">{tgo.description}</Text>
              </div>
            ))}
          </div>
        </WorkspaceCard>
      </div>

      <WorkspaceCard>
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <Subheading>Draft submission</Subheading>
            <Text className="mt-2">Paste plain text or markdown, or load a local draft file before requesting a review.</Text>
          </div>
          <div className="flex items-center gap-3">
            <label className="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-stone-300 px-3 py-2 text-sm font-medium text-zinc-700 hover:bg-stone-50 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5">
              <ArrowUpTrayIcon className="size-4" />
              Upload draft
              <input type="file" accept=".md,.txt,text/plain,text/markdown" className="hidden" onChange={handleFile} />
            </label>
            <Badge color="zinc">{wordCount} words</Badge>
          </div>
        </div>
        <div className="mt-5">
          <Textarea value={draft} onChange={(event) => setDraft(event.target.value)} rows={18} placeholder="Paste your draft here." />
        </div>
        <div className="mt-5 flex flex-wrap items-center justify-between gap-3">
          <Text>{workspace.submission ? `Latest saved draft: #${workspace.submission.draft_number}` : 'No draft submitted yet for this assignment.'}</Text>
          <Button onClick={handleReview} color="dark/zinc" disabled={submitting || draft.trim() === ''}>
            <SparklesIcon />
            {submitting ? 'Reviewing…' : 'Submit for review'}
          </Button>
        </div>
      </WorkspaceCard>

      {review ? (
        <WorkspaceCard>
          <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <Subheading>Latest coaching pass</Subheading>
              <Text className="mt-2">{review.summary}</Text>
            </div>
            <div className="flex gap-2">
              <Button href={`/reviews/${review.id}`} outline>
                Open review
              </Button>
              <Button href={`/compare/${workspace.submission?.id ?? 0}`} plain>
                Compare drafts
              </Button>
            </div>
          </div>
        </WorkspaceCard>
      ) : null}
    </div>
  )
}
