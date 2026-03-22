CREATE TABLE IF NOT EXISTS ai_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  tree_id INTEGER NOT NULL,
  enrollment_id INTEGER NOT NULL,
  kind TEXT NOT NULL,
  resource_key TEXT NOT NULL DEFAULT '',
  exercise_id INTEGER,
  submission_id INTEGER,
  review_id INTEGER,
  status TEXT NOT NULL DEFAULT 'queued',
  attempt_count INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL DEFAULT 3,
  last_error TEXT NOT NULL DEFAULT '',
  payload_json TEXT NOT NULL DEFAULT '{}',
  result_json TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (tree_id) REFERENCES trees(id),
  FOREIGN KEY (enrollment_id) REFERENCES user_tree_enrollments(id),
  FOREIGN KEY (exercise_id) REFERENCES exercises(id),
  FOREIGN KEY (submission_id) REFERENCES submissions(id),
  FOREIGN KEY (review_id) REFERENCES reviews(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_jobs_resource_key ON ai_jobs(resource_key) WHERE resource_key <> '';
CREATE INDEX IF NOT EXISTS idx_ai_jobs_status_updated ON ai_jobs(status, updated_at);
CREATE INDEX IF NOT EXISTS idx_ai_jobs_user_tree_kind ON ai_jobs(user_id, tree_id, kind, created_at);
CREATE INDEX IF NOT EXISTS idx_ai_jobs_submission_kind ON ai_jobs(submission_id, kind, created_at);
