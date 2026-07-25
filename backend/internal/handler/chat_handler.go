package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"noant/internal/infrastructure"
	"noant/internal/middleware"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	service *service.ChatService
	logger  *infrastructure.Logger
	wsHub   *WebSocketHub
}

func NewChatHandler(svc *service.ChatService, logger *infrastructure.Logger, wsHub *WebSocketHub) *ChatHandler {
	return &ChatHandler{service: svc, logger: logger, wsHub: wsHub}
}

// DirectChat handles one-off AI chat messages outside of a conversation context.
// It creates a temporary conversation, generates an AI response, and returns it
// without persisting to the conversation history.
func (h *ChatHandler) DirectChat(c *gin.Context) {
	var req struct {
		CustomerName string `json:"customer_name"`
		Message      string `json:"message" binding:"required"`
		Channel      string `json:"channel" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	middleware.AddBreadcrumb(c, "chat", "direct_chat request", map[string]interface{}{
		"channel": req.Channel,
		"user_id": userID,
	})
	conv, msg, err := h.service.DirectChat(c.Request.Context(), userID, req.CustomerName, req.CustomerName, req.Message, req.Channel, "")
	if err != nil {
		h.logger.Error("Direct chat failed", "error", err)
		middleware.AddBreadcrumb(c, "chat", "direct_chat failed", map[string]interface{}{
			"error": err.Error(),
		})
		utils.RespondInternalError(c, err.Error())
		return
	}

	middleware.AddBreadcrumb(c, "chat", "direct_chat success", map[string]interface{}{
		"conversation_id": conv.ID,
		"channel":         req.Channel,
	})

	c.JSON(http.StatusOK, gin.H{
		"conversation": conv,
		"message":      msg,
	})
}

func (h *ChatHandler) ClearChats(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	if err := h.service.ClearChats(c.Request.Context(), userID); err != nil {
		h.logger.Error("Clear chats failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chats cleared successfully"})
}

// ListConversations returns a paginated list of conversations for the authenticated user.
// Supports filtering by status (active, resolved, archived) and sorting by last message time.
func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	status := c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	conversations, total, err := h.service.ListConversations(c.Request.Context(), userID, status, page, limit)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	hasMore := page*limit < total

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
		"total":         total,
		"page":          page,
		"limit":         limit,
		"has_more":      hasMore,
	})
}

func (h *ChatHandler) GetConversation(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	if limit < 1 || limit > 100 {
		limit = 30
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	conv, messages, total, err := h.service.GetConversationPaginated(c.Request.Context(), userID, id, limit, offset)
	if err != nil {
		h.logger.Warn("Get conversation failed", "error", err, "conversation_id", id)
		utils.RespondNotFound(c, "Conversation not found")
		return
	}

	hasMore := page*limit < total

	c.JSON(http.StatusOK, gin.H{
		"conversation": conv,
		"messages":     messages,
		"total":        total,
		"has_more":     hasMore,
		"page":         page,
	})
}

// SendMessage sends a user message to an existing conversation and triggers an AI response.
// The AI response is generated asynchronously and broadcast via WebSocket to connected clients.
func (h *ChatHandler) SendMessage(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	// Store customer message
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	_, err := h.service.SendMessage(c.Request.Context(), userID, id, "customer", req.Content)
	if err != nil {
		h.logger.Error("Failed to store message", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	// Broadcast typing indicator - AI is thinking
	if h.wsHub != nil {
		h.wsHub.BroadcastMessage(WebSocketMessage{
			ConversationID: id,
			Type:           "typing_indicator",
			Data: map[string]interface{}{
				"conversation_id": id,
				"is_typing":       true,
			},
		})
	}

	// Generate AI response asynchronously
	go func() {
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error("Panic in AI response goroutine", "recover", r)
			}
		}()
		aiCtx, aiCancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer aiCancel()
		aiMsg, err := h.service.GenerateAIResponse(aiCtx, id, req.Content)
		if err != nil {
			h.logger.Error("AI generation failed in goroutine", "error", err)
			// Remove typing indicator on error
			if h.wsHub != nil {
				h.wsHub.BroadcastMessage(WebSocketMessage{
					ConversationID: id,
					Type:           "typing_indicator",
					Data: map[string]interface{}{
						"conversation_id": id,
						"is_typing":       false,
					},
				})
			}
			return
		}
		if aiMsg == nil {
			if h.wsHub != nil {
				h.wsHub.BroadcastMessage(WebSocketMessage{
					ConversationID: id,
					Type:           "typing_indicator",
					Data: map[string]interface{}{
						"conversation_id": id,
						"is_typing":       false,
					},
				})
			}
			return
		}
		if h.wsHub != nil {
			h.wsHub.BroadcastMessage(WebSocketMessage{
				ConversationID: id,
				Type:           "new_message",
				Data: map[string]interface{}{
					"id":              aiMsg.ID,
					"conversation_id": aiMsg.ConversationID,
					"content":         aiMsg.Content,
					"role":            aiMsg.Role,
					"created_at":      aiMsg.CreatedAt,
					"metadata":        aiMsg.Metadata,
					"confidence":      aiMsg.Confidence,
					"source":          aiMsg.Source,
				},
			})
			// Stop typing indicator when response arrives
			h.wsHub.BroadcastMessage(WebSocketMessage{
				ConversationID: id,
				Type:           "typing_indicator",
				Data: map[string]interface{}{
					"conversation_id": id,
					"is_typing":       false,
				},
			})
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Message sent"})
}

// StreamMessage sends a user message and streams the AI response back via Server-Sent Events (SSE).
// The response is generated token-by-token for real-time display in the frontend.
func (h *ChatHandler) StreamMessage(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	// Store customer message
	_, err := h.service.SendMessage(c.Request.Context(), userID, id, "customer", req.Content)
	if err != nil {
		h.logger.Error("Failed to store message", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	// Broadcast typing indicator
	if h.wsHub != nil {
		h.wsHub.BroadcastMessage(WebSocketMessage{
			ConversationID: id,
			Type:           "typing_indicator",
			Data: map[string]interface{}{
				"conversation_id": id,
				"is_typing":       true,
			},
		})
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	// Stream AI response
	var fullContent string
	aiMsg, err := h.service.GenerateAIStreamingResponse(c.Request.Context(), id, req.Content, func(chunk string) {
		fullContent += chunk
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
		flusher.Flush()
	})

	// Stop typing indicator
	if h.wsHub != nil {
		h.wsHub.BroadcastMessage(WebSocketMessage{
			ConversationID: id,
			Type:           "typing_indicator",
			Data: map[string]interface{}{
				"conversation_id": id,
				"is_typing":       false,
			},
		})
	}

	if err != nil {
		h.logger.Error("AI streaming failed", "error", err)
		_, _ = fmt.Fprintf(c.Writer, "data: [ERROR]\n\n")
		flusher.Flush()
		return
	}

	// Send completion signal with message metadata
	if aiMsg != nil {
		metaJSON, _ := json.Marshal(map[string]interface{}{
			"id":         aiMsg.ID,
			"created_at": aiMsg.CreatedAt,
			"confidence": aiMsg.Confidence,
			"source":     aiMsg.Source,
		})
		_, _ = fmt.Fprintf(c.Writer, "data: [DONE]%s\n\n", metaJSON)
	} else {
		_, _ = fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	}
	flusher.Flush()
}

// HumanTakeover switches a conversation from AI to human agent mode.
// The assigned agent receives real-time notifications via WebSocket.
func (h *ChatHandler) HumanTakeover(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	if err := h.service.HumanTakeover(c.Request.Context(), userID, id, userID); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation taken over by human agent"})
}

// RateConversation records a CSAT (Customer Satisfaction) rating for a conversation.
// Ratings are used in analytics dashboards and Prometheus metrics.
func (h *ChatHandler) RateConversation(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	var req struct {
		Score    int    `json:"score" binding:"required"`
		Feedback string `json:"feedback"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Score is required (1-5)")
		return
	}
	if req.Score < 1 || req.Score > 5 {
		utils.RespondValidationError(c, "Score must be between 1 and 5")
		return
	}
	utils.SanitizeStruct(&req)
	if err := h.service.RateConversation(c.Request.Context(), userID, id, req.Score, req.Feedback); err != nil {
		utils.RespondInternalError(c, "")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Rating submitted"})
}

func (h *ChatHandler) Escalate(c *gin.Context) {
	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	id := c.Param("id")
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	if err := h.service.Escalate(c.Request.Context(), userID, id, req.Reason); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation escalated"})
}
