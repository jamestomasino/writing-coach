CREATE TABLE IF NOT EXISTS playground_sessions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  tree_id INTEGER NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  writing_language TEXT NOT NULL DEFAULT '',
  writing_type TEXT NOT NULL DEFAULT '',
  assignment_format TEXT NOT NULL DEFAULT '',
  coaching_brief TEXT NOT NULL DEFAULT '',
  latest_review_id INTEGER,
  latest_review_at DATETIME,
  review_count INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (tree_id) REFERENCES trees(id),
  FOREIGN KEY (latest_review_id) REFERENCES playground_reviews(id)
);

CREATE TABLE IF NOT EXISTS playground_reviews (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  tree_id INTEGER NOT NULL,
  review_kind TEXT NOT NULL DEFAULT '',
  provider_note TEXT NOT NULL DEFAULT '',
  summary TEXT NOT NULL DEFAULT '',
  strengths_json TEXT NOT NULL DEFAULT '[]',
  weaknesses_json TEXT NOT NULL DEFAULT '[]',
  analyzer_findings_json TEXT NOT NULL DEFAULT '[]',
  next_focus TEXT NOT NULL DEFAULT '',
  metric_word_count INTEGER NOT NULL DEFAULT 0,
  skill_scores_json TEXT NOT NULL DEFAULT '[]',
  tgo_assessments_json TEXT NOT NULL DEFAULT '[]',
  completed_tgo_checks_json TEXT NOT NULL DEFAULT '[]',
  annotations_json TEXT NOT NULL DEFAULT '[]',
  analyzer_report_json TEXT NOT NULL DEFAULT '{}',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES playground_sessions(id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (tree_id) REFERENCES trees(id)
);

CREATE INDEX IF NOT EXISTS idx_playground_sessions_user_tree_updated
  ON playground_sessions(user_id, tree_id, updated_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_playground_sessions_latest_review
  ON playground_sessions(latest_review_id);
CREATE INDEX IF NOT EXISTS idx_playground_reviews_session_created
  ON playground_reviews(session_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_playground_reviews_user_tree
  ON playground_reviews(user_id, tree_id, created_at DESC, id DESC);
