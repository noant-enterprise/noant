package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

// CreditHandler handles credit and billing related endpoints
type CreditHandler struct {
	creditSvc *service.CreditService
	planSvc   *service.PlanService
	logger    *infrastructure.Logger
}

func NewCreditHandler(creditSvc *service.CreditService, planSvc *service.PlanService, logger *infrastructure.Logger) *CreditHandler {
	return &CreditHandler{
		creditSvc: creditSvc,
		planSvc:   planSvc,
		logger:    logger,
	}
}

// GetBalance returns the user's current credit balance and expiry
func (h *CreditHandler) GetBalance(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	balance, err := h.creditSvc.GetBalance(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get credit balance", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get credit balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance":      balance.Balance,
		"expires_at":   balance.ExpiresAt,
		"last_updated": balance.LastUpdatedAt,
	})
}

// GetLimits returns the current plan limits for the user
func (h *CreditHandler) GetLimits(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	limits, err := h.planSvc.GetLimitsByUserID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get plan limits", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get plan limits"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"limits": limits,
	})
}

// PurchasePack initiates a credit pack purchase by returning a Polar checkout URL
func (h *CreditHandler) PurchasePack(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	var req struct {
		PackType string `json:"pack_type" binding:"required,oneof=small medium large"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}
	utils.SanitizeStruct(&req)

	checkoutURL, err := h.creditSvc.PurchasePack(c.Request.Context(), userID, req.PackType)
	if err != nil {
		h.logger.Error("Failed to get checkout URL", "error", err, "userID", userID, "packType", req.PackType)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate purchase"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": checkoutURL,
	})
}

// GetHistory returns the user's credit purchase history
func (h *CreditHandler) GetHistory(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	history, err := h.creditSvc.GetPurchaseHistory(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get purchase history", "error", err, "userID", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get purchase history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
	})
}
