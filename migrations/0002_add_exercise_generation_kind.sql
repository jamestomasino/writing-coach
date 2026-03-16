ALTER TABLE exercises
ADD COLUMN generation_kind TEXT NOT NULL DEFAULT 'deterministic';
