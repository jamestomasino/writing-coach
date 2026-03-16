ALTER TABLE users
ADD COLUMN active_tree_slug TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS user_onboarding_profiles (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    writing_type TEXT NOT NULL,
    experience_level TEXT NOT NULL,
    desired_tone TEXT NOT NULL,
    biggest_weaknesses_json TEXT NOT NULL DEFAULT '[]',
    desired_outcomes_json TEXT NOT NULL DEFAULT '[]',
    difficulty_intensity TEXT NOT NULL,
    writing_goals TEXT NOT NULL,
    generated_tree_slug TEXT NOT NULL,
    template_key TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
