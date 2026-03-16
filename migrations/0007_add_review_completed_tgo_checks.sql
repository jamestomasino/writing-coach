ALTER TABLE reviews
ADD COLUMN completed_tgo_checks_json TEXT NOT NULL DEFAULT '[]';
