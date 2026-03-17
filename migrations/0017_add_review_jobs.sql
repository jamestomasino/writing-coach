CREATE TABLE IF NOT EXISTS review_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  tree_id INTEGER NOT NULL,
  enrollment_id INTEGER NOT NULL,
  submission_id INTEGER NOT NULL UNIQUE,
  review_id INTEGER,
  status TEXT NOT NULL DEFAULT 'queued',
  attempt_count INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL DEFAULT 3,
  last_error TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (tree_id) REFERENCES trees(id),
  FOREIGN KEY (enrollment_id) REFERENCES user_tree_enrollments(id),
  FOREIGN KEY (submission_id) REFERENCES submissions(id),
  FOREIGN KEY (review_id) REFERENCES reviews(id)
);

CREATE INDEX IF NOT EXISTS idx_review_jobs_status_updated ON review_jobs(status, updated_at);
