ALTER TABLE messages ADD COLUMN sequence INT DEFAULT NULL;

UPDATE messages m
JOIN (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY conversation_id ORDER BY created_at ASC, id ASC) AS seq
    FROM messages
) AS sub ON m.id = sub.id
SET m.sequence = sub.seq;

ALTER TABLE messages MODIFY COLUMN sequence INT NOT NULL;

CREATE INDEX idx_messages_conv_seq ON messages (conversation_id, sequence);
