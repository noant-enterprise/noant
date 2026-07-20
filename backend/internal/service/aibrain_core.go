package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"noant/config"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"
)

// ========== AI BRAIN SERVICE ==========

type CircuitBreaker struct {
	failures    int
	lastFailure time.Time
	state       string // closed, open, half-open
	mutex       sync.RWMutex
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	switch cb.state {
	case "open":
		if time.Since(cb.lastFailure) > 60*time.Second {
			cb.state = "half-open"
			cb.failures = 0
			return true
		}
		return false
	case "half-open":
		return true
	default: // closed
		return true
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.failures = 0
	cb.state = "closed"
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= 3 {
		cb.state = "open"
	}
}

type AIBrain struct {
	cfg         *config.Config
	repos       *repository.Repositories
	redis       *infrastructure.RedisClient
	logger      *infrastructure.Logger
	keyIndex    int
	keyMutex    sync.RWMutex
	cb          *CircuitBreaker
	broadcastFn func(convID string, msgType string, data interface{})
	embeddings  *EmbeddingService
	planSvc     *PlanService
	httpClient  *http.Client
}

// NewAIBrain creates the core AI orchestration engine. It manages intent classification,
// 3-tier semantic search (threshold 0.65 → category fallback → threshold 0.4),
// Groq-powered response humanization, sentiment analysis, and circuit breaker
// protection against API failures. The broadcastFn callback pushes real-time
// events to the WebSocket hub.
func NewAIBrain(cfg *config.Config, repos *repository.Repositories, redis *infrastructure.RedisClient, logger *infrastructure.Logger, broadcastFn func(convID string, msgType string, data interface{})) *AIBrain {
	transport := &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	return &AIBrain{
		cfg:         cfg,
		repos:       repos,
		redis:       redis,
		logger:      logger,
		keyIndex:    0,
		cb:          &CircuitBreaker{state: "closed"},
		broadcastFn: broadcastFn,
		embeddings:  NewEmbeddingService(cfg, repos, redis, logger),
		planSvc:     NewPlanService(cfg, repos, redis, logger, NewCreditService(cfg, repos, redis, logger)),
		httpClient:  &http.Client{Transport: transport, Timeout: 30 * time.Second},
	}
}

func (b *AIBrain) getNextAPIKey() string {
	b.keyMutex.Lock()
	defer b.keyMutex.Unlock()
	if len(b.cfg.GroqAPIKeys) == 0 {
		return ""
	}
	key := b.cfg.GroqAPIKeys[b.keyIndex]
	b.keyIndex = (b.keyIndex + 1) % len(b.cfg.GroqAPIKeys)
	return key
}

// LangChain-style prompt template
type PromptTemplate struct {
	SystemPrompt string
	Context      []MessageTurn
	UserQuery    string
	Language     string
	Tone         string
}

type MessageTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (b *AIBrain) BuildPrompt(template *PromptTemplate) []MessageTurn {
	var messages []MessageTurn
	systemContent := template.SystemPrompt
	if systemContent == "" {
		now := time.Now()
		systemContent = fmt.Sprintf(`You are a world-class customer service and sales assistant for a real business. You speak naturally, with emotional intelligence, adapting to the customer's tone and language.

DATE/TIME: %s

## CORE IDENTITY
- You are a real shop/customer service representative. Not an AI, not a bot.
- You represent the business with professionalism and warmth.
- You speak Nigerian English, Pidgin, Yoruba, Igbo, or Hausa naturally — matching the customer's language.

## EMOTIONAL INTELLIGENCE & SENTIMENT
Read the customer's sentiment and adapt your tone:
- 😊 **Positive/Neutral**: Warm, helpful, professional. Reinforce their good mood.
- 😟 **Negative/Frustrated**: Empathetic first ("I understand how you feel"), apologetic if needed, then solution-oriented. NEVER dismiss their concern.
- 😠 **Angry**: Stay calm. Acknowledge their frustration. Offer a concrete solution or escalate immediately.
- ❓ **Confused**: Patient, clear, step-by-step guidance. Ask clarifying questions.
- 🛒 **Buying-ready**: Enthusiastic, confirm details, guide to checkout/payment quickly.

## LANGUAGE AUTO-DETECTION
- Detect the customer's language from their message. Respond in the SAME language.
- Supported: English, Nigerian Pidgin ("How you dey?", "Wetin dey happen?"), Yoruba ("Bawo ni?", "E kaaro"), Igbo ("Kedu?", "Ndewo"), Hausa ("Sannu?", "Yaya kake?").
- Mix languages naturally the way real Nigerians speak (e.g., "Oya let's do business").
- Use Naira (₦) always.

## CONVERSATIONAL EXCELLENCE
1. **Listen first**: Understand before answering. Ask one question at a time.
2. **Be concise**: Like a real person texting — 2-3 sentences max per turn.
3. **Show personality**: Use "o", "o ya", "abi", "shebi", "na", "o ga" naturally.
4. **Use customer's name** if known.
5. **Never be robotic**: No numbered lists, no bullet points in casual chat.
6. **Proactive suggestions**: If customer seems interested in a product, offer relevant alternatives or upsells.
7. **Objection handling**: Price too high? Offer value. Need time? "No wahala, take your time o."
8. **Close naturally**: When they want to buy, guide them to the owner's WhatsApp.

## PRODUCT & INVENTORY RULES
- ONLY sell products from the "Available products/services" context below. NEVER invent products or prices.
- If no products match, say "Let me check with Oga" and escalate.
- Never quote below the listed Min price.
- Know stock levels — if low stock, mention it ("Only 2 left o!").
- For price negotiations, offer small discounts when possible, or say "Na the last price, I no fit go lower."

## CRITICAL RULES — NEVER VIOLATE
- NEVER say "I'm an AI", "as an AI", "I'm a bot", "Noant", or any platform name.
- NEVER make up answers outside the provided training data and inventory.
- NEVER hallucinate prices, product features, or policies.
- NEVER send account numbers or handle payments directly.
- NEVER say "based on my training data", "generally speaking", "typically".
- NEVER escalate unnecessarily — try to resolve first.
- If you cannot answer, say "Let me check with my manager" and escalate.

## HANDOFF PROTOCOL
When the customer wants to buy, escalate, or needs human help:
- "Perfect! Let me connect you with Oga to finalize this."
- "I don't have that info, but my manager can help you with that."
- Generate a brief summary of the conversation for the human agent.

## SENTIMENT ANALYSIS (for your response metadata)
At the end of your response, include a sentiment tag on a new line like:
[SENTIMENT:positive|negative|neutral|frustrated]
[LANGUAGE:en|pcm|yo|ha|ig]
[SUGGESTIONS:option 1|option 2|option 3]

Example:
"Good morning! How can I help you today?
[SENTIMENT:neutral]
[LANGUAGE:en]
[SUGGESTIONS:I want to buy something|I need help with an order|Tell me about your products]"`, now.Format("Monday, January 2, 2006 3:04 PM"))
	}
	messages = append(messages, MessageTurn{Role: "system", Content: systemContent})
	if len(template.Context) > 0 {
		messages = append(messages, template.Context...)
	}
	messages = append(messages, MessageTurn{Role: "user", Content: template.UserQuery})
	return messages
}

// classifyIntent uses LLM to classify user intent (replaces keyword matching)
func (b *AIBrain) classifyIntent(ctx context.Context, query string) string {
	// Fast path: check for clear handoff signals first (no LLM needed)
	lower := strings.ToLower(query)
	handoffTriggers := []string{"i want to buy", "i'll take it", "place the order", "checkout", "pay now", "send account number", "how do i pay", "send me your account"}
	for _, t := range handoffTriggers {
		if strings.Contains(lower, t) {
			return "handoff"
		}
	}

	salesTriggers := []string{"price", "cost", "how much", "available", "do you have", "show me", "recommend", "suggest", "discount", "cheaper", "stock", "product", "service", "package", "what do you sell"}
	for _, t := range salesTriggers {
		if strings.Contains(lower, t) {
			return "sales"
		}
	}

	supportTriggers := []string{"complaint", "problem", "issue", "refund", "return", "help", "support", "order status", "not working", "broken", "delay"}
	for _, t := range supportTriggers {
		if strings.Contains(lower, t) {
			return "support"
		}
	}

	// Use LLM for ambiguous cases
	prompt := []MessageTurn{
		{Role: "system", Content: `You are an intent classifier for a shop that sells products and services.
Classify the customer message into ONE intent:

- "sales": customer wants to BUY something, asks about products, prices, stock, availability, services, says "i need", "i want", "do you have", "show me", "how much"
- "handoff": customer is READY TO PAY, says "i want to buy", "deal", "checkout", "pay now", "send account"
- "support": customer has a complaint, asks about order status, return policy, or general help NOT about buying

IMPORTANT: "i need X" or "i want X" = sales (they want to buy).
Reply with ONLY the intent word.`},
		{Role: "user", Content: query},
	}

	response, _, err := b.callGroqWithFallback(ctx, prompt)
	if err != nil {
		// Fallback to keyword detection if LLM fails
		return b.keywordIntent(query)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if strings.Contains(response, "handoff") {
		return "handoff"
	}
	if strings.Contains(response, "sales") {
		return "sales"
	}
	return "support"
}

// keywordIntent is the fallback when LLM is unavailable
func (b *AIBrain) keywordIntent(query string) string {
	lower := strings.ToLower(query)
	salesTriggers := []string{"i need", "i want", "price", "how much", "cost", "stock", "available", "do you have", "what do you have", "show me", "product", "package", "do you sell", "discount", "cheaper", "last price", "service"}
	for _, t := range salesTriggers {
		if strings.Contains(lower, t) {
			return "sales"
		}
	}
	return "support"
}

func (b *AIBrain) isGreetingQuery(query string) bool {
	lower := strings.ToLower(strings.TrimSpace(query))
	if lower == "" {
		return false
	}

	greetings := []string{
		"hi",
		"hello",
		"hey",
		"good morning",
		"good afternoon",
		"good evening",
		"how are you",
	}
	for _, greeting := range greetings {
		if lower == greeting {
			return true
		}
	}

	words := strings.Fields(lower)
	if len(words) <= 2 {
		for _, greeting := range []string{"hi", "hello", "hey"} {
			if strings.Contains(lower, greeting) {
				return true
			}
		}
	}

	return false
}

func (b *AIBrain) searchKnowledgeBase(ctx context.Context, userID, query string, limit int) []domain.QAPair {
	// Try semantic search first (embeddings)
	if b.embeddings != nil {
		results, err := b.embeddings.SemanticSearchQAPairs(ctx, userID, query, limit, 0.65)
		if err == nil && len(results) > 0 {
			b.logger.Info("Semantic search found matches", "query", query, "results", len(results), "topScore", results[0].Score)
			qas := make([]domain.QAPair, len(results))
			for i, r := range results {
				qas[i] = domain.QAPair{
					ID:       r.ID,
					Question: r.Question,
					Answer:   r.Answer,
				}
			}
			return qas
		}
		if err != nil {
			b.logger.Warn("Semantic search failed, falling back to keyword", "error", err)
		}
	}

	// Fallback: keyword search (SQL LIKE)
	results, err := b.repos.QAPair.Search(ctx, userID, query)
	if err != nil {
		b.logger.Error("Failed to search Q&A pairs", "error", err)
		return nil
	}

	seen := make(map[string]struct{})
	merged := make([]domain.QAPair, 0, limit)
	for i := range results {
		if _, ok := seen[results[i].ID]; ok {
			continue
		}
		seen[results[i].ID] = struct{}{}
		merged = append(merged, results[i])
		if len(merged) >= limit {
			return merged
		}
	}

	// Word-by-word fallback
	for _, word := range strings.Fields(strings.ToLower(query)) {
		word = strings.Trim(word, "?!.,;:")
		if len(word) < 4 {
			continue
		}
		more, err := b.repos.QAPair.Search(ctx, userID, word)
		if err != nil {
			continue
		}
		for i := range more {
			if _, ok := seen[more[i].ID]; ok {
				continue
			}
			seen[more[i].ID] = struct{}{}
			merged = append(merged, more[i])
			if len(merged) >= limit {
				return merged
			}
		}
	}

	return merged
}

func (b *AIBrain) searchInventoryContext(ctx context.Context, userID, query string, limit int) []domain.InventoryItem {
	// Try semantic search first
	if b.embeddings != nil {
		results, err := b.embeddings.SemanticSearchInventory(ctx, userID, query, limit, 0.6)
		if err == nil && len(results) > 0 {
			var items []domain.InventoryItem
			for _, r := range results {
				item, err := b.repos.Inventory.GetByID(ctx, r.ID, userID)
				if err == nil && item != nil {
					items = append(items, *item)
				}
			}
			if len(items) > 0 {
				return items
			}
		}
	}

	// Fallback: keyword search
	items, err := b.repos.Inventory.Search(ctx, userID, query)
	if err != nil {
		b.logger.Error("Failed to search inventory", "error", err)
		return nil
	}

	seen := make(map[string]struct{})
	merged := make([]domain.InventoryItem, 0, limit)
	for i := range items {
		if _, ok := seen[items[i].ID]; ok {
			continue
		}
		seen[items[i].ID] = struct{}{}
		merged = append(merged, items[i])
		if len(merged) >= limit {
			return merged
		}
	}

	for _, word := range strings.Fields(strings.ToLower(query)) {
		word = strings.Trim(word, "?!.,;:")
		if len(word) < 4 {
			continue
		}
		more, err := b.repos.Inventory.Search(ctx, userID, word)
		if err != nil {
			continue
		}
		for i := range more {
			if _, ok := seen[more[i].ID]; ok {
				continue
			}
			seen[more[i].ID] = struct{}{}
			merged = append(merged, more[i])
			if len(merged) >= limit {
				return merged
			}
		}
	}

	return merged
}

// qaWordOverlap checks whether meaningful words in the query appear in the QA question.
// Returns 0.0-1.0 fraction of query words matched. Short queries (1-2 words) pass automatically.
func qaWordOverlap(query, question string) float64 {
	qWords := strings.Fields(strings.ToLower(query))
	if len(qWords) <= 2 {
		return 1.0 // short queries are direct intent signals
	}
	var meaningful []string
	for _, w := range qWords {
		w = strings.Trim(w, "?!.,;:")
		if len(w) < 3 {
			continue
		}
		meaningful = append(meaningful, w)
	}
	if len(meaningful) == 0 {
		return 1.0
	}
	qLower := strings.ToLower(question)
	matched := 0
	for _, w := range meaningful {
		if strings.Contains(qLower, w) {
			matched++
		}
	}
	return float64(matched) / float64(len(meaningful))
}

func (b *AIBrain) localPlatformAnswer(userID, query string, qaPairs []domain.QAPair, inventory []domain.InventoryItem) *AIResponse {
	lower := strings.ToLower(query)

	// Check if this is a negotiation follow-up (price-related short queries)
	isNegotiation := len(strings.Fields(query)) <= 5 && (strings.Contains(lower, "price") || strings.Contains(lower, "cheap") || strings.Contains(lower, "discount") || strings.Contains(lower, "last") || strings.Contains(lower, "negotiate") || strings.Contains(lower, "lower") || strings.Contains(lower, "reduce") || strings.Contains(lower, "offer"))

	// Training data takes priority — but only if the match is relevant
	if len(qaPairs) > 0 && qaWordOverlap(query, qaPairs[0].Question) >= 0.3 {
		first := qaPairs[0]
		content := strings.TrimSpace(first.Answer)
		if content == "" {
			content = "Let me know if you want me to explain anything else."
		} else if len(strings.Fields(query)) > 3 && !strings.HasSuffix(content, "?") {
			content += " If you want, I can also help you with the next best option."
		}
		return &AIResponse{
			Content:    content,
			Confidence: 0.9,
			MatchedQA:  &first.ID,
			Source:     "training",
		}
	}
	// Inventory is second priority
	if len(inventory) > 0 {
		if len(inventory) > 1 && isBroadInventoryQuery(query) && !isNegotiation {
			var lines []string
			for i := range inventory {
				if i >= 3 {
					break
				}
				lines = append(lines, "- "+inventorySummaryLine(&inventory[i]))
			}
			return &AIResponse{
				Content:    fmt.Sprintf("I found a few good options for you:\n%s\n\nWhich one should I break down for you?", strings.Join(lines, "\n")),
				Confidence: 0.9,
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
		// If it's a negotiation, handle it naturally
		if isNegotiation {
			if item.MinPrice != nil && *item.MinPrice > 0 && *item.MinPrice < item.Price {
				discount := item.Price - *item.MinPrice
				return &AIResponse{
					Content:    fmt.Sprintf("I can do ₦%.0f for you - that's ₦%.0f off. Best price I can offer. %s", *item.MinPrice, discount, stock),
					Confidence: 0.9,
					Source:     "inventory",
				}
			}
			return &AIResponse{
				Content:    fmt.Sprintf("₦%.0f is already my best price, o! %s", item.Price, stock),
				Confidence: 0.9,
				Source:     "inventory",
			}
		}
		// Normal product presentation
		answer := fmt.Sprintf("%s - ₦%.0f.%s", item.Name, item.Price, stock)
		if item.Description != "" {
			answer += " " + trimTurnText(item.Description, 120)
		}
		answer += " " + softCloseLine(query)
		return &AIResponse{
			Content:    answer,
			Confidence: 0.9,
			Source:     "inventory",
		}
	}
	// Nothing found
	return nil
}

// validateResponse checks if the AI response hallucinates prices or products
func (b *AIBrain) validateResponse(ctx context.Context, _, response string, _ []domain.QAPair, inventory []domain.InventoryItem) (resp string, conf float64) {
	if response == "" {
		return response, 0
	}

	lower := strings.ToLower(response)
	confidence := 0.95

	// If response contains phrases that suggest hallucination
	hallucinationSignals := []string{"according to my training", "based on my knowledge", "generally speaking", "typically"}
	for _, signal := range hallucinationSignals {
		if strings.Contains(lower, signal) {
			confidence *= 0.7
			break
		}
	}

	// If response is too long, it might be hallucinating
	if len(response) > 500 {
		confidence *= 0.8
	}

	// Extract price claims (e.g. "₦1,500", "₦50000") and validate against inventory
	if len(inventory) > 0 {
		re := regexp.MustCompile(`₦\s*([\d,]+(?:\.\d{1,2})?)`)
		matches := re.FindAllStringSubmatch(response, -1)
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
				return "I'm sorry, I don't have accurate pricing information for that. Please contact our sales team for exact prices.", 0.1
			}
		}
	}

	if confidence < 0.5 {
		return "I'm not confident about the answer to that. Let me transfer you to a human agent.", confidence
	}

	return response, confidence
}

// findSimilarForUnknown suggests similar Q&As when escalating an unknown question
func (b *AIBrain) findSimilarForUnknown(ctx context.Context, userID, query string) []domain.QAPair {
	if b.embeddings == nil {
		return nil
	}
	similar, _ := b.embeddings.FindSimilarQA(ctx, userID, query, 3)
	return similar
}

// allowGroqCall checks if the user has remaining Groq API calls in the current window.
// Returns true if the call is allowed, false if rate limited.
func (b *AIBrain) allowGroqCall(ctx context.Context, userID string) bool {
	if b.redis == nil || userID == "" {
		return true
	}
	allowed, err := b.redis.RateLimit(ctx, "groq_rate:"+userID, 20, time.Minute)
	if err != nil {
		b.logger.Warn("Groq rate limit check failed, allowing call", "error", err)
		return true
	}
	if !allowed {
		b.logger.Warn("Groq rate limit exceeded for user", "userID", userID)
		infrastructure.NoantGroqRateLimited.Inc()
	}
	return allowed
}

// intentCategoryFallback uses LLM to classify a query into one of the user's categories,
// then returns the best QA pair from that category. Returns nil if no category matches.
func (b *AIBrain) intentCategoryFallback(ctx context.Context, userID, userQuery string) *domain.QAPair {
	categories, err := b.repos.Category.List(ctx, userID)
	if err != nil || len(categories) == 0 {
		return nil
	}

	catNames := make([]string, len(categories))
	catMap := make(map[string]domain.Category)
	for i, cat := range categories {
		catNames[i] = cat.Name
		catMap[strings.ToLower(strings.TrimSpace(cat.Name))] = cat
	}

	prompt := []MessageTurn{
		{Role: "system", Content: fmt.Sprintf(`You are a strict query classifier. Your only job is to output the name of exactly one category from the list below, or the word NONE.

Customer question: "%s"

Categories:
%s

Rules:
- The question MUST logically belong to the category based on what the customer is asking about
- Do NOT guess. If no category clearly matches, output NONE
- Output ONLY the exact category name or NONE
- No punctuation, no explanation, no extra characters`, userQuery, strings.Join(catNames, ", "))},
		{Role: "user", Content: userQuery},
	}

	response, _, err := b.callGroqWithFallback(ctx, prompt)
	if err != nil {
		return nil
	}

	classified := strings.TrimSpace(response)
	classified = strings.TrimRight(classified, ".,;:!? \n\t")

	cat, ok := catMap[strings.ToLower(classified)]
	if !ok {
		return nil
	}

	qas, err := b.repos.QAPair.ListByCategoryAndUser(ctx, cat.ID, userID)
	if err != nil || len(qas) == 0 {
		return nil
	}

	return &qas[0]
}

// semanticFallback searches QA pairs with a lowered embedding threshold (0.4) to catch
// queries that are semantically related but didn't match the main search threshold.
func (b *AIBrain) semanticFallback(ctx context.Context, userID, userQuery string) *domain.QAPair {
	if b.embeddings == nil {
		return nil
	}
	results, err := b.embeddings.SemanticSearchQAPairs(ctx, userID, userQuery, 1, 0.4)
	if err != nil || len(results) == 0 {
		return nil
	}
	qa, err := b.repos.QAPair.GetByID(ctx, results[0].ID)
	if err != nil || qa == nil {
		return nil
	}
	return qa
}

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

// parseAIMetadata extracts [SENTIMENT], [LANGUAGE], [SUGGESTIONS] tags from AI response
func parseAIMetadata(content string) (clean, sentiment, language string, suggestions []string) {
	clean = content
	sentiment = "neutral"
	language = "en"

	re := regexp.MustCompile(`(?m)^\[SENTIMENT:([^\]]+)\]$`)
	if m := re.FindStringSubmatch(clean); len(m) > 1 {
		sentiment = strings.ToLower(strings.TrimSpace(m[1]))
		clean = re.ReplaceAllString(clean, "")
	}

	re = regexp.MustCompile(`(?m)^\[LANGUAGE:([^\]]+)\]$`)
	if m := re.FindStringSubmatch(clean); len(m) > 1 {
		language = strings.ToLower(strings.TrimSpace(m[1]))
		clean = re.ReplaceAllString(clean, "")
	}

	re = regexp.MustCompile(`(?m)^\[SUGGESTIONS:([^\]]+)\]$`)
	if m := re.FindStringSubmatch(clean); len(m) > 1 {
		for _, s := range strings.Split(m[1], "|") {
			s = strings.TrimSpace(s)
			if s != "" {
				suggestions = append(suggestions, s)
			}
		}
		clean = re.ReplaceAllString(clean, "")
	}

	clean = strings.TrimSpace(clean)
	return clean, sentiment, language, suggestions
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

func (b *AIBrain) GenerateResponse(ctx context.Context, conversationID, userQuery, language string) (aiResp *AIResponse, aiErr error) {
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
		infrastructure.AICallsTotal.WithLabelValues("llama-3.3-70b-versatile", status).Inc()
		infrastructure.AIDuration.WithLabelValues("llama-3.3-70b-versatile").Observe(duration)
	}()
	conv, err := b.repos.Conversation.GetByID(ctx, conversationID)
	if err != nil {
		b.logger.Warn("Failed to get conversation", "error", err)
	}
	userID := ""
	var user *domain.User
	if conv != nil {
		userID = conv.UserID
	}

	// Request trace
	traceID := fmt.Sprintf("trace_%d", time.Now().UnixNano())
	channel := ""
	if conv != nil {
		channel = conv.Channel
	}
	b.logger.Info("AI request started", "traceID", traceID, "userID", userID, "query", userQuery, "channel", channel)
	recentTurns := b.recentConversationTurns(ctx, conversationID, userQuery, 8)
	finalize := func(resp *AIResponse) (*AIResponse, error) {
		if resp != nil {
			if err := b.storeConversationTurn(ctx, conversationID, userQuery, resp.Content); err != nil {
				b.logger.Warn("Failed to store conversation turn", "error", err)
			}
		}
		return resp, nil
	}

	// Plan limit check (before intent classification)
	if userID != "" {
		if user == nil {
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

	// Per-user Groq rate limit check — if exceeded, skip Groq calls and use raw training data only
	groqLimited := !b.allowGroqCall(ctx, userID)
	if groqLimited {
		b.logger.Warn("Groq rate limited for user, using local-only mode", "traceID", traceID, "userID", userID)
	}

	// Short greetings should get a local response instead of burning an AI call.
	if b.isGreetingQuery(userQuery) {
		b.logger.Info("AI request completed (greeting)", "traceID", traceID, "duration", time.Since(startTime))
		return finalize(&AIResponse{
			Content:     "Hi! How can I help you today?",
			Confidence:  0.98,
			Sentiment:   "neutral",
			Language:    "en",
			Suggestions: []string{"I want to buy something", "I need help", "Tell me about your products"},
		})
	}

	// LLM-based intent classification (with keyword fallback)
	intent := b.classifyIntent(ctx, userQuery)
	b.logger.Info("Intent classified", "traceID", traceID, "intent", intent, "query", userQuery)

	// Handle handoff intent
	if intent == "handoff" {
		resp, err := b.handleHandoff(ctx, conversationID, userID, userQuery)
		b.logger.Info("AI request completed (handoff)", "traceID", traceID, "duration", time.Since(startTime))
		if err == nil && resp != nil {
			if err := b.storeConversationTurn(ctx, conversationID, userQuery, resp.Content); err != nil {
				b.logger.Warn("Failed to store conversation turn", "error", err)
			}
		}
		return resp, err
	}

	// Handle sales intent — search inventory first
	if intent == "sales" {
		resp, err := b.handleSalesMode(ctx, userID, conv, userQuery, language, recentTurns)
		b.logger.Info("AI request completed (sales)", "traceID", traceID, "duration", time.Since(startTime))
		if err == nil && resp != nil {
			if err := b.storeConversationTurn(ctx, conversationID, userQuery, resp.Content); err != nil {
				b.logger.Warn("Failed to store conversation turn", "error", err)
			}
		}
		return resp, err
	}

	// Default: support mode — search training data AND inventory
	qaPairs := b.searchKnowledgeBase(ctx, userID, userQuery, 6)
	inventory := b.searchInventoryContext(ctx, userID, userQuery, 3)

	b.logger.Info("Search completed", "traceID", traceID, "qaMatches", len(qaPairs), "inventoryMatches", len(inventory))

	// Try local answer from training data / inventory first
	local := b.localPlatformAnswer(userID, userQuery, qaPairs, inventory)
	if local != nil {
		if !groqLimited {
			// Humanize via Groq — make the training data answer sound natural, not bot-like
			humanized, err := b.humanizeResponse(ctx, local.Content, userQuery, qaPairs, inventory, recentTurns)
			if err == nil && humanized != "" {
				cleanContent, sentiment, detectedLang, suggestions := parseAIMetadata(humanized)
				confidence := local.Confidence * 0.95
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
		// Humanization skipped or failed — fall back to raw training data answer
		validatedContent, validatedConf := b.validateResponse(ctx, userID, local.Content, qaPairs, inventory)
		local.Content = validatedContent
		local.Confidence = validatedConf
		b.logger.Info("AI request completed (local fallback)", "traceID", traceID, "duration", time.Since(startTime), "confidence", local.Confidence)
		return finalize(local)
	}

	// No local answer — try intent-category fallback before escalating (skip if Groq rate limited)
	if !groqLimited {
		if qa := b.intentCategoryFallback(ctx, userID, userQuery); qa != nil {
			humanized, err := b.humanizeResponse(ctx, qa.Answer, userQuery, nil, nil, recentTurns)
			if err == nil && humanized != "" {
				cleanContent, sentiment, detectedLang, suggestions := parseAIMetadata(humanized)
				confidence := 0.75
				b.logger.Info("AI request completed (intent match)", "traceID", traceID, "duration", time.Since(startTime), "confidence", confidence, "sentiment", sentiment, "language", detectedLang, "categoryQA", qa.ID)
				infrastructure.AISentimentTotal.WithLabelValues(sentiment).Inc()
				infrastructure.AILanguageTotal.WithLabelValues(detectedLang).Inc()
				return finalize(&AIResponse{
					Content:     cleanContent,
					Confidence:  confidence,
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
			humanized, err := b.humanizeResponse(ctx, qa.Answer, userQuery, nil, nil, recentTurns)
			if err == nil && humanized != "" {
				cleanContent, sentiment, detectedLang, suggestions := parseAIMetadata(humanized)
				confidence := 0.65
				b.logger.Info("AI request completed (semantic fallback)", "traceID", traceID, "duration", time.Since(startTime), "confidence", confidence, "sentiment", sentiment, "language", detectedLang, "matchedQA", qa.ID)
				infrastructure.AISentimentTotal.WithLabelValues(sentiment).Inc()
				infrastructure.AILanguageTotal.WithLabelValues(detectedLang).Inc()
				return finalize(&AIResponse{
					Content:     cleanContent,
					Confidence:  confidence,
					Source:      "semantic_fallback",
					Sentiment:   sentiment,
					Language:    detectedLang,
					Suggestions: suggestions,
					MatchedQA:   &qa.ID,
				})
			}
		}
	}

	// Really no answer — escalate to admin (do NOT let Groq answer from its own knowledge)
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
			if i >= 3 {
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
		Confidence: 0.3,
		Escalate:   true,
		Reason:     "No matching training data or inventory",
	})
}

type AIResponse struct {
	Content     string            `json:"content"`
	Confidence  float64           `json:"confidence"`
	Escalate    bool              `json:"escalate"`
	Source      string            `json:"source,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	MatchedQA   *string           `json:"matched_qa,omitempty"`
	Sentiment   string            `json:"sentiment,omitempty"`   // positive, negative, neutral, frustrated
	Language    string            `json:"language,omitempty"`    // en, yo, ha, ig, pcm
	Suggestions []string          `json:"suggestions,omitempty"` // quick action chips
	Summary     string            `json:"summary,omitempty"`     // conversation summary (set on handoff)
}

func (b *AIBrain) callGroqWithFallback(ctx context.Context, messages []MessageTurn) (content string, confidence float64, err error) {
	if !b.cb.Allow() {
		return "", 0, fmt.Errorf("circuit breaker open: Groq API temporarily unavailable")
	}
	apiKey := b.getNextAPIKey()
	if apiKey == "" {
		b.cb.RecordFailure()
		return "", 0, fmt.Errorf("no Groq API keys configured")
	}
	payload := map[string]interface{}{
		"model":       "llama-3.3-70b-versatile",
		"messages":    messages,
		"temperature": 0.1,
		"max_tokens":  500,
		"top_p":       0.9,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.httpClient.Do(req)
	if err != nil {
		b.cb.RecordFailure()
		return "", 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		b.cb.RecordFailure()
		return "", 0, fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		b.cb.RecordFailure()
		snippet := string(body)
		if len(snippet) > 500 {
			snippet = snippet[:500]
		}
		return "", 0, fmt.Errorf("groq API error: %s - %s", resp.Status, snippet)
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		b.cb.RecordFailure()
		return "", 0, err
	}
	if len(result.Choices) == 0 {
		b.cb.RecordFailure()
		return "", 0, fmt.Errorf("no response from Groq")
	}
	content = result.Choices[0].Message.Content
	b.cb.RecordSuccess()
	confidence = 0.85
	if result.Choices[0].FinishReason != "stop" {
		confidence = 0.5
	}
	if result.Usage.CompletionTokens < 10 {
		confidence = 0.4
	}
	return content, confidence, nil
}

func (b *AIBrain) getConversationHistory(ctx context.Context, conversationID string) ([]MessageTurn, error) {
	if b.redis == nil {
		return nil, nil
	}
	key := fmt.Sprintf("conv:%s:history", conversationID)
	historyJSON, err := b.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var history []MessageTurn
	if err := json.Unmarshal([]byte(historyJSON), &history); err != nil {
		return nil, err
	}
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	return history, nil
}

func (b *AIBrain) storeConversationTurn(ctx context.Context, conversationID, userQuery, aiResponse string) error {
	if b.redis == nil {
		return nil
	}
	history, err := b.getConversationHistory(ctx, conversationID)
	if err != nil {
		b.logger.Warn("Failed to get conversation history for storage", "error", err)
	}
	history = append(history,
		MessageTurn{Role: "user", Content: userQuery},
		MessageTurn{Role: "assistant", Content: aiResponse},
	)
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	historyJSON, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}
	return b.redis.Set(ctx, fmt.Sprintf("conv:%s:history", conversationID), string(historyJSON), b.cfg.RedisShortTTL)
}
