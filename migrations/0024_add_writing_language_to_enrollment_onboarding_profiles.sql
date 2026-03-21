ALTER TABLE enrollment_onboarding_profiles
ADD COLUMN writing_language TEXT NOT NULL DEFAULT 'en';

UPDATE enrollment_onboarding_profiles
SET writing_language = 'en'
WHERE TRIM(COALESCE(writing_language, '')) = '';
