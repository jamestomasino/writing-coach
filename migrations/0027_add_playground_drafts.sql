CREATE TABLE IF NOT EXISTS playground_drafts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  tree_id INTEGER NOT NULL,
  parent_draft_id INTEGER,
  content TEXT NOT NULL DEFAULT '',
  word_count INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES playground_sessions(id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (tree_id) REFERENCES trees(id),
  FOREIGN KEY (parent_draft_id) REFERENCES playground_drafts(id)
);

ALTER TABLE playground_sessions ADD COLUMN latest_draft_id INTEGER REFERENCES playground_drafts(id);
ALTER TABLE playground_sessions ADD COLUMN draft_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE playground_reviews ADD COLUMN draft_id INTEGER REFERENCES playground_drafts(id);
ALTER TABLE playground_reviews ADD COLUMN comparison_json TEXT NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_playground_drafts_session_created
  ON playground_drafts(session_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_playground_drafts_user_tree
  ON playground_drafts(user_id, tree_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_playground_reviews_draft
  ON playground_reviews(draft_id, created_at DESC, id DESC);
