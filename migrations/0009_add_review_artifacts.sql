CREATE TABLE IF NOT EXISTS review_artifacts (
    review_id INTEGER PRIMARY KEY REFERENCES reviews(id),
    analyzer_report_json TEXT NOT NULL DEFAULT '{}',
    recommendation_json TEXT NOT NULL DEFAULT '{}',
    comparison_json TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
