package service

import (
	"context"
	"encoding/json"
	"strings"

	"noant/internal/infrastructure"
)

type AssistantAction struct {
	Type  string `json:"type"`
	Path  string `json:"path,omitempty"`
	Label string `json:"label,omitempty"`
}

type AssistantResponse struct {
	Content     string            `json:"content"`
	Action      *AssistantAction  `json:"action,omitempty"`
	Steps       []AssistantStep   `json:"steps,omitempty"`
	Suggestions []string          `json:"suggestions,omitempty"`
}

type AssistantStep struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Path        string `json:"path,omitempty"`
	Action      string `json:"action,omitempty"`
}

type AssistantService struct {
	brain  *AIBrain
	logger *infrastructure.Logger
}

func NewAssistantService(brain *AIBrain, logger *infrastructure.Logger) *AssistantService {
	return &AssistantService{brain: brain, logger: logger}
}

func (s *AssistantService) Chat(ctx context.Context, message string) (*AssistantResponse, error) {
	content, _, err := s.brain.callGroqWithFallback(ctx, []MessageTurn{
		{Role: "system", Content: assistantSystemPrompt},
		{Role: "user", Content: message},
	})
	if err != nil {
		s.logger.Warn("Assistant Groq call failed, using fallback", "error", err)
		return s.fallbackResponse(message)
	}

	var resp AssistantResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil || resp.Content == "" {
		return &AssistantResponse{Content: content}, nil
	}

	return &resp, nil
}

func (s *AssistantService) fallbackResponse(message string) (*AssistantResponse, error) {
	msg := strings.ToLower(message)

	if strings.Contains(msg, "chat") || strings.Contains(msg, "conversation") || strings.Contains(msg, "inbox") || strings.Contains(msg, "message") {
		return &AssistantResponse{
			Content: "The Chats page is where you manage all customer conversations. You can reply to messages, take over conversations from the AI, and escalate issues.",
			Action:  &AssistantAction{Type: "navigate", Path: "/chats", Label: "Go to Chats"},
			Steps: []AssistantStep{
				{Title: "View conversations", Description: "Go to the Chats page to see all active conversations", Path: "/chats"},
				{Title: "Reply to a customer", Description: "Click on any conversation and type your response in the input box", Path: "/chats"},
				{Title: "Take over from AI", Description: "Click the 'Take over' button to respond manually instead of the AI", Path: "/chats"},
			},
		}, nil
	}

	if strings.Contains(msg, "train") || strings.Contains(msg, "teach") || strings.Contains(msg, "qa") || strings.Contains(msg, "question") || strings.Contains(msg, "answer") || strings.Contains(msg, "knowledge") {
		return &AssistantResponse{
			Content: "The Teach page is where you train your AI by adding Q&A pairs. You can also import questions in bulk and review unknown questions that customers have asked.",
			Action:  &AssistantAction{Type: "navigate", Path: "/teach", Label: "Go to Teach"},
			Steps: []AssistantStep{
				{Title: "Add Q&A pairs", Description: "Go to Teach and add question-answer pairs to train your AI", Path: "/teach"},
				{Title: "Bulk import", Description: "Upload a CSV file with multiple Q&A pairs at once", Path: "/teach"},
				{Title: "Review unknown questions", Description: "Check what customers are asking that your AI can't answer yet", Path: "/teach"},
			},
		}, nil
	}

	if strings.Contains(msg, "channel") || strings.Contains(msg, "whatsapp") || strings.Contains(msg, "telegram") || strings.Contains(msg, "widget") || strings.Contains(msg, "integration") || strings.Contains(msg, "connect") {
		return &AssistantResponse{
			Content: "The Channels page lets you connect your AI to different communication platforms like WhatsApp, Telegram, and your own website via an embeddable widget.",
			Action:  &AssistantAction{Type: "navigate", Path: "/channels", Label: "Go to Channels"},
			Steps: []AssistantStep{
				{Title: "Connect a channel", Description: "Go to Channels and click 'Connect' on your preferred platform", Path: "/channels"},
				{Title: "Configure WhatsApp", Description: "Set up WhatsApp Cloud API to handle customer messages", Path: "/channels"},
				{Title: "Install web widget", Description: "Get the embed code to add the AI chat widget to your website", Path: "/widget"},
			},
		}, nil
	}

	if strings.Contains(msg, "analytics") || strings.Contains(msg, "insight") || strings.Contains(msg, "stats") || strings.Contains(msg, "statistics") || strings.Contains(msg, "trend") || strings.Contains(msg, "overview") {
		return &AssistantResponse{
			Content: "The Insights dashboard shows you analytics about your conversations, including volume trends, channel distribution, and response metrics.",
			Action:  &AssistantAction{Type: "navigate", Path: "/insights", Label: "Go to Insights"},
			Steps: []AssistantStep{
				{Title: "View overview", Description: "See key metrics and conversation statistics", Path: "/insights"},
				{Title: "Channel distribution", Description: "See which channels your customers use the most", Path: "/insights"},
				{Title: "Trends", Description: "Track conversation volume over time", Path: "/insights"},
			},
		}, nil
	}

	if strings.Contains(msg, "setting") || strings.Contains(msg, "profile") || strings.Contains(msg, "password") || strings.Contains(msg, "account") {
		return &AssistantResponse{
			Content: "Your account settings are in the Settings page. You can update your profile, change your password, manage API keys, and invite team members.",
			Action:  &AssistantAction{Type: "navigate", Path: "/settings", Label: "Go to Settings"},
			Steps: []AssistantStep{
				{Title: "Update profile", Description: "Change your name, email, and company details", Path: "/settings"},
				{Title: "Manage API keys", Description: "Create and revoke API keys for integrations", Path: "/settings/api-keys"},
				{Title: "Invite team", Description: "Add team members to collaborate", Path: "/settings/team"},
			},
		}, nil
	}

	if strings.Contains(msg, "billing") || strings.Contains(msg, "plan") || strings.Contains(msg, "pricing") || strings.Contains(msg, "subscription") || strings.Contains(msg, "upgrade") || strings.Contains(msg, "payment") {
		return &AssistantResponse{
			Content: "You can manage your subscription and billing on the Billing page. View your current plan, upgrade to unlock more features, and see payment history.",
			Action:  &AssistantAction{Type: "navigate", Path: "/billing", Label: "Go to Billing"},
		}, nil
	}

	if strings.Contains(msg, "lead") || strings.Contains(msg, "customer") {
		return &AssistantResponse{
			Content: "The Leads page shows you all your customer leads and their information. You can track and manage your sales pipeline here.",
			Action:  &AssistantAction{Type: "navigate", Path: "/leads", Label: "Go to Leads"},
		}, nil
	}

	if strings.Contains(msg, "inventory") || strings.Contains(msg, "product") || strings.Contains(msg, "stock") {
		return &AssistantResponse{
			Content: "The Inventory page lets you manage your products and stock levels. You can add, edit, and organize your inventory items.",
			Action:  &AssistantAction{Type: "navigate", Path: "/inventory", Label: "Go to Inventory"},
		}, nil
	}

	if strings.Contains(msg, "team") || strings.Contains(msg, "member") || strings.Contains(msg, "invite") || strings.Contains(msg, "collaborat") {
		return &AssistantResponse{
			Content: "You can manage your team members in the Team settings. Invite colleagues, assign roles, and collaborate on customer support.",
			Action:  &AssistantAction{Type: "navigate", Path: "/settings/team", Label: "Go to Team Settings"},
		}, nil
	}

	if strings.Contains(msg, "hello") || strings.Contains(msg, "hi ") || strings.Contains(msg, "hey") || strings.Contains(msg, "help") || msg == "hi" {
		return &AssistantResponse{
			Content: "Hi there! I'm your Noant guide. I can help you learn how to use the platform. Try asking me about:\n\n• Chats — managing conversations\n• Training the AI\n• Connecting channels\n• Analytics and insights\n• Settings and team management\n• Billing and plans",
		}, nil
	}

	return &AssistantResponse{
		Content: "I'm here to help you get started with Noant! Here are some things I can help you with:",
		Steps: []AssistantStep{
			{Title: "Manage conversations", Description: "Learn how to reply to customers and take over chats", Path: "/chats"},
			{Title: "Train your AI", Description: "Add Q&A pairs and improve your AI's knowledge", Path: "/teach"},
			{Title: "Connect channels", Description: "Set up WhatsApp, Telegram, or your website widget", Path: "/channels"},
			{Title: "View analytics", Description: "See your conversation metrics and trends", Path: "/insights"},
			{Title: "Manage settings", Description: "Update your profile, team, and API keys", Path: "/settings"},
		},
	}, nil
}

const assistantSystemPrompt = `You are an intelligent onboarding assistant for "Noant" — an AI-powered customer support platform.
Your role is to help users understand and use the software effectively.

## Your Capabilities:
1. **Answer questions** about how to use Noant's features (chats, training AI, analytics, integrations, channels, settings, etc.)
2. **Provide step-by-step guides** by returning a "steps" array in your JSON response
3. **Navigate users** to the right page by returning an "action" object with type "navigate" and the path

## Response Format:
Always respond with a valid JSON object:
{
  "content": "Your helpful response text here",
  "action": { "type": "navigate", "path": "/dashboard", "label": "Go to Dashboard" },
  "steps": [
    { "title": "Step 1", "description": "What to do", "path": "/chats", "action": "Go to Chats" }
  ],
  "suggestions": ["Tell me about chats", "How do I train the AI?", "Show me analytics"]
}

- If the user asks how to do something, provide "steps" with relevant page paths
- If the user wants to go somewhere, provide an "action" object
- If it's a general question, just provide "content"
- Always include 2-3 "suggestions" as quick action chips for the user to tap next
- Keep responses concise and helpful

## Available Pages:
- /dashboard — Overview dashboard with stats and quick actions
- /chats — Customer conversation inbox (main chat interface)
- /teach — Train the AI with Q&A pairs, view unknown questions
- /insights — Analytics dashboard
- /channels — Connect communication channels (WhatsApp, Telegram, Web Widget, Instagram)
- /settings — Profile, API keys, team management, audit logs
- /settings/team — Manage team members and invitations
- /settings/api-keys — API key management
- /settings/audit-logs — View audit logs
- /notifications — View notifications
- /billing — Subscription and billing information
- /inventory — Manage inventory items
- /widget — Embeddable web widget configuration
- /leads — View and manage leads
- /setup — Initial account setup

## Features you can explain:
- Conversations: Viewing, replying, taking over conversations, escalation
- Training AI: Adding Q&A pairs, bulk CSV import, handling unknown questions
- Channels: WhatsApp Cloud API, Telegram bot, Web Widget setup
- Analytics: Overview stats, channel distribution, insights, trends
- Integrations: Connecting external services
- Inventory management
- Team collaboration and role management
- Subscription plans and billing`
