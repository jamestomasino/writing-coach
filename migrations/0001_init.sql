CREATE TABLE IF NOT EXISTS writer_profile (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    aesthetic_target TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS exercises (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    brief TEXT NOT NULL,
    constraints_json TEXT NOT NULL,
    focus_skills_json TEXT NOT NULL,
    tgo_codes_json TEXT NOT NULL DEFAULT '[]',
    success_criteria_json TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id),
    parent_submission_id INTEGER REFERENCES submissions(id),
    draft_number INTEGER NOT NULL DEFAULT 1,
    content TEXT NOT NULL,
    word_count INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submission_id INTEGER NOT NULL REFERENCES submissions(id),
    review_kind TEXT NOT NULL,
    summary TEXT NOT NULL,
    strengths_json TEXT NOT NULL,
    weaknesses_json TEXT NOT NULL,
    analyzer_findings_json TEXT NOT NULL DEFAULT '[]',
    next_focus TEXT NOT NULL,
    metric_word_count INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS skill_dimensions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS submission_skill_scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    submission_id INTEGER NOT NULL REFERENCES submissions(id),
    skill_name TEXT NOT NULL,
    score INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS curriculum_state (
    id INTEGER PRIMARY KEY,
    current_focus TEXT NOT NULL,
    difficulty_level INTEGER NOT NULL,
    last_review_id INTEGER REFERENCES reviews(id),
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tgo_catalog (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    stage TEXT NOT NULL,
    stage_order INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS active_tgos (
    slot INTEGER PRIMARY KEY,
    tgo_code TEXT NOT NULL REFERENCES tgo_catalog(code),
    activated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS completed_tgos (
    tgo_code TEXT PRIMARY KEY REFERENCES tgo_catalog(code),
    completed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS review_tgo_assessments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    review_id INTEGER NOT NULL REFERENCES reviews(id),
    submission_id INTEGER NOT NULL REFERENCES submissions(id),
    tgo_code TEXT NOT NULL REFERENCES tgo_catalog(code),
    status TEXT NOT NULL,
    evidence TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
