ALTER TABLE calibration_runs
ADD COLUMN data_adequate INTEGER NOT NULL DEFAULT 0;

ALTER TABLE calibration_runs
ADD COLUMN approval_status TEXT NOT NULL DEFAULT 'pending';

ALTER TABLE calibration_runs
ADD COLUMN approved_by_user_id INTEGER;

ALTER TABLE calibration_runs
ADD COLUMN approved_at DATETIME;

ALTER TABLE calibration_runs
ADD COLUMN approval_notes TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_calibration_runs_approval_status_created
ON calibration_runs(approval_status, created_at DESC, id DESC);
