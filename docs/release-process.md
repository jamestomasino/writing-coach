# Release Process

This project uses a branch-and-PR workflow with squash merges and Calendar Versioning.

## Branching

- Create new work on a feature branch off `main`.
- Use one branch per feature, fix, or focused documentation change.
- Keep branch names descriptive, for example:
  - `feature/assignment-archive`
  - `fix/review-history-actions`
  - `docs/release-process`

## Pull Requests

- Open a pull request into `main`.
- Keep the PR scope coherent and reviewable.
- Update user-facing docs when behavior, setup, deployment, or workflow changes.
- Update `CHANGELOG.md` for changes that should appear in release history.
  - Add entries under `## Unreleased`.
  - Do not backfill older release history unless explicitly needed.

## Merge Strategy

- Merge approved PRs with **squash merge**.
- The squash commit should clearly describe the shipped change.
- `main` should remain the release-ready branch.

## Versioning

This project uses Calendar Versioning:

- format: `YYYY.MM.DD`
- example: `2026.03.20`

Tags are release markers, not branch names.

## Release Steps

When preparing a release from `main`:

1. Make sure intended changes are already merged through PRs.
2. Review `CHANGELOG.md` and finalize the `## Unreleased` section.
3. Create a new dated release section in `CHANGELOG.md`.
4. Commit the changelog update if needed.
5. Create an annotated tag using the CalVer date.
6. Push `main`.
7. Push the new tag.

Example:

```bash
git checkout main
git pull --ff-only
git tag -a 2026.03.20 -m "2026.03.20"
git push
git push origin 2026.03.20
```

## Changelog Rules

- `CHANGELOG.md` tracks release-facing changes moving forward.
- Keep entries concise and user-meaningful.
- Group related work together instead of listing every internal patch.
- Leave unreleased work under `## Unreleased` until the next tag is cut.

## Practical Defaults

- Build features in feature branches.
- Review through PRs.
- Squash merge to `main`.
- Tag releases from `main`.
- Record release-facing changes in `CHANGELOG.md`.
