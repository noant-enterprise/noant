package service

import (
	"context"
	"encoding/json"
	"fmt"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== TEMPLATE SERVICE ==========

type TemplateService struct {
	cfg    *config.Config
	openwa *OpenWAService
	redis  *infrastructure.RedisClient
	logger *infrastructure.Logger
	repos  *repository.Repositories
}

func NewTemplateService(cfg *config.Config, openwa *OpenWAService, redis *infrastructure.RedisClient, logger *infrastructure.Logger, repos *repository.Repositories) *TemplateService {
	return &TemplateService{
		cfg:    cfg,
		openwa: openwa,
		redis:  redis,
		logger: logger,
		repos:  repos,
	}
}

// TemplateButton represents a button in a template
type TemplateButton struct {
	Type    string `json:"type"`    // quick_reply, url, call, catalog
	Text    string `json:"text"`
	URL     string `json:"url,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

// CreateTemplateRequest is the request payload for creating a template
type CreateTemplateRequest struct {
	Name       string            `json:"name"`
	Language   string            `json:"language"`
	Category   string            `json:"category"`
	HeaderType string            `json:"header_type"`
	HeaderValue string           `json:"header_value"`
	BodyText   string            `json:"body_text"`
	FooterText string            `json:"footer_text"`
	Buttons    []TemplateButton  `json:"buttons"`
}

// SendTemplateRequest is the request payload for sending a template message
type SendTemplateRequest struct {
	SessionID string                 `json:"session_id"`
	ChatID    string                 `json:"chat_id"`
	TemplateID string                `json:"template_id"`
	Variables map[string]string      `json:"variables"` // {{1}} -> value
}

// OpenWATemplatePayload is the payload sent to OpenWA for template messages
type OpenWATemplatePayload struct {
	ChatID          string                 `json:"chatId"`
	Template        string                 `json:"template"`
	Language        string                 `json:"language,omitempty"`
	Namespace       string                 `json:"namespace,omitempty"`
	Variables       map[string]string      `json:"variables,omitempty"`
}

func (ts *TemplateService) Create(ctx context.Context, userID string, req *CreateTemplateRequest) (*domain.WhatsAppTemplate, error) {
	tpl := &domain.WhatsAppTemplate{
		UserID:     userID,
		Name:       req.Name,
		Language:   req.Language,
		Category:   req.Category,
		HeaderType: req.HeaderType,
		HeaderValue: req.HeaderValue,
		BodyText:   req.BodyText,
		FooterText: req.FooterText,
		Status:     "draft",
	}

	if req.Buttons != nil {
		btnJSON, err := json.Marshal(req.Buttons)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal buttons: %w", err)
		}
		tpl.Buttons = string(btnJSON)
	}

	if err := ts.repos.WhatsAppTemplate.Create(ctx, tpl); err != nil {
		return nil, err
	}

	return tpl, nil
}

func (ts *TemplateService) List(ctx context.Context, userID string) ([]domain.WhatsAppTemplate, error) {
	return ts.repos.WhatsAppTemplate.ListByOrg(ctx, userID)
}

func (ts *TemplateService) GetByID(ctx context.Context, id, userID string) (*domain.WhatsAppTemplate, error) {
	return ts.repos.WhatsAppTemplate.GetByID(ctx, id, userID)
}

func (ts *TemplateService) Update(ctx context.Context, userID string, tpl *domain.WhatsAppTemplate) error {
	tpl.UserID = userID
	return ts.repos.WhatsAppTemplate.Update(ctx, tpl)
}

func (ts *TemplateService) Delete(ctx context.Context, id, userID string) error {
	return ts.repos.WhatsAppTemplate.Delete(ctx, id, userID)
}

// SubmitForApproval submits a template to WhatsApp for approval via OpenWA
func (ts *TemplateService) SubmitForApproval(ctx context.Context, id, userID string) error {
	tpl, err := ts.repos.WhatsAppTemplate.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if tpl == nil {
		return fmt.Errorf("template not found")
	}

	tpl.Status = "pending"
	return ts.repos.WhatsAppTemplate.Update(ctx, tpl)
}

// SendTemplate sends a template message via OpenWA
func (ts *TemplateService) SendTemplate(ctx context.Context, req SendTemplateRequest) error {
	tpl, err := ts.repos.WhatsAppTemplate.GetByID(ctx, req.TemplateID, "")
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}
	if tpl == nil {
		return fmt.Errorf("template not found")
	}
	if tpl.Status != "approved" && tpl.Status != "draft" {
		return fmt.Errorf("template is not approved (status: %s)", tpl.Status)
	}

	// Validate required variables exist in body text
	variableCount := countTemplateVariables(tpl.BodyText)
	if variableCount > 0 && len(req.Variables) < variableCount {
		return fmt.Errorf("template requires %d variables, got %d", variableCount, len(req.Variables))
	}

	return ts.openwa.sendTemplateMessageInternal(req.SessionID, req.ChatID, map[string]interface{}{
		"template":  tpl.Name,
		"language":  tpl.Language,
		"namespace": tpl.Namespace,
		"variables": req.Variables,
	})
}

func countTemplateVariables(text string) int {
	count := 0
	for i := 1; i <= 100; i++ {
		placeholder := fmt.Sprintf("{{%d}}", i)
		if stringsContains(text, placeholder) {
			count++
		} else {
			break
		}
	}
	return count
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && containsString(s, substr)
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ========== INTERACTIVE MESSAGE SUPPORT ==========

type InteractiveMessage struct {
	Type    string        `json:"type"` // list, buttons, catalog
	Header  string        `json:"header,omitempty"`
	Body    string        `json:"body"`
	Footer  string        `json:"footer,omitempty"`
	Items   []interface{} `json:"items"`
}

type ListItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ReplyButton struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// SendInteractiveMessage sends an interactive message (list, buttons) via OpenWA
func (ts *TemplateService) SendInteractiveMessage(sessionID, chatID string, msg *InteractiveMessage) error {
	url := fmt.Sprintf("%s/api/sessions/%s/messages/send-interactive",
		ts.cfg.OpenWABaseURL, sessionID)

	payload := map[string]interface{}{
		"chatId": chatID,
		"type":   msg.Type,
		"body":   msg.Body,
	}
	if msg.Header != "" {
		payload["header"] = msg.Header
	}
	if msg.Footer != "" {
		payload["footer"] = msg.Footer
	}
	payload["items"] = msg.Items

	return ts.openwa.sendRawMessage(url, payload)
}

// SendListMessage sends a list message (up to 10 items)
func (ts *TemplateService) SendListMessage(sessionID, chatID, header, body, footer, buttonText string, items []ListItem) error {
	itemsPayload := make([]interface{}, len(items))
	for i, item := range items {
		itemsPayload[i] = map[string]interface{}{
			"id":          item.ID,
			"title":       item.Title,
			"description": item.Description,
		}
	}

	return ts.SendInteractiveMessage(sessionID, chatID, &InteractiveMessage{
		Type:   "list",
		Header: header,
		Body:   body,
		Footer: footer,
		Items:  []interface{}{map[string]interface{}{"buttonText": buttonText, "sections": []interface{}{map[string]interface{}{"title": "Options", "rows": itemsPayload}}}},
	})
}

// SendButtonsMessage sends a reply button message (up to 3 buttons)
func (ts *TemplateService) SendButtonsMessage(sessionID, chatID, body string, buttons []ReplyButton) error {
	itemsPayload := make([]interface{}, len(buttons))
	for i, btn := range buttons {
		itemsPayload[i] = map[string]interface{}{
			"id":    btn.ID,
			"title": btn.Title,
		}
	}

	return ts.SendInteractiveMessage(sessionID, chatID, &InteractiveMessage{
		Type:  "buttons",
		Body:  body,
		Items: itemsPayload,
	})
}

// ========== COMMON TEMPLATE LIBRARY ==========

var commonTemplates = []CreateTemplateRequest{
	{
		Name:       "order_confirmation",
		Language:   "en",
		Category:   "utility",
		BodyText:   "Hi {{1}}, your order #{{2}} has been confirmed! Total: {{3}}. We'll notify you when it ships.",
		FooterText: "Thank you for choosing NOANT",
	},
	{
		Name:       "shipping_update",
		Language:   "en",
		Category:   "utility",
		BodyText:   "Hi {{1}}, your order #{{2}} is {{3}}. Expected delivery: {{4}}.",
		FooterText: "Track your order on our website",
	},
	{
		Name:       "payment_received",
		Language:   "en",
		Category:   "utility",
		BodyText:   "Hi {{1}}, we've received your payment of {{2}} for order #{{3}}. Thank you!",
		FooterText: "NOANT Payments",
	},
	{
		Name:       "appointment_reminder",
		Language:   "en",
		Category:   "utility",
		BodyText:   "Reminder: You have an appointment with {{1}} on {{2}} at {{3}}. Reply C to confirm or R to reschedule.",
	},
	{
		Name:       "welcome_message",
		Language:   "en",
		Category:   "marketing",
		BodyText:   "Welcome to {{1}}, {{2}}! We're excited to have you. Reply HELP to see what we can do for you.",
		FooterText: "Powered by NOANT AI",
	},
	{
		Name:       "abandoned_cart",
		Language:   "en",
		Category:   "marketing",
		BodyText:   "Hi {{1}}, you left items in your cart! Use code COMEBACK for 10% off your order of {{2}}.",
		FooterText: "Offer expires in 24 hours",
	},
}

func GetCommonTemplates() []CreateTemplateRequest {
	return commonTemplates
}
