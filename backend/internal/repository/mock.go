package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"noant/internal/domain"
)

// ============================================================
// MockUserRepo
// ============================================================

type MockUserRepo struct {
	mu    sync.Mutex
	users map[string]*domain.User
	notifs map[string]*NotifPrefs
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		users:  make(map[string]*domain.User),
		notifs: make(map[string]*NotifPrefs),
	}
}

func (m *MockUserRepo) Create(ctx context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user.ID == "" {
		user.ID = generateUUID()
	}
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now
	cp := *user
	m.users[user.ID] = &cp
	return nil
}

func (m *MockUserRepo) RunInTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return fn(nil)
}

func (m *MockUserRepo) CreateTx(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	return m.Create(ctx, user)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (m *MockUserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil
	}
	now := time.Now()
	u.LastLoginAt = &now
	return nil
}

func (m *MockUserRepo) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil
	}
	u.Password = hashedPassword
	u.MustChangePassword = false
	return nil
}

func (m *MockUserRepo) UpdatePlan(ctx context.Context, userID, planID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return nil
	}
	u.PlanID = planID
	return nil
}

func (m *MockUserRepo) UpdateVerificationStatus(ctx context.Context, id string, verified bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil
	}
	u.IsVerified = verified
	u.VerificationCode = nil
	return nil
}

func (m *MockUserRepo) UpdateVerificationCode(ctx context.Context, id, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil
	}
	u.VerificationCode = &code
	return nil
}

func (m *MockUserRepo) GetOwnerWhatsApp(ctx context.Context, userID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return "", nil
	}
	if u.OwnerWhatsapp != nil {
		return *u.OwnerWhatsapp, nil
	}
	return "", nil
}

func (m *MockUserRepo) CleanupExpiredTrials(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	cutoff := time.Now().AddDate(0, 0, -days)
	for _, u := range m.users {
		if u.PlanID == "free" && u.TrialExpiresAt != nil && u.TrialExpiresAt.Before(cutoff) && u.IsActive {
			u.IsActive = false
			count++
		}
	}
	return count, nil
}

func (m *MockUserRepo) GetNotifPrefs(ctx context.Context, userID string) (*NotifPrefs, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.notifs[userID]
	if !ok {
		return &NotifPrefs{Escalation: true, UnknownQs: true, Payment: true, Security: true, TeamInvite: true, LanguagePref: "en"}, nil
	}
	cp := *p
	return &cp, nil
}

func (m *MockUserRepo) UpdateNotifPrefs(ctx context.Context, userID string, prefs *NotifPrefs) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *prefs
	m.notifs[userID] = &cp
	return nil
}

func (m *MockUserRepo) Delete(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, userID)
	delete(m.notifs, userID)
	return nil
}

func (m *MockUserRepo) ExportUserData(ctx context.Context, userID string) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	cp := *u
	return map[string]interface{}{
		"user":        &cp,
		"exported_at": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (m *MockUserRepo) UpdateProfile(ctx context.Context, userID, firstName, lastName, companyName, phone string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return nil
	}
	u.FirstName = firstName
	u.LastName = lastName
	u.CompanyName = companyName
	u.Phone = phone
	u.UpdatedAt = time.Now()
	return nil
}

func (m *MockUserRepo) GetOnboardingStatus(ctx context.Context, userID string) (*string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return nil, nil
	}
	return u.OnboardingStatus, nil
}

func (m *MockUserRepo) UpdateOnboardingStatus(ctx context.Context, userID, status string, industry *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return nil
	}
	u.OnboardingStatus = &status
	if industry != nil {
		u.Industry = industry
	}
	u.UpdatedAt = time.Now()
	return nil
}

// ============================================================
// MockConversationRepo
// ============================================================

type MockConversationRepo struct {
	mu            sync.Mutex
	conversations map[string]*domain.Conversation
	messages      []*domain.Message
	csatRatings   []csatRecord
}

type csatRecord struct {
	userID         string
	conversationID string
	score          int
	comment        *string
	createdAt      time.Time
}

func NewMockConversationRepo() *MockConversationRepo {
	return &MockConversationRepo{
		conversations: make(map[string]*domain.Conversation),
	}
}

func (m *MockConversationRepo) GetByID(ctx context.Context, id string) (*domain.Conversation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.conversations[id]
	if !ok {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

func (m *MockConversationRepo) Create(ctx context.Context, conv *domain.Conversation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if conv.ID == "" {
		conv.ID = generateUUID()
	}
	now := time.Now()
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = now
	}
	conv.UpdatedAt = now
	cp := *conv
	m.conversations[conv.ID] = &cp
	return nil
}

func (m *MockConversationRepo) List(ctx context.Context, userID, status string, limit, offset int) ([]domain.Conversation, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []domain.Conversation
	for _, c := range m.conversations {
		if c.UserID != userID {
			continue
		}
		if status != "" && c.Status != status {
			continue
		}
		cp := *c
		all = append(all, cp)
	}
	total := len(all)
	if offset >= total {
		return []domain.Conversation{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (m *MockConversationRepo) GetByIDAndUser(ctx context.Context, id, userID string) (*domain.Conversation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.conversations[id]
	if !ok || c.UserID != userID {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

func (m *MockConversationRepo) UpdateStatus(ctx context.Context, id, userID, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.conversations[id]
	if !ok || c.UserID != userID {
		return nil
	}
	c.Status = status
	if status == "resolved" {
		now := time.Now()
		c.ResolvedAt = &now
	}
	return nil
}

func (m *MockConversationRepo) UpdateCustomerInfo(ctx context.Context, id, name, avatar string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.conversations[id]
	if !ok {
		return nil
	}
	c.CustomerName = name
	c.CustomerAvatar = avatar
	return nil
}

func (m *MockConversationRepo) FindActiveByCustomer(ctx context.Context, userID, customerName, channel string) (*domain.Conversation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.conversations {
		if c.UserID == userID && (c.CustomerPhone == customerName || c.CustomerName == customerName) && c.Channel == channel && c.Status == "active" {
			cp := *c
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *MockConversationRepo) Takeover(ctx context.Context, id, userID, agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.conversations[id]
	if !ok || c.UserID != userID {
		return nil
	}
	c.Status = "escalated"
	c.TakenOverBy = &agentID
	now := time.Now()
	c.TakenOverAt = &now
	return nil
}

func (m *MockConversationRepo) ClearChats(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, c := range m.conversations {
		if c.UserID == userID {
			delete(m.conversations, id)
		}
	}
	var msgs []*domain.Message
	for _, msg := range m.messages {
		if c, ok := m.conversations[msg.ConversationID]; ok && c.UserID != userID {
			msgs = append(msgs, msg)
		}
	}
	m.messages = msgs
	return nil
}

func (m *MockConversationRepo) GetOverview(ctx context.Context, userID string) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var total, active, escalated int
	today := time.Now().Format("2006-01-02")
	for _, c := range m.conversations {
		if c.UserID != userID {
			continue
		}
		total++
		if c.Status == "active" {
			active++
		}
		if c.Status == "escalated" {
			escalated++
		}
	}
	_ = today
	return map[string]interface{}{
		"total_conversations":  total,
		"conversations_today":  0,
		"active_conversations": active,
		"resolved_today":       0,
		"ai_resolution_rate":   0.0,
		"escalated_count":      escalated,
	}, nil
}

func (m *MockConversationRepo) CountByChannel(ctx context.Context, userID string) (map[string]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]int)
	for _, c := range m.conversations {
		if c.UserID == userID {
			result[c.Channel]++
		}
	}
	return result, nil
}

func (m *MockConversationRepo) CountByIntent(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	counts := make(map[string]int)
	for _, c := range m.conversations {
		if c.UserID == userID {
			counts[c.Intent]++
		}
	}
	result := make([]map[string]interface{}, 0, len(counts))
	for intent, count := range counts {
		result = append(result, map[string]interface{}{"intent": intent, "count": count})
	}
	return result, nil
}

func (m *MockConversationRepo) CountByHour(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockConversationRepo) CountByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockConversationRepo) RecordCSAT(ctx context.Context, userID, conversationID string, score int, comment *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.csatRatings = append(m.csatRatings, csatRecord{
		userID:         userID,
		conversationID: conversationID,
		score:          score,
		comment:        comment,
		createdAt:      time.Now(),
	})
	return nil
}

func (m *MockConversationRepo) GetCSATAverage(ctx context.Context, userID string) (avg float64, total int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var sum, count int
	for _, r := range m.csatRatings {
		if r.userID == userID {
			sum += r.score
			count++
		}
	}
	if count == 0 {
		return 0, 0, nil
	}
	return float64(sum) / float64(count), count, nil
}

func (m *MockConversationRepo) GetCSATDistribution(ctx context.Context, userID string) (map[int]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	dist := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	for _, r := range m.csatRatings {
		if r.userID == userID {
			dist[r.score]++
		}
	}
	return dist, nil
}

func (m *MockConversationRepo) GetCSATTrend(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockConversationRepo) CountMessagesByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockConversationRepo) GetUptimeStats(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockConversationRepo) CleanupOldResolved(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var count int64
	for id, c := range m.conversations {
		if c.Status == "resolved" && c.UpdatedAt.Before(cutoff) {
			delete(m.conversations, id)
			count++
		}
	}
	return count, nil
}

func (m *MockConversationRepo) CleanupAbandoned(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var count int64
	for _, c := range m.conversations {
		if c.Status == "active" && c.UpdatedAt.Before(cutoff) {
			c.Status = "resolved"
			now := time.Now()
			c.ResolvedAt = &now
			count++
		}
	}
	return count, nil
}

// ============================================================
// MockMessageRepo
// ============================================================

type MockMessageRepo struct {
	mu       sync.Mutex
	messages []*domain.Message
}

func NewMockMessageRepo() *MockMessageRepo {
	return &MockMessageRepo{}
}

func (m *MockMessageRepo) Create(ctx context.Context, msg *domain.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if msg.ID == "" {
		msg.ID = generateUUID()
	}
	msg.Sequence = len(m.messages) + 1
	now := time.Now()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	cp := *msg
	m.messages = append(m.messages, &cp)
	return nil
}

func (m *MockMessageRepo) ListByConversation(ctx context.Context, conversationID string, limit int) ([]domain.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.Message
	for _, msg := range m.messages {
		if msg.ConversationID == conversationID {
			cp := *msg
			result = append(result, cp)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockMessageRepo) ListByConversationPaginated(ctx context.Context, conversationID string, limit, offset int) ([]domain.Message, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []domain.Message
	for _, msg := range m.messages {
		if msg.ConversationID == conversationID {
			cp := *msg
			all = append(all, cp)
		}
	}
	total := len(all)
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	if offset >= total {
		return []domain.Message{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (m *MockMessageRepo) GetLastMessage(ctx context.Context, conversationID string) (*domain.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var last *domain.Message
	for _, msg := range m.messages {
		if msg.ConversationID == conversationID {
			cp := *msg
			last = &cp
		}
	}
	return last, nil
}

func (m *MockMessageRepo) CountUnread(ctx context.Context, conversationID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, msg := range m.messages {
		if msg.ConversationID == conversationID && !msg.IsRead && msg.Role == "customer" {
			count++
		}
	}
	return count, nil
}

func (m *MockMessageRepo) MarkRead(ctx context.Context, conversationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, msg := range m.messages {
		if msg.ConversationID == conversationID && !msg.IsRead {
			msg.IsRead = true
		}
	}
	return nil
}

func (m *MockMessageRepo) CleanupOrphaned(ctx context.Context) (int64, error) {
	return 0, nil
}

// ============================================================
// MockQAPairRepo
// ============================================================

type MockQAPairRepo struct {
	mu   sync.Mutex
	qas  map[string]*domain.QAPair
}

func NewMockQAPairRepo() *MockQAPairRepo {
	return &MockQAPairRepo{qas: make(map[string]*domain.QAPair)}
}

func (m *MockQAPairRepo) Create(ctx context.Context, qa *domain.QAPair) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if qa.ID == "" {
		qa.ID = generateUUID()
	}
	now := time.Now()
	if qa.CreatedAt.IsZero() {
		qa.CreatedAt = now
	}
	qa.UpdatedAt = now
	if qa.Variations == nil {
		qa.Variations = []string{}
	}
	cp := *qa
	m.qas[qa.ID] = &cp
	return nil
}

func (m *MockQAPairRepo) BulkCreate(ctx context.Context, qas []domain.QAPair) error {
	for i := range qas {
		if err := m.Create(ctx, &qas[i]); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockQAPairRepo) ListByCategory(ctx context.Context, categoryID string) ([]domain.QAPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.QAPair
	for _, qa := range m.qas {
		if qa.CategoryID == categoryID && qa.IsActive {
			cp := *qa
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockQAPairRepo) ListByCategoryAndUser(ctx context.Context, categoryID, userID string) ([]domain.QAPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.QAPair
	for _, qa := range m.qas {
		if qa.CategoryID == categoryID && qa.IsActive {
			cp := *qa
			cp.UserID = userID
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockQAPairRepo) Search(ctx context.Context, userID, query string) ([]domain.QAPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	lq := strings.ToLower(query)
	var result []domain.QAPair
	for _, qa := range m.qas {
		if !qa.IsActive {
			continue
		}
		if strings.Contains(strings.ToLower(qa.Question), lq) || strings.Contains(strings.ToLower(qa.Answer), lq) {
			cp := *qa
			result = append(result, cp)
			if len(result) >= 10 {
				break
			}
		}
	}
	return result, nil
}

func (m *MockQAPairRepo) ListByUser(ctx context.Context, userID, categoryID string) ([]domain.QAPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.QAPair
	for _, qa := range m.qas {
		if qa.IsActive {
			if categoryID == "" || qa.CategoryID == categoryID {
				cp := *qa
				result = append(result, cp)
			}
		}
	}
	return result, nil
}

func (m *MockQAPairRepo) GetByID(ctx context.Context, id string) (*domain.QAPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	qa, ok := m.qas[id]
	if !ok {
		return nil, nil
	}
	cp := *qa
	return &cp, nil
}

func (m *MockQAPairRepo) GetByQuestion(ctx context.Context, userID, question string) (*domain.QAPair, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, qa := range m.qas {
		if qa.Question == question {
			cp := *qa
			cp.UserID = userID
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *MockQAPairRepo) Update(ctx context.Context, qa *domain.QAPair) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.qas[qa.ID]
	if !ok {
		return nil
	}
	existing.CategoryID = qa.CategoryID
	existing.Question = qa.Question
	existing.Answer = qa.Answer
	existing.Variations = qa.Variations
	existing.IsActive = qa.IsActive
	existing.UpdatedAt = time.Now()
	return nil
}

func (m *MockQAPairRepo) IncrementUsage(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if qa, ok := m.qas[id]; ok {
		qa.UsageCount++
	}
	return nil
}

func (m *MockQAPairRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, qa := range m.qas {
		if qa.IsActive {
			count++
		}
	}
	return count, nil
}

func (m *MockQAPairRepo) Delete(ctx context.Context, id, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.qas, id)
	return nil
}

// ============================================================
// MockCategoryRepo
// ============================================================

type MockCategoryRepo struct {
	mu         sync.Mutex
	categories map[string]*domain.Category
}

func NewMockCategoryRepo() *MockCategoryRepo {
	return &MockCategoryRepo{categories: make(map[string]*domain.Category)}
}

func (m *MockCategoryRepo) GetByName(ctx context.Context, userID, name string) (*domain.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cat := range m.categories {
		if cat.Name == name {
			cp := *cat
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *MockCategoryRepo) Create(ctx context.Context, cat *domain.Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cat.ID == "" {
		cat.ID = generateUUID()
	}
	now := time.Now()
	if cat.CreatedAt.IsZero() {
		cat.CreatedAt = now
	}
	cp := *cat
	m.categories[cat.ID] = &cp
	return nil
}

func (m *MockCategoryRepo) List(ctx context.Context, userID string) ([]domain.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var cats []domain.Category
	for _, cat := range m.categories {
		if cat.UserID != userID {
			continue
		}
		cp := *cat
		cats = append(cats, cp)
	}
	return cats, nil
}

func (m *MockCategoryRepo) Delete(ctx context.Context, id, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.categories, id)
	return nil
}

// ============================================================
// MockUnknownQuestionRepo
// ============================================================

type MockUnknownQuestionRepo struct {
	mu        sync.Mutex
	questions map[string]*domain.UnknownQuestion
}

func NewMockUnknownQuestionRepo() *MockUnknownQuestionRepo {
	return &MockUnknownQuestionRepo{questions: make(map[string]*domain.UnknownQuestion)}
}

func (m *MockUnknownQuestionRepo) Create(ctx context.Context, uq *domain.UnknownQuestion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if uq.ID == "" {
		uq.ID = generateUUID()
	}
	now := time.Now()
	if uq.CreatedAt.IsZero() {
		uq.CreatedAt = now
	}
	cp := *uq
	m.questions[uq.ID] = &cp
	return nil
}

func (m *MockUnknownQuestionRepo) GetByIDAndUser(ctx context.Context, id, userID string) (*domain.UnknownQuestion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	uq, ok := m.questions[id]
	if !ok || uq.UserID != userID {
		return nil, nil
	}
	cp := *uq
	return &cp, nil
}

func (m *MockUnknownQuestionRepo) List(ctx context.Context, userID, status string, limit, offset int) ([]domain.UnknownQuestion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []domain.UnknownQuestion
	for _, uq := range m.questions {
		if uq.UserID != userID {
			continue
		}
		if status != "" && uq.Status != status {
			continue
		}
		cp := *uq
		all = append(all, cp)
	}
	if offset >= len(all) {
		return []domain.UnknownQuestion{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *MockUnknownQuestionRepo) BatchTrain(ctx context.Context, userID, answer, categoryID string, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		if uq, ok := m.questions[id]; ok && uq.UserID == userID {
			uq.Status = "trained"
			uq.SuggestedAnswer = &answer
			uq.CategoryID = &categoryID
		}
	}
	return nil
}

func (m *MockUnknownQuestionRepo) BatchIgnore(ctx context.Context, userID string, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range ids {
		if uq, ok := m.questions[id]; ok && uq.UserID == userID {
			uq.Status = "ignored"
		}
	}
	return nil
}

func (m *MockUnknownQuestionRepo) ExistsPending(ctx context.Context, userID, question string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, uq := range m.questions {
		if uq.UserID == userID && strings.EqualFold(uq.Question, question) && uq.Status == "pending" {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUnknownQuestionRepo) UpdateStatus(ctx context.Context, id, userID, status string, answer, categoryID *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	uq, ok := m.questions[id]
	if !ok || uq.UserID != userID {
		return nil
	}
	uq.Status = status
	uq.SuggestedAnswer = answer
	uq.CategoryID = categoryID
	return nil
}

func (m *MockUnknownQuestionRepo) Clear(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, uq := range m.questions {
		if uq.UserID == userID {
			delete(m.questions, id)
		}
	}
	return nil
}

func (m *MockUnknownQuestionRepo) CountByStatus(ctx context.Context, userID string) (map[string]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := map[string]int{"pending": 0, "trained": 0, "ignored": 0}
	for _, uq := range m.questions {
		if uq.UserID == userID {
			result[uq.Status]++
		}
	}
	return result, nil
}

func (m *MockUnknownQuestionRepo) MostPopular(ctx context.Context, userID string, limit int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockUnknownQuestionRepo) CountByFilter(ctx context.Context, userID, status string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, uq := range m.questions {
		if uq.UserID != userID {
			continue
		}
		if status == "" || uq.Status == status {
			count++
		}
	}
	return count, nil
}

func (m *MockUnknownQuestionRepo) CountByDate(ctx context.Context, userID string, days int) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockUnknownQuestionRepo) CleanupStale(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var count int64
	for id, uq := range m.questions {
		if (uq.Status == "trained" || uq.Status == "ignored") && uq.CreatedAt.Before(cutoff) {
			delete(m.questions, id)
			count++
		}
	}
	return count, nil
}

// ============================================================
// MockIntegrationRepo
// ============================================================

type MockIntegrationRepo struct {
	mu           sync.Mutex
	integrations map[string]*domain.Integration
}

func NewMockIntegrationRepo() *MockIntegrationRepo {
	return &MockIntegrationRepo{integrations: make(map[string]*domain.Integration)}
}

func (m *MockIntegrationRepo) Create(ctx context.Context, integration *domain.Integration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if integration.ID == "" {
		integration.ID = generateUUID()
	}
	now := time.Now()
	if integration.CreatedAt.IsZero() {
		integration.CreatedAt = now
	}
	integration.UpdatedAt = now
	cp := *integration
	m.integrations[integration.ID] = &cp
	return nil
}

func (m *MockIntegrationRepo) ListByUser(ctx context.Context, userID string) ([]domain.Integration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.Integration
	for _, i := range m.integrations {
		if i.UserID == userID {
			cp := *i
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockIntegrationRepo) ListActive(ctx context.Context) ([]domain.Integration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.Integration
	for _, i := range m.integrations {
		if i.Status == "active" {
			cp := *i
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockIntegrationRepo) ListByChannel(ctx context.Context, channel string) ([]domain.Integration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.Integration
	for _, i := range m.integrations {
		if i.Channel == channel {
			cp := *i
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockIntegrationRepo) UpdateStatus(ctx context.Context, id, status string, lastError *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	i, ok := m.integrations[id]
	if !ok {
		return nil
	}
	i.Status = status
	i.LastError = lastError
	i.UpdatedAt = time.Now()
	return nil
}

func (m *MockIntegrationRepo) GetByUserAndChannel(ctx context.Context, userID, channel string) (*domain.Integration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, i := range m.integrations {
		if i.UserID == userID && i.Channel == channel {
			cp := *i
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *MockIntegrationRepo) GetByChannelAndSessionID(ctx context.Context, channel, sessionID string) (*domain.Integration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, i := range m.integrations {
		if i.Channel == channel {
			if sid, ok := i.Config["session_id"].(string); ok && sid == sessionID {
				cp := *i
				return &cp, nil
			}
		}
	}
	return nil, nil
}

func (m *MockIntegrationRepo) GetByChannelAndWebhookSecret(ctx context.Context, channel, secret string) (*domain.Integration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, i := range m.integrations {
		if i.Channel == channel && (i.Status == "active" || i.Status == "connected") {
			if cfgSecret, ok := i.Config["webhook_secret"].(string); ok && cfgSecret == secret {
				cp := *i
				return &cp, nil
			}
		}
	}
	return nil, nil
}

func (m *MockIntegrationRepo) Update(ctx context.Context, integration *domain.Integration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.integrations[integration.ID]
	if !ok {
		return nil
	}
	existing.Status = integration.Status
	existing.Config = integration.Config
	existing.WebhookURL = integration.WebhookURL
	existing.UpdatedAt = time.Now()
	return nil
}

func (m *MockIntegrationRepo) Disconnect(ctx context.Context, userID, channel string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, i := range m.integrations {
		if i.UserID == userID && i.Channel == channel {
			i.Status = "inactive"
			i.UpdatedAt = time.Now()
		}
	}
	return nil
}

func (m *MockIntegrationRepo) CleanupStaleInactive(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var count int64
	for id, i := range m.integrations {
		if i.Status == "inactive" && i.UpdatedAt.Before(cutoff) {
			delete(m.integrations, id)
			count++
		}
	}
	return count, nil
}

// ============================================================
// MockTeamRepo
// ============================================================

type MockTeamRepo struct {
	mu      sync.Mutex
	members map[string]*domain.TeamMember
	owners  map[string]string
}

func NewMockTeamRepo() *MockTeamRepo {
	return &MockTeamRepo{
		members: make(map[string]*domain.TeamMember),
		owners:  make(map[string]string),
	}
}

func (m *MockTeamRepo) ListByUser(ctx context.Context, ownerID string) ([]domain.TeamMember, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.TeamMember
	for id, member := range m.members {
		if m.owners[id] == ownerID {
			cp := *member
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockTeamRepo) Create(ctx context.Context, ownerID string, member *domain.TeamMember) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if member.ID == "" {
		member.ID = generateUUID()
	}
	now := time.Now()
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}
	cp := *member
	m.members[member.ID] = &cp
	m.owners[member.ID] = ownerID
	return nil
}

// ============================================================
// MockAPIKeyRepo
// ============================================================

type MockAPIKeyRepo struct {
	mu   sync.Mutex
	keys map[string]*domain.APIKey
}

func NewMockAPIKeyRepo() *MockAPIKeyRepo {
	return &MockAPIKeyRepo{keys: make(map[string]*domain.APIKey)}
}

func (m *MockAPIKeyRepo) Create(ctx context.Context, key *domain.APIKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if key.ID == "" {
		key.ID = generateUUID()
	}
	now := time.Now()
	if key.CreatedAt.IsZero() {
		key.CreatedAt = now
	}
	cp := *key
	m.keys[key.ID] = &cp
	return nil
}

func (m *MockAPIKeyRepo) ListByUser(ctx context.Context, userID string) ([]domain.APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.APIKey
	for _, k := range m.keys {
		if k.UserID == userID && k.IsActive {
			cp := *k
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockAPIKeyRepo) Revoke(ctx context.Context, id, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.keys[id]
	if ok && k.UserID == userID {
		k.IsActive = false
	}
	return nil
}

// ============================================================
// MockArchiveRepo
// ============================================================

type MockArchiveRepo struct {
	mu      sync.Mutex
	folders map[string]*domain.ArchiveFolder
}

func NewMockArchiveRepo() *MockArchiveRepo {
	return &MockArchiveRepo{folders: make(map[string]*domain.ArchiveFolder)}
}

func (m *MockArchiveRepo) CreateFolder(ctx context.Context, folder *domain.ArchiveFolder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if folder.ID == "" {
		folder.ID = generateUUID()
	}
	now := time.Now()
	if folder.CreatedAt.IsZero() {
		folder.CreatedAt = now
	}
	cp := *folder
	m.folders[folder.ID] = &cp
	return nil
}

func (m *MockArchiveRepo) ListFolders(ctx context.Context, userID, folderType string) ([]domain.ArchiveFolder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.ArchiveFolder
	for _, f := range m.folders {
		if f.UserID == userID {
			if folderType == "" || f.Type == folderType {
				cp := *f
				result = append(result, cp)
			}
		}
	}
	return result, nil
}

func (m *MockArchiveRepo) MoveChat(ctx context.Context, conversationID, userID, folderID string) error {
	return nil
}

// ============================================================
// MockSubscriptionRepo
// ============================================================

type MockSubscriptionRepo struct {
	mu  sync.Mutex
	subs map[string]*domain.Subscription
}

func NewMockSubscriptionRepo() *MockSubscriptionRepo {
	return &MockSubscriptionRepo{subs: make(map[string]*domain.Subscription)}
}

func (m *MockSubscriptionRepo) GetActive(ctx context.Context, userID string) (*domain.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.subs {
		if s.UserID == userID && s.Status == "active" {
			cp := *s
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *MockSubscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sub.ID == "" {
		sub.ID = generateUUID()
	}
	now := time.Now()
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = now
	}
	sub.UpdatedAt = now
	cp := *sub
	m.subs[sub.ID] = &cp
	return nil
}

func (m *MockSubscriptionRepo) CreateOrUpdate(ctx context.Context, sub *domain.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing := (*domain.Subscription)(nil)
	for _, s := range m.subs {
		if s.UserID == sub.UserID && s.Status == "active" {
			existing = s
			break
		}
	}
	if existing != nil {
		existing.PlanID = sub.PlanID
		existing.Status = sub.Status
		existing.CurrentPeriodStart = sub.CurrentPeriodStart
		existing.CurrentPeriodEnd = sub.CurrentPeriodEnd
		existing.UpdatedAt = time.Now()
		return nil
	}
	if sub.ID == "" {
		sub.ID = generateUUID()
	}
	now := time.Now()
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = now
	}
	sub.UpdatedAt = now
	cp := *sub
	m.subs[sub.ID] = &cp
	return nil
}

func (m *MockSubscriptionRepo) Cancel(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.subs {
		if s.UserID == userID && s.Status == "active" {
			s.Status = "canceled" //nolint:misspell // external API status value
			s.UpdatedAt = time.Now()
		}
	}
	return nil
}

// ============================================================
// MockAuditRepo
// ============================================================

type MockAuditRepo struct {
	mu   sync.Mutex
	logs []*domain.AuditLog
}

func NewMockAuditRepo() *MockAuditRepo {
	return &MockAuditRepo{}
}

func (m *MockAuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if log.ID == "" {
		log.ID = generateUUID()
	}
	now := time.Now()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = now
	}
	cp := *log
	m.logs = append(m.logs, &cp)
	return nil
}

func (m *MockAuditRepo) ListByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.AuditLog
	for i := len(m.logs) - 1; i >= 0; i-- {
		if m.logs[i].UserID == userID {
			cp := *m.logs[i]
			result = append(result, cp)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	if result == nil {
		result = []domain.AuditLog{}
	}
	return result, nil
}

func (m *MockAuditRepo) CleanupOld(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var kept []*domain.AuditLog
	var count int64
	for _, log := range m.logs {
		if log.CreatedAt.Before(cutoff) {
			count++
		} else {
			kept = append(kept, log)
		}
	}
	m.logs = kept
	return count, nil
}

// ============================================================
// MockNotificationRepo
// ============================================================

type MockNotificationRepo struct {
	mu            sync.Mutex
	notifications map[string]*domain.Notification
}

func NewMockNotificationRepo() *MockNotificationRepo {
	return &MockNotificationRepo{notifications: make(map[string]*domain.Notification)}
}

func (m *MockNotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if n.ID == "" {
		n.ID = generateUUID()
	}
	now := time.Now()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}
	cp := *n
	m.notifications[n.ID] = &cp
	return nil
}

func (m *MockNotificationRepo) ListByUser(ctx context.Context, userID string, limit int) ([]*domain.Notification, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID {
			cp := *n
			result = append(result, &cp)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockNotificationRepo) UnreadCount(ctx context.Context, userID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, n := range m.notifications {
		if n.UserID == userID && !n.IsRead {
			count++
		}
	}
	return count, nil
}

func (m *MockNotificationRepo) MarkRead(ctx context.Context, id, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.notifications[id]
	if ok && n.UserID == userID {
		n.IsRead = true
	}
	return nil
}

func (m *MockNotificationRepo) MarkAllRead(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, n := range m.notifications {
		if n.UserID == userID {
			n.IsRead = true
		}
	}
	return nil
}

func (m *MockNotificationRepo) CleanupOld(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var count int64
	for id, n := range m.notifications {
		if n.CreatedAt.Before(cutoff) {
			delete(m.notifications, id)
			count++
		}
	}
	return count, nil
}

// ============================================================
// MockWidgetConfigRepo
// ============================================================

type MockWidgetConfigRepo struct {
	mu    sync.Mutex
	cfgs  map[string]*domain.WidgetConfig
}

func NewMockWidgetConfigRepo() *MockWidgetConfigRepo {
	return &MockWidgetConfigRepo{cfgs: make(map[string]*domain.WidgetConfig)}
}

func (m *MockWidgetConfigRepo) Get(ctx context.Context, userID string) (*domain.WidgetConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cfg := range m.cfgs {
		if cfg.UserID == userID {
			cp := *cfg
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *MockWidgetConfigRepo) GetByAPIKey(ctx context.Context, apiKey string) (*domain.WidgetConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cfg := range m.cfgs {
		if cfg.WidgetAPIKey == apiKey && cfg.IsActive {
			cp := *cfg
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *MockWidgetConfigRepo) Upsert(ctx context.Context, cfg *domain.WidgetConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cfg.ID == "" {
		cfg.ID = generateUUID()
	}
	now := time.Now()
	cfg.UpdatedAt = now
	for _, existing := range m.cfgs {
		if existing.UserID != cfg.UserID {
			continue
		}
		existing.BrandColor = cfg.BrandColor
		existing.Greeting = cfg.Greeting
		existing.BotName = cfg.BotName
		existing.Position = cfg.Position
		existing.WidgetAPIKey = cfg.WidgetAPIKey
		existing.IsActive = cfg.IsActive
		existing.UpdatedAt = now
		return nil
	}
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = now
	}
	cp := *cfg
	m.cfgs[cfg.ID] = &cp
	return nil
}

// ============================================================
// MockInventoryRepo
// ============================================================

type MockInventoryRepo struct {
	mu    sync.Mutex
	items map[string]*domain.InventoryItem
}

func NewMockInventoryRepo() *MockInventoryRepo {
	return &MockInventoryRepo{items: make(map[string]*domain.InventoryItem)}
}

func (m *MockInventoryRepo) Create(ctx context.Context, item *domain.InventoryItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if item.ID == "" {
		item.ID = generateUUID()
	}
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	cp := *item
	m.items[item.ID] = &cp
	return nil
}

func (m *MockInventoryRepo) GetByID(ctx context.Context, id, userID string) (*domain.InventoryItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok || item.UserID != userID {
		return nil, nil
	}
	cp := *item
	return &cp, nil
}

func (m *MockInventoryRepo) List(ctx context.Context, userID, itemType string, activeOnly bool) ([]domain.InventoryItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.InventoryItem
	for _, item := range m.items {
		if item.UserID != userID {
			continue
		}
		if itemType != "" && item.Type != itemType {
			continue
		}
		if activeOnly && !item.IsActive {
			continue
		}
		cp := *item
		result = append(result, cp)
	}
	return result, nil
}

func (m *MockInventoryRepo) Search(ctx context.Context, userID, q string) ([]domain.InventoryItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	lq := strings.ToLower(q)
	var result []domain.InventoryItem
	for _, item := range m.items {
		if item.UserID != userID || !item.IsActive {
			continue
		}
		if strings.Contains(strings.ToLower(item.Name), lq) || strings.Contains(strings.ToLower(item.Description), lq) {
			cp := *item
			result = append(result, cp)
			if len(result) >= 10 {
				break
			}
		}
	}
	return result, nil
}

func (m *MockInventoryRepo) Update(ctx context.Context, item *domain.InventoryItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.items[item.ID]
	if !ok || existing.UserID != item.UserID {
		return nil
	}
	existing.Type = item.Type
	existing.Name = item.Name
	existing.Description = item.Description
	existing.Price = item.Price
	existing.MinPrice = item.MinPrice
	existing.StockQuantity = item.StockQuantity
	existing.ImageURL = item.ImageURL
	existing.IsActive = item.IsActive
	existing.UpdatedAt = time.Now()
	return nil
}

func (m *MockInventoryRepo) Delete(ctx context.Context, id, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if item, ok := m.items[id]; ok && item.UserID == userID {
		delete(m.items, id)
	}
	return nil
}

func (m *MockInventoryRepo) DecreaseStock(ctx context.Context, itemID string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[itemID]
	if !ok {
		return nil
	}
	if item.StockQuantity != nil {
		newQty := *item.StockQuantity - quantity
		if newQty < 0 {
			newQty = 0
		}
		item.StockQuantity = &newQty
	}
	return nil
}

func (m *MockInventoryRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, item := range m.items {
		if item.UserID == userID {
			count++
		}
	}
	return count, nil
}

// ============================================================
// MockHandoffRepo
// ============================================================

type MockHandoffRepo struct {
	mu        sync.Mutex
	handoffs  map[string]*domain.Handoff
}

func NewMockHandoffRepo() *MockHandoffRepo {
	return &MockHandoffRepo{handoffs: make(map[string]*domain.Handoff)}
}

func (m *MockHandoffRepo) Create(ctx context.Context, h *domain.Handoff) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if h.ID == "" {
		h.ID = generateUUID()
	}
	if h.Quantity == 0 {
		h.Quantity = 1
	}
	now := time.Now()
	if h.CreatedAt.IsZero() {
		h.CreatedAt = now
	}
	h.UpdatedAt = now
	cp := *h
	m.handoffs[h.ID] = &cp
	return nil
}

func (m *MockHandoffRepo) GetByID(ctx context.Context, id, userID string) (*domain.Handoff, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.handoffs[id]
	if !ok || h.UserID != userID {
		return nil, nil
	}
	cp := *h
	return &cp, nil
}

func (m *MockHandoffRepo) List(ctx context.Context, userID, status string, limit int) ([]domain.Handoff, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.Handoff
	for _, h := range m.handoffs {
		if h.UserID != userID {
			continue
		}
		if status != "" && h.Status != status {
			continue
		}
		cp := *h
		result = append(result, cp)
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockHandoffRepo) UpdateStatus(ctx context.Context, id, userID, status, notes string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.handoffs[id]
	if !ok || h.UserID != userID {
		return nil
	}
	h.Status = status
	h.OwnerNotes = notes
	h.UpdatedAt = time.Now()
	return nil
}

func (m *MockHandoffRepo) GetPending(ctx context.Context, userID string) ([]domain.Handoff, error) {
	return m.List(ctx, userID, "pending", 100)
}

func (m *MockHandoffRepo) GetReadyForReminder(ctx context.Context) ([]domain.Handoff, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.Handoff
	now := time.Now()
	for _, h := range m.handoffs {
		if h.Status == "pending" && h.NextReminderAt != nil && !h.NextReminderAt.After(now) && h.ReminderCount < 3 {
			cp := *h
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockHandoffRepo) IncrementReminder(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.handoffs[id]
	if !ok {
		return nil
	}
	h.ReminderCount++
	next := time.Now().Add(15 * time.Minute)
	h.NextReminderAt = &next
	now := time.Now()
	h.OwnerNotifiedAt = &now
	h.UpdatedAt = now
	return nil
}

func (m *MockHandoffRepo) Expire(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.handoffs[id]
	if !ok {
		return nil
	}
	h.Status = "expired"
	h.UpdatedAt = time.Now()
	return nil
}

func (m *MockHandoffRepo) CleanupExpired(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var count int64
	for _, h := range m.handoffs {
		if h.Status == "pending" && h.CreatedAt.Before(cutoff) {
			h.Status = "expired"
			h.UpdatedAt = time.Now()
			count++
		}
	}
	return count, nil
}

// ============================================================
// MockCreditRepo
// ============================================================

type MockCreditRepo struct {
	mu        sync.Mutex
	credits   map[string]*domain.UserCredit
	purchases map[string]*domain.CreditPurchase
}

func NewMockCreditRepo() *MockCreditRepo {
	return &MockCreditRepo{
		credits:   make(map[string]*domain.UserCredit),
		purchases: make(map[string]*domain.CreditPurchase),
	}
}

func (m *MockCreditRepo) GetByUserID(ctx context.Context, userID string) (*domain.UserCredit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.credits {
		if c.UserID == userID {
			cp := *c
			return &cp, nil
		}
	}
	return &domain.UserCredit{
		UserID:        userID,
		Balance:       0,
		LastUpdatedAt: time.Now(),
	}, nil
}

func (m *MockCreditRepo) Upsert(ctx context.Context, credit *domain.UserCredit) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if credit.ID == "" {
		credit.ID = generateUUID()
	}
	credit.LastUpdatedAt = time.Now()
	cp := *credit
	m.credits[credit.ID] = &cp
	return nil
}

func (m *MockCreditRepo) Deduct(ctx context.Context, userID string, amount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.credits {
		if c.UserID != userID {
			continue
		}
		if c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now()) {
			return fmt.Errorf("credit balance has expired for user %s", userID)
		}
		if c.Balance < amount {
			return fmt.Errorf("insufficient credit balance: have %d, need %d", c.Balance, amount)
		}
		c.Balance -= amount
		c.LastUpdatedAt = time.Now()
		return nil
	}
	return fmt.Errorf("no credit balance found for user %s", userID)
}

func (m *MockCreditRepo) GetExpiring(ctx context.Context, days int) ([]domain.UserCredit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	expiryDate := time.Now().AddDate(0, 0, days)
	var result []domain.UserCredit
	for _, c := range m.credits {
		if c.ExpiresAt != nil && c.ExpiresAt.Before(expiryDate) && c.ExpiresAt.After(time.Now()) {
			cp := *c
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockCreditRepo) CreatePurchase(ctx context.Context, p *domain.CreditPurchase) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p.ID == "" {
		p.ID = generateUUID()
	}
	cp := *p
	m.purchases[p.ID] = &cp
	return nil
}

func (m *MockCreditRepo) GetPurchaseHistory(ctx context.Context, userID string) ([]domain.CreditPurchase, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.CreditPurchase
	for _, p := range m.purchases {
		if p.UserID == userID {
			cp := *p
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockCreditRepo) CleanupExpired(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	var count int64
	for id, c := range m.credits {
		if c.ExpiresAt != nil && c.ExpiresAt.Before(now) {
			delete(m.credits, id)
			count++
		}
	}
	return count, nil
}

func (m *MockCreditRepo) CleanupStalePurchases(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var count int64
	for id, p := range m.purchases {
		if p.PurchasedAt.Before(cutoff) {
			delete(m.purchases, id)
			count++
		}
	}
	return count, nil
}

// ============================================================
// MockCampaignRepo
// ============================================================

type MockCampaignRepo struct {
	mu        sync.Mutex
	campaigns map[string]*domain.CampaignSchedule
}

func NewMockCampaignRepo() *MockCampaignRepo {
	return &MockCampaignRepo{campaigns: make(map[string]*domain.CampaignSchedule)}
}

func (m *MockCampaignRepo) Create(ctx context.Context, campaign *domain.CampaignSchedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if campaign.ID == "" {
		campaign.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	now := time.Now()
	if campaign.CreatedAt.IsZero() {
		campaign.CreatedAt = now
	}
	campaign.UpdatedAt = now
	cp := *campaign
	m.campaigns[campaign.ID] = &cp
	return nil
}

func (m *MockCampaignRepo) ListByUser(ctx context.Context, userID string) ([]domain.CampaignSchedule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.CampaignSchedule
	for _, c := range m.campaigns {
		if c.UserID == userID {
			cp := *c
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockCampaignRepo) GetScheduledForToday(ctx context.Context) ([]domain.CampaignSchedule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	var result []domain.CampaignSchedule
	for _, c := range m.campaigns {
		if c.StartDate == today && c.Status == "draft" {
			cp := *c
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockCampaignRepo) GetEndingToday(ctx context.Context) ([]domain.CampaignSchedule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	var result []domain.CampaignSchedule
	for _, c := range m.campaigns {
		if c.EndDate == today && c.Status == "active" {
			cp := *c
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockCampaignRepo) UpdateStatus(ctx context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.campaigns[id]
	if !ok {
		return nil
	}
	c.Status = status
	c.UpdatedAt = time.Now()
	return nil
}

func (m *MockCampaignRepo) CleanupCompleted(ctx context.Context, days int) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().AddDate(0, 0, -days)
	var count int64
	for id, c := range m.campaigns {
		if (c.Status == "completed" || c.Status == "canceled") && c.UpdatedAt.Before(cutoff) {
			delete(m.campaigns, id)
			count++
		}
	}
	return count, nil
}

// ============================================================
// MockWhatsAppTemplateRepo
// ============================================================

type MockWhatsAppTemplateRepo struct {
	mu        sync.Mutex
	templates map[string]*domain.WhatsAppTemplate
}

func NewMockWhatsAppTemplateRepo() *MockWhatsAppTemplateRepo {
	return &MockWhatsAppTemplateRepo{templates: make(map[string]*domain.WhatsAppTemplate)}
}

func (m *MockWhatsAppTemplateRepo) Create(ctx context.Context, tpl *domain.WhatsAppTemplate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if tpl.ID == "" {
		tpl.ID = generateUUID()
	}
	now := time.Now()
	if tpl.CreatedAt.IsZero() {
		tpl.CreatedAt = now
	}
	tpl.UpdatedAt = now
	cp := *tpl
	m.templates[tpl.ID] = &cp
	return nil
}

func (m *MockWhatsAppTemplateRepo) ListByUser(ctx context.Context, userID string) ([]domain.WhatsAppTemplate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.WhatsAppTemplate
	for _, t := range m.templates {
		if t.UserID == userID {
			cp := *t
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockWhatsAppTemplateRepo) GetByID(ctx context.Context, id, userID string) (*domain.WhatsAppTemplate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.templates[id]
	if !ok {
		return nil, nil
	}
	if userID != "" && t.UserID != userID {
		return nil, nil
	}
	cp := *t
	return &cp, nil
}

func (m *MockWhatsAppTemplateRepo) Update(ctx context.Context, tpl *domain.WhatsAppTemplate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.templates[tpl.ID]
	if !ok {
		return nil
	}
	existing.Name = tpl.Name
	existing.Language = tpl.Language
	existing.Category = tpl.Category
	existing.Status = tpl.Status
	existing.HeaderType = tpl.HeaderType
	existing.HeaderValue = tpl.HeaderValue
	existing.BodyText = tpl.BodyText
	existing.FooterText = tpl.FooterText
	existing.Buttons = tpl.Buttons
	existing.Namespace = tpl.Namespace
	existing.RejectionReason = tpl.RejectionReason
	existing.UpdatedAt = time.Now()
	return nil
}

func (m *MockWhatsAppTemplateRepo) Delete(ctx context.Context, id, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.templates[id]; ok && t.UserID == userID {
		delete(m.templates, id)
	}
	return nil
}

func (m *MockWhatsAppTemplateRepo) GetByStatus(ctx context.Context, status string) ([]domain.WhatsAppTemplate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.WhatsAppTemplate
	for _, t := range m.templates {
		if t.Status == status {
			cp := *t
			result = append(result, cp)
		}
	}
	return result, nil
}

// ============================================================
// MockCampaignRecipientRepo
// ============================================================

type MockCampaignRecipientRepo struct {
	mu          sync.Mutex
	recipients  map[string]*domain.CampaignRecipient
}

func NewMockCampaignRecipientRepo() *MockCampaignRecipientRepo {
	return &MockCampaignRecipientRepo{recipients: make(map[string]*domain.CampaignRecipient)}
}

func (m *MockCampaignRecipientRepo) Create(ctx context.Context, cr *domain.CampaignRecipient) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cr.ID == "" {
		cr.ID = generateUUID()
	}
	now := time.Now()
	if cr.CreatedAt.IsZero() {
		cr.CreatedAt = now
	}
	cp := *cr
	m.recipients[cr.ID] = &cp
	return nil
}

func (m *MockCampaignRecipientRepo) ListByCampaign(ctx context.Context, campaignID string) ([]domain.CampaignRecipient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.CampaignRecipient
	for _, r := range m.recipients {
		if r.CampaignID == campaignID {
			cp := *r
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockCampaignRecipientRepo) UpdateStatus(ctx context.Context, id, status string, errInfo *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.recipients[id]
	if !ok {
		return nil
	}
	r.Status = status
	r.Error = ""
	if errInfo != nil {
		r.Error = *errInfo
	}
	now := time.Now()
	switch status {
	case "sent":
		r.SentAt = &now
	case "delivered":
		r.DeliveredAt = &now
	case "read":
		r.ReadAt = &now
	}
	return nil
}

func (m *MockCampaignRecipientRepo) MarkOptedOut(ctx context.Context, userID, phone string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.recipients {
		if r.UserID == userID && r.Phone == phone && (r.Status == "pending" || r.Status == "sent") {
			r.Status = "opted_out"
		}
	}
	return nil
}

func (m *MockCampaignRecipientRepo) IsOptedOut(ctx context.Context, userID, phone string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.recipients {
		if r.UserID == userID && r.Phone == phone && r.Status == "opted_out" {
			return true, nil
		}
	}
	return false, nil
}

// ============================================================
// MockMediaMessageRepo
// ============================================================

type MockMediaMessageRepo struct {
	mu      sync.Mutex
	media   map[string]*domain.MediaMessage
}

func NewMockMediaMessageRepo() *MockMediaMessageRepo {
	return &MockMediaMessageRepo{media: make(map[string]*domain.MediaMessage)}
}

func (m *MockMediaMessageRepo) Create(ctx context.Context, msg *domain.MediaMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if msg.ID == "" {
		msg.ID = generateUUID()
	}
	now := time.Now()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	cp := *msg
	m.media[msg.ID] = &cp
	return nil
}

func (m *MockMediaMessageRepo) GetByConversation(ctx context.Context, conversationID string) ([]domain.MediaMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.MediaMessage
	for _, msg := range m.media {
		if msg.ConversationID == conversationID {
			cp := *msg
			result = append(result, cp)
		}
	}
	return result, nil
}

func (m *MockMediaMessageRepo) CleanupExpired(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	var count int64
	for id, msg := range m.media {
		if msg.ExpiresAt.Before(now) {
			delete(m.media, id)
			count++
		}
	}
	return count, nil
}

// ============================================================
// MockPushSubscriptionRepo
// ============================================================

type MockPushSubscriptionRepo struct {
	mu   sync.Mutex
	subs []*domain.PushSubscription
}

func NewMockPushSubscriptionRepo() *MockPushSubscriptionRepo {
	return &MockPushSubscriptionRepo{}
}

func (m *MockPushSubscriptionRepo) Create(ctx context.Context, sub *domain.PushSubscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sub.ID == "" {
		sub.ID = generateUUID()
	}
	now := time.Now()
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = now
	}
	sub.UpdatedAt = now
	for i, existing := range m.subs {
		if existing.UserID == sub.UserID && existing.Endpoint == sub.Endpoint {
			cp := *sub
			m.subs[i] = &cp
			return nil
		}
	}
	cp := *sub
	m.subs = append(m.subs, &cp)
	return nil
}

func (m *MockPushSubscriptionRepo) Delete(ctx context.Context, userID, endpoint string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, sub := range m.subs {
		if sub.UserID == userID && sub.Endpoint == endpoint {
			m.subs = append(m.subs[:i], m.subs[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockPushSubscriptionRepo) DeleteAllByUser(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var kept []*domain.PushSubscription
	for _, sub := range m.subs {
		if sub.UserID != userID {
			kept = append(kept, sub)
		}
	}
	m.subs = kept
	return nil
}

func (m *MockPushSubscriptionRepo) ListByUser(ctx context.Context, userID string) ([]*domain.PushSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*domain.PushSubscription
	for _, sub := range m.subs {
		if sub.UserID == userID {
			cp := *sub
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (m *MockPushSubscriptionRepo) ListByUserIDs(ctx context.Context, userIDs []string) ([]*domain.PushSubscription, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	idSet := make(map[string]bool, len(userIDs))
	for _, id := range userIDs {
		idSet[id] = true
	}
	var result []*domain.PushSubscription
	for _, sub := range m.subs {
		if idSet[sub.UserID] {
			cp := *sub
			result = append(result, &cp)
		}
	}
	return result, nil
}

func (m *MockPushSubscriptionRepo) DeleteByID(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, sub := range m.subs {
		if sub.ID == id {
			m.subs = append(m.subs[:i], m.subs[i+1:]...)
			break
		}
	}
	return nil
}

// ============================================================
// MockRepositories — groups all mock repos
// ============================================================

type MockRepositories struct {
	User              *MockUserRepo
	Conversation      *MockConversationRepo
	Message           *MockMessageRepo
	QAPair            *MockQAPairRepo
	Category          *MockCategoryRepo
	UnknownQ          *MockUnknownQuestionRepo
	Integration       *MockIntegrationRepo
	Team              *MockTeamRepo
	APIKey            *MockAPIKeyRepo
	Archive           *MockArchiveRepo
	Subscription      *MockSubscriptionRepo
	Audit             *MockAuditRepo
	Notification      *MockNotificationRepo
	WidgetConfig      *MockWidgetConfigRepo
	Inventory         *MockInventoryRepo
	Handoff           *MockHandoffRepo
	Credit            *MockCreditRepo
	Campaign          *MockCampaignRepo
	WhatsAppTemplate  *MockWhatsAppTemplateRepo
	CampaignRecipient *MockCampaignRecipientRepo
	MediaMessage      *MockMediaMessageRepo
	PushSubscription  *MockPushSubscriptionRepo
}

func NewMockRepositories() *MockRepositories {
	return &MockRepositories{
		User:              NewMockUserRepo(),
		Conversation:      NewMockConversationRepo(),
		Message:           NewMockMessageRepo(),
		QAPair:            NewMockQAPairRepo(),
		Category:          NewMockCategoryRepo(),
		UnknownQ:          NewMockUnknownQuestionRepo(),
		Integration:       NewMockIntegrationRepo(),
		Team:              NewMockTeamRepo(),
		APIKey:            NewMockAPIKeyRepo(),
		Archive:           NewMockArchiveRepo(),
		Subscription:      NewMockSubscriptionRepo(),
		Audit:             NewMockAuditRepo(),
		Notification:      NewMockNotificationRepo(),
		WidgetConfig:      NewMockWidgetConfigRepo(),
		Inventory:         NewMockInventoryRepo(),
		Handoff:           NewMockHandoffRepo(),
		Credit:            NewMockCreditRepo(),
		Campaign:          NewMockCampaignRepo(),
		WhatsAppTemplate:  NewMockWhatsAppTemplateRepo(),
		CampaignRecipient: NewMockCampaignRecipientRepo(),
		MediaMessage:      NewMockMediaMessageRepo(),
		PushSubscription:  NewMockPushSubscriptionRepo(),
	}
}
