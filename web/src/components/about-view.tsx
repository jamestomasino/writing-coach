'use client'

import { Button } from '@/components/button'
import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Strong, Text } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import { getSession } from '@/lib/api'
import { useEffect, useState } from 'react'

type SessionState = {
  checked: boolean
  authenticated: boolean
}

function AboutCard({ eyebrow, title, body }: { eyebrow: string; title: string; body: string }) {
  return (
    <WorkspaceCard className="border-stone-200/80 bg-white/90 dark:border-white/10 dark:bg-zinc-900">
      <Eyebrow>{eyebrow}</Eyebrow>
      <Subheading className="mt-3 text-xl/7 sm:text-lg/7">{title}</Subheading>
      <Text className="mt-3">{body}</Text>
    </WorkspaceCard>
  )
}

export function AboutView() {
  const [session, setSession] = useState<SessionState>({ checked: false, authenticated: false })

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const next = await getSession()
        if (!cancelled) {
          setSession({ checked: true, authenticated: next.authenticated })
        }
      } catch {
        if (!cancelled) {
          setSession({ checked: true, authenticated: false })
        }
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="space-y-8">
      <section className="overflow-hidden rounded-[2rem] border border-stone-200 bg-linear-to-br from-stone-100 via-white to-sky-50 p-8 shadow-sm ring-1 ring-black/3 dark:border-white/10 dark:from-zinc-900 dark:via-zinc-900 dark:to-sky-950/40 dark:ring-white/10">
        <div className="max-w-3xl">
          <Eyebrow className="tracking-[0.22em]">About Writing Coach</Eyebrow>
          <Heading className="mt-4 text-4xl/11 sm:text-3xl/10">Focused practice for writers who want to improve on purpose</Heading>
          <Text className="mt-4 text-lg/8 text-zinc-600 dark:text-zinc-300">
            Writing Coach gives you one assignment, a few clear craft goals, and feedback you can use right away. It is
            designed to help you build skill over time, not just produce one lucky draft.
          </Text>
          <div className="mt-6 flex flex-wrap gap-3">
            {session.authenticated ? (
              <>
                <Button href="/" color="dark/zinc">
                  Open current assignment
                </Button>
                <Button href="/progress" outline>
                  View progress
                </Button>
              </>
            ) : (
              <>
                <Button href="/register" color="dark/zinc">
                  Register
                </Button>
                <Button href="/login" outline>
                  Sign in
                </Button>
              </>
            )}
          </div>
        </div>
      </section>

      <div className="grid gap-6 xl:grid-cols-[1.3fr_0.9fr]">
        <WorkspaceCard className="border-stone-200/80 bg-stone-50 dark:border-white/10 dark:bg-white/5">
          <Subheading className="text-xl/7 sm:text-lg/7">How the skill system works</Subheading>
          <div className="mt-4 space-y-4">
            <Text>
              Writing is broken into small, teachable skills. A skill might be <Strong>causal clarity</Strong>,
              <Strong> scene architecture</Strong>, or <Strong>prose precision</Strong>. You do not work on everything
              at once.
            </Text>
            <Text>
              Each assignment focuses on just a few skills. Those become the goals for that piece. The review then
              checks the draft mainly against those goals, so the feedback stays specific.
            </Text>
            <Text>
              When you show the same skill consistently, it moves toward mastery. Then the coach can unlock the next
              challenge. That keeps the work focused and makes progress easier to see.
            </Text>
          </div>
        </WorkspaceCard>

        <WorkspaceCard className="border-stone-200/80 bg-zinc-950 text-white dark:border-white/10">
          <Subheading className="text-xl/7 text-white sm:text-lg/7">What you get</Subheading>
          <ul className="mt-4 space-y-3 text-sm/6 text-zinc-300">
            <li>Assignments built around the skills you need most right now.</li>
            <li>Feedback tied to clear goals instead of vague general advice.</li>
            <li>Revision briefs when a draft deserves another pass.</li>
            <li>A progress view that shows where your control is improving.</li>
            <li>A visible path to the next skill, not just the next prompt.</li>
          </ul>
        </WorkspaceCard>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <AboutCard
          eyebrow="Focus"
          title="You practice one thing at a time"
          body="Most writers get overloaded when every draft is judged on everything. Writing Coach narrows the target. That makes it easier to see what changed and easier to repeat the improvement in later work."
        />
        <AboutCard
          eyebrow="Feedback"
          title="The feedback is meant to be usable"
          body="The goal is not to sound clever. The goal is to help you revise. Reviews point to a few strengths to keep, a few weaknesses to fix, and the next place to put your attention."
        />
        <AboutCard
          eyebrow="Progress"
          title="The system remembers what happened"
          body="Old assignments, drafts, reviews, and revisions stay connected. That history helps the coach see what is improving, what keeps slipping, and what to assign next."
        />
      </div>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading className="text-xl/7 sm:text-lg/7">What makes this different</Subheading>
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">It is more than a prompt generator</div>
            <Text className="mt-2">
              The assignment matters, but the real value is the training structure behind it. The coach remembers what
              was practiced, what improved, and what still needs work.
            </Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">It is feedback with continuity</div>
            <Text className="mt-2">
              Each review is tied to the goals of that assignment. Later, you can look back at older work and still see
              what it trained. Your history stays useful.
            </Text>
          </div>
        </div>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-linear-to-br from-stone-50 via-white to-sky-50 dark:border-white/10 dark:from-zinc-900 dark:via-zinc-900 dark:to-sky-950/30">
        <Eyebrow>How It Works</Eyebrow>
        <Subheading className="mt-3 text-xl/7 sm:text-lg/7">A guided review, not just a chatbot answer</Subheading>
        <div className="mt-4 max-w-3xl space-y-4">
          <Text>
            Writing Coach does not simply send your draft to a language model and accept whatever comes back. It starts
            with concrete checks for things like grammar problems, style issues, sentence strain, and repeated
            patterns. Tools such as <Strong>LanguageTool</Strong>, <Strong>Vale</Strong>, and the app&apos;s own
            sentence-analysis checks help build that first layer.
          </Text>
          <Text>
            After that, the coach checks your <Strong>active skills</Strong>. These are the few craft goals your
            assignment is meant to train right now. Instead of trying to judge every part of the draft at once, the
            review stays focused on those goals.
          </Text>
          <Text>
            When a language model is enabled, it comes in after those checks. The model helps write the assignment,
            shape the coaching language, and turn the raw signals into feedback that reads more like a teacher. It is
            working from structure and evidence, not guessing from scratch.
          </Text>
          <Text>
            Over time, the app compares reviews, tracks which skills are improving, watches for old weaknesses coming
            back, and decides what to practice next. That is how it aims to build repeatable skill, not just produce
            one interesting comment on one draft.
          </Text>
        </div>

        <div className="mt-6 grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">1. Check the draft</div>
            <Text className="mt-2">
              The system scans for concrete issues it can measure reliably.
            </Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">2. Focus the review</div>
            <Text className="mt-2">
              The coach matches those signals to the few skills that matter on this assignment.
            </Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">3. Shape the next step</div>
            <Text className="mt-2">
              The review, revision brief, and next assignment are built from that evidence.
            </Text>
          </div>
        </div>
      </WorkspaceCard>
    </div>
  )
}
