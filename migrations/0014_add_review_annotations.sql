ALTER TABLE review_artifacts
ADD COLUMN annotations_json TEXT NOT NULL DEFAULT '[]';
