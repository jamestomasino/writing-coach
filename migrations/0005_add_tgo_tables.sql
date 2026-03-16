ALTER TABLE exercises
ADD COLUMN tgo_codes_json TEXT NOT NULL DEFAULT '[]';

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
