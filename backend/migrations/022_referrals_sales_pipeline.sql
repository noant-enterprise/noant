-- Referral tracking for field sales
CREATE TABLE IF NOT EXISTS referrals (
    id VARCHAR(36) PRIMARY KEY,
    referrer_user_id VARCHAR(36) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    clicks INT DEFAULT 0,
    signups INT DEFAULT 0,
    conversions INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_referrer (referrer_user_id)
);

CREATE TABLE IF NOT EXISTS referral_events (
    id VARCHAR(36) PRIMARY KEY,
    referral_id VARCHAR(36) NOT NULL,
    event_type ENUM('click', 'signup', 'conversion') NOT NULL,
    visitor_ip VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_referral_id (referral_id),
    FOREIGN KEY (referral_id) REFERENCES referrals(id)
);

-- Sales pipeline for tracking field meetings
CREATE TABLE IF NOT EXISTS sales_leads (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(30),
    contact_email VARCHAR(255),
    business_name VARCHAR(255),
    business_type VARCHAR(100),
    status ENUM('contacted', 'demo_sent', 'signed_up', 'paying', 'lost') DEFAULT 'contacted',
    notes TEXT,
    meeting_location VARCHAR(255),
    referral_code VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
