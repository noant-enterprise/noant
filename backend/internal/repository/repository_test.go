package repository

import (
	"context"
	"testing"

	"noant/internal/domain"
)

// Compile-time interface checks (verifies all 22 repos implement their interfaces).
var (
	_ IUserRepo              = (*MockUserRepo)(nil)
	_ IConversationRepo      = (*MockConversationRepo)(nil)
	_ IMessageRepo           = (*MockMessageRepo)(nil)
	_ IQAPairRepo            = (*MockQAPairRepo)(nil)
	_ ICategoryRepo          = (*MockCategoryRepo)(nil)
	_ IUnknownQuestionRepo   = (*MockUnknownQuestionRepo)(nil)
	_ IIntegrationRepo       = (*MockIntegrationRepo)(nil)
	_ ITeamRepo              = (*MockTeamRepo)(nil)
	_ IAPIKeyRepo            = (*MockAPIKeyRepo)(nil)
	_ IArchiveRepo           = (*MockArchiveRepo)(nil)
	_ ISubscriptionRepo      = (*MockSubscriptionRepo)(nil)
	_ IAuditRepo             = (*MockAuditRepo)(nil)
	_ INotificationRepo      = (*MockNotificationRepo)(nil)
	_ IWidgetConfigRepo      = (*MockWidgetConfigRepo)(nil)
	_ IInventoryRepo         = (*MockInventoryRepo)(nil)
	_ IHandoffRepo           = (*MockHandoffRepo)(nil)
	_ ICreditRepo            = (*MockCreditRepo)(nil)
	_ ICampaignRepo          = (*MockCampaignRepo)(nil)
	_ IWhatsAppTemplateRepo  = (*MockWhatsAppTemplateRepo)(nil)
	_ ICampaignRecipientRepo = (*MockCampaignRecipientRepo)(nil)
	_ IMediaMessageRepo      = (*MockMediaMessageRepo)(nil)
	_ IPushSubscriptionRepo  = (*MockPushSubscriptionRepo)(nil)
)

func TestMockRepositories_AllInstantiable(t *testing.T) {
	m := NewMockRepositories()
	if m == nil {
		t.Fatal("NewMockRepositories returned nil")
	}
	// Verify all 22 repos are non-nil
	if m.User == nil { t.Error("User repo is nil") }
	if m.Conversation == nil { t.Error("Conversation repo is nil") }
	if m.Message == nil { t.Error("Message repo is nil") }
	if m.QAPair == nil { t.Error("QAPair repo is nil") }
	if m.Category == nil { t.Error("Category repo is nil") }
	if m.UnknownQ == nil { t.Error("UnknownQ repo is nil") }
	if m.Integration == nil { t.Error("Integration repo is nil") }
	if m.Team == nil { t.Error("Team repo is nil") }
	if m.APIKey == nil { t.Error("APIKey repo is nil") }
	if m.Archive == nil { t.Error("Archive repo is nil") }
	if m.Subscription == nil { t.Error("Subscription repo is nil") }
	if m.Audit == nil { t.Error("Audit repo is nil") }
	if m.Notification == nil { t.Error("Notification repo is nil") }
	if m.WidgetConfig == nil { t.Error("WidgetConfig repo is nil") }
	if m.Inventory == nil { t.Error("Inventory repo is nil") }
	if m.Handoff == nil { t.Error("Handoff repo is nil") }
	if m.Credit == nil { t.Error("Credit repo is nil") }
	if m.Campaign == nil { t.Error("Campaign repo is nil") }
	if m.WhatsAppTemplate == nil { t.Error("WhatsAppTemplate repo is nil") }
	if m.CampaignRecipient == nil { t.Error("CampaignRecipient repo is nil") }
	if m.MediaMessage == nil { t.Error("MediaMessage repo is nil") }
	if m.PushSubscription == nil { t.Error("PushSubscription repo is nil") }
}

// ============================================================
// User Repository Tests
// ============================================================

func TestMockUserRepo_CRUD(t *testing.T) {
	repo := NewMockUserRepo()
	ctx := context.Background()

	user := &domain.User{
		ID:    "user-1",
		Email: "test@example.com",
		PlanID: "pulse",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, "user-1")
	if err != nil { t.Fatalf("GetByID: %v", err) }
	if got == nil { t.Fatal("GetByID returned nil") }
	if got.Email != "test@example.com" {
		t.Errorf("Email = %q, want test@example.com", got.Email)
	}

	gotByEmail, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil { t.Fatalf("GetByEmail: %v", err) }
	if gotByEmail == nil { t.Fatal("GetByEmail returned nil") }
	if gotByEmail.ID != "user-1" {
		t.Errorf("GetByEmail ID = %q, want user-1", gotByEmail.ID)
	}

	if err := repo.UpdatePlan(ctx, "user-1", "pro"); err != nil {
		t.Fatalf("UpdatePlan: %v", err)
	}
	got, _ = repo.GetByID(ctx, "user-1")
	if got.PlanID != "pro" {
		t.Errorf("PlanID = %q, want pro", got.PlanID)
	}

	if err := repo.Delete(ctx, "user-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, _ = repo.GetByID(ctx, "user-1")
	if got != nil {
		t.Error("user should be deleted")
	}
}

func TestMockUserRepo_GetByEmail_NotFound(t *testing.T) {
	repo := NewMockUserRepo()
	got, err := repo.GetByEmail(context.Background(), "nonexistent@example.com")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if got != nil { t.Error("expected nil for non-existent email") }
}

// ============================================================
// Conversation Repository Tests
// ============================================================

func TestMockConversationRepo_CRUD(t *testing.T) {
	repo := NewMockConversationRepo()
	ctx := context.Background()

	conv := &domain.Conversation{
		ID:           "conv-1",
		UserID:       "user-1",
		CustomerName: "Test Customer",
		Channel:      "web",
		Status:       "active",
	}
	if err := repo.Create(ctx, conv); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, "conv-1")
	if err != nil { t.Fatalf("GetByID: %v", err) }
	if got == nil { t.Fatal("GetByID returned nil") }
	if got.CustomerName != "Test Customer" {
		t.Errorf("CustomerName = %q, want Test Customer", got.CustomerName)
	}

	convs, total, err := repo.List(ctx, "user-1", "", 10, 0)
	if err != nil { t.Fatalf("List: %v", err) }
	if total != 1 { t.Errorf("total = %d, want 1", total) }
	if len(convs) != 1 { t.Errorf("len = %d, want 1", len(convs)) }

	if err := repo.UpdateStatus(ctx, "conv-1", "user-1", "resolved"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ = repo.GetByID(ctx, "conv-1")
	if got.Status != "resolved" {
		t.Errorf("Status = %q, want resolved", got.Status)
	}
}

// ============================================================
// Message Repository Tests
// ============================================================

func TestMockMessageRepo_CreateAndList(t *testing.T) {
	repo := NewMockMessageRepo()
	ctx := context.Background()

	msg := &domain.Message{
		ID:             "msg-1",
		ConversationID: "conv-1",
		Role:           "customer",
		Content:        "Hello!",
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatalf("Create: %v", err)
	}

	msgs, err := repo.ListByConversation(ctx, "conv-1", 10)
	if err != nil { t.Fatalf("ListByConversation: %v", err) }
	if len(msgs) != 1 { t.Errorf("len = %d, want 1", len(msgs)) }
	if msgs[0].Content != "Hello!" {
		t.Errorf("Content = %q, want Hello!", msgs[0].Content)
	}

	msg2 := &domain.Message{
		ID:             "msg-2",
		ConversationID: "conv-1",
		Role:           "ai",
		Content:        "Hi there!",
	}
	_ = repo.Create(ctx, msg2)

	last, err := repo.GetLastMessage(ctx, "conv-1")
	if err != nil { t.Fatalf("GetLastMessage: %v", err) }
	if last == nil { t.Fatal("GetLastMessage returned nil") }
	if last.Content != "Hi there!" {
		t.Errorf("Last message Content = %q, want Hi there!", last.Content)
	}
}

// ============================================================
// QAPair Repository Tests
// ============================================================

func TestMockQAPairRepo_CRUD(t *testing.T) {
	repo := NewMockQAPairRepo()
	ctx := context.Background()

	qa := &domain.QAPair{
		ID:         "qa-1",
		UserID:     "user-1",
		CategoryID: "cat-1",
		Question:   "What is your return policy?",
		Answer:     "30 day returns",
		IsActive:   true,
	}
	if err := repo.Create(ctx, qa); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, "qa-1")
	if err != nil { t.Fatalf("GetByID: %v", err) }
	if got == nil { t.Fatal("GetByID returned nil") }
	if got.Question != "What is your return policy?" {
		t.Errorf("Question = %q", got.Question)
	}

	qas, err := repo.ListByUser(ctx, "user-1", "")
	if err != nil { t.Fatalf("ListByUser: %v", err) }
	if len(qas) != 1 { t.Errorf("len = %d, want 1", len(qas)) }

	if err := repo.Update(ctx, &domain.QAPair{ID: "qa-1", UserID: "user-1", Question: "Updated?", Answer: "Updated!", IsActive: true}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = repo.GetByID(ctx, "qa-1")
	if got.Question != "Updated?" {
		t.Errorf("Question after update = %q, want Updated?", got.Question)
	}

	count, err := repo.CountByUser(ctx, "user-1")
	if err != nil { t.Fatalf("CountByUser: %v", err) }
	if count != 1 { t.Errorf("count = %d, want 1", count) }

	if err := repo.Delete(ctx, "qa-1", "user-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, _ = repo.GetByID(ctx, "qa-1")
	if got != nil { t.Error("QA pair should be deleted") }
}

func TestMockQAPairRepo_Search(t *testing.T) {
	repo := NewMockQAPairRepo()
	ctx := context.Background()

	_ = repo.Create(ctx, &domain.QAPair{ID: "qa-1", UserID: "user-1", Question: "How much does it cost?", Answer: "₦5000", IsActive: true})
	_ = repo.Create(ctx, &domain.QAPair{ID: "qa-2", UserID: "user-1", Question: "What colors are available?", Answer: "Red and Blue", IsActive: true})

	results, err := repo.Search(ctx, "user-1", "cost")
	if err != nil { t.Fatalf("Search: %v", err) }
	if len(results) == 0 { t.Error("Search returned no results for 'cost'") }
}

// ============================================================
// Category Repository Tests
// ============================================================

func TestMockCategoryRepo_CRUD(t *testing.T) {
	repo := NewMockCategoryRepo()
	ctx := context.Background()

	cat := &domain.Category{ID: "cat-1", UserID: "user-1", Name: "Pricing"}
	if err := repo.Create(ctx, cat); err != nil {
		t.Fatalf("Create: %v", err)
	}

	cats, err := repo.List(ctx, "user-1")
	if err != nil { t.Fatalf("List: %v", err) }
	if len(cats) != 1 { t.Errorf("len = %d, want 1", len(cats)) }

	got, err := repo.GetByName(ctx, "user-1", "Pricing")
	if err != nil { t.Fatalf("GetByName: %v", err) }
	if got == nil { t.Fatal("GetByName returned nil") }

	if err := repo.Delete(ctx, "cat-1", "user-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	cats, _ = repo.List(ctx, "user-1")
	if len(cats) != 0 { t.Errorf("after delete len = %d, want 0", len(cats)) }
}

// ============================================================
// Notification Repository Tests
// ============================================================

func TestMockNotificationRepo_CRUD(t *testing.T) {
	repo := NewMockNotificationRepo()
	ctx := context.Background()

	n := &domain.Notification{ID: "notif-1", UserID: "user-1", Type: "info", Title: "Test", Body: "Hello"}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("Create: %v", err)
	}

	notifs, err := repo.ListByUser(ctx, "user-1", 10)
	if err != nil { t.Fatalf("ListByUser: %v", err) }
	if len(notifs) != 1 { t.Errorf("len = %d, want 1", len(notifs)) }

	count, err := repo.UnreadCount(ctx, "user-1")
	if err != nil { t.Fatalf("UnreadCount: %v", err) }
	if count != 1 { t.Errorf("UnreadCount = %d, want 1", count) }

	if err := repo.MarkRead(ctx, "notif-1", "user-1"); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	count, _ = repo.UnreadCount(ctx, "user-1")
	if count != 0 { t.Errorf("UnreadCount after mark = %d, want 0", count) }
}

// ============================================================
// Inventory Repository Tests
// ============================================================

func TestMockInventoryRepo_CRUD(t *testing.T) {
	repo := NewMockInventoryRepo()
	ctx := context.Background()

	item := &domain.InventoryItem{ID: "inv-1", UserID: "user-1", Name: "Shirt", Price: 5000, Type: "product", IsActive: true}
	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, "inv-1", "user-1")
	if err != nil { t.Fatalf("GetByID: %v", err) }
	if got == nil { t.Fatal("GetByID returned nil") }
	if got.Name != "Shirt" {
		t.Errorf("Name = %q, want Shirt", got.Name)
	}

	items, err := repo.List(ctx, "user-1", "", true)
	if err != nil { t.Fatalf("List: %v", err) }
	if len(items) != 1 { t.Errorf("len = %d, want 1", len(items)) }

	if err := repo.Delete(ctx, "inv-1", "user-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, _ = repo.GetByID(ctx, "inv-1", "user-1")
	if got != nil { t.Error("item should be deleted") }
}

// ============================================================
// Integration Repository Tests
// ============================================================

func TestMockIntegrationRepo_CRUD(t *testing.T) {
	repo := NewMockIntegrationRepo()
	ctx := context.Background()

	integ := &domain.Integration{ID: "int-1", UserID: "user-1", Channel: "whatsapp", Status: "connected"}
	if err := repo.Create(ctx, integ); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByUserAndChannel(ctx, "user-1", "whatsapp")
	if err != nil { t.Fatalf("GetByUserAndChannel: %v", err) }
	if got == nil { t.Fatal("GetByUserAndChannel returned nil") }

	integrations, err := repo.ListByUser(ctx, "user-1")
	if err != nil { t.Fatalf("ListByUser: %v", err) }
	if len(integrations) != 1 { t.Errorf("len = %d, want 1", len(integrations)) }
}

// ============================================================
// Handoff Repository Tests
// ============================================================

func TestMockHandoffRepo_CRUD(t *testing.T) {
	repo := NewMockHandoffRepo()
	ctx := context.Background()

	h := &domain.Handoff{ID: "h-1", UserID: "user-1", ConversationID: "conv-1", Status: "pending"}
	if err := repo.Create(ctx, h); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, "h-1", "user-1")
	if err != nil { t.Fatalf("GetByID: %v", err) }
	if got == nil { t.Fatal("GetByID returned nil") }

	handoffs, err := repo.List(ctx, "user-1", "pending", 10)
	if err != nil { t.Fatalf("List: %v", err) }
	if len(handoffs) != 1 { t.Errorf("len = %d, want 1", len(handoffs)) }

	if err := repo.UpdateStatus(ctx, "h-1", "user-1", "resolved", "Fixed it"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ = repo.GetByID(ctx, "h-1", "user-1")
	if got.Status != "resolved" {
		t.Errorf("Status = %q, want resolved", got.Status)
	}
}

// ============================================================
// Credit Repository Tests
// ============================================================

func TestMockCreditRepo_UpsertAndGet(t *testing.T) {
	repo := NewMockCreditRepo()
	ctx := context.Background()

	credit := &domain.UserCredit{UserID: "user-1", Balance: 100}
	if err := repo.Upsert(ctx, credit); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := repo.GetByUserID(ctx, "user-1")
	if err != nil { t.Fatalf("GetByUserID: %v", err) }
	if got == nil { t.Fatal("GetByUserID returned nil") }
	if got.Balance != 100 {
		t.Errorf("Balance = %d, want 100", got.Balance)
	}

	if err := repo.Deduct(ctx, "user-1", 30); err != nil {
		t.Fatalf("Deduct: %v", err)
	}
	got, _ = repo.GetByUserID(ctx, "user-1")
	if got.Balance != 70 {
		t.Errorf("Balance after deduct = %d, want 70", got.Balance)
	}
}

// ============================================================
// Campaign Repository Tests
// ============================================================

func TestMockCampaignRepo_CRUD(t *testing.T) {
	repo := NewMockCampaignRepo()
	ctx := context.Background()

	c := &domain.CampaignSchedule{ID: "camp-1", UserID: "user-1", Name: "Test Campaign", Status: "draft"}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}

	camps, err := repo.ListByUser(ctx, "user-1")
	if err != nil { t.Fatalf("ListByUser: %v", err) }
	if len(camps) != 1 { t.Errorf("len = %d, want 1", len(camps)) }
}

// ============================================================
// Audit Repository Tests
// ============================================================

func TestMockAuditRepo_CreateAndList(t *testing.T) {
	repo := NewMockAuditRepo()
	ctx := context.Background()

	log := &domain.AuditLog{ID: "audit-1", UserID: "user-1", Action: "login"}
	if err := repo.Create(ctx, log); err != nil {
		t.Fatalf("Create: %v", err)
	}

	logs, err := repo.ListByUser(ctx, "user-1", 10)
	if err != nil { t.Fatalf("ListByUser: %v", err) }
	if len(logs) != 1 { t.Errorf("len = %d, want 1", len(logs)) }
}

// ============================================================
// Widget Config Repository Tests
// ============================================================

func TestMockWidgetConfigRepo_UpsertAndGet(t *testing.T) {
	repo := NewMockWidgetConfigRepo()
	ctx := context.Background()

	cfg := &domain.WidgetConfig{UserID: "user-1", WidgetAPIKey: "key-123", BrandColor: "#0ea5e9", IsActive: true}
	if err := repo.Upsert(ctx, cfg); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := repo.Get(ctx, "user-1")
	if err != nil { t.Fatalf("Get: %v", err) }
	if got == nil { t.Fatal("Get returned nil") }
	if got.BrandColor != "#0ea5e9" {
		t.Errorf("BrandColor = %q, want #0ea5e9", got.BrandColor)
	}

	gotByKey, err := repo.GetByAPIKey(ctx, "key-123")
	if err != nil { t.Fatalf("GetByAPIKey: %v", err) }
	if gotByKey == nil { t.Fatal("GetByAPIKey returned nil") }
}

// ============================================================
// Push Subscription Repository Tests
// ============================================================

func TestMockPushSubscriptionRepo_CRUD(t *testing.T) {
	repo := NewMockPushSubscriptionRepo()
	ctx := context.Background()

	sub := &domain.PushSubscription{ID: "sub-1", UserID: "user-1", Endpoint: "https://example.com/push"}
	if err := repo.Create(ctx, sub); err != nil {
		t.Fatalf("Create: %v", err)
	}

	subs, err := repo.ListByUser(ctx, "user-1")
	if err != nil { t.Fatalf("ListByUser: %v", err) }
	if len(subs) != 1 { t.Errorf("len = %d, want 1", len(subs)) }

	if err := repo.Delete(ctx, "user-1", "https://example.com/push"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	subs, _ = repo.ListByUser(ctx, "user-1")
	if len(subs) != 0 { t.Errorf("after delete len = %d, want 0", len(subs)) }
}

// ============================================================
// WhatsApp Template Repository Tests
// ============================================================

func TestMockWhatsAppTemplateRepo_CRUD(t *testing.T) {
	repo := NewMockWhatsAppTemplateRepo()
	ctx := context.Background()

	tpl := &domain.WhatsAppTemplate{ID: "tpl-1", UserID: "user-1", Name: "Welcome", Status: "approved"}
	if err := repo.Create(ctx, tpl); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, "tpl-1", "user-1")
	if err != nil { t.Fatalf("GetByID: %v", err) }
	if got == nil { t.Fatal("GetByID returned nil") }

	tpls, err := repo.ListByUser(ctx, "user-1")
	if err != nil { t.Fatalf("ListByUser: %v", err) }
	if len(tpls) != 1 { t.Errorf("len = %d, want 1", len(tpls)) }

	if err := repo.Delete(ctx, "tpl-1", "user-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got, _ = repo.GetByID(ctx, "tpl-1", "user-1")
	if got != nil { t.Error("template should be deleted") }
}

// ============================================================
// Media Message Repository Tests
// ============================================================

func TestMockMediaMessageRepo_CreateAndList(t *testing.T) {
	repo := NewMockMediaMessageRepo()
	ctx := context.Background()

	m := &domain.MediaMessage{ID: "media-1", ConversationID: "conv-1", MediaType: "image", RemoteURL: "https://example.com/img.jpg"}
	if err := repo.Create(ctx, m); err != nil {
		t.Fatalf("Create: %v", err)
	}

	msgs, err := repo.GetByConversation(ctx, "conv-1")
	if err != nil { t.Fatalf("GetByConversation: %v", err) }
	if len(msgs) != 1 { t.Errorf("len = %d, want 1", len(msgs)) }
}

// ============================================================
// Team Repository Tests
// ============================================================

func TestMockTeamRepo_CreateAndList(t *testing.T) {
	repo := NewMockTeamRepo()
	ctx := context.Background()

	member := &domain.TeamMember{ID: "tm-1", UserID: "user-2", Role: "agent"}
	if err := repo.Create(ctx, "org-1", member); err != nil {
		t.Fatalf("Create: %v", err)
	}

	members, err := repo.ListByOrg(ctx, "org-1")
	if err != nil { t.Fatalf("ListByOrg: %v", err) }
	if len(members) != 1 { t.Errorf("len = %d, want 1", len(members)) }
}

// ============================================================
// API Key Repository Tests
// ============================================================

func TestMockAPIKeyRepo_CreateAndList(t *testing.T) {
	repo := NewMockAPIKeyRepo()
	ctx := context.Background()

	key := &domain.APIKey{ID: "key-1", UserID: "user-1", Key: "ak_test123", IsActive: true}
	if err := repo.Create(ctx, key); err != nil {
		t.Fatalf("Create: %v", err)
	}

	keys, err := repo.ListByUser(ctx, "user-1")
	if err != nil { t.Fatalf("ListByUser: %v", err) }
	if len(keys) != 1 { t.Errorf("len = %d, want 1", len(keys)) }

	if err := repo.Revoke(ctx, "key-1", "user-1"); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
}

// ============================================================
// Archive Repository Tests
// ============================================================

func TestMockArchiveRepo_FoldersAndMove(t *testing.T) {
	repo := NewMockArchiveRepo()
	ctx := context.Background()

	folder := &domain.ArchiveFolder{ID: "f-1", UserID: "user-1", Name: "Old Chats", Type: "chat"}
	if err := repo.CreateFolder(ctx, folder); err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}

	folders, err := repo.ListFolders(ctx, "user-1", "chat")
	if err != nil { t.Fatalf("ListFolders: %v", err) }
	if len(folders) != 1 { t.Errorf("len = %d, want 1", len(folders)) }

	if err := repo.MoveChat(ctx, "conv-1", "user-1", "f-1"); err != nil {
		t.Fatalf("MoveChat: %v", err)
	}
}

// ============================================================
// Subscription Repository Tests
// ============================================================

func TestMockSubscriptionRepo_CRUD(t *testing.T) {
	repo := NewMockSubscriptionRepo()
	ctx := context.Background()

	sub := &domain.Subscription{ID: "sub-1", UserID: "user-1", PlanID: "pro", Status: "active"}
	if err := repo.Create(ctx, sub); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetActive(ctx, "user-1")
	if err != nil { t.Fatalf("GetActive: %v", err) }
	if got == nil { t.Fatal("GetActive returned nil") }

	if err := repo.Cancel(ctx, "user-1"); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
}

// ============================================================
// Campaign Recipient Repository Tests
// ============================================================

func TestMockCampaignRecipientRepo_CRUD(t *testing.T) {
	repo := NewMockCampaignRecipientRepo()
	ctx := context.Background()

	cr := &domain.CampaignRecipient{ID: "cr-1", CampaignID: "camp-1", Phone: "+2348012345678", Status: "pending"}
	if err := repo.Create(ctx, cr); err != nil {
		t.Fatalf("Create: %v", err)
	}

	recipients, err := repo.ListByCampaign(ctx, "camp-1")
	if err != nil { t.Fatalf("ListByCampaign: %v", err) }
	if len(recipients) != 1 { t.Errorf("len = %d, want 1", len(recipients)) }

	if err := repo.UpdateStatus(ctx, "cr-1", "sent", nil); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
}

// ============================================================
// Unknown Question Repository Tests
// ============================================================

func TestMockUnknownQuestionRepo_CRUD(t *testing.T) {
	repo := NewMockUnknownQuestionRepo()
	ctx := context.Background()

	uq := &domain.UnknownQuestion{ID: "uq-1", UserID: "user-1", Question: "What is AI?", Status: "pending"}
	if err := repo.Create(ctx, uq); err != nil {
		t.Fatalf("Create: %v", err)
	}

	exists, err := repo.ExistsPending(ctx, "user-1", "what is ai?")
	if err != nil { t.Fatalf("ExistsPending: %v", err) }
	if !exists { t.Error("ExistsPending should return true") }

	uqs, err := repo.List(ctx, "user-1", "pending", 10, 0)
	if err != nil { t.Fatalf("List: %v", err) }
	if len(uqs) != 1 { t.Errorf("len = %d, want 1", len(uqs)) }

	counts, err := repo.CountByStatus(ctx, "user-1")
	if err != nil { t.Fatalf("CountByStatus: %v", err) }
	if counts["pending"] != 1 {
		t.Errorf("pending count = %d, want 1", counts["pending"])
	}
}
