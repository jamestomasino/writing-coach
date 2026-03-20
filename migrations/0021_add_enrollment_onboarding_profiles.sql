CREATE TABLE IF NOT EXISTS enrollment_onboarding_profiles (
    enrollment_id INTEGER PRIMARY KEY REFERENCES user_tree_enrollments(id),
    writing_type TEXT NOT NULL,
    assignment_format TEXT NOT NULL DEFAULT '',
    target_audience TEXT NOT NULL DEFAULT '',
    subject_matter TEXT NOT NULL DEFAULT '',
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

INSERT INTO enrollment_onboarding_profiles (
    enrollment_id,
    writing_type,
    assignment_format,
    target_audience,
    subject_matter,
    experience_level,
    desired_tone,
    biggest_weaknesses_json,
    desired_outcomes_json,
    difficulty_intensity,
    writing_goals,
    generated_tree_slug,
    template_key,
    created_at,
    updated_at
)
SELECT
    e.id,
    p.writing_type,
    p.assignment_format,
    p.target_audience,
    p.subject_matter,
    p.experience_level,
    p.desired_tone,
    p.biggest_weaknesses_json,
    p.desired_outcomes_json,
    p.difficulty_intensity,
    p.writing_goals,
    p.generated_tree_slug,
    p.template_key,
    p.created_at,
    p.updated_at
FROM user_onboarding_profiles p
JOIN users u ON u.id = p.user_id
JOIN tgo_trees t ON t.slug = CASE
    WHEN TRIM(u.active_tree_slug) <> '' THEN u.active_tree_slug
    ELSE p.generated_tree_slug
END
JOIN user_tree_enrollments e ON e.user_id = p.user_id AND e.tree_id = t.id
WHERE NOT EXISTS (
    SELECT 1
    FROM enrollment_onboarding_profiles ep
    WHERE ep.enrollment_id = e.id
);
