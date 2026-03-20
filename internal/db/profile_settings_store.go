package db

import (
	"context"
	"database/sql"

	"github.com/tomasino/writing-coach/internal/domain"
)

func (s *Store) OnboardingProfileByEnrollmentID(ctx context.Context, enrollmentID int64) (domain.OnboardingProfile, error) {
	var profile domain.OnboardingProfile
	var weaknessesJSON, outcomesJSON string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT e.id, e.user_id, p.writing_type, p.assignment_format, p.target_audience, p.subject_matter, p.experience_level, p.desired_tone, p.biggest_weaknesses_json, p.desired_outcomes_json, p.difficulty_intensity, p.writing_goals, p.generated_tree_slug, p.template_key
		FROM enrollment_onboarding_profiles p
		JOIN user_tree_enrollments e ON e.id = p.enrollment_id
		WHERE p.enrollment_id = ?
	`, enrollmentID).Scan(
		&profile.EnrollmentID,
		&profile.UserID,
		&profile.WritingType,
		&profile.AssignmentFormat,
		&profile.TargetAudience,
		&profile.SubjectMatter,
		&profile.ExperienceLevel,
		&profile.DesiredTone,
		&weaknessesJSON,
		&outcomesJSON,
		&profile.DifficultyIntensity,
		&profile.WritingGoals,
		&profile.GeneratedTreeSlug,
		&profile.TemplateKey,
	)
	if err != nil {
		return domain.OnboardingProfile{}, err
	}
	if profile.BiggestWeaknesses, err = DecodeStringSlice(weaknessesJSON); err != nil {
		return domain.OnboardingProfile{}, err
	}
	if profile.DesiredOutcomes, err = DecodeStringSlice(outcomesJSON); err != nil {
		return domain.OnboardingProfile{}, err
	}
	return profile, nil
}

func (s *Store) OnboardingProfileByUserID(ctx context.Context, userID int64) (domain.OnboardingProfile, error) {
	enrollmentID, err := s.ActiveEnrollmentIDByUserID(ctx, userID)
	if err != nil {
		return domain.OnboardingProfile{}, err
	}
	return s.OnboardingProfileByEnrollmentID(ctx, enrollmentID)
}

func (s *Store) SaveOnboardingProfile(ctx context.Context, profile domain.OnboardingProfile) error {
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO enrollment_onboarding_profiles (
			enrollment_id, writing_type, assignment_format, target_audience, subject_matter, experience_level, desired_tone, biggest_weaknesses_json, desired_outcomes_json,
			difficulty_intensity, writing_goals, generated_tree_slug, template_key, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(enrollment_id) DO UPDATE SET
			writing_type = excluded.writing_type,
			assignment_format = excluded.assignment_format,
			target_audience = excluded.target_audience,
			subject_matter = excluded.subject_matter,
			experience_level = excluded.experience_level,
			desired_tone = excluded.desired_tone,
			biggest_weaknesses_json = excluded.biggest_weaknesses_json,
			desired_outcomes_json = excluded.desired_outcomes_json,
			difficulty_intensity = excluded.difficulty_intensity,
			writing_goals = excluded.writing_goals,
			generated_tree_slug = excluded.generated_tree_slug,
			template_key = excluded.template_key,
			updated_at = CURRENT_TIMESTAMP
	`, profile.EnrollmentID, profile.WritingType, profile.AssignmentFormat, profile.TargetAudience, profile.SubjectMatter, profile.ExperienceLevel, profile.DesiredTone, mustJSON(profile.BiggestWeaknesses), mustJSON(profile.DesiredOutcomes), profile.DifficultyIntensity, profile.WritingGoals, profile.GeneratedTreeSlug, profile.TemplateKey)
	return err
}

func (s *Store) AIProviderSettingsByUserID(ctx context.Context, userID int64) (domain.AIProviderSettings, error) {
	var settings domain.AIProviderSettings
	var enabled int
	var validatedAt sql.NullTime
	err := s.SQL.QueryRowContext(ctx, `
		SELECT user_id, provider, api_key_encrypted, api_key_last4, base_url_override, prompt_model_override, review_model_override, enabled, validated_at, last_validation_error, created_at, updated_at
		FROM user_ai_provider_settings
		WHERE user_id = ?
	`, userID).Scan(
		&settings.UserID,
		&settings.Provider,
		&settings.APIKeyEncrypted,
		&settings.APIKeyLast4,
		&settings.BaseURLOverride,
		&settings.PromptModelOverride,
		&settings.ReviewModelOverride,
		&enabled,
		&validatedAt,
		&settings.LastValidationError,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		return domain.AIProviderSettings{}, err
	}
	settings.Enabled = enabled != 0
	if validatedAt.Valid {
		settings.ValidatedAt = validatedAt.Time
	}
	return settings, nil
}

func (s *Store) SaveAIProviderSettings(ctx context.Context, settings domain.AIProviderSettings) error {
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO user_ai_provider_settings (
			user_id, provider, api_key_encrypted, api_key_last4, base_url_override, prompt_model_override, review_model_override, enabled, validated_at, last_validation_error, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
			provider = excluded.provider,
			api_key_encrypted = excluded.api_key_encrypted,
			api_key_last4 = excluded.api_key_last4,
			base_url_override = excluded.base_url_override,
			prompt_model_override = excluded.prompt_model_override,
			review_model_override = excluded.review_model_override,
			enabled = excluded.enabled,
			validated_at = excluded.validated_at,
			last_validation_error = excluded.last_validation_error,
			updated_at = CURRENT_TIMESTAMP
	`,
		settings.UserID,
		settings.Provider,
		settings.APIKeyEncrypted,
		settings.APIKeyLast4,
		settings.BaseURLOverride,
		settings.PromptModelOverride,
		settings.ReviewModelOverride,
		boolToInt(settings.Enabled),
		nullTime(settings.ValidatedAt),
		settings.LastValidationError,
	)
	return err
}

func (s *Store) DeleteAIProviderSettings(ctx context.Context, userID int64) error {
	_, err := s.SQL.ExecContext(ctx, `
		DELETE FROM user_ai_provider_settings
		WHERE user_id = ?
	`, userID)
	return err
}
