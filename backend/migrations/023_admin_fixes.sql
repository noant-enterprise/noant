-- Add reason column to payments for failed payment tracking
ALTER TABLE payments ADD COLUMN IF NOT EXISTS reason TEXT NULL AFTER status;

-- Add onboarding_step to users for funnel tracking
ALTER TABLE users ADD COLUMN IF NOT EXISTS onboarding_step VARCHAR(50) DEFAULT 'started' AFTER plan_id;

-- Add user_sessions table for bounce rate / session tracking
CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP NULL,
    duration_seconds INT DEFAULT 0,
    page_views INT DEFAULT 1,
    source VARCHAR(100),
    utm_source VARCHAR(100),
    utm_medium VARCHAR(100),
    utm_campaign VARCHAR(100),
    INDEX idx_user (user_id),
    INDEX idx_started (started_at)
);