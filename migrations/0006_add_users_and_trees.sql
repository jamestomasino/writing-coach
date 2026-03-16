CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tgo_trees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tree_tgos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tree_id INTEGER NOT NULL REFERENCES tgo_trees(id),
    tgo_code TEXT NOT NULL REFERENCES tgo_catalog(code),
    UNIQUE(tree_id, tgo_code)
);

CREATE TABLE IF NOT EXISTS user_tree_enrollments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    tree_id INTEGER NOT NULL REFERENCES tgo_trees(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, tree_id)
);

CREATE TABLE IF NOT EXISTS user_curriculum_state (
    enrollment_id INTEGER PRIMARY KEY REFERENCES user_tree_enrollments(id),
    current_focus TEXT NOT NULL,
    difficulty_level INTEGER NOT NULL,
    last_review_id INTEGER REFERENCES reviews(id),
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS enrollment_active_tgos (
    enrollment_id INTEGER NOT NULL REFERENCES user_tree_enrollments(id),
    slot INTEGER NOT NULL,
    tgo_code TEXT NOT NULL REFERENCES tgo_catalog(code),
    activated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (enrollment_id, slot)
);

CREATE TABLE IF NOT EXISTS enrollment_completed_tgos (
    enrollment_id INTEGER NOT NULL REFERENCES user_tree_enrollments(id),
    tgo_code TEXT NOT NULL REFERENCES tgo_catalog(code),
    completed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (enrollment_id, tgo_code)
);

ALTER TABLE exercises
ADD COLUMN user_id INTEGER REFERENCES users(id);

ALTER TABLE exercises
ADD COLUMN tree_id INTEGER REFERENCES tgo_trees(id);

ALTER TABLE submissions
ADD COLUMN user_id INTEGER REFERENCES users(id);

ALTER TABLE submissions
ADD COLUMN tree_id INTEGER REFERENCES tgo_trees(id);

ALTER TABLE reviews
ADD COLUMN user_id INTEGER REFERENCES users(id);

ALTER TABLE reviews
ADD COLUMN tree_id INTEGER REFERENCES tgo_trees(id);
