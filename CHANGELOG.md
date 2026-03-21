# Changelog

This project uses Calendar Versioning (`YYYY.MM.DD`) for release tags.

This changelog starts after the initial launch state. Early development and launch work are intentionally not backfilled here.

Development workflow moving forward:

- build changes on feature branches
- merge through pull requests
- squash merge into `main`
- cut release tags from `main`
- move release-facing notes from `Unreleased` into the dated release section when tagging

For the full workflow, see [docs/release-process.md](/home/tomasino/writing-coach/docs/release-process.md).

## Unreleased

### Added

- Added a spaCy/TextDescriptives NLP analyzer sidecar for richer deterministic review signals.
- Added localization infrastructure with `next-intl` and an English message catalog.
- Added domain-aware Vale style selection across fiction, technical, academic, professional, marketing, and thought-leadership writing.
- Added progress dashboard stats for drafts, revisions, and completed assignment chains.

### Changed

- Rewrote user-facing copy across the app to use clearer terminology, including `Practice Path` and `skills`.
- Moved more About, Admin, Progress, and error-page UI copy into localization messages.
- Made deterministic heuristics, NLP analysis, and fallback review language domain-aware across writing types.
- Improved admin AI provider event categories and user labeling.
- Updated README and architecture docs for the analyzer pipeline and archive/progress UX.

### Fixed

- Fixed a `/progress` crash caused by missing history skill tag data.
- Fixed progress dashboard counts so revision rounds no longer inflate draft and completed-assignment totals.
- Fixed archive assignment summaries so revision submissions no longer inflate draft totals.

## 2026.03.21

- Added multi-track coaching so each user can create, switch, and archive independent writing tracks with separate progress and history.
- Added a guided first-run setup funnel covering AI readiness, first-track creation, and first-assignment creation.
- Hardened backend request handling, auth defaults, submission ownership checks, and shutdown behavior.
- Refactored major frontend and backend controller/API boundaries to reduce branch-specific structural drift.
- Added targeted browser integration coverage for the core multi-track onboarding, switching, and archiving flows.

## 2026.03.20

- Established repository licensing structure for a mixed-license codebase.
- Added a root GPL license for original project code.
- Added repository-level licensing guidance in `NOTICE.md` and `docs/licensing.md`.
- Reworked the README to better support deployment and self-hosting.
