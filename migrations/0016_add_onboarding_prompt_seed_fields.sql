ALTER TABLE user_onboarding_profiles
ADD COLUMN assignment_format TEXT NOT NULL DEFAULT '';

ALTER TABLE user_onboarding_profiles
ADD COLUMN target_audience TEXT NOT NULL DEFAULT '';

ALTER TABLE user_onboarding_profiles
ADD COLUMN subject_matter TEXT NOT NULL DEFAULT '';
