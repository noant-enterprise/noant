-- Noant Database Schema
-- TiDB Cloud Compatible

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role ENUM('owner', 'admin', 'agent') DEFAULT 'owner',
    company_name VARCHAR(255),
    phone VARCHAR(20),
    avatar VARCHAR(500),
    plan_id VARCHAR(50) DEFAULT 'free',
    is_active BOOLEAN DEFAULT true,
    must_change_password BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_plan (plan_id)
);

CREATE TABLE IF NOT EXISTS conversations (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    customer_name VARCHAR(100),
    customer_phone VARCHAR(20),
    customer_email VARCHAR(255),
    channel ENUM('telegram', 'whatsapp', 'web', 'instagram', 'messenger', 'email') NOT NULL,
    status ENUM('active', 'resolved', 'escalated', 'archived') DEFAULT 'active',
    intent ENUM('buying', 'complaining', 'inquiry', 'support', 'other') DEFAULT 'inquiry',
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    is_ai_transferred BOOLEAN DEFAULT false,
    taken_over_by VARCHAR(36) NULL,
    taken_over_at TIMESTAMP NULL,
    resolved_at TIMESTAMP NULL,
    folder_id VARCHAR(36) NULL,
    location_lat DECIMAL(10,8) NULL,
    location_lng DECIMAL(11,8) NULL,
    location_city VARCHAR(100),
    location_country VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_status (user_id, status),
    INDEX idx_channel (channel),
    INDEX idx_created (created_at)
);

CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(36) PRIMARY KEY,
    conversation_id VARCHAR(36) NOT NULL,
    sender_type ENUM('ai', 'human', 'customer', 'system') NOT NULL,
    sender_id VARCHAR(36) NULL,
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    confidence DECIMAL(5,4) NULL,
    matched_qa_id VARCHAR(36) NULL,
    escalation_reason VARCHAR(255),
    language VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_conversation (conversation_id, created_at),
    INDEX idx_sender (sender_id)
);

CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#3b82f6',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_name (name)
);

CREATE TABLE IF NOT EXISTS qa_pairs (
    id VARCHAR(36) PRIMARY KEY,
    category_id VARCHAR(36) NOT NULL,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    variations JSON,
    embedding JSON,
    is_active BOOLEAN DEFAULT true,
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category_id),
    FULLTEXT INDEX idx_question (question)
);

CREATE TABLE IF NOT EXISTS unknown_questions (
    id VARCHAR(36) PRIMARY KEY,
    question TEXT NOT NULL,
    conversation_id VARCHAR(36),
    channel VARCHAR(50),
    status ENUM('pending', 'trained', 'ignored') DEFAULT 'pending',
    suggested_answer TEXT,
    category_id VARCHAR(36) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_created (created_at)
);

CREATE TABLE IF NOT EXISTS integrations (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    channel VARCHAR(50) NOT NULL,
    status ENUM('active', 'inactive', 'error') DEFAULT 'inactive',
    config JSON,
    webhook_url VARCHAR(500),
    last_error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_channel (channel)
);

CREATE TABLE IF NOT EXISTS team_members (
    id VARCHAR(36) PRIMARY KEY,
    owner_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    role ENUM('admin', 'agent') DEFAULT 'agent',
    is_active BOOLEAN DEFAULT true,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_owner (owner_id),
    INDEX idx_user (user_id)
);

CREATE TABLE IF NOT EXISTS api_keys (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    last_used TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_key (key_hash)
);

CREATE TABLE IF NOT EXISTS archive_folders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    type ENUM('chats', 'contacts', 'locations') DEFAULT 'chats',
    color VARCHAR(7) DEFAULT '#6b7280',
    item_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_type (user_id, type)
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    plan_id VARCHAR(50) NOT NULL,
    status ENUM('active', 'cancelled', 'expired') DEFAULT 'active',
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_status (user_id, status),
    INDEX idx_period (current_period_end)
);

CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    subscription_id VARCHAR(36),
    amount INT NOT NULL,
    currency VARCHAR(3) DEFAULT 'NGN',
    status ENUM('pending', 'completed', 'failed', 'refunded') DEFAULT 'pending',
    provider VARCHAR(50) DEFAULT 'polar',
    provider_payment_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_status (status)
);
