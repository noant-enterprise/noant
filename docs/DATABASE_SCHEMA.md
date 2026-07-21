# NOANT Database Schema Reference

> TiDB Cloud compatible. All primary keys are `VARCHAR(36)` UUIDs. Timestamps use `TIMESTAMP` with `CURRENT_TIMESTAMP` defaults. JSON columns store flexible metadata. Engine: InnoDB, charset: `utf8mb4`.

---

## Table of Contents

- [users](#users)
- [conversations](#conversations)
- [messages](#messages)
- [categories](#categories)
- [qa_pairs](#qa_pairs)
- [unknown_questions](#unknown_questions)
- [integrations](#integrations)
- [team_members](#team_members)
- [api_keys](#api_keys)
- [archive_folders](#archive_folders)
- [subscriptions](#subscriptions)
- [payments](#payments)
- [audit_logs](#audit_logs)
- [notifications](#notifications)
- [widget_configs](#widget_configs)
- [inventory_items](#inventory_items)
- [handoffs](#handoffs)
- [user_credits](#user_credits)
- [credit_purchases](#credit_purchases)
- [campaign_schedules](#campaign_schedules)
- [campaign_recipients](#campaign_recipients)
- [whatsapp_templates](#whatsapp_templates)
- [media_messages](#media_messages)
- [csat_ratings](#csat_ratings)
- [push_subscriptions](#push_subscriptions)
- [Entity Relationships](#entity-relationships)
- [Migration Version History](#migration-version-history)

---

## users

Core user accounts. Each user owns a workspace (conversations, QA pairs, integrations, etc.).

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| email | VARCHAR(255) | NO | — | Unique login email |
| password_hash | VARCHAR(255) | NO | — | Bcrypt hashed password |
| first_name | VARCHAR(100) | NO | — | User's first name |
| last_name | VARCHAR(100) | NO | — | User's last name |
| role | ENUM('owner','admin','agent') | YES | 'owner' | Workspace role |
| company_name | VARCHAR(255) | YES | NULL | Company or org name |
| phone | VARCHAR(20) | YES | NULL | Contact phone |
| avatar | VARCHAR(500) | YES | NULL | Avatar image URL |
| plan_id | VARCHAR(50) | NO | 'free' | Current subscription plan |
| credit_balance | INT | NO | 0 | Pulse AI response credits |
| last_credit_purchase_at | TIMESTAMP | YES | NULL | Last credit purchase time |
| is_active | BOOLEAN | YES | true | Account enabled flag |
| must_change_password | BOOLEAN | YES | true | Force password change on login |
| is_verified | BOOLEAN | YES | FALSE | Email verification status |
| verification_code | VARCHAR(6) | YES | NULL | 6-digit email verification code |
| onboarding_status | VARCHAR(20) | YES | 'pending' | Onboarding wizard status |
| industry | VARCHAR(100) | YES | NULL | User's industry vertical |
| owner_whatsapp | VARCHAR(50) | YES | NULL | Owner's WhatsApp number for handoff alerts |
| notif_escalation | BOOLEAN | YES | TRUE | Notification preference |
| notif_unknown_questions | BOOLEAN | YES | TRUE | Notification preference |
| notif_payment | BOOLEAN | YES | TRUE | Notification preference |
| notif_security | BOOLEAN | YES | TRUE | Notification preference |
| notif_team_invite | BOOLEAN | YES | TRUE | Notification preference |
| language_preference | VARCHAR(10) | YES | 'en' | UI language code |
| last_login_at | TIMESTAMP | YES | NULL | Last successful login timestamp |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Account creation time |
| updated_at | TIMESTAMP | YES | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_email (email)`, `idx_plan (plan_id)`

---

## conversations

Customer support conversations. Each conversation belongs to one user workspace and has a channel source.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owner workspace user |
| customer_name | VARCHAR(100) | YES | NULL | Customer display name |
| customer_phone | VARCHAR(20) | YES | NULL | Customer phone number |
| customer_email | VARCHAR(255) | YES | NULL | Customer email address |
| customer_avatar | VARCHAR(500) | YES | NULL | Customer avatar URL |
| channel | ENUM('telegram','whatsapp','web','instagram','messenger','email') | NO | — | Source channel |
| status | ENUM('active','resolved','escalated','archived') | YES | 'active' | Conversation state |
| intent | ENUM('buying','complaining','inquiry','support','other') | YES | 'inquiry' | Detected customer intent |
| priority | ENUM('low','medium','high','urgent') | YES | 'medium' | Priority level |
| is_ai_transferred | BOOLEAN | YES | false | Handoff from AI to human occurred |
| taken_over_by | VARCHAR(36) | YES | NULL | Agent user ID who took over |
| taken_over_at | TIMESTAMP | YES | NULL | When human takeover started |
| resolved_at | TIMESTAMP | YES | NULL | When conversation was resolved |
| folder_id | VARCHAR(36) | YES | NULL | Archive folder assignment |
| location_lat | DECIMAL(10,8) | YES | NULL | Customer latitude |
| location_lng | DECIMAL(11,8) | YES | NULL | Customer longitude |
| location_city | VARCHAR(100) | YES | NULL | Customer city |
| location_country | VARCHAR(100) | YES | NULL | Customer country |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | YES | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_user_status (user_id, status)`, `idx_channel (channel)`, `idx_created (created_at)`

---

## messages

Individual messages within a conversation. Tracks sender, content, AI confidence, and sequence order.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| conversation_id | VARCHAR(36) | NO | — | Parent conversation |
| sender_type | ENUM('ai','human','customer','system') | NO | — | Who sent the message |
| sender_id | VARCHAR(36) | YES | NULL | User ID of sender (for human/agent) |
| content | TEXT | NO | — | Message body |
| is_read | BOOLEAN | YES | false | Read receipt status |
| confidence | DECIMAL(5,4) | YES | NULL | AI response confidence score (0-1) |
| matched_qa_id | VARCHAR(36) | YES | NULL | QA pair that generated AI response |
| escalation_reason | VARCHAR(255) | YES | NULL | Reason for escalation |
| language | VARCHAR(10) | YES | 'en' | Detected message language |
| source | VARCHAR(50) | YES | NULL | Message source (e.g. channel plugin) |
| sequence | INT | NO | — | Ordered position within conversation |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Message timestamp |

**Indexes:** `idx_conversation (conversation_id, created_at)`, `idx_sender (sender_id)`, `idx_messages_conv_seq (conversation_id, sequence)`

---

## categories

QA training categories. Each category groups related QA pairs for a user workspace.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| name | VARCHAR(100) | NO | — | Category name |
| description | TEXT | YES | NULL | Category description |
| color | VARCHAR(7) | YES | '#3b82f6' | Hex color for UI display |
| user_id | VARCHAR(36) | YES | NULL | Owning workspace user |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |

**Indexes:** `idx_name (name)`, `idx_categories_user (user_id)`

---

## qa_pairs

Question-and-answer training data. These drive the AI response engine. Supports semantic embeddings and variations for fuzzy matching.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| category_id | VARCHAR(36) | NO | — | Parent category |
| question | TEXT | NO | — | Question text (full-text indexed) |
| answer | TEXT | NO | — | Answer text |
| variations | JSON | YES | NULL | Alternative phrasings of the question |
| embedding | JSON | YES | NULL | Vector embedding for semantic search |
| is_active | BOOLEAN | YES | true | Whether this QA is active |
| usage_count | INT | YES | 0 | Times this QA was matched |
| user_id | VARCHAR(36) | YES | NULL | Owning workspace user |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | YES | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_category (category_id)`, `idx_qa_user (user_id)`, `FULLTEXT idx_question (question)`

---

## unknown_questions

Questions from customers that could not be matched to any QA pair. Used for training pipeline improvement.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| question | TEXT | NO | — | Unmatched question text |
| conversation_id | VARCHAR(36) | YES | NULL | Source conversation |
| channel | VARCHAR(50) | YES | NULL | Channel it came from |
| status | ENUM('pending','trained','ignored') | YES | 'pending' | Processing status |
| suggested_answer | TEXT | YES | NULL | AI-suggested answer for review |
| category_id | VARCHAR(36) | YES | NULL | Suggested category assignment |
| user_id | VARCHAR(36) | YES | NULL | Owning workspace user |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | When the question was logged |

**Indexes:** `idx_status (status)`, `idx_created (created_at)`, `idx_uq_user_status (user_id, status)`

---

## integrations

Channel integrations (Telegram, WhatsApp, Instagram, etc.) configured per workspace.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning workspace user |
| channel | VARCHAR(50) | NO | — | Channel type (telegram, whatsapp, etc.) |
| status | ENUM('active','inactive','error') | YES | 'inactive' | Connection status |
| config | JSON | YES | NULL | Channel-specific configuration |
| webhook_url | VARCHAR(500) | YES | NULL | Registered webhook URL |
| last_error | TEXT | YES | NULL | Last connection error message |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | YES | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_user (user_id)`, `idx_channel (channel)`

---

## team_members

Workspace team membership. Links owner workspaces to agent/admin users.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| owner_id | VARCHAR(36) | NO | — | Workspace owner user ID |
| user_id | VARCHAR(36) | NO | — | Team member user ID |
| role | ENUM('admin','agent') | YES | 'agent' | Role within workspace |
| is_active | BOOLEAN | YES | true | Membership active flag |
| joined_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | When member joined |

**Indexes:** `idx_owner (owner_id)`, `idx_user (user_id)`

---

## api_keys

Programmatic API keys for external integrations. Stored as hashes.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning user |
| name | VARCHAR(100) | NO | — | Human-readable key name |
| key_hash | VARCHAR(255) | NO | — | SHA-256 hash of the API key |
| last_used | TIMESTAMP | YES | NULL | Last usage timestamp |
| is_active | BOOLEAN | YES | true | Key enabled flag |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |

**Indexes:** `idx_user (user_id)`, `idx_key (key_hash)`

---

## archive_folders

Folders for archiving conversations, contacts, and locations.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning workspace user |
| name | VARCHAR(100) | NO | — | Folder name |
| type | ENUM('chats','contacts','locations') | YES | 'chats' | Archive content type |
| color | VARCHAR(7) | YES | '#6b7280' | UI display color |
| item_count | INT | YES | 0 | Number of items in folder |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |

**Indexes:** `idx_user_type (user_id, type)`

---

## subscriptions

Active and historical subscriptions per user.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning user |
| plan_id | VARCHAR(50) | NO | — | Plan identifier |
| status | ENUM('active','cancelled','expired') | YES | 'active' | Subscription status |
| current_period_start | TIMESTAMP | NO | — | Period start timestamp |
| current_period_end | TIMESTAMP | NO | — | Period end timestamp |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | YES | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_user_status (user_id, status)`, `idx_period (current_period_end)`

---

## payments

Payment transaction records.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning user |
| subscription_id | VARCHAR(36) | YES | NULL | Related subscription |
| amount | INT | NO | — | Amount in smallest currency unit |
| currency | VARCHAR(3) | YES | 'NGN' | ISO 4217 currency code |
| status | ENUM('pending','completed','failed','refunded') | YES | 'pending' | Payment status |
| provider | VARCHAR(50) | YES | 'polar' | Payment provider name |
| provider_payment_id | VARCHAR(255) | YES | NULL | Provider's transaction ID |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Transaction time |

**Indexes:** `idx_user (user_id)`, `idx_status (status)`

---

## audit_logs

Immutable audit trail for all user actions.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Acting user |
| action | VARCHAR(255) | NO | — | Action identifier (e.g. `login`, `create_qa`) |
| resource_type | VARCHAR(100) | YES | NULL | Resource type affected |
| resource_id | VARCHAR(36) | YES | NULL | ID of affected resource |
| details | JSON | YES | NULL | Action-specific metadata |
| ip_address | VARCHAR(45) | YES | NULL | Client IP (supports IPv6) |
| user_agent | TEXT | YES | NULL | Client user agent string |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Event timestamp |

**Indexes:** `idx_audit_user (user_id)`, `idx_audit_created (created_at)`, `idx_audit_action (action)`

---

## notifications

In-app notification entries per user.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Recipient user |
| type | VARCHAR(100) | NO | — | Notification type category |
| title | VARCHAR(255) | NO | — | Notification title |
| body | TEXT | NO | — | Notification body text |
| link | VARCHAR(500) | YES | NULL | Deep link for click-through |
| is_read | BOOLEAN | YES | FALSE | Read status |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |

**Indexes:** `idx_notif_user (user_id)`, `idx_notif_read (user_id, is_read)`, `idx_notif_created (created_at)`

---

## widget_configs

Embedded chat widget configuration per workspace.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | UNIQUE | Owning workspace (one config per user) |
| brand_color | VARCHAR(20) | YES | '#3b82f6' | Widget brand color |
| greeting | VARCHAR(500) | YES | 'Hello! How can I help you today?' | Initial greeting message |
| bot_name | VARCHAR(100) | YES | 'Noant AI' | Displayed bot name |
| position | VARCHAR(20) | YES | 'bottom-right' | Widget position on page |
| widget_api_key | VARCHAR(100) | NO | — | Public API key for embed |
| is_active | BOOLEAN | YES | TRUE | Widget enabled flag |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | YES | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_widget_user (user_id)`, `idx_widget_key (widget_api_key)`

---

## inventory_items

Product, service, or package catalog for sales handoff workflows.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning workspace user |
| type | ENUM('product','service','package') | YES | 'product' | Item category |
| name | VARCHAR(255) | NO | — | Item name |
| description | TEXT | YES | NULL | Item description |
| price | DECIMAL(15,2) | NO | — | Listed price |
| min_price | DECIMAL(15,2) | YES | NULL | Minimum acceptable price |
| stock_quantity | INT | YES | NULL | Available stock count |
| image_url | VARCHAR(500) | YES | NULL | Product image URL |
| is_active | BOOLEAN | YES | TRUE | Listed flag |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | YES | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_user_active (user_id, is_active)`

---

## handoffs

Sales handoff records when AI transfers a buying conversation to a human agent.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning workspace user |
| conversation_id | VARCHAR(36) | NO | — | Source conversation |
| customer_name | VARCHAR(100) | YES | NULL | Customer name |
| customer_phone | VARCHAR(50) | YES | NULL | Customer phone |
| customer_whatsapp | VARCHAR(50) | YES | NULL | Customer WhatsApp number |
| customer_location | TEXT | YES | NULL | Customer location string |
| product_name | VARCHAR(255) | YES | NULL | Product being discussed |
| original_price | DECIMAL(15,2) | YES | NULL | List price |
| agreed_price | DECIMAL(15,2) | YES | NULL | Agreed sale price |
| quantity | INT | YES | 1 | Number of units |
| status | ENUM('pending','sold','lost','expired') | YES | 'pending' | Handoff outcome |
| final_price | DECIMAL(15,2) | YES | NULL | Final sale price |
| owner_notes | TEXT | YES | NULL | Agent notes |
| owner_notified_at | TIMESTAMP | YES | NULL | When owner was notified |
| reminder_count | INT | YES | 0 | Number of reminders sent |
| next_reminder_at | TIMESTAMP | YES | NULL | Next scheduled reminder |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | YES | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_user_status (user_id, status)`

---

## user_credits

Pulse AI response credit balance per user.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning user (FK -> users.id CASCADE) |
| balance | INT | NO | 0 | Remaining credit count |
| expires_at | TIMESTAMP | YES | NULL | Credit expiration time |
| last_updated_at | TIMESTAMP | NO | CURRENT_TIMESTAMP ON UPDATE | Last balance change |

**Indexes:** `idx_user_id (user_id)`, `idx_expires_at (expires_at)`, `idx_user_credits_expiring (expires_at)`

**Foreign Keys:** `user_id -> users(id) ON DELETE CASCADE`

---

## credit_purchases

Historical record of credit pack purchases.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Purchasing user (FK -> users.id CASCADE) |
| checkout_id | VARCHAR(100) | NO | UNIQUE | Payment provider checkout ID |
| pack_type | VARCHAR(20) | NO | — | Pack tier identifier |
| amount | INT | NO | — | Credits purchased |
| status | VARCHAR(20) | NO | 'pending' | Purchase status |
| purchased_at | TIMESTAMP | NO | CURRENT_TIMESTAMP | Purchase timestamp |
| expires_at | TIMESTAMP | NO | — | When these credits expire |

**Indexes:** `idx_user_id (user_id)`, `idx_checkout_id (checkout_id)`, `idx_status (status)`, `idx_purchased_at (purchased_at)`, `idx_credit_purchases_user_month (user_id, purchased_at)`

**Foreign Keys:** `user_id -> users(id) ON DELETE CASCADE`

---

## campaign_schedules

Broadcast campaign definitions.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning user (FK -> users.id CASCADE) |
| name | VARCHAR(100) | NO | — | Campaign name |
| start_date | DATE | NO | — | Scheduled start date |
| end_date | DATE | NO | — | Scheduled end date |
| status | VARCHAR(20) | NO | 'draft' | Campaign status |
| created_at | TIMESTAMP | NO | CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | NO | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_user_id (user_id)`, `idx_status (status)`, `idx_start_date (start_date)`, `idx_end_date (end_date)`

**Foreign Keys:** `user_id -> users(id) ON DELETE CASCADE`

---

## campaign_recipients

Individual recipients within a broadcast campaign.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| campaign_id | VARCHAR(36) | NO | — | Parent campaign (FK -> campaign_schedules.id CASCADE) |
| user_id | VARCHAR(36) | NO | — | Owning user (FK -> users.id CASCADE) |
| phone | VARCHAR(20) | NO | — | Recipient phone number |
| name | VARCHAR(100) | YES | NULL | Recipient name |
| status | VARCHAR(20) | NO | 'pending' | Delivery status |
| error | TEXT | YES | NULL | Delivery error message |
| sent_at | TIMESTAMP | YES | NULL | When message was sent |
| delivered_at | TIMESTAMP | YES | NULL | Delivery confirmation time |
| read_at | TIMESTAMP | YES | NULL | Read receipt time |
| created_at | TIMESTAMP | NO | CURRENT_TIMESTAMP | Creation time |

**Indexes:** `idx_campaign_id (campaign_id)`, `idx_user_id (user_id)`, `idx_phone (phone)`, `idx_status (status)`

**Foreign Keys:** `campaign_id -> campaign_schedules(id) ON DELETE CASCADE`, `user_id -> users(id) ON DELETE CASCADE`

---

## whatsapp_templates

WhatsApp HSM (Highly Structured Message) templates.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning user (FK -> users.id CASCADE) |
| name | VARCHAR(100) | NO | — | Template name |
| language | VARCHAR(10) | NO | 'en' | Template language code |
| category | VARCHAR(20) | NO | 'utility' | HSM category |
| status | VARCHAR(20) | NO | 'draft' | Approval status |
| header_type | VARCHAR(20) | NO | 'none' | Header media type |
| header_value | TEXT | YES | NULL | Header content |
| body_text | TEXT | NO | — | Template body text |
| footer_text | TEXT | YES | NULL | Template footer text |
| buttons | JSON | YES | NULL | Interactive button definitions |
| namespace | VARCHAR(100) | YES | NULL | Meta template namespace |
| rejection_reason | TEXT | YES | NULL | Meta rejection reason |
| created_at | TIMESTAMP | NO | CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP | NO | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `idx_user_id (user_id)`, `idx_status (status)`, `idx_name (name)`

**Foreign Keys:** `user_id -> users(id) ON DELETE CASCADE`

---

## media_messages

Uploaded media attachments for conversations (images, video, audio, documents).

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Owning user (FK -> users.id CASCADE) |
| conversation_id | VARCHAR(36) | NO | — | Parent conversation (FK -> conversations.id CASCADE) |
| message_id | VARCHAR(36) | YES | NULL | Associated message |
| session_id | VARCHAR(100) | YES | NULL | OpenWA session identifier |
| media_type | VARCHAR(20) | NO | — | Media category (image, video, audio, document) |
| mime_type | VARCHAR(100) | YES | NULL | MIME type |
| file_size | BIGINT | NO | 0 | File size in bytes |
| file_name | VARCHAR(255) | YES | NULL | Original file name |
| file_path | VARCHAR(500) | YES | NULL | Stored file path |
| thumb_path | VARCHAR(500) | YES | NULL | Thumbnail file path |
| width | INT | NO | 0 | Image/video width in pixels |
| height | INT | NO | 0 | Image/video height in pixels |
| duration | INT | NO | 0 | Audio/video duration in seconds |
| caption | TEXT | YES | NULL | Media caption text |
| remote_url | VARCHAR(500) | YES | NULL | Remote source URL |
| created_at | TIMESTAMP | NO | CURRENT_TIMESTAMP | Upload time |
| expires_at | TIMESTAMP | NO | — | When the media file expires |

**Indexes:** `idx_user_id (user_id)`, `idx_conversation_id (conversation_id)`, `idx_media_type (media_type)`, `idx_expires_at (expires_at)`

**Foreign Keys:** `user_id -> users(id) ON DELETE CASCADE`, `conversation_id -> conversations(id) ON DELETE CASCADE`

---

## csat_ratings

Customer satisfaction ratings for resolved conversations.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Workspace owner |
| conversation_id | VARCHAR(36) | NO | — | Rated conversation |
| score | TINYINT | NO | — | Rating 1-5 (CHECK constraint) |
| comment | TEXT | YES | NULL | Optional feedback text |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | Rating time |

**Indexes:** `idx_csat_user (user_id)`, `idx_csat_conv (conversation_id)`, `idx_csat_created (created_at)`

---

## push_subscriptions

Web Push notification subscriptions for PWA clients.

| Column | Type | Nullable | Default | Description |
|---|---|---|---|---|
| id | VARCHAR(36) | NO | — | Primary key (UUID) |
| user_id | VARCHAR(36) | NO | — | Subscribing user |
| endpoint | TEXT | NO | — | Push service endpoint URL |
| auth | VARCHAR(255) | NO | — | Push subscription auth key |
| p256dh | VARCHAR(255) | NO | — | Push subscription encryption key |
| user_agent | VARCHAR(500) | YES | NULL | Client user agent |
| created_at | DATETIME | NO | CURRENT_TIMESTAMP | Subscription time |
| updated_at | DATETIME | NO | CURRENT_TIMESTAMP ON UPDATE | Last update time |

**Indexes:** `uq_endpoint (endpoint(255))` UNIQUE, `idx_push_user (user_id)`

---

## Entity Relationships

```
users (1) ──────── (*) conversations
users (1) ──────── (*) categories
users (1) ──────── (*) qa_pairs
users (1) ──────── (*) unknown_questions
users (1) ──────── (*) integrations
users (1) ──────── (*) team_members (as owner_id)
users (1) ──────── (*) team_members (as user_id)
users (1) ──────── (*) api_keys
users (1) ──────── (*) archive_folders
users (1) ──────── (*) subscriptions
users (1) ──────── (*) payments
users (1) ──────── (*) audit_logs
users (1) ──────── (*) notifications
users (1) ────── (1) widget_configs
users (1) ──────── (*) inventory_items
users (1) ──────── (*) handoffs
users (1) ────── (1) user_credits
users (1) ──────── (*) credit_purchases
users (1) ──────── (*) campaign_schedules
users (1) ──────── (*) campaign_recipients
users (1) ──────── (*) whatsapp_templates
users (1) ──────── (*) media_messages
users (1) ──────── (*) csat_ratings
users (1) ──────── (*) push_subscriptions

conversations (1) ──────── (*) messages
conversations (1) ──────── (*) handoffs
conversations (1) ──────── (*) media_messages
conversations (1) ──────── (*) csat_ratings

categories (1) ──────── (*) qa_pairs

campaign_schedules (1) ──── (*) campaign_recipients
```

---

## Migration Version History

| Version | File | Description |
|---|---|---|
| 001 | `001_init.sql` | Initial schema: users, conversations, messages, categories, qa_pairs, unknown_questions, integrations, team_members, api_keys, archive_folders, subscriptions, payments |
| 003 | `003_audit_logs.sql` | Create audit_logs table |
| 004 | `004_audit_logs.sql` | Re-apply audit_logs table (idempotent) |
| 005 | `005_user_isolation.sql` | Add user_id columns to categories, unknown_questions, qa_pairs for workspace isolation with backfill |
| 006 | `006_notifications_widget.sql` | Create notifications and widget_configs tables; add notification preference columns to users |
| 007 | `007_message_source.sql` | Add source column to messages |
| 008 | `008_inventory_leads.sql` | Create inventory_items and handoffs tables; add owner_whatsapp to users |
| 009 | `009_inventory_leads_fix.sql` | Re-apply inventory_items and handoffs tables (repair) |
| 010 | `010_billing.sql` | Create user_credits, credit_purchases, campaign_schedules tables; add credit_balance to users |
| 011 | `011_email_verification.sql` | Add is_verified and verification_code columns to users |
| 012 | `012_openwa_hardening.sql` | Create whatsapp_templates, campaign_recipients, media_messages tables; add customer_avatar to conversations |
| 013 | `013_message_sequence.sql` | Add sequence column to messages; backfill and add composite index |
| 014 | `014_onboarding.sql` | Add onboarding_status and industry columns to users |
| 015 | `015_csat_analytics.sql` | Create csat_ratings table; add composite index on unknown_questions |
| 016 | `016_push_subscriptions.sql` | Create push_subscriptions table for PWA Web Push |
| 017 | `017_schema_repair.sql` | Consolidate inline DDL: ensure inventory_items, handoffs, audit_logs, owner_whatsapp exist |
