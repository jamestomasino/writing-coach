CREATE TABLE IF NOT EXISTS decision_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    tree_id INTEGER NOT NULL REFERENCES tgo_trees(id),
    enrollment_id INTEGER NOT NULL REFERENCES user_tree_enrollments(id),
    review_id INTEGER REFERENCES reviews(id),
    submission_id INTEGER REFERENCES submissions(id),
    event_type TEXT NOT NULL,
    decision_payload_json TEXT NOT NULL DEFAULT '{}',
    rule_version TEXT NOT NULL DEFAULT '',
    evidence_refs_json TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_decision_events_review_event
ON decision_events(review_id, event_type);

CREATE INDEX IF NOT EXISTS idx_decision_events_enrollment_created
ON decision_events(enrollment_id, created_at DESC);
