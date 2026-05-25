-- 006: Notifications, Widget Config, Notification Preferences

CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    link VARCHAR(500),
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_notif_user (user_id),
    INDEX idx_notif_read (user_id, is_read),
    INDEX idx_notif_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS widget_configs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL UNIQUE,
    brand_color VARCHAR(20) DEFAULT '#3b82f6',
    greeting VARCHAR(500) DEFAULT 'Hello! How can I help you today?',
    bot_name VARCHAR(100) DEFAULT 'Noant AI',
    position VARCHAR(20) DEFAULT 'bottom-right',
    widget_api_key VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_widget_user (user_id),
    INDEX idx_widget_key (widget_api_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS notif_escalation BOOLEAN DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS notif_unknown_questions BOOLEAN DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS notif_payment BOOLEAN DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS notif_security BOOLEAN DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS notif_team_invite BOOLEAN DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS language_preference VARCHAR(10) DEFAULT 'en';
