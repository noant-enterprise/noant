-- Migration 020: Delivery tracking + queue reliability improvements

-- Add delivery_status column to messages table
ALTER TABLE messages ADD COLUMN IF NOT EXISTS delivery_status VARCHAR(20) DEFAULT 'sent';
ALTER TABLE messages ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);

-- Index for fast lookups by external_id (delivery status updates)
CREATE INDEX IF NOT EXISTS idx_messages_external_id ON messages(external_id);
CREATE INDEX IF NOT EXISTS idx_messages_delivery_status ON messages(delivery_status);
