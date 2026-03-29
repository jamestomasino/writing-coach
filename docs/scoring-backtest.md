# Scoring Backtest

Use the backtest tool to compare stored deterministic scores against current rubric output.

## Run

```bash
go run ./cmd/scoring-backtest --db ./.writing-coach/writing-coach.db --limit-per-track 80 --out reports/scoring-backtest-$(date +%F).md
```

If `--db` is omitted, the tool first tries the configured `database_url`, then falls back to `./.writing-coach/writing-coach.db` if present.
By default, the tool includes score sources `deterministic,llm_or_legacy` to support mixed historical datasets.

## Output

The report includes:

- per-track and per-domain `% of 5s` before vs after
- average score shift
- top `5 -> 4` downgrade candidates with score-drop reasons
- a data sufficiency warning when the sample size is below the chosen threshold

## Flags

- `--db`: sqlite path
- `--limit-per-track`: max submissions per track (newest first)
- `--min-samples`: minimum total submissions before treating results as stable
- `--score-sources`: comma-separated `score_source` values to include, or `any`
- `--out`: write report to file (optional)
