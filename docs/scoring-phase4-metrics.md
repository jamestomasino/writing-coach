# Phase 4 Deterministic Metric Contract

This document freezes the deterministic Phase 4 metric keys emitted by the NLP analyzer sidecar and consumed by rubric scoring.

## Metrics

| Metric | Range | Meaning | Suggested interpretation |
|---|---:|---|---|
| `nlp_claim_count` | 0+ | Count of sentences with explicit claim markers (`should`, `must`, `recommend`, etc.) | Higher means more explicit argumentative/recommendation load. |
| `nlp_evidence_marker_count` | 0+ | Count of sentences with evidence markers (`because`, `for example`, `data`, `according to`, etc.) | Higher means more explicit support language. |
| `nlp_claim_evidence_coverage` | 0..100 | Percent of claims with corresponding evidence-marker coverage | `<40` weak support signal; `>=60` generally healthy support signal. |
| `nlp_coref_ambiguity_count` | 0+ | Count of ambiguous referent events (heuristic + optional CoreNLP chain signal) | `>=3` often indicates referent instability; `<=1` usually stable. |
| `nlp_semantic_repetition_ratio` | 0..100 | Percent of similar sentence pairs (high means repeated semantic content) | `>=60` likely redundancy; `<=35` generally acceptable repetition. |
| `nlp_topic_drift_score` | 0..100 | Paragraph-level lexical drift from anchor section | `>=70` strong drift risk; `<=45` generally cohesive progression. |

## Stability Rules

- Keys above are additive and must remain backward-compatible once released.
- If optional tooling is unavailable (for example CoreNLP), the analyzer must still emit keys with deterministic fallback values.
- Rubric logic should treat absent metrics as no-op, not fatal.
