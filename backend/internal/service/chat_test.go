package service

import (
	"context"
	"testing"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

func newTestChatService() *ChatService {
	repos := &repository.Repositories{
		Conversation: repository.NewMockConversationRepo(),
		Message:      repository.NewMockMessageRepo(),
		Integration:  repository.NewMockIntegrationRepo(),
		Handoff:      repository.NewMockHandoffRepo(),
		MediaMessage: repository.NewMockMediaMessageRepo(),
		User:         repository.NewMockUserRepo(),
	}
	return &ChatService{
		cfg:      &config.Config{},
		repos:    repos,
		redis:    nil,
		aiBrain:  nil,
		logger:   infrastructure.NewNullLogger(),
		openwa:   nil,
		telegram: nil,
		replies:  make(map[string]*replyGateState),
	}
}

// ============================================================
// TestNormalizeReplyKey
// ============================================================

func TestNormalizeReplyKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"whitespace only", "   ", ""},
		{"tabs and newlines", "\t\n  \r", ""},
		{"simple lowercase", "hello", "hello"},
		{"uppercase input", "HELLO", "hello"},
		{"mixed case", "HeLLo WoRLd", "hello world"},
		{"extra spaces", "  hello   world  ", "hello world"},
		{"tabs between words", "hello\tworld", "hello world"},
		{"single word", "Hi", "hi"},
		{"special characters", "hello!@#$%^&*()", "hello!@#$%^&*()"},
		{"numeric", "12345", "12345"},
		{"alphanumeric", "test123 msg456", "test123 msg456"},
		{"unicode", "Café Résumé", "café résumé"},
		{"leading trailing mixed whitespace", " \t hello \n ", "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeReplyKey(tt.input)
			if got != tt.want {
				t.Errorf("normalizeReplyKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// TestBeginAIReply
// ============================================================

func TestBeginAIReply_FirstCallAllowed(t *testing.T) {
	svc := newTestChatService()
	if !svc.beginAIReply("conv-1", "hello") {
		t.Fatal("first call should be allowed")
	}
	svc.completeAIReply("conv-1", "hello")
}

func TestBeginAIReply_EmptyMessageAlwaysAllowed(t *testing.T) {
	svc := newTestChatService()
	if !svc.beginAIReply("conv-1", "") {
		t.Fatal("empty message should always be allowed")
	}
	if !svc.beginAIReply("conv-1", "   ") {
		t.Fatal("whitespace-only message should always be allowed")
	}
}

func TestBeginAIReply_DuplicateSuppressionWhileInFlight(t *testing.T) {
	svc := newTestChatService()

	if !svc.beginAIReply("conv-1", "help me") {
		t.Fatal("first call should be allowed")
	}
	if svc.beginAIReply("conv-1", "help me") {
		t.Fatal("duplicate while in-flight should be blocked")
	}
	if svc.beginAIReply("conv-1", "Help  Me") {
		t.Fatal("case-insensitive duplicate while in-flight should be blocked")
	}
	svc.completeAIReply("conv-1", "help me")
}

func TestBeginAIReply_CooldownAfterCompletion(t *testing.T) {
	svc := newTestChatService()

	if !svc.beginAIReply("conv-1", "urgent") {
		t.Fatal("first call should be allowed")
	}
	svc.completeAIReply("conv-1", "urgent")

	if svc.beginAIReply("conv-1", "urgent") {
		t.Fatal("same message within cooldown should be blocked after completion")
	}
}

func TestBeginAIReply_DifferentMessagesAllowed(t *testing.T) {
	svc := newTestChatService()

	if !svc.beginAIReply("conv-1", "hello") {
		t.Fatal("first message should be allowed")
	}
	if !svc.beginAIReply("conv-1", "world") {
		t.Fatal("different message should be allowed even while another is in-flight")
	}
	svc.abortAIReply("conv-1")
	svc.completeAIReply("conv-1", "hello")
	svc.completeAIReply("conv-1", "world")
}

func TestBeginAIReply_DifferentConversationsIndependent(t *testing.T) {
	svc := newTestChatService()

	if !svc.beginAIReply("conv-1", "help") {
		t.Fatal("conv-1 first call should be allowed")
	}
	if !svc.beginAIReply("conv-2", "help") {
		t.Fatal("conv-2 same message should be allowed independently")
	}
	svc.abortAIReply("conv-1")
	svc.abortAIReply("conv-2")
	svc.completeAIReply("conv-1", "help")
	svc.completeAIReply("conv-2", "help")
}

func TestBeginAIReply_AfterCooldownExpires(t *testing.T) {
	svc := newTestChatService()

	svc.beginAIReply("conv-1", "test")
	svc.completeAIReply("conv-1", "test")

	svc.replyMu.Lock()
	if state, ok := svc.replies["conv-1"]; ok {
		state.lastReplyAt = time.Now().Add(-10 * time.Second)
	}
	svc.replyMu.Unlock()

	if !svc.beginAIReply("conv-1", "test") {
		t.Fatal("should be allowed after cooldown expires")
	}
	svc.completeAIReply("conv-1", "test")
}

func TestBeginAIReply_InFlightCooldownExpiry(t *testing.T) {
	svc := newTestChatService()

	svc.beginAIReply("conv-1", "msg")

	svc.replyMu.Lock()
	if state, ok := svc.replies["conv-1"]; ok {
		state.inFlightAt = time.Now().Add(-10 * time.Second)
	}
	svc.replyMu.Unlock()

	if !svc.beginAIReply("conv-1", "msg") {
		t.Fatal("should be allowed after in-flight cooldown expires")
	}
	svc.completeAIReply("conv-1", "msg")
}

// ============================================================
// TestCompleteAIReply
// ============================================================

func TestCompleteAIReply_ResetsInFlightKey(t *testing.T) {
	svc := newTestChatService()

	svc.beginAIReply("conv-1", "question")

	svc.replyMu.Lock()
	if svc.replies["conv-1"].inFlightKey == "" {
		svc.replyMu.Unlock()
		t.Fatal("inFlightKey should be set after beginAIReply")
	}
	svc.replyMu.Unlock()

	svc.completeAIReply("conv-1", "question")

	svc.replyMu.Lock()
	state := svc.replies["conv-1"]
	if state.inFlightKey != "" {
		svc.replyMu.Unlock()
		t.Fatalf("inFlightKey should be empty after completeAIReply, got %q", state.inFlightKey)
	}
	if state.lastKey == "" {
		svc.replyMu.Unlock()
		t.Fatal("lastKey should be set after completeAIReply")
	}
	svc.replyMu.Unlock()
}

func TestCompleteAIReply_CreatesStateIfMissing(t *testing.T) {
	svc := newTestChatService()

	svc.completeAIReply("new-conv", "hello")

	svc.replyMu.Lock()
	defer svc.replyMu.Unlock()
	state, ok := svc.replies["new-conv"]
	if !ok {
		t.Fatal("completeAIReply should create state if none exists")
	}
	if state.lastKey != "hello" {
		t.Errorf("lastKey = %q, want %q", state.lastKey, "hello")
	}
}

func TestCompleteAIReply_NormalizesKey(t *testing.T) {
	svc := newTestChatService()

	svc.beginAIReply("conv-1", "HELLO WORLD")
	svc.completeAIReply("conv-1", "HELLO  WORLD")

	svc.replyMu.Lock()
	defer svc.replyMu.Unlock()
	if svc.replies["conv-1"].lastKey != "hello world" {
		t.Errorf("lastKey = %q, want %q", svc.replies["conv-1"].lastKey, "hello world")
	}
}

func TestCompleteAIReply_EnablesSameMessageAfterCooldownReset(t *testing.T) {
	svc := newTestChatService()

	svc.beginAIReply("conv-1", "msg")
	svc.completeAIReply("conv-1", "msg")

	svc.replyMu.Lock()
	svc.replies["conv-1"].lastReplyAt = time.Now().Add(-10 * time.Second)
	svc.replyMu.Unlock()

	if !svc.beginAIReply("conv-1", "msg") {
		t.Fatal("should allow same message after complete + cooldown bypass")
	}
	svc.completeAIReply("conv-1", "msg")
}

// ============================================================
// TestAbortAIReply
// ============================================================

func TestAbortAIReply_ClearsInFlightKey(t *testing.T) {
	svc := newTestChatService()

	svc.beginAIReply("conv-1", "message")

	svc.replyMu.Lock()
	if svc.replies["conv-1"].inFlightKey == "" {
		svc.replyMu.Unlock()
		t.Fatal("inFlightKey should be set")
	}
	svc.replyMu.Unlock()

	svc.abortAIReply("conv-1")

	svc.replyMu.Lock()
	if svc.replies["conv-1"].inFlightKey != "" {
		svc.replyMu.Unlock()
		t.Fatal("inFlightKey should be cleared after abort")
	}
	svc.replyMu.Unlock()
}

func TestAbortAIReply_NoopIfNoState(t *testing.T) {
	svc := newTestChatService()
	svc.abortAIReply("nonexistent")
}

func TestAbortAIReply_AllowsNewReply(t *testing.T) {
	svc := newTestChatService()

	svc.beginAIReply("conv-1", "hello")
	svc.abortAIReply("conv-1")

	if !svc.beginAIReply("conv-1", "hello") {
		t.Fatal("should be allowed after abort clears in-flight state")
	}
	svc.completeAIReply("conv-1", "hello")
}

// ============================================================
// TestIsAllDigits
// ============================================================

func TestIsAllDigits(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"12345", true},
		{"0", true},
		{"123a5", false},
		{"abc", false},
		{"12 34", false},
		{"+12345", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isAllDigits(tt.input)
			if got != tt.want {
				t.Errorf("isAllDigits(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// TestCleanPhoneNumber
// ============================================================

func TestCleanPhoneNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"digits only", "1234567890", "1234567890"},
		{"with dashes", "123-456-7890", "1234567890"},
		{"with plus prefix", "+1234567890", "1234567890"},
		{"with spaces", "123 456 7890", "1234567890"},
		{"with parentheses", "(123) 456-7890", "1234567890"},
		{"empty string", "", ""},
		{"no digits", "abc-xyz", ""},
		{"mixed special", "+1 (234) 567-8901", "12345678901"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanPhoneNumber(tt.input)
			if got != tt.want {
				t.Errorf("CleanPhoneNumber(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// TestCleanWhatsAppID
// ============================================================

func TestCleanWhatsAppIDTable(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain number", "1234567890", "1234567890"},
		{"waid prefix", "waid:1234567890", "1234567890"},
		{"phone with dashes", "123-456-7890", "1234567890"},
		{"empty string", "", ""},
		{"waid prefix with special chars", "waid:+1 (234) 567-8901", "12345678901"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanWhatsAppID(tt.input)
			if got != tt.want {
				t.Errorf("cleanWhatsAppID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// TestFirstNonEmpty
// ============================================================

func TestFirstNonEmptyTable(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{"first non-empty", []string{"", "hello", "world"}, "hello"},
		{"first is non-empty", []string{"hello", "world"}, "hello"},
		{"all empty", []string{"", "", ""}, ""},
		{"single value", []string{"only"}, "only"},
		{"empty slice", []string{}, ""},
		{"whitespace only", []string{"  ", "", "real"}, "real"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstNonEmpty(tt.values...)
			if got != tt.want {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}

// ============================================================
// TestFormatChatID
// ============================================================

func TestFormatChatID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1234567890", "1234567890@s.whatsapp.net"},
		{"+1 (234) 567-8901", "12345678901@s.whatsapp.net"},
		{"", "@s.whatsapp.net"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FormatChatID(tt.input)
			if got != tt.want {
				t.Errorf("FormatChatID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// TestFormatContactID
// ============================================================

func TestFormatContactID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1234567890", "1234567890@c.us"},
		{"+1234567890", "1234567890@c.us"},
		{"", "@c.us"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FormatContactID(tt.input)
			if got != tt.want {
				t.Errorf("FormatContactID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// TestListConversations
// ============================================================

func TestListConversations_Success(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		err := svc.repos.Conversation.Create(ctx, &domain.Conversation{
			ID:           "conv-" + string(rune('0'+i)),
			UserID:       "user-1",
			CustomerName: "Customer " + string(rune('0'+i)),
			Channel:      "web",
			Status:       "active",
		})
		if err != nil {
			t.Fatalf("failed to create conversation %d: %v", i, err)
		}
	}

	convs, total, err := svc.ListConversations(ctx, "user-1", "", 1, 10)
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(convs) != 3 {
		t.Errorf("returned %d conversations, want 3", len(convs))
	}
	for _, c := range convs {
		if c.UserID != "user-1" {
			t.Errorf("conversation UserID = %q, want %q", c.UserID, "user-1")
		}
	}
}

func TestListConversations_Empty(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	convs, total, err := svc.ListConversations(ctx, "user-1", "", 1, 10)
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(convs) != 0 {
		t.Errorf("returned %d conversations, want 0", len(convs))
	}
}

func TestListConversations_FilterByStatus(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c2", UserID: "user-1", Channel: "web", Status: "escalated",
	})
	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c3", UserID: "user-1", Channel: "web", Status: "active",
	})

	convs, total, err := svc.ListConversations(ctx, "user-1", "active", 1, 10)
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(convs) != 2 {
		t.Errorf("returned %d conversations, want 2", len(convs))
	}
}

func TestListConversations_Pagination(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.repos.Conversation.Create(ctx, &domain.Conversation{
			UserID: "user-1", Channel: "web", Status: "active",
		})
	}

	convs, total, err := svc.ListConversations(ctx, "user-1", "", 1, 2)
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(convs) != 2 {
		t.Errorf("page 1 returned %d conversations, want 2", len(convs))
	}

	convs2, total2, err := svc.ListConversations(ctx, "user-1", "", 2, 2)
	if err != nil {
		t.Fatalf("ListConversations page 2 returned error: %v", err)
	}
	if total2 != 5 {
		t.Errorf("total2 = %d, want 5", total2)
	}
	if len(convs2) != 2 {
		t.Errorf("page 2 returned %d conversations, want 2", len(convs2))
	}

	convs3, _, err := svc.ListConversations(ctx, "user-1", "", 3, 2)
	if err != nil {
		t.Fatalf("ListConversations page 3 returned error: %v", err)
	}
	if len(convs3) != 1 {
		t.Errorf("page 3 returned %d conversations, want 1", len(convs3))
	}
}

func TestListConversations_PopulatesLastMessage(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "Hello there!",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "ai", Content: "Hi! How can I help?",
	})

	convs, _, err := svc.ListConversations(ctx, "user-1", "", 1, 10)
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(convs))
	}
	if convs[0].LastMessage != "Hi! How can I help?" {
		t.Errorf("LastMessage = %q, want %q", convs[0].LastMessage, "Hi! How can I help?")
	}
}

func TestListConversations_PopulatesUnreadCount(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "msg1", IsRead: false,
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "msg2", IsRead: false,
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "msg3", IsRead: true,
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "ai", Content: "ai msg", IsRead: false,
	})

	convs, _, err := svc.ListConversations(ctx, "user-1", "", 1, 10)
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(convs))
	}
	if convs[0].Unread != 2 {
		t.Errorf("Unread = %d, want 2 (only unread customer messages)", convs[0].Unread)
	}
}

func TestListConversations_OtherUserIsolation(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c2", UserID: "user-2", Channel: "web", Status: "active",
	})

	convs, total, err := svc.ListConversations(ctx, "user-1", "", 1, 10)
	if err != nil {
		t.Fatalf("ListConversations returned error: %v", err)
	}
	if total != 1 {
		t.Errorf("total = %d, want 1", total)
	}
	if len(convs) != 1 {
		t.Errorf("returned %d conversations, want 1", len(convs))
	}
}

// ============================================================
// TestGetConversation (via GetConversation / GetConversationMessages)
// ============================================================

func TestGetConversationMessages_Success(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "Hello",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "ai", Content: "Hi there!",
	})

	conv, messages, err := svc.GetConversation(ctx, "user-1", "c1")
	if err != nil {
		t.Fatalf("GetConversation returned error: %v", err)
	}
	if conv == nil {
		t.Fatal("conversation should not be nil")
	}
	if conv.ID != "c1" {
		t.Errorf("conv.ID = %q, want %q", conv.ID, "c1")
	}
	if len(messages) != 2 {
		t.Errorf("returned %d messages, want 2", len(messages))
	}
}

func TestGetConversationMessages_NotFound(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	_, _, err := svc.GetConversation(ctx, "user-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent conversation")
	}
}

func TestGetConversationMessages_WrongUser(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})

	_, _, err := svc.GetConversation(ctx, "user-2", "c1")
	if err == nil {
		t.Fatal("expected error when wrong user accesses conversation")
	}
}

func TestGetConversationMessages_Empty(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})

	_, messages, err := svc.GetConversation(ctx, "user-1", "c1")
	if err != nil {
		t.Fatalf("GetConversation returned error: %v", err)
	}
	if len(messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(messages))
	}
}

func TestGetConversation_MarksMessagesAsRead(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "unread msg", IsRead: false,
	})

	_, _, err := svc.GetConversation(ctx, "user-1", "c1")
	if err != nil {
		t.Fatalf("GetConversation returned error: %v", err)
	}

	messages, _ := svc.repos.Message.ListByConversation(ctx, "c1", 100)
	for _, m := range messages {
		if !m.IsRead {
			t.Errorf("message %q should be marked as read after GetConversation", m.ID)
		}
	}
}

// ============================================================
// TestGetConversationPaginated
// ============================================================

func TestGetConversationPaginated_Success(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	for i := 0; i < 5; i++ {
		svc.repos.Message.Create(ctx, &domain.Message{
			ConversationID: "c1", Role: "customer", Content: "msg",
		})
	}

	conv, messages, total, err := svc.GetConversationPaginated(ctx, "user-1", "c1", 2, 0)
	if err != nil {
		t.Fatalf("GetConversationPaginated returned error: %v", err)
	}
	if conv == nil {
		t.Fatal("conversation should not be nil")
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(messages) != 2 {
		t.Errorf("returned %d messages, want 2", len(messages))
	}
}

func TestGetConversationPaginated_NotFound(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	_, _, _, err := svc.GetConversationPaginated(ctx, "user-1", "nonexistent", 10, 0)
	if err == nil {
		t.Fatal("expected error for nonexistent conversation")
	}
}

func TestGetConversationPaginated_MarksAsRead(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "unread", IsRead: false,
	})

	_, _, _, err := svc.GetConversationPaginated(ctx, "user-1", "c1", 10, 0)
	if err != nil {
		t.Fatalf("GetConversationPaginated returned error: %v", err)
	}

	msgs, _ := svc.repos.Message.ListByConversation(ctx, "c1", 100)
	for _, m := range msgs {
		if !m.IsRead {
			t.Errorf("message should be read after GetConversationPaginated")
		}
	}
}

// ============================================================
// TestMarkRead
// ============================================================

func TestMarkRead_MarksAllUnread(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "msg1", IsRead: false,
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "msg2", IsRead: false,
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "ai", Content: "reply", IsRead: false,
	})

	err := svc.repos.Message.MarkRead(ctx, "c1")
	if err != nil {
		t.Fatalf("MarkRead returned error: %v", err)
	}

	messages, _ := svc.repos.Message.ListByConversation(ctx, "c1", 100)
	for _, m := range messages {
		if !m.IsRead {
			t.Errorf("message %q should be read after MarkRead", m.ID)
		}
	}
}

func TestMarkRead_AlreadyReadUnchanged(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "already read", IsRead: true,
	})

	err := svc.repos.Message.MarkRead(ctx, "c1")
	if err != nil {
		t.Fatalf("MarkRead returned error: %v", err)
	}

	messages, _ := svc.repos.Message.ListByConversation(ctx, "c1", 100)
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if !messages[0].IsRead {
		t.Error("message that was already read should remain read")
	}
}

func TestMarkRead_OtherConversationsUnaffected(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c2", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "msg1", IsRead: false,
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c2", Role: "customer", Content: "msg2", IsRead: false,
	})

	svc.repos.Message.MarkRead(ctx, "c1")

	msgs1, _ := svc.repos.Message.ListByConversation(ctx, "c1", 100)
	msgs2, _ := svc.repos.Message.ListByConversation(ctx, "c2", 100)

	if len(msgs1) != 1 || !msgs1[0].IsRead {
		t.Error("c1 message should be read")
	}
	if len(msgs2) != 1 || msgs2[0].IsRead {
		t.Error("c2 message should NOT be affected by MarkRead on c1")
	}
}

func TestMarkRead_CountUnreadReflectsMarkRead(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "msg1", IsRead: false,
	})
	svc.repos.Message.Create(ctx, &domain.Message{
		ConversationID: "c1", Role: "customer", Content: "msg2", IsRead: false,
	})

	count, _ := svc.repos.Message.CountUnread(ctx, "c1")
	if count != 2 {
		t.Fatalf("CountUnread = %d, want 2 before MarkRead", count)
	}

	svc.repos.Message.MarkRead(ctx, "c1")

	count, _ = svc.repos.Message.CountUnread(ctx, "c1")
	if count != 0 {
		t.Errorf("CountUnread = %d, want 0 after MarkRead", count)
	}
}

// ============================================================
// TestGetConversationOnly
// ============================================================

func TestGetConversationOnly_Success(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", CustomerName: "Alice", Channel: "web", Status: "active",
	})

	conv, err := svc.GetConversationOnly(ctx, "c1", "user-1")
	if err != nil {
		t.Fatalf("GetConversationOnly returned error: %v", err)
	}
	if conv == nil {
		t.Fatal("conversation should not be nil")
	}
	if conv.CustomerName != "Alice" {
		t.Errorf("CustomerName = %q, want %q", conv.CustomerName, "Alice")
	}
}

func TestGetConversationOnly_WrongUser(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})

	conv, err := svc.GetConversationOnly(ctx, "c1", "user-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv != nil {
		t.Fatal("should return nil for wrong user")
	}
}

// ============================================================
// TestEnsureConversation
// ============================================================

func TestEnsureConversation_CreatesNew(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	conv, err := svc.EnsureConversation(ctx, "user-1", "Bob", "12345", "web", "")
	if err != nil {
		t.Fatalf("EnsureConversation returned error: %v", err)
	}
	if conv == nil {
		t.Fatal("conversation should not be nil")
	}
	if conv.CustomerName != "Bob" {
		t.Errorf("CustomerName = %q, want %q", conv.CustomerName, "Bob")
	}
	if conv.Channel != "web" {
		t.Errorf("Channel = %q, want %q", conv.Channel, "web")
	}
	if conv.Status != "active" {
		t.Errorf("Status = %q, want %q", conv.Status, "active")
	}
}

func TestEnsureConversation_ReturnsExisting(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	conv1, err := svc.EnsureConversation(ctx, "user-1", "Bob", "12345", "web", "")
	if err != nil {
		t.Fatalf("first EnsureConversation returned error: %v", err)
	}

	conv2, err := svc.EnsureConversation(ctx, "user-1", "Bob", "12345", "web", "")
	if err != nil {
		t.Fatalf("second EnsureConversation returned error: %v", err)
	}

	if conv1.ID != conv2.ID {
		t.Errorf("expected same conversation ID, got %q and %q", conv1.ID, conv2.ID)
	}
}

func TestEnsureConversation_DifferentChannels(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	conv1, err := svc.EnsureConversation(ctx, "user-1", "Bob", "12345", "web", "")
	if err != nil {
		t.Fatalf("first EnsureConversation returned error: %v", err)
	}

	conv2, err := svc.EnsureConversation(ctx, "user-1", "Bob", "12345", "whatsapp", "")
	if err != nil {
		t.Fatalf("second EnsureConversation returned error: %v", err)
	}

	if conv1.ID == conv2.ID {
		t.Error("different channels should produce different conversations")
	}
}

// ============================================================
// TestHumanTakeover
// ============================================================

func TestHumanTakeover_Success(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})

	err := svc.HumanTakeover(ctx, "user-1", "c1", "agent-1")
	if err != nil {
		t.Fatalf("HumanTakeover returned error: %v", err)
	}

	conv, err := svc.repos.Conversation.GetByID(ctx, "c1")
	if err != nil || conv == nil {
		t.Fatal("conversation should still exist")
	}
}

func TestHumanTakeover_NotFound(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	err := svc.HumanTakeover(ctx, "user-1", "nonexistent", "agent-1")
	if err == nil {
		t.Fatal("expected error for nonexistent conversation")
	}
}

func TestHumanTakeover_WrongUser(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})

	err := svc.HumanTakeover(ctx, "user-2", "c1", "agent-1")
	if err == nil {
		t.Fatal("expected error for wrong user")
	}
}

// ============================================================
// TestEscalate
// ============================================================

func TestEscalate_Success(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})

	err := svc.Escalate(ctx, "user-1", "c1", "complex question")
	if err != nil {
		t.Fatalf("Escalate returned error: %v", err)
	}

	messages, _ := svc.repos.Message.ListByConversation(ctx, "c1", 100)
	if len(messages) != 1 {
		t.Fatalf("expected 1 system message, got %d", len(messages))
	}
	if messages[0].Role != "system" {
		t.Errorf("message Role = %q, want %q", messages[0].Role, "system")
	}
	if messages[0].Content != "Conversation escalated. Reason: complex question" {
		t.Errorf("message Content = %q, want %q", messages[0].Content, "Conversation escalated. Reason: complex question")
	}
}

func TestEscalate_NotFound(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	err := svc.Escalate(ctx, "user-1", "nonexistent", "reason")
	if err == nil {
		t.Fatal("expected error for nonexistent conversation")
	}
}

// ============================================================
// TestClearChats
// ============================================================

func TestClearChats_RemovesAllConversations(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c2", UserID: "user-1", Channel: "web", Status: "active",
	})

	err := svc.ClearChats(ctx, "user-1")
	if err != nil {
		t.Fatalf("ClearChats returned error: %v", err)
	}

	convs, total, _ := svc.repos.Conversation.List(ctx, "user-1", "", 100, 0)
	if total != 0 || len(convs) != 0 {
		t.Errorf("expected 0 conversations after ClearChats, got %d", total)
	}
}

func TestClearChats_OtherUserUnaffected(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})
	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c2", UserID: "user-2", Channel: "web", Status: "active",
	})

	svc.ClearChats(ctx, "user-1")

	conv, _ := svc.repos.Conversation.GetByID(ctx, "c2")
	if conv == nil {
		t.Fatal("user-2 conversation should still exist")
	}
}

// ============================================================
// TestStoreWhatsAppIntegration
// ============================================================

func TestStoreWhatsAppIntegration_CreatesNew(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.StoreWhatsAppIntegration(ctx, "user-1", "session-123", "+1234567890")

	integration, err := svc.repos.Integration.GetByOrgAndChannel(ctx, "user-1", "whatsapp")
	if err != nil {
		t.Fatalf("GetByOrgAndChannel returned error: %v", err)
	}
	if integration == nil {
		t.Fatal("integration should exist after StoreWhatsAppIntegration")
	}
	if integration.Status != "connected" {
		t.Errorf("Status = %q, want %q", integration.Status, "connected")
	}
}

func TestStoreWhatsAppIntegration_WithStatus(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.StoreWhatsAppIntegrationWithStatus(ctx, "user-1", "session-123", "+1234567890", "connecting")

	integration, _ := svc.repos.Integration.GetByOrgAndChannel(ctx, "user-1", "whatsapp")
	if integration == nil {
		t.Fatal("integration should exist")
	}
	if integration.Status != "connecting" {
		t.Errorf("Status = %q, want %q", integration.Status, "connecting")
	}
}

func TestStoreWhatsAppIntegration_UpdatesExisting(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.StoreWhatsAppIntegration(ctx, "user-1", "session-old", "+1234567890")
	svc.StoreWhatsAppIntegration(ctx, "user-1", "session-new", "+1234567890")

	integration, _ := svc.repos.Integration.GetByOrgAndChannel(ctx, "user-1", "whatsapp")
	if integration == nil {
		t.Fatal("integration should exist")
	}
	if integration.Config["session_id"] != "session-new" {
		t.Errorf("session_id = %v, want %q", integration.Config["session_id"], "session-new")
	}
}

// ============================================================
// TestGetWhatsAppIntegration
// ============================================================

func TestGetWhatsAppIntegration_Connected(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.StoreWhatsAppIntegration(ctx, "user-1", "session-123", "+1234567890")

	integration, err := svc.GetWhatsAppIntegration(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetWhatsAppIntegration returned error: %v", err)
	}
	if integration == nil {
		t.Fatal("integration should exist")
	}
}

func TestGetWhatsAppIntegration_ErrorStatus(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.StoreWhatsAppIntegrationWithStatus(ctx, "user-1", "session-123", "+1234567890", "error")

	integration, err := svc.GetWhatsAppIntegration(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integration != nil {
		t.Fatal("integration with error status should not be returned")
	}
}

func TestGetWhatsAppIntegration_DisconnectedStatus(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.StoreWhatsAppIntegrationWithStatus(ctx, "user-1", "session-123", "+1234567890", "disconnected")

	integration, err := svc.GetWhatsAppIntegration(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integration != nil {
		t.Fatal("integration with disconnected status should not be returned")
	}
}

func TestGetWhatsAppIntegration_InactiveStatus(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.StoreWhatsAppIntegrationWithStatus(ctx, "user-1", "session-123", "+1234567890", "inactive")

	integration, err := svc.GetWhatsAppIntegration(ctx, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integration != nil {
		t.Fatal("integration with inactive status should not be returned")
	}
}

func TestGetWhatsAppIntegration_NotFound(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	integration, err := svc.GetWhatsAppIntegration(ctx, "user-unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integration != nil {
		t.Fatal("should return nil for nonexistent integration")
	}
}

// ============================================================
// TestRemoveWhatsAppIntegration
// ============================================================

func TestRemoveWhatsAppIntegration(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.StoreWhatsAppIntegration(ctx, "user-1", "session-123", "+1234567890")
	svc.RemoveWhatsAppIntegration(ctx, "user-1")

	integration, _ := svc.repos.Integration.GetByOrgAndChannel(ctx, "user-1", "whatsapp")
	if integration != nil && integration.Status != "inactive" {
		t.Errorf("integration status = %q, want %q after RemoveWhatsAppIntegration", integration.Status, "inactive")
	}
}

// ============================================================
// TestResolveWhatsAppIdentity (basic fallback without openwa)
// ============================================================

func TestResolveWhatsAppIdentity_NilMessage(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	_, err := svc.ResolveWhatsAppIdentity(ctx, "user-1", "session-1", nil)
	if err == nil {
		t.Fatal("expected error for nil message")
	}
}

func TestResolveWhatsAppIdentity_FromSenderPayload(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	msg := &OpenWAMessageData{
		From: "1234567890@s.whatsapp.net",
		Sender: OpenWASender{
			ID:            "1234567890@c.us",
			Pushname:      "Alice",
			FormattedName: "Alice Smith",
			ProfilePicThumbObj: OpenWAProfilePic{
				Eurl: "https://example.com/pic.jpg",
			},
		},
	}

	identity, err := svc.ResolveWhatsAppIdentity(ctx, "user-1", "session-1", msg)
	if err != nil {
		t.Fatalf("ResolveWhatsAppIdentity returned error: %v", err)
	}
	if identity.Name != "Alice" {
		t.Errorf("Name = %q, want %q (Pushname should be preferred)", identity.Name, "Alice")
	}
	if identity.Phone != "1234567890" {
		t.Errorf("Phone = %q, want %q", identity.Phone, "1234567890")
	}
	if identity.Avatar != "https://example.com/pic.jpg" {
		t.Errorf("Avatar = %q, want %q", identity.Avatar, "https://example.com/pic.jpg")
	}
}

func TestResolveWhatsAppIdentity_SenderIDFallback(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	msg := &OpenWAMessageData{
		From:   "",
		Sender: OpenWASender{ID: "9876543210@c.us"},
	}

	identity, err := svc.ResolveWhatsAppIdentity(ctx, "user-1", "session-1", msg)
	if err != nil {
		t.Fatalf("ResolveWhatsAppIdentity returned error: %v", err)
	}
	if identity.Phone != "9876543210" {
		t.Errorf("Phone = %q, want %q", identity.Phone, "9876543210")
	}
	if identity.Name != "9876543210" {
		t.Errorf("Name = %q, want %q (should fallback to phone)", identity.Name, "9876543210")
	}
}

func TestResolveWhatsAppIdentity_FallbackToWhatsAppUser(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	msg := &OpenWAMessageData{
		From:   "",
		Sender: OpenWASender{},
	}

	identity, err := svc.ResolveWhatsAppIdentity(ctx, "user-1", "session-1", msg)
	if err != nil {
		t.Fatalf("ResolveWhatsAppIdentity returned error: %v", err)
	}
	if identity.Name != "WhatsApp User" {
		t.Errorf("Name = %q, want %q (final fallback)", identity.Name, "WhatsApp User")
	}
}

func TestResolveWhatsAppIdentity_ExistingConversationFallback(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID:           "c1",
		UserID:       "user-1",
		CustomerName: "Returning Bob",
		CustomerPhone: "5551234567",
		Channel:      "whatsapp",
		Status:       "active",
	})

	msg := &OpenWAMessageData{
		From: "5551234567@s.whatsapp.net",
		Sender: OpenWASender{
			Pushname: "",
		},
	}

	identity, err := svc.ResolveWhatsAppIdentity(ctx, "user-1", "session-1", msg)
	if err != nil {
		t.Fatalf("ResolveWhatsAppIdentity returned error: %v", err)
	}
	if identity.Name != "Returning Bob" {
		t.Errorf("Name = %q, want %q (should use existing conversation name)", identity.Name, "Returning Bob")
	}
}

func TestResolveWhatsAppIdentity_PushnameHierarchy(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	tests := []struct {
		name      string
		sender    OpenWASender
		wantName  string
	}{
		{
			"pushname preferred",
			OpenWASender{Pushname: "PN", Name: "N", FormattedName: "FN", ShortName: "SN"},
			"PN",
		},
		{
			"name if no pushname",
			OpenWASender{Name: "N", FormattedName: "FN", ShortName: "SN"},
			"N",
		},
		{
			"formatted name if no pushname or name",
			OpenWASender{FormattedName: "FN", ShortName: "SN"},
			"FN",
		},
		{
			"short name as last resort",
			OpenWASender{ShortName: "SN"},
			"SN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &OpenWAMessageData{
				From:   "1112223333@s.whatsapp.net",
				Sender: tt.sender,
			}
			identity, err := svc.ResolveWhatsAppIdentity(ctx, "user-1", "session-1", msg)
			if err != nil {
				t.Fatalf("ResolveWhatsAppIdentity returned error: %v", err)
			}
			if identity.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", identity.Name, tt.wantName)
			}
		})
	}
}

func TestResolveWhatsAppIdentity_MethodsTracked(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	msg := &OpenWAMessageData{
		From: "1234567890@s.whatsapp.net",
		Sender: OpenWASender{
			ID:       "1234567890@c.us",
			Pushname: "Alice",
		},
	}

	identity, err := svc.ResolveWhatsAppIdentity(ctx, "user-1", "session-1", msg)
	if err != nil {
		t.Fatalf("ResolveWhatsAppIdentity returned error: %v", err)
	}
	if len(identity.Methods) == 0 {
		t.Fatal("Methods should be non-empty")
	}
	hasSenderPayload := false
	for _, m := range identity.Methods {
		if m == "sender_payload" {
			hasSenderPayload = true
		}
	}
	if !hasSenderPayload {
		t.Error("Methods should contain 'sender_payload'")
	}
}

// ============================================================
// TestSendMessage (basic, no external services)
// ============================================================

func TestSendMessage_AgentReply(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})

	msg, err := svc.SendMessage(ctx, "user-1", "c1", "agent", "Hello customer!")
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if msg == nil {
		t.Fatal("message should not be nil")
	}
	if msg.Role != "agent" {
		t.Errorf("Role = %q, want %q", msg.Role, "agent")
	}
	if msg.Content != "Hello customer!" {
		t.Errorf("Content = %q, want %q", msg.Content, "Hello customer!")
	}
	if !msg.IsRead {
		t.Error("agent messages should be read by default")
	}
}

func TestSendMessage_CustomerFromDashboardTreatedAsAgent(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", CustomerName: "Real Customer", Channel: "web", Status: "active",
	})

	msg, err := svc.SendMessage(ctx, "user-1", "c1", "customer", "reply from dashboard")
	if err != nil {
		t.Fatalf("SendMessage returned error: %v", err)
	}
	if msg.Role != "agent" {
		t.Errorf("Role = %q, want %q (dashboard customer should be treated as agent)", msg.Role, "agent")
	}
}

func TestSendMessage_NotFound(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	_, err := svc.SendMessage(ctx, "user-1", "nonexistent", "agent", "msg")
	if err == nil {
		t.Fatal("expected error for nonexistent conversation")
	}
}

// ============================================================

// TestGetMediaByConversation

// ============================================================

func TestGetMediaByConversation_Success(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})

	media, err := svc.GetMediaByConversation(ctx, "c1", "user-1")
	if err != nil {
		t.Fatalf("GetMediaByConversation returned error: %v", err)
	}
	if len(media) != 0 {
		t.Errorf("expected 0 media, got %d", len(media))
	}
}

func TestGetMediaByConversation_WrongUser(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	svc.repos.Conversation.Create(ctx, &domain.Conversation{
		ID: "c1", UserID: "user-1", Channel: "web", Status: "active",
	})

	_, err := svc.GetMediaByConversation(ctx, "c1", "user-2")
	if err == nil {
		t.Fatal("expected error for wrong user")
	}
}

func TestGetMediaByConversation_NotFound(t *testing.T) {
	svc := newTestChatService()
	ctx := context.Background()

	_, err := svc.GetMediaByConversation(ctx, "nonexistent", "user-1")
	if err == nil {
		t.Fatal("expected error for nonexistent conversation")
	}
}
