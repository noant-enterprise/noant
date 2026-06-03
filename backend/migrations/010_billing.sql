-- Create user_credits table for Pulse response balance tracking
CREATE TABLE IF NOT EXISTS user_credits (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    balance INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMP NULL,
    last_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
);

-- Create credit_purchases table for purchase history
CREATE TABLE IF NOT EXISTS credit_purchases (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    checkout_id VARCHAR(100) NOT NULL UNIQUE,
    pack_type VARCHAR(20) NOT NULL,
    amount INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    purchased_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_checkout_id (checkout_id),
    INDEX idx_status (status),
    INDEX idx_purchased_at (purchased_at)
);

-- Create campaign_schedules table for Campaign Mode
CREATE TABLE IF NOT EXISTS campaign_schedules (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_start_date (start_date),
    INDEX idx_end_date (end_date)
);

-- Alter users table to add billing columns
ALTER TABLE users MODIFY COLUMN plan_id VARCHAR(50) NOT NULL DEFAULT 'free';
ALTER TABLE users ADD COLUMN credit_balance INT NOT NULL DEFAULT 0 AFTER plan_id;
ALTER TABLE users ADD COLUMN last_credit_purchase_at TIMESTAMP NULL AFTER credit_balance;

-- Add index for efficient querying of expiring credits
CREATE INDEX idx_user_credits_expiring ON user_credits(expires_at);

-- Add index for efficient querying of recent purchases for nudge logic
CREATE INDEX idx_credit_purchases_user_month ON credit_purchases(user_id, purchased_at);