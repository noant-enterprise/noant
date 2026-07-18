-- Create whatsapp_templates table for HSM message templates
CREATE TABLE IF NOT EXISTS whatsapp_templates (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    category VARCHAR(20) NOT NULL DEFAULT 'utility',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    header_type VARCHAR(20) NOT NULL DEFAULT 'none',
    header_value TEXT,
    body_text TEXT NOT NULL,
    footer_text TEXT,
    buttons JSON,
    namespace VARCHAR(100),
    rejection_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_name (name)
);

-- Create campaign_recipients table for broadcast recipients
CREATE TABLE IF NOT EXISTS campaign_recipients (
    id VARCHAR(36) PRIMARY KEY,
    campaign_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error TEXT,
    sent_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    read_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (campaign_id) REFERENCES campaign_schedules(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_campaign_id (campaign_id),
    INDEX idx_user_id (user_id),
    INDEX idx_phone (phone),
    INDEX idx_status (status)
);

-- Create media_messages table for tracking uploaded media
CREATE TABLE IF NOT EXISTS media_messages (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    conversation_id VARCHAR(36) NOT NULL,
    message_id VARCHAR(36),
    session_id VARCHAR(100),
    media_type VARCHAR(20) NOT NULL,
    mime_type VARCHAR(100),
    file_size BIGINT NOT NULL DEFAULT 0,
    file_name VARCHAR(255),
    file_path VARCHAR(500),
    thumb_path VARCHAR(500),
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    duration INT NOT NULL DEFAULT 0,
    caption TEXT,
    remote_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_conversation_id (conversation_id),
    INDEX idx_media_type (media_type),
    INDEX idx_expires_at (expires_at)
);

-- Add customer_avatar column to conversations if not present (moved from auto-alter in main.go)
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS customer_avatar VARCHAR(500) AFTER customer_email;
