ALTER TABLE user_curriculum_state
ADD COLUMN progression_hold_active INTEGER NOT NULL DEFAULT 0;

ALTER TABLE user_curriculum_state
ADD COLUMN progression_hold_reason_code TEXT NOT NULL DEFAULT '';

ALTER TABLE user_curriculum_state
ADD COLUMN hold_trigger_review_id INTEGER REFERENCES reviews(id);

ALTER TABLE user_curriculum_state
ADD COLUMN hold_cleared_review_id INTEGER REFERENCES reviews(id);

ALTER TABLE user_curriculum_state
ADD COLUMN hold_updated_at DATETIME;

CREATE INDEX IF NOT EXISTS idx_user_curriculum_state_hold_active
ON user_curriculum_state(progression_hold_active);
