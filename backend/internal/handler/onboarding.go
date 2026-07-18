package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type OnboardingHandler struct {
	svc    *service.OnboardingService
	logger *infrastructure.Logger
}

func NewOnboardingHandler(svc *service.OnboardingService, logger *infrastructure.Logger) *OnboardingHandler {
	return &OnboardingHandler{svc: svc, logger: logger}
}

func (h *OnboardingHandler) GetStatus(c *gin.Context) {
	userID, _ := c.Get("userID")

	status, err := h.svc.GetStatus(c.Request.Context(), userID.(string))
	if err != nil {
		h.logger.Error("Failed to get onboarding status", "error", err, "user_id", userID)
		utils.RespondInternalError(c, "")
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *OnboardingHandler) CompleteStep(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		Step     string  `json:"step" binding:"required"`
		Industry *string `json:"industry"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Step is required")
		return
	}

	if err := h.svc.CompleteStep(c.Request.Context(), userID.(string), req.Step, req.Industry); err != nil {
		h.logger.Error("Failed to complete onboarding step", "error", err, "user_id", userID, "step", req.Step)
		utils.RespondInternalError(c, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Step completed", "step": req.Step})
}

func (h *OnboardingHandler) AutoCreateCategories(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		IndustryID string `json:"industry_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "Industry ID is required")
		return
	}

	cats, err := h.svc.AutoCreateCategories(c.Request.Context(), userID.(string), req.IndustryID)
	if err != nil {
		h.logger.Error("Failed to auto-create categories", "error", err, "user_id", userID, "industry", req.IndustryID)
		utils.RespondInternalError(c, "")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Categories created",
		"categories": cats,
		"count":      len(cats),
	})
}

func (h *OnboardingHandler) GetIndustryTemplates(c *gin.Context) {
	templates := h.svc.GetIndustryTemplates()
	c.JSON(http.StatusOK, gin.H{"industries": templates})
}
