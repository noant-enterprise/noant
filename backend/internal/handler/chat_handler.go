package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	service *service.ChatService
	logger  *infrastructure.Logger
	wsHub   *WebSocketHub
}

func NewChatHandler(service *service.ChatService, logger *infrastructure.Logger, wsHub *WebSocketHub) *ChatHandler {
	return &ChatHandler{service: service, logger: logger, wsHub: wsHub}
}

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

	userID, _ := c.Get("userID")
	conv, msg, err := h.service.DirectChat(c.Request.Context(), userID.(string), req.CustomerName, req.CustomerName, req.Message, req.Channel, "")
	if err != nil {
		h.logger.Error("Direct chat failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": conv,
		"message":      msg,
	})
}

func (h *ChatHandler) ClearChats(c *gin.Context) {
	userID, _ := c.Get("userID")

	if err := h.service.ClearChats(c.Request.Context(), userID.(string)); err != nil {
		h.logger.Error("Clear chats failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chats cleared successfully"})
}

func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID, _ := c.Get("userID")
	status := c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	conversations, total, err := h.service.ListConversations(c.Request.Context(), userID.(string), status, page, limit)
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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

	conv, messages, total, err := h.service.GetConversationPaginated(c.Request.Context(), userID.(string), id, limit, offset)
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_, err := h.service.SendMessage(c.Request.Context(), userID.(string), id, "customer", req.Content)
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

func (h *ChatHandler) HumanTakeover(c *gin.Context) {
	id := c.Param("id")
	agentID, _ := c.Get("userID")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if err := h.service.HumanTakeover(c.Request.Context(), userID.(string), id, agentID.(string)); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation taken over by human agent"})
}

func (h *ChatHandler) RateConversation(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
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
	if err := h.service.RateConversation(c.Request.Context(), userID.(string), id, req.Score, req.Feedback); err != nil {
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if err := h.service.Escalate(c.Request.Context(), userID.(string), id, req.Reason); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation escalated"})
}
