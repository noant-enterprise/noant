CREATE TABLE IF NOT EXISTS csat_ratings (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    conversation_id VARCHAR(36) NOT NULL,
    score TINYINT NOT NULL CHECK (score >= 1 AND score <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_csat_user (user_id),
    INDEX idx_csat_conv (conversation_id),
    INDEX idx_csat_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE unknown_questions ADD INDEX IF NOT EXISTS idx_uq_user_status (user_id, status);
