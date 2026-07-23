package handler

import (
	"net/http"

	"noant/internal/domain"
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
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	profile, err := h.service.GetProfile(c.Request.Context(), userID)
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

	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	if err := h.service.UpdateProfile(c.Request.Context(), userID, req); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

func (h *SettingsHandler) ListAPIKeys(c *gin.Context) {
	userID := getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	keys, err := h.service.ListAPIKeys(c.Request.Context(), userID)
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

	userID := getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	key, err := h.service.CreateAPIKey(c.Request.Context(), userID, req.Name)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"api_key": key, "id": key.ID})
}

func (h *SettingsHandler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")
	userID := getScopeID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	if err := h.service.RevokeAPIKey(c.Request.Context(), userID, id); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked"})
}

func (h *SettingsHandler) ListTeam(c *gin.Context) {
	orgID := getOrgID(c)
	if orgID == "" {
		c.JSON(http.StatusOK, gin.H{"team": []domain.TeamMember{}})
		return
	}
	members, err := h.service.ListTeam(c.Request.Context(), orgID)
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

	orgID := getOrgID(c)
	if orgID == "" {
		utils.RespondValidationError(c, "Organization required for team management")
		return
	}
	member, err := h.service.InviteTeamMember(c.Request.Context(), orgID, req.Email, req.Role)
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
