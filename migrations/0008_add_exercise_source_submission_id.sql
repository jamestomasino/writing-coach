ALTER TABLE exercises
ADD COLUMN source_submission_id INTEGER REFERENCES submissions(id);
