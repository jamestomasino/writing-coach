CREATE TABLE IF NOT EXISTS calibration_runs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_kind TEXT NOT NULL,
  status TEXT NOT NULL,
  triggered_by_user_id INTEGER,
  min_samples INTEGER NOT NULL,
  limit_per_track INTEGER NOT NULL,
  submission_count INTEGER NOT NULL DEFAULT 0,
  deterministic_score_count INTEGER NOT NULL DEFAULT 0,
  track_learnings_json TEXT NOT NULL DEFAULT '[]',
  domain_learnings_json TEXT NOT NULL DEFAULT '[]',
  highlights_json TEXT NOT NULL DEFAULT '[]',
  recommendations_json TEXT NOT NULL DEFAULT '[]',
  error_text TEXT NOT NULL DEFAULT '',
  started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  completed_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (triggered_by_user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_calibration_runs_status_created
ON calibration_runs(status, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS admin_notifications (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  payload_json TEXT NOT NULL DEFAULT '{}',
  related_run_id INTEGER,
  is_read INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  read_at DATETIME,
  FOREIGN KEY (related_run_id) REFERENCES calibration_runs(id)
);

CREATE INDEX IF NOT EXISTS idx_admin_notifications_read_created
ON admin_notifications(is_read, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_admin_notifications_kind_run
ON admin_notifications(kind, related_run_id);
