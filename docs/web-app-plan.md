# Web App Plan

## Product Model

The web client is a guided writing workshop, not a general dashboard.

Primary user flow:

1. authenticate
2. land on the current assignment
3. read the prompt and the 3 active TGOs
4. paste or upload a draft
5. submit for review
6. read the review with inline and summary feedback
7. choose to revise or move on
8. inspect the full assignment timeline or archive later if needed

The home screen is therefore the assignment workspace rather than a reporting page.

## Information Architecture

Top-level routes for the first web slice:

- `/` current assignment workspace
- `/new-assignment` create the next assignment from unlocked TGOs
- `/reviews/[id]` full review workspace
- `/compare/[submissionId]` revision comparison
- `/progress` TGO progress and recent coaching history
- `/admin` admin-only view for invite and account management
- `/login` auth entry point
- `/register` auth registration entry point

Future routes:

- `/onboarding`
- `/tree`
- `/tree/refresh`

Current history routes:

- `/assignments` assignment archive
- `/assignments/[id]` assignment timeline

## Current Assignment UX

The default page should answer four questions immediately:

1. What am I working on?
2. Which 3 TGOs matter most on this assignment?
3. Where do I submit the draft?
4. What happened on the latest attempt?

The page layout should therefore include:

- prompt card
- active TGO strip
- draft paste/upload area
- recent submission and review status
- next action buttons

If no assignment exists:

- show the active TGOs
- explain that no assignment is active
- route the user to `New Assignment`

## Review UX

Review pages should contain:

- executive summary
- active TGO assessment cards
- completed TGO regression check
- strengths and weaknesses
- analyzer findings
- inline markup panel
- revision compare CTA

The UI should present active TGOs as the primary rubric and completed TGO checks as a lighter maintenance pass.

Queue-state behavior:

- review submission can return immediately while a background job runs
- revision brief generation can also take a short background turn before redirect
- those states should use the same progress treatment so the product feels coherent across assignment generation, review, and revision work

## Assignment Archive UX

The archive should present past assignments as history, not as active workflow stages.

Rules:

- the current item should be labeled `Current assignment`, not `Current chain`
- archive cards should lead with the assignment title instead of an internal stage eyebrow
- skill tags should reflect current skill state where possible so old assignments remain useful as progress artifacts
- the current assignment page CTA is not needed from the archive header if the current timeline card already exposes the relevant action

## TGO Selection UX

New assignments are created from exactly 3 TGOs.

Selection rules:

- choices must come from the unlocked set
- already active TGOs remain selectable
- completed TGOs are not selectable
- the user must end with exactly 3

Once selected, prompt generation uses those TGOs plus the user’s writing goals.

## Admin Scope

Admin remains intentionally narrow:

- invite or provision users
- inspect current users

Tree editing and advanced curriculum administration remain backend features for now, but they are not part of the first browser slice.

## Design Direction

The interface should feel:

- scholastic
- workshop-like
- structured
- calm

Implementation direction:

- Catalyst drives the app shell, forms, cards, and dense workflow surfaces
- Compass remains a secondary reference for future curriculum and explanatory pages
- the visual hierarchy should privilege assignment flow over status tiles

## Frontend Contract Expectations

The first slice relies on:

- `GET /api/auth/session`
- `GET /api/dashboard`
- `GET /api/exercises`
- `POST /api/prompts/next`
- `POST /api/prompts/revise`
- `GET /api/submissions`
- `POST /api/submissions`
- `GET /api/reviews`
- `POST /api/reviews`
- `GET /api/compare`
- `GET /api/admins`
- `GET /api/users`

The web app should run as the only published service in Compose and proxy:

- `/api/*` to the Go API
- `/.ory/kratos/public/*` to Kratos public
- branded auth routes in the web app, backed by Kratos browser flows

That keeps host nginx simple: one upstream for the entire product.
