package service

import (
	"context"
	"fmt"
	"strings"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

// handleSalesMode searches inventory and builds a sales-oriented response
func (b *AIBrain) handleSalesMode(ctx context.Context, userID string, conversation *domain.Conversation, query, language string, history []MessageTurn) (*AIResponse, error) {
	inventory := b.searchInventoryContext(ctx, userID, query, 5)
	user, err := b.repos.User.GetByID(ctx, userID)
	if err != nil {
		b.logger.Warn("Failed to get user in sales mode", "error", err)
	}
	var contextMessages []MessageTurn
	if len(inventory) > 0 {
		contextMessages = append(contextMessages, MessageTurn{
			Role:    "system",
			Content: "Available products/services from the store:",
		})
		for i := range inventory[:min(5, len(inventory))] {
			item := &inventory[i]
			stockInfo := ""
			if item.StockQuantity != nil {
				stockInfo = fmt.Sprintf(" (Stock: %d)", *item.StockQuantity)
			}
			minPriceInfo := ""
			if item.MinPrice != nil && *item.MinPrice > 0 && *item.MinPrice < item.Price {
				minPriceInfo = fmt.Sprintf(" | Min price: ₦%.0f (DO NOT go below this)", *item.MinPrice)
			}
			contextMessages = append(contextMessages, MessageTurn{
				Role:    "system",
				Content: fmt.Sprintf("%s — ₦%.0f%s%s\n%s", item.Name, item.Price, stockInfo, minPriceInfo, item.Description),
			})
		}
	} else {
		// No inventory found — tell the AI we have nothing to sell
		contextMessages = append(contextMessages, MessageTurn{
			Role:    "system",
			Content: "CRITICAL: No matching products or services found in inventory. Do NOT invent products or prices. Offer to connect the customer with a human agent.",
		})
	}
	contextMessages = append(contextMessages,
		b.buildSalesCoachMessage(conversation, user, query, history, inventory, nil),
		salesVoiceExamplesMessage(query, inventory),
		MessageTurn{
			Role:    "system",
			Content: salesReplyStyleLine(query),
		},
	)
	contextMessages = append(contextMessages, history...)
	ownerName := ""
	ownerWhatsApp := ""
	if user != nil {
		ownerName = user.FirstName
		whatsapp, err := b.repos.User.GetOwnerWhatsApp(ctx, userID)
		if err == nil {
			ownerWhatsApp = whatsapp
		}
	}
	prompt := b.BuildPrompt(&PromptTemplate{
		SystemPrompt: fmt.Sprintf(`You are a friendly sales assistant for %s. You work for %s (the owner). You sell products and negotiate prices naturally — like a real shop assistant texting a customer.

Owner: %s | WhatsApp: %s

CRITICAL RULES:
1. ONLY show products that appear in the "Available products/services" context below. NEVER invent products or prices.
2. If no products are listed, say: "Let me check what we have available" and escalate.
3. NEVER quote a price lower than the "Min price" listed.
4. When customer asks for discount: offer a SMALL discount if possible, or say "Let me check with Oga on that".
5. When customer is ready to buy: say "Perfect! Message %s on WhatsApp: %s to finalize".
6. NEVER send account numbers, handle payment, or promise delivery.
7. NEVER say "I'm an AI" or mention any platform name.

NEGOTIATION STYLE — be natural, like a real Nigerian shop assistant:
- Customer: "How much?" → You: "It's ₦X. Very good quality o!"
- Customer: "Can you do cheaper?" → You: "I can do ₦X — that's my best price." (only if above min_price)
- Customer: "That's too expensive" → You: "Oga, this one na quality o! But I fit do ₦X for you." (small discount)
- Customer: "Last price?" → You: "₦X na the last. I no fit go lower than that."
- Customer: "I want to buy" → You: "Great! Message %%s on WhatsApp: %%s"

LANGUAGE: English, Pidgin, Yoruba, Igbo, Hausa. Naira (₦) only.
Be warm, short, and natural.`, user.CompanyName, ownerName, ownerName, ownerWhatsApp, ownerName, ownerWhatsApp),
		Context:   contextMessages,
		UserQuery: query,
		Language:  language,
		Tone:      "friendly",
	})
	response, confidence, err := b.callGroqWithFallback(ctx, prompt)
	if err != nil {
		b.logger.Error("Groq API failed in sales mode", "error", err)
		// Local fallback — if we have inventory, show it directly
		if len(inventory) > 0 {
			items := make([]string, 0, min(3, len(inventory)))
			for i := range inventory[:min(3, len(inventory))] {
				items = append(items, "- "+inventorySummaryLine(&inventory[i]))
			}
			return &AIResponse{
				Content:    fmt.Sprintf("Here's what we have:\n%s\n\nAsk me about any item for more details.", strings.Join(items, "\n")),
				Confidence: 0.8,
				Source:     "inventory",
			}, nil
		}
		return &AIResponse{
			Content:    "I don't have that information yet, but I'll escalate this to a human agent who can help you.",
			Confidence: 0,
			Escalate:   true,
			Reason:     "AI service unavailable",
			Source:     "fallback",
		}, nil
	}
	cleanContent, sentiment, detectedLang, suggestions := parseAIMetadata(response)
	infrastructure.AISentimentTotal.WithLabelValues(sentiment).Inc()
	infrastructure.AILanguageTotal.WithLabelValues(detectedLang).Inc()
	return &AIResponse{
		Content:     cleanContent,
		Confidence:  confidence,
		Source:      "groq",
		Sentiment:   sentiment,
		Language:    detectedLang,
		Suggestions: suggestions,
	}, nil
}

// handleHandoff creates a handoff record and notifies the owner
func (b *AIBrain) handleHandoff(ctx context.Context, conversationID, userID, query string) (*AIResponse, error) {
	conv, err := b.repos.Conversation.GetByID(ctx, conversationID)
	if err != nil {
		b.logger.Warn("Failed to get conversation in handoff", "error", err)
	}
	customerName := ""
	customerPhone := ""
	if conv != nil {
		customerName = conv.CustomerName
		customerPhone = conv.CustomerPhone
	}
	// Search inventory for the product the customer wants to buy
	productName := ""
	var price float64
	inventory := b.searchInventoryContext(ctx, userID, query, 3)
	if len(inventory) > 0 {
		// Pick the most relevant match (first result from search)
		productName = inventory[0].Name
		price = inventory[0].Price
	} else {
		// No inventory match — use generic product from query
		productName = query
	}
	// Generate a summary from recent conversation turns
	recentTurns := b.recentConversationTurns(ctx, conversationID, query, 8)
	summary := b.summarizeConversation(ctx, query, recentTurns)
	if summary == "" {
		summary = fmt.Sprintf("Customer interested in: %s (₦%.0f)", productName, price)
	}

	handoff := &domain.Handoff{
		UserID:         userID,
		ConversationID: conversationID,
		CustomerName:   customerName,
		CustomerPhone:  customerPhone,
		ProductName:    productName,
		OriginalPrice:  price,
		AgreedPrice:    price,
		Quantity:       1,
		Summary:        summary,
	}

	// Check if this plan gets notifications
	user, err := b.repos.User.GetByID(ctx, userID)
	if err != nil {
		b.logger.Warn("Failed to get user in handoff", "error", err)
	}
	var hasNotification bool
	if user != nil {
		_, hasNotification, err = b.planSvc.CanCreateHandoff(ctx, userID, user.PlanID)
		if err != nil {
			b.logger.Warn("Failed to check handoff plan", "error", err)
		}
		// For free plan specifically, we know it doesn't get notifications
		if user.PlanID == "free" {
			hasNotification = false
		}
	}

	// Create handoff record (no notification if plan doesn't allow it)
	_ = b.repos.Handoff.Create(ctx, handoff)

	var ownerName string
	var ownerWhatsApp string
	if user != nil {
		ownerName = user.FirstName
		whatsapp, err := b.repos.User.GetOwnerWhatsApp(ctx, userID)
		if err == nil {
			ownerWhatsApp = whatsapp
		}

		// Notify owner via WebSocket only if plan allows it
		if hasNotification && b.broadcastFn != nil {
			b.broadcastFn(conversationID, "new_handoff", map[string]interface{}{
				"handoff_id":      handoff.ID,
				"customer_name":   customerName,
				"product_name":    productName,
				"agreed_price":    price,
				"conversation_id": conversationID,
				"summary":         summary,
			})
		}

		// Create notification only if plan allows it
		if hasNotification {
			notif := &domain.Notification{
				UserID: userID,
				Type:   "handoff",
				Title:  "New Sale Handoff",
				Body:   fmt.Sprintf("%s wants to buy %s for ₦%.0f. %s", customerName, productName, price, summary),
				Link:   "/leads",
				IsRead: false,
			}

			_ = b.repos.Notification.Create(ctx, notif)
		}
	}

	// Build response — if we found the product in inventory, confirm details
	if productName != query && price > 0 {
		response := fmt.Sprintf("Great! I see you're interested in %s for ₦%.0f. I'm connecting you with %s to complete your order. Message %s on WhatsApp: %s. Thank you!", productName, price, ownerName, ownerName, ownerWhatsApp)
		return &AIResponse{
			Content:    response,
			Confidence: 0.95,
		}, nil
	}
	response := fmt.Sprintf("Great! I'm connecting you with %s to complete your order. Message %s on WhatsApp: %s. Thank you!", ownerName, ownerName, ownerWhatsApp)
	return &AIResponse{
		Content:    response,
		Confidence: 0.95,
	}, nil
}

// summarizeConversation generates a brief summary of the conversation for handoff
func (b *AIBrain) summarizeConversation(ctx context.Context, query string, turns []MessageTurn) string {
	if len(turns) == 0 {
		return ""
	}
	// Build conversation text
	var parts []string
	for _, t := range turns {
		switch t.Role {
		case "user":
			parts = append(parts, "Customer: "+t.Content)
		case "assistant":
			parts = append(parts, "AI: "+t.Content)
		}
	}
	convText := strings.Join(parts, "\n")
	if len(convText) > 1500 {
		convText = convText[len(convText)-1500:]
	}

	summaryPrompt := []MessageTurn{
		{Role: "system", Content: "Summarize this customer service conversation in 1-2 sentences for a human agent taking over. Include: what the customer wants, any key details mentioned (product, price, issue), and the customer's sentiment."},
		{Role: "user", Content: convText},
	}
	summary, _, err := b.callGroqWithFallback(ctx, summaryPrompt)
	if err != nil || summary == "" {
		// Simple fallback
		return fmt.Sprintf("Customer query: %s. Conversation length: %d turns.", query, len(turns))
	}
	return strings.TrimSpace(summary)
}

// humanizeResponse uses Groq to rephrase a training-data answer to sound like a natural human,
// without changing facts or adding new information. Falls back to original on error.
func (b *AIBrain) humanizeResponse(ctx context.Context, answer, query string, qaPairs []domain.QAPair, inventory []domain.InventoryItem, history []MessageTurn) (string, error) {
	if answer == "" {
		return "", nil
	}
	// Short answers don't need humanizing
	if len(strings.Fields(answer)) <= 5 {
		return answer, nil
	}
	var contextItems []string
	for i := range qaPairs {
		if i >= 3 {
			break
		}
		contextItems = append(contextItems, fmt.Sprintf("Q: %s\nA: %s", qaPairs[i].Question, qaPairs[i].Answer))
	}
	var contextBuf string
	if len(contextItems) > 0 {
		contextBuf = "\nTraining data context:\n" + strings.Join(contextItems, "\n\n")
	}
	prompt := []MessageTurn{
		{Role: "system", Content: fmt.Sprintf(`You are a humanizer. Your ONLY job is to rephrase the provided answer so it sounds like a warm, natural human — like a Nigerian shop assistant texting.

RULES:
1. Keep ALL facts, prices, and information EXACTLY as given. Never add or remove information.
2. Make it sound natural — use Nigerian Pidgin/Yoruba/English expressions naturally (o, abeg, oya, na, shebi).
3. Keep the same length and meaning. Just change the VOICE.
4. NEVER say "I'm an AI" or "I'm a bot" or "according to my training".
5. If the customer seems to say hi, respond naturally.
6. Output ONLY the rephrased answer — no prefixes, no explanations.%s`, contextBuf)},
	}
	if len(history) > 0 {
		cut := history
		if len(cut) > 4 {
			cut = cut[len(cut)-4:]
		}
		var historyLines []string
		for _, t := range cut {
			historyLines = append(historyLines, t.Role+": "+t.Content)
		}
		prompt = append(prompt, MessageTurn{Role: "system", Content: "Recent conversation:\n" + strings.Join(historyLines, "\n")})
	}
	prompt = append(prompt, MessageTurn{Role: "user", Content: fmt.Sprintf("Customer asked: %s\n\nAnswer to humanize: %s", query, answer)})

	humanized, _, err := b.callGroqWithFallback(ctx, prompt)
	if err != nil || humanized == "" {
		return answer, err
	}
	return strings.TrimSpace(humanized), nil
}

// humanizeStreaming is like humanizeResponse but streams the Groq response via onChunk.
func (b *AIBrain) humanizeStreaming(ctx context.Context, answer, query string, qaPairs []domain.QAPair, inventory []domain.InventoryItem, history []MessageTurn, onChunk func(chunk string)) (string, error) {
	var contextBuf strings.Builder
	if len(qaPairs) > 0 {
		contextBuf.WriteString("Relevant Q&A:\n")
		for _, qa := range qaPairs {
			contextBuf.WriteString(fmt.Sprintf("Q: %s\nA: %s\n\n", qa.Question, qa.Answer))
		}
	}
	if len(inventory) > 0 {
		contextBuf.WriteString("Relevant products:\n")
		for _, item := range inventory {
			contextBuf.WriteString(fmt.Sprintf("- %s: ₦%.0f", item.Name, item.Price))
			if item.Description != "" {
				contextBuf.WriteString(" - " + trimTurnText(item.Description, 80))
			}
			contextBuf.WriteString("\n")
		}
	}

	prompt := []MessageTurn{
		{Role: "system", Content: fmt.Sprintf(`You are a friendly Nigerian shop assistant rephrasing answers to sound natural.

Rules:
1. Keep ALL facts, prices, and information EXACTLY as given. Never add or remove information.
2. Make it sound natural — use Nigerian Pidgin/Yoruba/English expressions naturally (o, abeg, oya, na, shebi).
3. Keep the same length and meaning. Just change the VOICE.
4. NEVER say "I'm an AI" or "I'm a bot" or "according to my training".
5. If the customer seems to say hi, respond naturally.
6. Output ONLY the rephrased answer — no prefixes, no explanations.%s`, contextBuf.String())},
	}
	if len(history) > 0 {
		cut := history
		if len(cut) > 4 {
			cut = cut[len(cut)-4:]
		}
		var historyLines []string
		for _, t := range cut {
			historyLines = append(historyLines, t.Role+": "+t.Content)
		}
		prompt = append(prompt, MessageTurn{Role: "system", Content: "Recent conversation:\n" + strings.Join(historyLines, "\n")})
	}
	prompt = append(prompt, MessageTurn{Role: "user", Content: fmt.Sprintf("Customer asked: %s\n\nAnswer to humanize: %s", query, answer)})

	humanized, _, err := b.callGroqStreaming(ctx, prompt, onChunk)
	if err != nil || humanized == "" {
		return answer, err
	}
	return strings.TrimSpace(humanized), nil
}
