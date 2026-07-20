package service

import (
	"context"
	"fmt"
	"strings"

	"noant/internal/domain"
)

func normalizeTurnRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "customer", "user":
		return "user"
	case "ai", "assistant", "human":
		return "assistant"
	case "system":
		return "system"
	default:
		return "user"
	}
}

func trimTurnText(text string, maxLen int) string {
	text = strings.TrimSpace(text)
	if maxLen <= 0 || len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return strings.TrimSpace(text[:maxLen-3]) + "..."
}

func containsAny(text string, needles ...string) bool {
	lower := strings.ToLower(text)
	for _, needle := range needles {
		if strings.Contains(lower, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func (b *AIBrain) recentConversationTurns(ctx context.Context, conversationID, currentQuery string, limit int) []MessageTurn {
	if limit <= 0 {
		limit = 8
	}

	if b.redis != nil {
		if history, err := b.getConversationHistory(ctx, conversationID); err == nil && len(history) > 0 {
			if len(history) > limit {
				history = history[len(history)-limit:]
			}
			return history
		}
	}

	messages, err := b.repos.Message.ListByConversation(ctx, conversationID, limit+1)
	if err != nil || len(messages) == 0 {
		return nil
	}

	turns := make([]MessageTurn, 0, len(messages))
	for i := range messages {
		if strings.TrimSpace(messages[i].Content) == "" {
			continue
		}
		turns = append(turns, MessageTurn{
			Role:    normalizeTurnRole(messages[i].Role),
			Content: strings.TrimSpace(messages[i].Content),
		})
	}

	if len(turns) > 0 {
		last := turns[len(turns)-1]
		if last.Role == "user" && strings.EqualFold(strings.TrimSpace(last.Content), strings.TrimSpace(currentQuery)) {
			turns = turns[:len(turns)-1]
		}
	}

	if len(turns) > limit {
		turns = turns[len(turns)-limit:]
	}

	return turns
}

func summarizeTurns(turns []MessageTurn, maxPairs int) string {
	if len(turns) == 0 {
		return "No prior conversation yet."
	}
	if maxPairs <= 0 {
		maxPairs = 3
	}

	start := 0
	if len(turns) > maxPairs*2 {
		start = len(turns) - maxPairs*2
	}

	var lines []string
	for _, turn := range turns[start:] {
		label := "Customer"
		switch turn.Role {
		case "assistant":
			label = "Assistant"
		case "system":
			label = "System"
		}
		lines = append(lines, fmt.Sprintf("%s: %s", label, trimTurnText(turn.Content, 180)))
	}
	return strings.Join(lines, "\n")
}

func detectSalesStage(query string, inventory []domain.InventoryItem, qaPairs []domain.QAPair, history []MessageTurn) string {
	lower := strings.ToLower(query)
	switch {
	case containsAny(lower, "price", "cost", "cheap", "discount", "last price", "cheaper", "offer", "reduce", "budget"):
		return "price negotiation"
	case containsAny(lower, "recommend", "suggest", "best one", "which one", "help me choose", "compare", "difference"):
		return "comparison and recommendation"
	case containsAny(lower, "buy", "order", "checkout", "pay now", "i'll take it", "place the order", "send account"):
		return "ready to buy"
	case containsAny(lower, "delivery", "ship", "arrive", "when", "timeline", "urgent", "same day"):
		return "logistics and reassurance"
	case len(inventory) == 0 && len(qaPairs) == 0:
		return "needs human clarification"
	default:
		if len(history) > 0 {
			last := history[len(history)-1]
			if last.Role == "user" && containsAny(last.Content, "price", "cost", "cheap", "discount", "help me choose") {
				return "follow-up sales question"
			}
		}
		return "product discovery"
	}
}

func (b *AIBrain) buildSalesCoachMessage(conv *domain.Conversation, user *domain.User, query string, history []MessageTurn, inventory []domain.InventoryItem, qaPairs []domain.QAPair) MessageTurn {
	customerName := ""
	channel := ""
	companyName := ""
	ownerName := ""
	ownerWhatsApp := ""
	if conv != nil {
		customerName = conv.CustomerName
		channel = conv.Channel
	}
	if user != nil {
		companyName = user.CompanyName
		ownerName = user.FirstName
		if whatsapp, err := b.repos.User.GetOwnerWhatsApp(context.Background(), user.ID); err == nil && whatsapp != "" {
			ownerWhatsApp = whatsapp
		}
	}

	stage := detectSalesStage(query, inventory, qaPairs, history)
	recent := summarizeTurns(history, 3)
	inventoryHint := "No matching inventory items."
	if len(inventory) > 0 {
		parts := make([]string, 0, min(3, len(inventory)))
		for i := range inventory {
			if i >= 3 {
				break
			}
			stock := ""
			if inventory[i].StockQuantity != nil {
				stock = fmt.Sprintf(" stock=%d", *inventory[i].StockQuantity)
			}
			minPrice := ""
			if inventory[i].MinPrice != nil && *inventory[i].MinPrice > 0 {
				minPrice = fmt.Sprintf(" min=%.0f", *inventory[i].MinPrice)
			}
			parts = append(parts, fmt.Sprintf("%s (%.0f%s%s)", inventory[i].Name, inventory[i].Price, stock, minPrice))
		}
		inventoryHint = strings.Join(parts, "; ")
	}

	return MessageTurn{
		Role: "system",
		Content: fmt.Sprintf(`Sales coaching:
- Business: %s
- Customer: %s
- Channel: %s
- Stage: %s
- Owner: %s
- Owner WhatsApp: %s
- Response goal: answer clearly, explain value, sound caring, and move the customer one step forward.
- Style goal: warm, soft, professional, human, and lightly playful only when appropriate.
- Conversation memory:
%s
- Relevant inventory snapshot: %s
- Important behavior:
  * If several items fit, recommend the best fit and ask one short follow-up question.
  * If the customer seems uncertain, help them compare options instead of forcing a hard close.
  * If the customer is ready to buy, make the next step very clear and easy.
  * Never sound automated or overly formal.`, companyName, customerName, channel, stage, ownerName, ownerWhatsApp, recent, inventoryHint),
	}
}

func isBroadInventoryQuery(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower == "" {
		return true
	}
	broadSignals := []string{
		"what do you have",
		"what do you sell",
		"available",
		"options",
		"recommend",
		"suggest",
		"show me",
		"what are your products",
		"how much",
		"price",
		"cheap",
		"budget",
		"best",
	}
	for _, signal := range broadSignals {
		if strings.Contains(lower, signal) {
			return true
		}
	}
	for _, word := range strings.Fields(lower) {
		if strings.IndexFunc(word, func(r rune) bool { return r >= '0' && r <= '9' }) != -1 {
			return false
		}
	}
	return len(strings.Fields(lower)) <= 4
}

func inventorySummaryLine(item *domain.InventoryItem) string {
	line := fmt.Sprintf("%s - %.0f", item.Name, item.Price)
	if item.StockQuantity != nil {
		line += fmt.Sprintf(" (stock: %d)", *item.StockQuantity)
	}
	if item.Description != "" {
		line += " " + trimTurnText(item.Description, 100)
	}
	return line
}

func softCloseLine(query string) string {
	lower := strings.ToLower(query)
	switch {
	case containsAny(lower, "price", "cost", "cheap", "discount", "last price"):
		return "If you want, I can also help you compare options or see the best value."
	case containsAny(lower, "recommend", "suggest", "help me choose", "best one"):
		return "If you want, I can narrow it down based on budget, quality, or speed."
	default:
		return "If you want, I can explain any option in more detail."
	}
}

func salesVoiceExamplesMessage(query string, inventory []domain.InventoryItem) MessageTurn {
	stage := detectSalesStage(query, inventory, nil, nil)
	examples := map[string][]string{
		"price negotiation": {
			"Customer: Can you do cheaper?",
			"Assistant: I can adjust a little, but I want to keep the quality right for you. This one is already a strong price.",
			"Customer: What is your last price?",
			"Assistant: That is the best I can do on this one, but I can show you another option if you want something lower.",
		},
		"comparison and recommendation": {
			"Customer: Which one should I pick?",
			"Assistant: If you want the best value, I would go with this one. If budget is tighter, I can show you a simpler option too.",
		},
		"ready to buy": {
			"Customer: I want to buy now.",
			"Assistant: Perfect. I can help you finish this quickly. Let me connect you to the owner so we keep it smooth.",
		},
		"logistics and reassurance": {
			"Customer: When will it arrive?",
			"Assistant: Let me give you the honest timeline so you can decide with confidence.",
		},
		"product discovery": {
			"Customer: What do you have?",
			"Assistant: I can show you a few good options and help you choose the best one for your need.",
		},
		"needs human clarification": {
			"Customer: Do you have something special?",
			"Assistant: Let me check the best option for you so I do not mislead you.",
		},
	}
	lines := examples[stage]
	if len(lines) == 0 {
		lines = examples["product discovery"]
	}
	return MessageTurn{
		Role:    "system",
		Content: "Voice examples for this sales moment:\n- " + strings.Join(lines, "\n- ") + "\n- Keep the language short, warm, and human. Never sound scripted.",
	}
}

func salesReplyStyleLine(query string) string {
	stage := detectSalesStage(query, nil, nil, nil)
	switch stage {
	case "price negotiation":
		return "Style: calm, confident, and value-protecting. Do not over-explain."
	case "comparison and recommendation":
		return "Style: helpful and advisory. Make the choice feel easy."
	case "ready to buy":
		return "Style: efficient, upbeat, and closing-oriented."
	case "logistics and reassurance":
		return "Style: reassuring, honest, and clear."
	default:
		return "Style: warm, friendly, and lightly persuasive."
	}
}
