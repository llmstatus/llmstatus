-- Add sponsor review workflow: status column and is_admin flag.

ALTER TABLE sponsors ADD COLUMN status TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'approved', 'rejected'));

-- Existing active sponsors are already approved.
UPDATE sponsors SET status = 'approved' WHERE active = TRUE;
UPDATE sponsors SET status = 'rejected' WHERE active = FALSE;

-- New sponsors start inactive until an admin approves.
ALTER TABLE sponsors ALTER COLUMN active SET DEFAULT FALSE;

ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT FALSE;
