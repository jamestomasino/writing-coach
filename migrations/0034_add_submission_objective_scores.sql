CREATE TABLE IF NOT EXISTS submission_objective_scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submission_id INTEGER NOT NULL REFERENCES submissions(id),
    tgo_code TEXT NOT NULL REFERENCES tgo_catalog(code),
    score INTEGER NOT NULL,
    score_source TEXT NOT NULL DEFAULT 'deterministic',
    score_version TEXT NOT NULL DEFAULT 'obj-det-v1',
    score_evidence_json TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_submission_objective_scores_submission
    ON submission_objective_scores(submission_id);

CREATE INDEX IF NOT EXISTS idx_submission_objective_scores_tgo
    ON submission_objective_scores(tgo_code);
