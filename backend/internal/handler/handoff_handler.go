package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type HandoffHandler struct {
	service *service.HandoffService
	logger  *infrastructure.Logger
}

func NewHandoffHandler(svc *service.HandoffService, logger *infrastructure.Logger) *HandoffHandler {
	return &HandoffHandler{service: svc, logger: logger}
}

func (h *HandoffHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	status := c.Query("status")

	handoffs, err := h.service.List(c.Request.Context(), userID.(string), status)
	if err != nil {
		h.logger.Error("Failed to list handoffs", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"handoffs": handoffs, "count": len(handoffs)})
}

func (h *HandoffHandler) GetByID(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	handoff, err := h.service.GetByID(c.Request.Context(), id, userID.(string))
	if err != nil {
		h.logger.Error("Failed to get handoff", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}
	if handoff == nil {
		utils.RespondNotFound(c, "Handoff not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"handoff": handoff})
}

func (h *HandoffHandler) UpdateStatus(c *gin.Context) {
	var req struct {
		ID         string   `json:"id" binding:"required"`
		Status     string   `json:"status" binding:"required"`
		Notes      string   `json:"notes"`
		FinalPrice *float64 `json:"final_price"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	if err := h.service.UpdateStatus(c.Request.Context(), req.ID, userID.(string), req.Status, req.Notes, req.FinalPrice); err != nil {
		h.logger.Error("Failed to update handoff status", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Handoff updated"})
}
