package handler

import (
	"net/http"

	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type ArchiveHandler struct {
	service *service.ArchiveService
	logger  *infrastructure.Logger
}

func NewArchiveHandler(svc *service.ArchiveService, logger *infrastructure.Logger) *ArchiveHandler {
	return &ArchiveHandler{service: svc, logger: logger}
}

func (h *ArchiveHandler) ListFolders(c *gin.Context) {
	userID, _ := c.Get("userID")
	folderType := c.Query("type")

	folders, err := h.service.ListFolders(c.Request.Context(), userID.(string), folderType)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"folders": folders})
}

func (h *ArchiveHandler) CreateFolder(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Type  string `json:"type"`
		Color string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	if req.Type == "" {
		req.Type = "custom"
	}

	userID, _ := c.Get("userID")
	folder, err := h.service.CreateFolder(c.Request.Context(), userID.(string), req.Name, req.Type, req.Color)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Folder created", "id": folder.ID})
}

func (h *ArchiveHandler) DeleteFolder(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteFolder(c.Request.Context(), id); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted"})
}

func (h *ArchiveHandler) MoveChat(c *gin.Context) {
	var req struct {
		ConversationID string `json:"conversation_id" binding:"required"`
		FolderID       string `json:"folder_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	if err := h.service.MoveChat(c.Request.Context(), userID.(string), req.ConversationID, req.FolderID); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat moved"})
}

func (h *ArchiveHandler) RemoveFromArchive(c *gin.Context) {
	var req struct {
		ConversationID string `json:"conversation_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	if err := h.service.RemoveFromArchive(c.Request.Context(), userID.(string), req.ConversationID); err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat removed from archive"})
}

func (h *ArchiveHandler) GetStatus(c *gin.Context) {
	userID, _ := c.Get("userID")
	status, err := h.service.GetStatus(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, status)
}
