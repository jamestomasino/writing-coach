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

- Added a Dockerized spaCy/TextDescriptives analyzer sidecar and wired its findings into the deterministic review pipeline.
- Expanded the About page with a plain-language explanation of deterministic analysis, active skills, and model-assisted coaching.
- Unified assignment, review, and revision queue loaders around the same progress treatment.
- Refined the assignment archive language and visuals so current work and historical assignments read more naturally to users.

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
