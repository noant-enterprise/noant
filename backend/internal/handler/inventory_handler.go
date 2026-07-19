package handler

import (
	"net/http"

	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	service *service.InventoryService
	logger  *infrastructure.Logger
}

func NewInventoryHandler(service *service.InventoryService, logger *infrastructure.Logger) *InventoryHandler {
	return &InventoryHandler{service: service, logger: logger}
}

func (h *InventoryHandler) Create(c *gin.Context) {
	var req struct {
		Type          string   `json:"type" binding:"required"`
		Name          string   `json:"name" binding:"required"`
		Description   string   `json:"description"`
		Price         float64  `json:"price" binding:"required"`
		MinPrice      *float64 `json:"min_price"`
		StockQuantity *int     `json:"stock_quantity"`
		ImageURL      *string  `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	item := &domain.InventoryItem{
		Type:          req.Type,
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		MinPrice:      req.MinPrice,
		StockQuantity: req.StockQuantity,
		ImageURL:      req.ImageURL,
	}

	if err := h.service.Create(c.Request.Context(), userID.(string), item); err != nil {
		h.logger.Error("Failed to create inventory item", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *InventoryHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	itemType := c.Query("type")

	items, err := h.service.List(c.Request.Context(), userID.(string), itemType)
	if err != nil {
		h.logger.Error("Failed to list inventory", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

func (h *InventoryHandler) GetByID(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	item, err := h.service.GetByID(c.Request.Context(), id, userID.(string))
	if err != nil {
		h.logger.Error("Failed to get inventory item", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}
	if item == nil {
		utils.RespondNotFound(c, "Item not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *InventoryHandler) Update(c *gin.Context) {
	var req struct {
		ID            string   `json:"id" binding:"required"`
		Type          string   `json:"type"`
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		Price         float64  `json:"price"`
		MinPrice      *float64 `json:"min_price"`
		StockQuantity *int     `json:"stock_quantity"`
		ImageURL      *string  `json:"image_url"`
		IsActive      *bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	item, err := h.service.GetByID(c.Request.Context(), req.ID, userID.(string))
	if err != nil || item == nil {
		utils.RespondNotFound(c, "Item not found")
		return
	}

	if req.Type != "" {
		item.Type = req.Type
	}
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.Price > 0 {
		item.Price = req.Price
	}
	if req.MinPrice != nil {
		item.MinPrice = req.MinPrice
	}
	if req.StockQuantity != nil {
		item.StockQuantity = req.StockQuantity
	}
	if req.ImageURL != nil {
		item.ImageURL = req.ImageURL
	}
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}

	if err := h.service.Update(c.Request.Context(), item); err != nil {
		h.logger.Error("Failed to update inventory item", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *InventoryHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	id := c.Param("id")

	if err := h.service.Delete(c.Request.Context(), id, userID.(string)); err != nil {
		h.logger.Error("Failed to delete inventory item", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}

func (h *InventoryHandler) Search(c *gin.Context) {
	userID, _ := c.Get("userID")
	q := c.Query("q")

	items, err := h.service.Search(c.Request.Context(), userID.(string), q)
	if err != nil {
		h.logger.Error("Failed to search inventory", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}
