-- Multi-tenancy: organizations table and org_id columns

-- 1. Create organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    owner_id VARCHAR(36) NOT NULL,
    plan_id VARCHAR(50) DEFAULT 'free',
    settings JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_org_owner (owner_id),
    INDEX idx_org_slug (slug),
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 2. Add org_id to users
ALTER TABLE users ADD COLUMN org_id VARCHAR(36) AFTER id;
CREATE INDEX idx_user_org ON users(org_id);

-- 3. Backfill: create an organization for each existing user and set their org_id
SET @org_counter = 0;
INSERT INTO organizations (id, name, slug, owner_id, plan_id, created_at, updated_at)
SELECT UUID(), COALESCE(u.company_name, CONCAT(u.first_name, "'s Workspace")),
       CONCAT('org-', LOWER(REPLACE(UUID(), '-', ''))), u.id, u.plan_id, NOW(), NOW()
FROM users u;

UPDATE users u
JOIN organizations o ON o.owner_id = u.id
SET u.org_id = o.id;

-- 4. Add org_id to key data tables
ALTER TABLE conversations ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE categories ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE qa_pairs ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE unknown_questions ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE integrations ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE inventory_items ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE handoffs ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE audit_logs ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE campaign_schedules ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE widget_configs ADD COLUMN org_id VARCHAR(36) AFTER user_id;

CREATE INDEX idx_conv_org ON conversations(org_id);
CREATE INDEX idx_cat_org ON categories(org_id);
CREATE INDEX idx_qa_org ON qa_pairs(org_id);
CREATE INDEX idx_uq_org ON unknown_questions(org_id);
CREATE INDEX idx_integ_org ON integrations(org_id);
CREATE INDEX idx_inv_org ON inventory_items(org_id);
CREATE INDEX idx_handoff_org ON handoffs(org_id);
CREATE INDEX idx_audit_org ON audit_logs(org_id);
CREATE INDEX idx_campaign_org ON campaign_schedules(org_id);
CREATE INDEX idx_widget_org ON widget_configs(org_id);

-- 5. Backfill org_id on data tables from users.org_id
UPDATE conversations c JOIN users u ON c.user_id = u.id SET c.org_id = u.org_id WHERE c.org_id IS NULL;
UPDATE categories c JOIN users u ON c.user_id = u.id SET c.org_id = u.org_id WHERE c.org_id IS NULL;
UPDATE qa_pairs q JOIN users u ON q.user_id = u.id SET q.org_id = u.org_id WHERE q.org_id IS NULL;
UPDATE unknown_questions uq JOIN users u ON uq.user_id = u.id SET uq.org_id = u.org_id WHERE uq.org_id IS NULL;
UPDATE integrations i JOIN users u ON i.user_id = u.id SET i.org_id = u.org_id WHERE i.org_id IS NULL;
UPDATE inventory_items inv JOIN users u ON inv.user_id = u.id SET inv.org_id = u.org_id WHERE inv.org_id IS NULL;
UPDATE handoffs h JOIN users u ON h.user_id = u.id SET h.org_id = u.org_id WHERE h.org_id IS NULL;
UPDATE audit_logs a JOIN users u ON a.user_id = u.id SET a.org_id = u.org_id WHERE a.org_id IS NULL;
UPDATE campaign_schedules cs JOIN users u ON cs.user_id = u.id SET cs.org_id = u.org_id WHERE cs.org_id IS NULL;
UPDATE widget_configs wc JOIN users u ON wc.user_id = u.id SET wc.org_id = u.org_id WHERE wc.org_id IS NULL;

-- 6. Update team_members: replace owner_id with org_id
ALTER TABLE team_members ADD COLUMN org_id VARCHAR(36) AFTER id;
CREATE INDEX idx_team_org ON team_members(org_id);

UPDATE team_members tm
JOIN users u ON tm.owner_id = u.id
SET tm.org_id = u.org_id;

-- Add unique constraint: one membership per user per org
ALTER TABLE team_members ADD UNIQUE KEY uk_org_user (org_id, user_id);
