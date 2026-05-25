-- Step 1: Add user_id columns
ALTER TABLE categories ADD COLUMN IF NOT EXISTS user_id VARCHAR(36) NULL;
ALTER TABLE unknown_questions ADD COLUMN IF NOT EXISTS user_id VARCHAR(36) NULL;
ALTER TABLE qa_pairs ADD COLUMN IF NOT EXISTS user_id VARCHAR(36) NULL;

-- Step 2: Create indexes
CREATE INDEX IF NOT EXISTS idx_categories_user ON categories(user_id);
CREATE INDEX IF NOT EXISTS idx_unknown_user ON unknown_questions(user_id);
CREATE INDEX IF NOT EXISTS idx_qa_user ON qa_pairs(user_id);

-- Step 3: Backfill unknown_questions from conversations
UPDATE unknown_questions uq
JOIN conversations c ON uq.conversation_id = c.id
SET uq.user_id = c.user_id
WHERE uq.user_id IS NULL;

-- Step 4: Backfill qa_pairs from categories
UPDATE qa_pairs qp
JOIN categories cat ON qp.category_id = cat.id
SET qp.user_id = cat.user_id
WHERE qp.user_id IS NULL AND cat.user_id IS NOT NULL;