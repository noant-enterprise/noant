-- Schema repair migration
-- Consolidates inline DDL that was previously in main.go

-- Ensure inventory_items table exists (idempotent)
CREATE TABLE IF NOT EXISTS inventory_items (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type ENUM('product','service','package') DEFAULT 'product',
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(15,2) NOT NULL,
    min_price DECIMAL(15,2),
    stock_quantity INT,
    image_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_active (user_id, is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Ensure handoffs table exists (idempotent)
CREATE TABLE IF NOT EXISTS handoffs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    conversation_id VARCHAR(36) NOT NULL,
    customer_name VARCHAR(100),
    customer_phone VARCHAR(50),
    customer_whatsapp VARCHAR(50),
    customer_location TEXT,
    product_name VARCHAR(255),
    original_price DECIMAL(15,2),
    agreed_price DECIMAL(15,2),
    quantity INT DEFAULT 1,
    status ENUM('pending','sold','lost','expired') DEFAULT 'pending',
    final_price DECIMAL(15,2),
    owner_notes TEXT,
    owner_notified_at TIMESTAMP,
    reminder_count INT DEFAULT 0,
    next_reminder_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_status (user_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Ensure owner_whatsapp column exists on users
SET @sql = (SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = DATABASE()
     AND TABLE_NAME = 'users'
     AND COLUMN_NAME = 'owner_whatsapp') > 0,
    'SELECT 1',
    'ALTER TABLE users ADD COLUMN owner_whatsapp VARCHAR(50)'
));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Ensure audit_logs table exists (idempotent)
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(36),
    details JSON,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_audit_user (user_id),
    INDEX idx_audit_created (created_at),
    INDEX idx_audit_action (action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
