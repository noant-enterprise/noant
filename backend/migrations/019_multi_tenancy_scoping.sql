-- Multi-tenancy scoping: add org_id to remaining tables

-- 1. Add org_id to tables that don't have it yet
ALTER TABLE user_credits ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE credit_purchases ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE subscriptions ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE whatsapp_templates ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE archive_folders ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE api_keys ADD COLUMN org_id VARCHAR(36) AFTER user_id;
ALTER TABLE campaign_recipients ADD COLUMN org_id VARCHAR(36) AFTER user_id;

-- 2. Create indexes
CREATE INDEX idx_uc_org ON user_credits(org_id);
CREATE INDEX idx_cp_org ON credit_purchases(org_id);
CREATE INDEX idx_sub_org ON subscriptions(org_id);
CREATE INDEX idx_wat_org ON whatsapp_templates(org_id);
CREATE INDEX idx_af_org ON archive_folders(org_id);
CREATE INDEX idx_ak_org ON api_keys(org_id);
CREATE INDEX idx_cr_org ON campaign_recipients(org_id);

-- 3. Backfill org_id from users.org_id
UPDATE user_credits uc JOIN users u ON uc.user_id = u.id SET uc.org_id = u.org_id WHERE uc.org_id IS NULL;
UPDATE credit_purchases cp JOIN users u ON cp.user_id = u.id SET cp.org_id = u.org_id WHERE cp.org_id IS NULL;
UPDATE subscriptions s JOIN users u ON s.user_id = u.id SET s.org_id = u.org_id WHERE s.org_id IS NULL;
UPDATE whatsapp_templates wt JOIN users u ON wt.user_id = u.id SET wt.org_id = u.org_id WHERE wt.org_id IS NULL;
UPDATE archive_folders af JOIN users u ON af.user_id = u.id SET af.org_id = u.org_id WHERE af.org_id IS NULL;
UPDATE api_keys ak JOIN users u ON ak.user_id = u.id SET ak.org_id = u.org_id WHERE ak.org_id IS NULL;
UPDATE campaign_recipients cr JOIN users u ON cr.user_id = u.id SET cr.org_id = u.org_id WHERE cr.org_id IS NULL;

-- 4. Make org_id NOT NULL after backfill
ALTER TABLE user_credits MODIFY COLUMN org_id VARCHAR(36) NOT NULL;
ALTER TABLE credit_purchases MODIFY COLUMN org_id VARCHAR(36) NOT NULL;
ALTER TABLE subscriptions MODIFY COLUMN org_id VARCHAR(36) NOT NULL;
ALTER TABLE whatsapp_templates MODIFY COLUMN org_id VARCHAR(36) NOT NULL;
ALTER TABLE archive_folders MODIFY COLUMN org_id VARCHAR(36) NOT NULL;
ALTER TABLE api_keys MODIFY COLUMN org_id VARCHAR(36) NOT NULL;
ALTER TABLE campaign_recipients MODIFY COLUMN org_id VARCHAR(36) NOT NULL;
