package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type IntegrationHandler struct {
	service *service.IntegrationService
	logger  *infrastructure.Logger
}

func NewIntegrationHandler(svc *service.IntegrationService, logger *infrastructure.Logger) *IntegrationHandler {
	return &IntegrationHandler{service: svc, logger: logger}
}

func (h *IntegrationHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	integrations, err := h.service.List(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"integrations": integrations})
}

func (h *IntegrationHandler) Connect(c *gin.Context) {
	var req struct {
		Channel string                 `json:"channel" binding:"required"`
		Config  map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	integration, err := h.service.Connect(c.Request.Context(), userID.(string), req.Channel, req.Config)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"integration": integration})
}

func (h *IntegrationHandler) Disconnect(c *gin.Context) {
	channel := c.Param("channel")
	userID, _ := c.Get("userID")
	if err := h.service.Disconnect(c.Request.Context(), userID.(string), channel); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration disconnected"})
}

func (h *IntegrationHandler) Test(c *gin.Context) {
	channel := c.Param("channel")

	// Optionally parse config credentials from the request body (for pre-connect testing)
	var req struct {
		Config map[string]interface{} `json:"config"`
	}
	// Ignore bind errors – the body is optional
	_ = c.ShouldBindJSON(&req)
	utils.SanitizeStruct(&req)

	success, message := h.service.Test(c.Request.Context(), channel, req.Config)

	if success {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": message})
	}
}
