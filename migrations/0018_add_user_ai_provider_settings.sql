CREATE TABLE IF NOT EXISTS user_ai_provider_settings (
  user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  provider TEXT NOT NULL,
  api_key_encrypted TEXT NOT NULL DEFAULT '',
  api_key_last4 TEXT NOT NULL DEFAULT '',
  base_url_override TEXT NOT NULL DEFAULT '',
  prompt_model_override TEXT NOT NULL DEFAULT '',
  review_model_override TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 1,
  validated_at DATETIME,
  last_validation_error TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_ai_provider_settings_provider
ON user_ai_provider_settings(provider);
