package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service *service.AnalyticsService
	logger  *infrastructure.Logger
}

func NewAnalyticsHandler(service *service.AnalyticsService, logger *infrastructure.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{service: service, logger: logger}
}

func (h *AnalyticsHandler) Overview(c *gin.Context) {
	userID, _ := c.Get("userID")
	overview, err := h.service.Overview(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, overview)
}

func (h *AnalyticsHandler) ChannelDistribution(c *gin.Context) {
	userID, _ := c.Get("userID")
	distribution, err := h.service.ChannelDistribution(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"distribution": distribution})
}

func (h *AnalyticsHandler) Insights(c *gin.Context) {
	userID, _ := c.Get("userID")
	insights, err := h.service.Insights(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, insights)
}

func (h *AnalyticsHandler) Trends(c *gin.Context) {
	userID, _ := c.Get("userID")
	days := 7

	trends, err := h.service.Trends(c.Request.Context(), userID.(string), days)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"trends": trends})
}

func (h *AnalyticsHandler) Satisfaction(c *gin.Context) {
	userID, _ := c.Get("userID")
	data, err := h.service.Satisfaction(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) UnknownQuestions(c *gin.Context) {
	userID, _ := c.Get("userID")
	data, err := h.service.UnknownQuestionsStats(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AnalyticsHandler) PopularQuestions(c *gin.Context) {
	userID, _ := c.Get("userID")
	data, err := h.service.PopularQuestions(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": data})
}

func (h *AnalyticsHandler) MessagesTrend(c *gin.Context) {
	userID, _ := c.Get("userID")
	days := 7
	data, err := h.service.MessagesTrend(c.Request.Context(), userID.(string), days)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"trends": data})
}

func (h *AnalyticsHandler) Uptime(c *gin.Context) {
	userID, _ := c.Get("userID")
	data, err := h.service.Uptime(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, data)
}
