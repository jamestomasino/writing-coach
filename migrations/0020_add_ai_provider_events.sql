CREATE TABLE IF NOT EXISTS ai_provider_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider TEXT NOT NULL DEFAULT '',
  event TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT '',
  status_code INTEGER NOT NULL DEFAULT 0,
  detail_json TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ai_provider_events_created_at
  ON ai_provider_events(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ai_provider_events_user_id_created_at
  ON ai_provider_events(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ai_provider_events_event_created_at
  ON ai_provider_events(event, created_at DESC);
