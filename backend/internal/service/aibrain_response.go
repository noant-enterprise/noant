package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"noant/internal/domain"
	"noant/internal/infrastructure"
)

// Pre-compiled regex patterns (avoids recompilation on every AI response).
var (
	priceRegex     = regexp.MustCompile(`₦\s*([\d,]+(?:\.\d{1,2})?)`)
	sentimentRegex = regexp.MustCompile(`(?m)^\[SENTIMENT:([^\]]+)\]$`)
	languageRegex  = regexp.MustCompile(`(?m)^\[LANGUAGE:([^\]]+)\]$`)
	suggestionsRegex = regexp.MustCompile(`(?m)^\[SUGGESTIONS:([^\]]+)\]$`)
)

// validateResponse checks if the AI response hallucinates prices or products.
func (b *AIBrain) validateResponse(ctx context.Context, _, response string, _ []domain.QAPair, inventory []domain.InventoryItem) (resp string, conf float64) {
	if response == "" {
		return response, 0
	}

	lower := strings.ToLower(response)
	confidence := 0.95

	hallucinationSignals := []string{"according to my training", "based on my knowledge", "generally speaking", "typically"}
	for _, signal := range hallucinationSignals {
		if strings.Contains(lower, signal) {
			confidence *= hallucinationPenalty
			break
		}
	}

	if len(response) > maxResponseLength {
		confidence *= longResponsePenalty
	}

	if len(inventory) > 0 {
		matches := priceRegex.FindAllStringSubmatch(response, -1)
		for _, match := range matches {
			priceStr := strings.ReplaceAll(match[1], ",", "")
			claimedPrice, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				continue
			}
			found := false
			for i := range inventory {
				if inventory[i].Price == claimedPrice || (inventory[i].MinPrice != nil && *inventory[i].MinPrice <= claimedPrice && claimedPrice <= inventory[i].Price) {
					found = true
					break
				}
			}
			if !found {
				b.logger.Warn("AI hallucinated price", "claimedPrice", claimedPrice, "inventoryItems", len(inventory))
				return "I'm sorry, I don't have accurate pricing information for that. Please contact our sales team for exact prices.", confidenceHallucinated
			}
		}
	}

	if confidence < lowConfidenceThreshold {
		return "I'm not confident about the answer to that. Let me transfer you to a human agent.", confidence
	}

	return response, confidence
}

// parseAIMetadata extracts [SENTIMENT], [LANGUAGE], [SUGGESTIONS] tags from AI response.
func parseAIMetadata(content string) (clean, sentiment, language string, suggestions []string) {
	clean = content
	sentiment = "neutral"
	language = "en"

	if m := sentimentRegex.FindStringSubmatch(clean); len(m) > 1 {
		sentiment = strings.ToLower(strings.TrimSpace(m[1]))
		clean = sentimentRegex.ReplaceAllString(clean, "")
	}

	if m := languageRegex.FindStringSubmatch(clean); len(m) > 1 {
		language = strings.ToLower(strings.TrimSpace(m[1]))
		clean = languageRegex.ReplaceAllString(clean, "")
	}

	if m := suggestionsRegex.FindStringSubmatch(clean); len(m) > 1 {
		for _, s := range strings.Split(m[1], "|") {
			s = strings.TrimSpace(s)
			if s != "" {
				suggestions = append(suggestions, s)
			}
		}
		clean = suggestionsRegex.ReplaceAllString(clean, "")
	}

	clean = strings.TrimSpace(clean)
	return clean, sentiment, language, suggestions
}

func (b *AIBrain) localPlatformAnswer(userID, query string, qaPairs []domain.QAPair, inventory []domain.InventoryItem) *AIResponse {
	lower := strings.ToLower(query)

	isNegotiation := len(strings.Fields(query)) <= maxNegotiationWords && (strings.Contains(lower, "price") || strings.Contains(lower, "cheap") || strings.Contains(lower, "discount") || strings.Contains(lower, "last") || strings.Contains(lower, "negotiate") || strings.Contains(lower, "lower") || strings.Contains(lower, "reduce") || strings.Contains(lower, "offer"))

	if len(qaPairs) > 0 && qaWordOverlap(query, qaPairs[0].Question) >= localAnswerOverlapThreshold {
		first := qaPairs[0]
		content := strings.TrimSpace(first.Answer)
		if content == "" {
			content = "Let me know if you want me to explain anything else."
		} else if len(strings.Fields(query)) > 3 && !strings.HasSuffix(content, "?") {
			content += " If you want, I can also help you with the next best option."
		}
		return &AIResponse{
			Content:    content,
			Confidence: confidenceLocalMatch,
			MatchedQA:  &first.ID,
			Source:     "training",
		}
	}

	if len(inventory) > 0 {
		if len(inventory) > 1 && isBroadInventoryQuery(query) && !isNegotiation {
			var lines []string
			for i := range inventory {
				if i >= maxBroadInventory {
					break
				}
				lines = append(lines, "- "+inventorySummaryLine(&inventory[i]))
			}
			return &AIResponse{
				Content:    fmt.Sprintf("I found a few good options for you:\n%s\n\nWhich one should I break down for you?", strings.Join(lines, "\n")),
				Confidence: confidenceLocalMatch,
				Source:     "inventory",
			}
		}

		item := inventory[0]
		stock := ""
		if item.StockQuantity != nil {
			if *item.StockQuantity <= 5 {
				stock = fmt.Sprintf(" Only %d left!", *item.StockQuantity)
			} else {
				stock = fmt.Sprintf(" (Stock: %d)", *item.StockQuantity)
			}
		}
		if isNegotiation {
			if item.MinPrice != nil && *item.MinPrice > 0 && *item.MinPrice < item.Price {
				discount := item.Price - *item.MinPrice
				return &AIResponse{
					Content:    fmt.Sprintf("I can do ₦%.0f for you - that's ₦%.0f off. Best price I can offer. %s", *item.MinPrice, discount, stock),
					Confidence: confidenceLocalMatch,
					Source:     "inventory",
				}
			}
			return &AIResponse{
				Content:    fmt.Sprintf("₦%.0f is already my best price, o! %s", item.Price, stock),
				Confidence: confidenceLocalMatch,
				Source:     "inventory",
			}
		}
		answer := fmt.Sprintf("%s - ₦%.0f.%s", item.Name, item.Price, stock)
		if item.Description != "" {
			answer += " " + trimTurnText(item.Description, maxDescriptionPreview)
		}
		answer += " " + softCloseLine(query)
		return &AIResponse{
			Content:    answer,
			Confidence: confidenceLocalMatch,
			Source:     "inventory",
		}
	}
	return nil
}

// generateResponseCore is the single implementation shared by GenerateResponse and
// GenerateStreamingResponse. If onChunk is non-nil, the final humanization step
// streams tokens via Groq SSE; otherwise it uses the synchronous path.
func (b *AIBrain) generateResponseCore(ctx context.Context, conversationID, userQuery, language string, onChunk func(chunk string)) (aiResp *AIResponse, aiErr error) {
	startTime := time.Now()
	status := "success"
	defer func() {
		duration := time.Since(startTime).Seconds()
		switch {
		case aiErr != nil:
			status = "error"
		case aiResp != nil && aiResp.Escalate:
			status = "escalated"
		case aiResp != nil && aiResp.Source == "plan_limit":
			status = "plan_limited"
		case aiResp != nil && aiResp.Source == "greeting":
			status = "greeting"
		}
		infrastructure.AICallsTotal.WithLabelValues(groqModel, status).Inc()
		infrastructure.AIDuration.WithLabelValues(groqModel).Observe(duration)
	}()

	// Parallel DB lookups: conversation + history (no dependencies between them)
	type convResult struct {
		conv *domain.Conversation
		err  error
	}
	type historyResult struct {
		turns []MessageTurn
	}
	var convRes convResult
	var histRes historyResult
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.logger.Error("Panic in conversation fetch goroutine", "error", r)
			}
			wg.Done()
		}()
		convRes.conv, convRes.err = b.repos.Conversation.GetByID(ctx, conversationID)
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.logger.Error("Panic in history fetch goroutine", "error", r)
			}
			wg.Done()
		}()
		histRes.turns = b.recentConversationTurns(ctx, conversationID, userQuery, conversationTurnLimit)
	}()
	wg.Wait()

	conv := convRes.conv
	if convRes.err != nil {
		b.logger.Warn("Failed to get conversation", "error", convRes.err)
	}
	recentTurns := histRes.turns

	userID := ""
	var user *domain.User
	if conv != nil {
		userID = conv.UserID
	}

	traceID := fmt.Sprintf("trace_%d", time.Now().UnixNano())
	channel := ""
	if conv != nil {
		channel = conv.Channel
	}
	b.logger.Info("AI request started", "traceID", traceID, "userID", userID, "query", userQuery, "channel", channel)

	finalize := func(resp *AIResponse) (*AIResponse, error) {
		if resp != nil {
			if err := b.storeConversationTurn(ctx, conversationID, userQuery, resp.Content); err != nil {
				b.logger.Warn("Failed to store conversation turn", "error", err)
			}
		}
		return resp, nil
	}

	// Plan limit check
	if userID != "" {
		if user == nil {
			var err error
			user, err = b.repos.User.GetByID(ctx, userID)
			if err != nil {
				b.logger.Warn("Failed to get user for plan check", "error", err)
			}
		}
		if user != nil {
			canRespond, reason, err := b.planSvc.CanGenerateResponse(ctx, userID, user.PlanID)
			if err != nil {
				b.logger.Warn("Failed to check plan limit", "error", err)
			}
			if !canRespond {
				return finalize(&AIResponse{Content: reason, Source: "plan_limit"})
			}
		}
	}

	groqLimited := !b.allowGroqCall(ctx, userID)
	if groqLimited {
		b.logger.Warn("Groq rate limited for user, using local-only mode", "traceID", traceID, "userID", userID)
	}

	// Short greetings → local response (no AI call needed)
	if b.isGreetingQuery(userQuery) {
		b.logger.Info("AI request completed (greeting)", "traceID", traceID, "duration", time.Since(startTime))
		return finalize(&AIResponse{
			Content:     "Hi! How can I help you today?",
			Confidence:  confidenceGreeting,
			Sentiment:   "neutral",
			Language:    "en",
			Suggestions: []string{"I want to buy something", "I need help", "Tell me about your products"},
		})
	}

	// Intent classification: use keyword fast-path when streaming (saves ~500ms),
	// full LLM classification when not streaming.
	var intent string
	if onChunk != nil {
		intent = b.keywordIntent(userQuery)
		b.logger.Info("Intent classified (keyword fast-path)", "traceID", traceID, "intent", intent)
	} else {
		intent = b.classifyIntent(ctx, userQuery)
		b.logger.Info("Intent classified", "traceID", traceID, "intent", intent, "query", userQuery)
	}

	if intent == "handoff" {
		resp, err := b.handleHandoff(ctx, conversationID, userID, userQuery)
		b.logger.Info("AI request completed (handoff)", "traceID", traceID, "duration", time.Since(startTime))
		if err == nil && resp != nil {
			_ = b.storeConversationTurn(ctx, conversationID, userQuery, resp.Content)
		}
		return resp, err
	}

	if intent == "sales" {
		resp, err := b.handleSalesMode(ctx, userID, conv, userQuery, language, recentTurns)
		b.logger.Info("AI request completed (sales)", "traceID", traceID, "duration", time.Since(startTime))
		if err == nil && resp != nil {
			_ = b.storeConversationTurn(ctx, conversationID, userQuery, resp.Content)
		}
		return resp, err
	}

	// Support mode — search training data AND inventory in parallel
	var qaPairs []domain.QAPair
	var inventory []domain.InventoryItem
	var searchWg sync.WaitGroup
	searchWg.Add(2)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.logger.Error("Panic in knowledge base search goroutine", "error", r)
			}
			searchWg.Done()
		}()
		qaPairs = b.searchKnowledgeBase(ctx, userID, userQuery, qaSearchLimit)
	}()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.logger.Error("Panic in inventory search goroutine", "error", r)
			}
			searchWg.Done()
		}()
		inventory = b.searchInventoryContext(ctx, userID, userQuery, inventorySearchLimit)
	}()
	searchWg.Wait()

	b.logger.Info("Search completed", "traceID", traceID, "qaMatches", len(qaPairs), "inventoryMatches", len(inventory))

	// Try local answer from training data / inventory first
	local := b.localPlatformAnswer(userID, userQuery, qaPairs, inventory)
	if local != nil {
		if !groqLimited {
			var humanized string
			var err error
			if onChunk != nil {
				humanized, err = b.humanizeStreaming(ctx, local.Content, userQuery, qaPairs, inventory, recentTurns, onChunk)
			} else {
				humanized, err = b.humanizeResponse(ctx, local.Content, userQuery, qaPairs, inventory, recentTurns)
			}
			if err == nil && humanized != "" {
				cleanContent, sentiment, detectedLang, suggestions := parseAIMetadata(humanized)
				confidence := local.Confidence * confidenceHumanized
				b.logger.Info("AI request completed (humanized)", "traceID", traceID, "duration", time.Since(startTime), "confidence", confidence, "sentiment", sentiment, "language", detectedLang)
				infrastructure.AISentimentTotal.WithLabelValues(sentiment).Inc()
				infrastructure.AILanguageTotal.WithLabelValues(detectedLang).Inc()
				return finalize(&AIResponse{
					Content:     cleanContent,
					Confidence:  confidence,
					Source:      local.Source,
					Sentiment:   sentiment,
					Language:    detectedLang,
					Suggestions: suggestions,
					MatchedQA:   local.MatchedQA,
				})
			}
		}
		validatedContent, validatedConf := b.validateResponse(ctx, userID, local.Content, qaPairs, inventory)
		local.Content = validatedContent
		local.Confidence = validatedConf
		b.logger.Info("AI request completed (local fallback)", "traceID", traceID, "duration", time.Since(startTime), "confidence", local.Confidence)
		return finalize(local)
	}

	// No local answer — try intent-category fallback (skip if Groq rate limited)
	if !groqLimited {
		if qa := b.intentCategoryFallback(ctx, userID, userQuery); qa != nil {
			var humanized string
			var err error
			if onChunk != nil {
				humanized, err = b.humanizeStreaming(ctx, qa.Answer, userQuery, nil, nil, recentTurns, onChunk)
			} else {
				humanized, err = b.humanizeResponse(ctx, qa.Answer, userQuery, nil, nil, recentTurns)
			}
			if err == nil && humanized != "" {
				cleanContent, sentiment, detectedLang, suggestions := parseAIMetadata(humanized)
				b.logger.Info("AI request completed (intent match)", "traceID", traceID, "duration", time.Since(startTime), "confidence", confidenceIntentMatch, "sentiment", sentiment, "language", detectedLang, "categoryQA", qa.ID)
				infrastructure.AISentimentTotal.WithLabelValues(sentiment).Inc()
				infrastructure.AILanguageTotal.WithLabelValues(detectedLang).Inc()
				return finalize(&AIResponse{
					Content:     cleanContent,
					Confidence:  confidenceIntentMatch,
					Source:      "intent_match",
					Sentiment:   sentiment,
					Language:    detectedLang,
					Suggestions: suggestions,
					MatchedQA:   &qa.ID,
				})
			}
		}
	}

	// Try lowered semantic search as second fallback (skip if Groq rate limited)
	if !groqLimited {
		if qa := b.semanticFallback(ctx, userID, userQuery); qa != nil {
			var humanized string
			var err error
			if onChunk != nil {
				humanized, err = b.humanizeStreaming(ctx, qa.Answer, userQuery, nil, nil, recentTurns, onChunk)
			} else {
				humanized, err = b.humanizeResponse(ctx, qa.Answer, userQuery, nil, nil, recentTurns)
			}
			if err == nil && humanized != "" {
				cleanContent, sentiment, detectedLang, suggestions := parseAIMetadata(humanized)
				b.logger.Info("AI request completed (semantic fallback)", "traceID", traceID, "duration", time.Since(startTime), "confidence", confidenceSemantic, "sentiment", sentiment, "language", detectedLang, "matchedQA", qa.ID)
				infrastructure.AISentimentTotal.WithLabelValues(sentiment).Inc()
				infrastructure.AILanguageTotal.WithLabelValues(detectedLang).Inc()
				return finalize(&AIResponse{
					Content:     cleanContent,
					Confidence:  confidenceSemantic,
					Source:      "semantic_fallback",
					Sentiment:   sentiment,
					Language:    detectedLang,
					Suggestions: suggestions,
					MatchedQA:   &qa.ID,
				})
			}
		}
	}

	// Really no answer — escalate to admin
	similar := b.findSimilarForUnknown(ctx, userID, userQuery)
	escChannel := ""
	if conv != nil {
		escChannel = conv.Channel
	}
	normalizedQuery := strings.ToLower(strings.TrimSpace(userQuery))
	alreadyPending, checkErr := b.repos.UnknownQ.ExistsPending(ctx, userID, normalizedQuery)
	if checkErr != nil {
		b.logger.Error("Failed to check existing unknown question", "error", checkErr, "conversationID", conversationID)
	}
	if !alreadyPending {
		err := b.repos.UnknownQ.Create(ctx, &domain.UnknownQuestion{
			UserID:         userID,
			Question:       normalizedQuery,
			ConversationID: conversationID,
			Channel:        escChannel,
			Status:         "pending",
		})
		if err != nil {
			b.logger.Error("Failed to create unknown question", "error", err, "conversationID", conversationID)
		}
	}

	escalationMsg := "I don't have that information yet, but I'll escalate this to a human agent who can help you."
	if len(similar) > 0 {
		escalationMsg += "\n\nDid you mean:"
		for i := range similar {
			if i >= maxSimilarSuggestions {
				break
			}
			escalationMsg += "\n• " + similar[i].Question
		}
	}

	notif := &domain.Notification{
		UserID: userID,
		Type:   "unknown_question",
		Title:  "New Unknown Question",
		Body:   fmt.Sprintf("AI could not answer: %q", userQuery),
		Link:   "/teach?tab=unknown",
		IsRead: false,
	}
	if err := b.repos.Notification.Create(ctx, notif); err != nil {
		b.logger.Error("Failed to create notification", "error", err)
	}
	if b.broadcastFn != nil {
		b.broadcastFn(conversationID, "unknown_question", map[string]interface{}{
			"question":        userQuery,
			"conversation_id": conversationID,
			"channel":         escChannel,
			"created_at":      time.Now(),
		})
	}

	b.logger.Info("AI request completed (escalated)", "traceID", traceID, "duration", time.Since(startTime), "similarFound", len(similar))
	return finalize(&AIResponse{
		Content:    escalationMsg,
		Confidence: confidenceEscalation,
		Escalate:   true,
		Reason:     "No matching training data or inventory",
	})
}

// GenerateResponse processes a user query and returns a complete AI response.
func (b *AIBrain) GenerateResponse(ctx context.Context, conversationID, userQuery, language string) (aiResp *AIResponse, aiErr error) {
	return b.generateResponseCore(ctx, conversationID, userQuery, language, nil)
}

// GenerateStreamingResponse processes a user query and streams the final humanization
// step via Groq SSE. The onChunk callback receives tokens as they arrive.
func (b *AIBrain) GenerateStreamingResponse(ctx context.Context, conversationID, userQuery, language string, onChunk func(chunk string)) (aiResp *AIResponse, aiErr error) {
	return b.generateResponseCore(ctx, conversationID, userQuery, language, onChunk)
}
