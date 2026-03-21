# Release Process

This project uses Calendar Versioning and keeps release notes in `CHANGELOG.md`.

## Current Convention

- `main` is the release line
- release tags use CalVer
- unreleased notes live under `## Unreleased` in `CHANGELOG.md`

Current tag examples:

- `2026.03.21`
- `2026.03.20`

## Versioning

CalVer format:

- `YYYY.MM.DD`

Example:

- `2026.03.21`

Tags are release markers, not branch names.

## Changelog Rules

- update `CHANGELOG.md` for user-facing or release-facing changes
- add new work under `## Unreleased`
- keep entries concise and grouped by outcome
- do not backfill old release sections unless there is a specific reason

## Release Steps

When preparing a release:

1. Make sure `main` contains the intended shipped state.
2. Finalize the `## Unreleased` section in `CHANGELOG.md`.
3. Move those notes into a dated release section.
4. Commit the changelog update if needed.
5. Create an annotated tag for the release date.
6. Push `main`.
7. Push the new tag.

Example:

```bash
git checkout main
git pull --ff-only
git tag -a 2026.03.21 -m "2026.03.21"
git push
git push origin 2026.03.21
```

## Working Practice

Feature branches and PRs are preferred when they help keep work reviewable, but the release process itself depends on the state of `main`, the changelog, and the CalVer tag.
