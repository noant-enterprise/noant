package handler

import (
	"errors"
	"net/http"

	"noant/internal/infrastructure"
	apperrors "noant/internal/errors"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

// CampaignHandler handles campaign mode related endpoints
type CampaignHandler struct {
	campaignSvc *service.CampaignService
	logger      *infrastructure.Logger
}

func NewCampaignHandler(campaignSvc *service.CampaignService, logger *infrastructure.Logger) *CampaignHandler {
	return &CampaignHandler{
		campaignSvc: campaignSvc,
		logger:      logger,
	}
}

// List returns all campaigns for the current user
func (h *CampaignHandler) List(c *gin.Context) {
	userID := getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	campaigns, err := h.campaignSvc.List(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list campaigns", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list campaigns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaigns": campaigns,
	})
}

// Create creates a new campaign schedule
func (h *CampaignHandler) Create(c *gin.Context) {
	userID := getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	var req service.CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	utils.SanitizeStruct(&req)

	campaign, err := h.campaignSvc.Create(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("Failed to create campaign", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create campaign"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"campaign": campaign,
	})
}

// Cancel cancels a campaign by ID
func (h *CampaignHandler) Cancel(c *gin.Context) {
	userID := getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign ID is required"})
		return
	}

	if err := h.campaignSvc.Cancel(c.Request.Context(), id, userID); err != nil {
		h.logger.Error("Failed to cancel campaign", "error", err, "userID", userID, "campaignID", id)
		if errors.Is(err, apperrors.ErrCampaign) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel campaign"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
