-- Migration 016: Push notification subscriptions for PWA
-- Creates tables and indexes for Web Push subscription storage

CREATE TABLE IF NOT EXISTS push_subscriptions (
    id          VARCHAR(36)  NOT NULL PRIMARY KEY,
    user_id     VARCHAR(36)  NOT NULL,
    endpoint    TEXT         NOT NULL,
    auth        VARCHAR(255) NOT NULL,
    p256dh      VARCHAR(255) NOT NULL,
    user_agent  VARCHAR(500) NULL,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_endpoint (endpoint(255)),
    INDEX idx_push_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
