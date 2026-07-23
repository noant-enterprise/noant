-- Backfill org_id for training tables where org_id is empty but user_id exists.
-- This fixes data created before the OrgID scoping fix.

UPDATE categories SET org_id = user_id WHERE org_id = '' OR org_id IS NULL;
UPDATE qa_pairs SET org_id = user_id WHERE org_id = '' OR org_id IS NULL;
UPDATE unknown_questions SET org_id = user_id WHERE org_id = '' OR org_id IS NULL;
