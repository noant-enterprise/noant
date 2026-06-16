-- Email Verification Support for User Registration

ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_code VARCHAR(6) DEFAULT NULL;

UPDATE users SET is_verified = TRUE;
