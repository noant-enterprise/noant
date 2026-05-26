-- Migration to add source column to messages table
ALTER TABLE messages ADD COLUMN IF NOT EXISTS source VARCHAR(50) NULL;
