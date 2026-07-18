ALTER TABLE users ADD COLUMN onboarding_status VARCHAR(20) DEFAULT 'pending';
ALTER TABLE users ADD COLUMN industry VARCHAR(100) DEFAULT NULL;

UPDATE users SET onboarding_status = 'complete' WHERE onboarding_status IS NULL AND is_active = true;
