-- Impersonation tokens for admin "impersonate user" feature
CREATE TABLE IF NOT EXISTS impersonation_tokens (
    id VARCHAR(36) PRIMARY KEY,
    admin_user_id VARCHAR(36) NOT NULL,
    target_user_id VARCHAR(36) NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMP NULL,
    INDEX idx_token (token),
    INDEX idx_admin (admin_user_id),
    INDEX idx_target (target_user_id),
    FOREIGN KEY (admin_user_id) REFERENCES users(id),
    FOREIGN KEY (target_user_id) REFERENCES users(id)
);