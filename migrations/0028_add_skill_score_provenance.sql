ALTER TABLE submission_skill_scores
ADD COLUMN score_source TEXT NOT NULL DEFAULT 'deterministic';

ALTER TABLE submission_skill_scores
ADD COLUMN score_version TEXT NOT NULL DEFAULT 'legacy-unknown';

ALTER TABLE submission_skill_scores
ADD COLUMN score_evidence_json TEXT NOT NULL DEFAULT '{}';

UPDATE submission_skill_scores
SET score_source = 'llm_or_legacy'
WHERE score_source = 'deterministic' AND score_version = 'legacy-unknown';

CREATE INDEX IF NOT EXISTS idx_submission_skill_scores_submission_skill
ON submission_skill_scores (submission_id, skill_name);

CREATE INDEX IF NOT EXISTS idx_submission_skill_scores_source_version
ON submission_skill_scores (score_source, score_version);

CREATE INDEX IF NOT EXISTS idx_submission_skill_scores_skill_source
ON submission_skill_scores (skill_name, score_source);
