ALTER TABLE submissions
ADD COLUMN parent_submission_id INTEGER REFERENCES submissions(id);

ALTER TABLE submissions
ADD COLUMN draft_number INTEGER NOT NULL DEFAULT 1;
