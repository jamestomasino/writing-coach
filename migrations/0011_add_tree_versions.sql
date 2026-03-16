CREATE TABLE IF NOT EXISTS tree_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tree_id INTEGER NOT NULL REFERENCES tgo_trees(id),
    version INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    seed_codes_json TEXT NOT NULL,
    priority_skills_json TEXT NOT NULL,
    tgos_json TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tree_id, version)
);
