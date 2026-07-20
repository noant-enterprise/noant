package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	service *service.SettingsService
	logger  *infrastructure.Logger
}

func NewSettingsHandler(svc *service.SettingsService, logger *infrastructure.Logger) *SettingsHandler {
	return &SettingsHandler{service: svc, logger: logger}
}

func (h *SettingsHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	profile, err := h.service.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *SettingsHandler) UpdateProfile(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	if err := h.service.UpdateProfile(c.Request.Context(), userID.(string), req); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

func (h *SettingsHandler) ListAPIKeys(c *gin.Context) {
	userID, _ := c.Get("userID")
	keys, err := h.service.ListAPIKeys(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": keys})
}

func (h *SettingsHandler) CreateAPIKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	key, err := h.service.CreateAPIKey(c.Request.Context(), userID.(string), req.Name)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"api_key": key, "id": key.ID})
}

func (h *SettingsHandler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")
	if err := h.service.RevokeAPIKey(c.Request.Context(), userID.(string), id); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked"})
}

func (h *SettingsHandler) ListTeam(c *gin.Context) {
	userID, _ := c.Get("userID")
	members, err := h.service.ListTeam(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": members})
}

func (h *SettingsHandler) InviteTeamMember(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	member, err := h.service.InviteTeamMember(c.Request.Context(), userID.(string), req.Email, req.Role)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation sent", "id": member.ID})
}

func (h *SettingsHandler) RemoveTeamMember(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.RemoveTeamMember(c.Request.Context(), id); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team member removed"})
}
