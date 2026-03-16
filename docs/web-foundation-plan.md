# Web Foundation Plan

## Decision

Build the web app on:

- `Next.js 16`
- `React 19`
- `Tailwind CSS 4`
- `TypeScript`

Use:

- `Catalyst` as the primary application UI system
- `Compass` as a secondary source for editorial and curriculum-navigation patterns

Do not build from the Compass template wholesale. Its content model is too course/article oriented for the core app. Instead, lift only the layout and content-navigation ideas that fit the coaching product.

## Available Source Kits

### Catalyst

Source location:

- [tailwind-plus/catalyst-ui-kit/catalyst-ui-kit/typescript](/home/tomasino/writing-coach/tailwind-plus/catalyst-ui-kit/catalyst-ui-kit/typescript)
- [tailwind-plus/catalyst-ui-kit/catalyst-ui-kit/demo/typescript](/home/tomasino/writing-coach/tailwind-plus/catalyst-ui-kit/catalyst-ui-kit/demo/typescript)

Primary strengths:

- authenticated app shell
- sidebar and navbar patterns
- forms and fieldsets
- dialogs and dropdowns
- tables and pagination
- stable app primitives for all data-entry and review workflows

Important components:

- `sidebar-layout.tsx`
- `sidebar.tsx`
- `navbar.tsx`
- `table.tsx`
- `dialog.tsx`
- `fieldset.tsx`
- `input.tsx`
- `textarea.tsx`
- `select.tsx`
- `combobox.tsx`
- `badge.tsx`
- `alert.tsx`

### Compass

Source location:

- [tailwind-plus/compass-template/tailwind-plus-compass/compass-ts](/home/tomasino/writing-coach/tailwind-plus/compass-template/tailwind-plus-compass/compass-ts)

Primary strengths:

- content-heavy information architecture
- centered reading layouts
- sidebar knowledge navigation
- breadcrumbs and next-step navigation
- curriculum/lesson presentation patterns

Important components:

- `centered-layout.tsx`
- `sidebar-layout.tsx`
- `breadcrumbs.tsx`
- `content-link.tsx`
- `page-section.tsx`
- `table-of-contents.tsx`
- `next-page-link.tsx`

## Recommended Product Structure

Use three UI modes:

1. `Application Mode`
For daily work.
Powered by Catalyst.

Screens:

- dashboard
- prompt detail
- submission editor
- review detail
- revision compare
- progress board
- user/tree management

2. `Curriculum Mode`
For browsing TGO trees, completed objectives, and learning paths.
Catalyst shell plus Compass content/navigation patterns.

Screens:

- tree overview
- TGO detail
- enrollment board
- mastery history
- regression history

3. `Reference Mode`
For longer explanations of TGOs, pedagogy, and writing guidance.
Mostly Compass-inspired.

Screens:

- methodology pages
- “how reviews work”
- “how TGOs advance”
- author onboarding/explanations

## Screen Mapping

### Use Catalyst Directly

- `/app/dashboard`
- `/app/prompts/[id]`
- `/app/submissions/[id]`
- `/app/reviews/[id]`
- `/app/compare/[submissionId]`
- `/app/progress`
- `/app/users`
- `/app/enrollments`
- `/app/settings`

Reason:

- these are dense, transactional, stateful screens
- Catalyst already solves the shell, controls, and data display patterns

### Use Compass Patterns Selectively

- `/curriculum`
- `/curriculum/trees/[slug]`
- `/curriculum/tgos/[code]`
- `/curriculum/enrollments/[id]/board`
- `/docs/method`

Reason:

- these are explanatory and navigational
- Compass is better for reading flow, sectioning, and next-step guidance

## API To Screen Mapping

Current backend endpoints now support the first UI slice:

- `GET /api/dashboard`
- `POST /api/prompts/next`
- `POST /api/prompts/revise`
- `POST /api/submissions`
- `GET /api/submissions/{id}`
- `POST /api/reviews`
- `GET /api/compare`
- `GET /api/reviews/{id}`
- `GET /api/users`
- `GET /api/users/{slug}`
- `GET /api/trees`
- `GET /api/trees/{slug}`
- `GET /api/enrollments`
- `POST /api/enrollments`
- `GET /api/enrollments/{id}/board`

This is enough for:

- dashboard
- prompt generation
- submission/review loop
- enrollment switching
- curriculum browsing

The current first slice is now implemented in `web/`:

- `/onboarding`
- `/`
- `/new-assignment`
- `/reviews/[id]`
- `/compare/[submissionId]`
- `/progress`
- `/admin`
- `/login`
- `/register`

Remaining gaps before the workflow feels fully complete:

- a real invite flow instead of admin-side user provisioning
- true inline review annotations
- richer tree browsing and editing views

## Dependency Strategy

Base dependencies to plan for:

- `next`
- `react`
- `react-dom`
- `tailwindcss`
- `@tailwindcss/postcss`
- `@headlessui/react`
- `@heroicons/react`
- `clsx`
- `motion`

Optional dependencies from Compass:

- `@mdx-js/react`
- `@next/mdx`
- `shiki`
- `geist`

Recommendation:

- include the base Catalyst stack immediately
- do not include MDX/content tooling unless we actually decide to ship reference/docs pages in the first UI phase
- do not inherit Compass wholesale dependencies unless the screen plan requires them

## Import Strategy

Do not import directly from `tailwind-plus/` at runtime.

Instead:

1. create a dedicated frontend app directory when ready
2. copy the Catalyst TypeScript components you actually use into that app
3. create a small internal design-system layer around them
4. selectively adapt Compass patterns into app-specific components

Reason:

- avoids coupling the product to the raw vendor dump
- keeps imports stable
- makes refactoring and styling consistent
- reduces dead code and unclear provenance

## Styling Direction

Default visual direction should come from Catalyst, not Compass.

Why:

- Catalyst is the cleaner base for a serious app
- Compass is useful for content presentation but too editorial to govern the entire product

Suggested split:

- shell, controls, tables, forms: Catalyst
- curriculum reading surfaces: Compass-inspired
- typography and spacing tokens: align to Catalyst first, then selectively borrow Compass reading treatments

## What Else Is Needed From Tailwind Plus

Nothing required right now.

Catalyst plus Compass already provide enough licensed source to build:

- the app shell
- workflow screens
- curriculum views
- auth entry screens
- documentation/reference surfaces

Additional Tailwind Plus templates would only be optional references, not blockers.

## Recommended First Web Build Slice

When frontend work begins, build only this thin vertical slice:

1. authenticated app shell
2. dashboard using `GET /api/dashboard`
3. generate prompt using `POST /api/prompts/next`
4. create submission using `POST /api/submissions`
5. run review using `POST /api/reviews`
6. show compare/progress links

Do not begin with the curriculum docs or richer editorial screens. Those can wait until the daily loop is stable in the browser.
