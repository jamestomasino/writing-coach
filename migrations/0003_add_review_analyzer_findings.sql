ALTER TABLE reviews
ADD COLUMN analyzer_findings_json TEXT NOT NULL DEFAULT '[]';
