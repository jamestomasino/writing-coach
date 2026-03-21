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
          <Heading className="mt-4 text-4xl/11 sm:text-3xl/10">Practice that builds real writing skill</Heading>
          <Text className="mt-4 text-lg/8 text-zinc-600 dark:text-zinc-300">
            Writing Coach is built around a skill tree. Each assignment gives you one clear place to start. Each review
            checks a small set of craft goals. Over time, the work gets harder in a steady way.
          </Text>
          <div className="mt-6 flex flex-wrap gap-3">
            {session.authenticated ? (
              <>
                <Button href="/" color="dark/zinc">
                  Open active track
                </Button>
                <Button href="/progress" outline>
                  View progress board
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
          <Subheading className="text-xl/7 sm:text-lg/7">How the skill tree works</Subheading>
          <div className="mt-4 space-y-4">
            <Text>
              The tree breaks writing into small skills. A skill might be <Strong>causal clarity</Strong>,
              <Strong> scene architecture</Strong>, or <Strong>prose precision</Strong>. You do not try to fix
              everything at once.
            </Text>
            <Text>
              The coach activates a few skills at a time. Those are your current goals. New assignments are shaped by
              your track, audience, and subject matter. Reviews then judge the draft against the active skill goals.
            </Text>
            <Text>
              When a skill is strong again and again, it moves toward mastery. Then the coach can unlock the next part
              of the tree. That keeps the work focused and keeps progress easy to see.
            </Text>
          </div>
        </WorkspaceCard>

        <WorkspaceCard className="border-stone-200/80 bg-zinc-950 text-white dark:border-white/10">
          <Subheading className="text-xl/7 text-white sm:text-lg/7">What the coach offers</Subheading>
          <ul className="mt-4 space-y-3 text-sm/6 text-zinc-300">
            <li>Targeted assignments shaped by the active track.</li>
            <li>Reviews tied to a small set of active skills.</li>
            <li>Revision briefs when a draft needs another pass.</li>
            <li>A progress board that shows growth over time.</li>
            <li>A skill map that explains what comes next.</li>
          </ul>
        </WorkspaceCard>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <AboutCard
          eyebrow="Pedagogy"
          title="Why this method works"
          body="People improve faster when practice is specific, feedback is quick, and the next step is clear. Writing Coach uses that pattern on purpose. It narrows the target, asks for a fresh draft, then checks only the skills that matter most in that moment."
        />
        <AboutCard
          eyebrow="Science"
          title="Built on deliberate practice"
          body="Research on skill growth shows that strong practice is not just repetition. It is focused work with feedback, reflection, and a rising challenge level. The coach uses those ideas to keep you working near the edge of your current control without turning every draft into a full rewrite of every weakness."
        />
        <AboutCard
          eyebrow="Goal"
          title="Made for durable progress"
          body="The point is not to get one good draft by luck. The point is to build habits that keep showing up in new work. That is why the coach tracks mastered skills, watches for slips, and stores the assignments that helped you grow."
        />
      </div>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading className="text-xl/7 sm:text-lg/7">What makes this different</Subheading>
        <div className="mt-4 grid gap-4 md:grid-cols-2">
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">
              Assignment-first, but not assignment-only
            </div>
            <Text className="mt-2">
              Prompts give you a strong starting point. The real teaching happens in the skill system behind them. The
              coach remembers what was practiced, what improved, and what still needs work.
            </Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">A review system with memory</div>
            <Text className="mt-2">
              Each review looks at the active goals for that assignment. Later, you can look back at old assignments and
              see which skills they trained. That makes your history useful, not just stored.
            </Text>
          </div>
        </div>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-linear-to-br from-stone-50 via-white to-sky-50 dark:border-white/10 dark:from-zinc-900 dark:via-zinc-900 dark:to-sky-950/30">
        <Eyebrow>How It Works</Eyebrow>
        <Subheading className="mt-3 text-xl/7 sm:text-lg/7">A guided review, not just a chatbot reply</Subheading>
        <div className="mt-4 max-w-3xl space-y-4">
          <Text>
            Writing Coach does not hand your draft to a model and hope for the best. It starts with a set of
            rule-based checks that look for concrete things like grammar trouble, style problems, sentence strain, and
            repeated patterns. Tools such as <Strong>LanguageTool</Strong>, <Strong>Vale</Strong>, and the app&apos;s
            own sentence-analysis checks help build that first layer.
          </Text>
          <Text>
            After that, the coach looks at your <Strong>active skills</Strong>. Those are the few craft goals your
            assignment is meant to train right now. Instead of trying to judge everything at once, the review asks
            whether this draft got clearer in the specific skills you were working on.
          </Text>
          <Text>
            When a language model is enabled, it comes in after those checks. The model helps write the assignment
            brief, shape the coaching language, and turn the raw signals into feedback that reads more like a human
            teacher. It is working with structure and evidence from the system, not starting from a blank guess.
          </Text>
          <Text>
            Over time, the app compares reviews, tracks which skills are improving, watches for old weaknesses coming
            back, and decides what to practice next. That is how it aims to build steady skill, not just generate one
            clever comment on one draft.
          </Text>
        </div>

        <div className="mt-6 grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">1. Check the draft</div>
            <Text className="mt-2">
              Deterministic tools scan for concrete issues the system can measure reliably.
            </Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">2. Match the current skills</div>
            <Text className="mt-2">
              The coach cross-references those signals with the few skills that matter for your current assignment.
            </Text>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-white/80 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">3. Shape the next step</div>
            <Text className="mt-2">
              The review, revision brief, and next assignment are built from that evidence so practice stays focused.
            </Text>
          </div>
        </div>
      </WorkspaceCard>
    </div>
  )
}
