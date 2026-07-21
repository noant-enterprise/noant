package service

// AI model configuration constants.
const (
	groqModel       = "llama-3.3-70b-versatile"
	groqTemperature = 0.1
	groqMaxTokens   = 500
	groqTopP        = 0.9
)

// Search configuration constants.
const (
	qaSearchLimit         = 6
	inventorySearchLimit  = 3
	similarUnknownLimit   = 3
	conversationTurnLimit = 8
	humanizeHistoryLimit  = 4
	maxSimilarSuggestions = 3
	maxBroadInventory     = 3
)

// Semantic search thresholds.
const (
	semanticSearchThreshold       = 0.65
	semanticFallbackThreshold     = 0.4
	inventorySemanticThreshold    = 0.6
	localAnswerOverlapThreshold   = 0.3
)

// Response confidence scores.
const (
	confidenceGreeting     = 0.98
	confidenceLocalMatch   = 0.9
	confidenceIntentMatch  = 0.75
	confidenceSemantic     = 0.65
	confidenceEscalation   = 0.3
	confidenceHallucinated = 0.1
	confidenceLowLength    = 0.4
	confidenceHumanized    = 0.95
)

// Response validation thresholds.
const (
	maxResponseLength       = 500
	hallucinationPenalty    = 0.7
	longResponsePenalty     = 0.8
	lowConfidenceThreshold  = 0.5
)

// Chat rate limit constants.
const (
	chatRateLimitPulse    = 500
	chatRateLimitUnlimited = 999999
	groqRateLimitPerMin   = 20
	circuitBreakerThreshold = 3
	circuitBreakerRecovery  = 60 // seconds
)

// Conversation history limits.
const (
	maxConversationHistory = 10
	maxNegotiationWords    = 5
	maxDescriptionPreview  = 120
	maxNegotiationDesc     = 80
)
