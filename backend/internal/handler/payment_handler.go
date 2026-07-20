package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	service *service.PaymentService
	logger  *infrastructure.Logger
}

func NewPaymentHandler(svc *service.PaymentService, logger *infrastructure.Logger) *PaymentHandler {
	return &PaymentHandler{service: svc, logger: logger}
}

func (h *PaymentHandler) ListPlans(c *gin.Context) {
	plans, err := h.service.ListPlans(c.Request.Context())
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *PaymentHandler) Subscribe(c *gin.Context) {
	// Accept both 'plan_id' and 'plan' keys for frontend compatibility
	var req struct {
		PlanID   string `json:"plan_id"`
		Plan     string `json:"plan"`
		Currency string `json:"currency"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	// Prefer plan_id, fall back to plan
	planID := req.PlanID
	if planID == "" {
		planID = req.Plan
	}
	if planID == "" {
		utils.RespondValidationError(c, "plan or plan_id is required")
		return
	}

	userID, _ := c.Get("userID")
	checkoutURL, err := h.service.Subscribe(c.Request.Context(), userID.(string), planID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	resp := gin.H{"message": "Subscription initiated"}
	if checkoutURL != "" {
		resp["checkout_url"] = checkoutURL
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PaymentHandler) Webhook(c *gin.Context) {
	payload, _ := c.GetRawData()

	headers := map[string]string{
		"webhook-id":        c.GetHeader("webhook-id"),
		"webhook-timestamp": c.GetHeader("webhook-timestamp"),
		"webhook-signature": c.GetHeader("webhook-signature"),
	}

	if err := h.service.Webhook(c.Request.Context(), payload, headers); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *PaymentHandler) Status(c *gin.Context) {
	userID, _ := c.Get("userID")
	status, err := h.service.Status(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, status)
}
