package handler

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"noant/internal/domain"
	apperrors "noant/internal/errors"
	"noant/internal/infrastructure"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type TrainingHandler struct {
	service *service.TrainingService
	logger  *infrastructure.Logger
}

func NewTrainingHandler(svc *service.TrainingService, logger *infrastructure.Logger) *TrainingHandler {
	return &TrainingHandler{service: svc, logger: logger}
}

func (h *TrainingHandler) ListCategories(c *gin.Context) {
	userID, _ := c.Get("userID")
	categories, err := h.service.ListCategories(c.Request.Context(), userID.(string))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *TrainingHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	category, err := h.service.CreateCategory(c.Request.Context(), userID.(string), req.Name, req.Description, req.Color)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Category created", "id": category.ID})
}

// BulkImport imports multiple QA pairs from a JSON array in the request body.
// Useful for migrating training data from other platforms.
func (h *TrainingHandler) BulkImport(c *gin.Context) {
	var req struct {
		CategoryID string `json:"category_id" binding:"required"`
		QAPairs    []struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
		} `json:"qa_pairs" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	var pairs []domain.QAPair
	for _, p := range req.QAPairs {
		pairs = append(pairs, domain.QAPair{
			Question: p.Question,
			Answer:   p.Answer,
		})
	}

	userID, _ := c.Get("userID")
	if err := h.service.BulkImport(c.Request.Context(), userID.(string), req.CategoryID, pairs); err != nil {
		h.logger.Error("Bulk import failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bulk import successful", "count": len(pairs)})
}

// UploadCSV imports QA pairs from a CSV file upload.
// Expects columns: category, question, answer (with header row).
func (h *TrainingHandler) UploadCSV(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer func() { _ = file.Close() }()

	// 1. File Size Guard: limit to 2 MB to prevent OOM / Denial of Service
	const maxFileSize = 2 * 1024 * 1024 // 2 MB
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds the 2 MB limit"})
		return
	}

	// 2. Extension Guard: only allow .csv
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only CSV files are allowed"})
		return
	}

	// 3. Content Type Verification: read first 512 bytes to verify it's not a binary/executable
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)
	// Seek back to the beginning of the file so that ReadAll reads from start
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		h.logger.Warn("Failed to seek file, skipping content type detection")
	} else {
		contentType := http.DetectContentType(buffer[:n])
		// Check for common malicious/binary formats
		blacklistedTypes := []string{"executable", "dosexec", "elf", "zip", "pdf", "image", "msdownload", "octet-stream"}
		for _, bt := range blacklistedTypes {
			if strings.Contains(strings.ToLower(contentType), bt) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file content. Binary and compressed formats are not allowed"})
				return
			}
		}
	}

	categoryID := c.PostForm("category_id")
	if categoryID == "" {
		categoryID = "default"
	}

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	userID, _ := c.Get("userID")
	count, err := h.service.UploadCSV(c.Request.Context(), userID.(string), categoryID, data)
	if err != nil {
		h.logger.Error("CSV upload failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CSV uploaded successfully", "count": count})
}

func (h *TrainingHandler) ListUnknownQuestions(c *gin.Context) {
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	userID, _ := c.Get("userID")
	questions, err := h.service.ListUnknownQuestions(c.Request.Context(), userID.(string), status, limit, offset)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	total, _ := h.service.CountUnknownQuestions(c.Request.Context(), userID.(string), status)

	c.JSON(http.StatusOK, gin.H{"questions": questions, "total": total, "limit": limit, "offset": offset})
}

func (h *TrainingHandler) BatchTrainUnknown(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req struct {
		IDs        []string `json:"ids" binding:"required,min=1"`
		Answer     string   `json:"answer" binding:"required"`
		CategoryID string   `json:"category_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "IDs, answer, and category_id are required")
		return
	}

	if err := h.service.BatchTrainUnknown(c.Request.Context(), userID.(string), req.Answer, req.CategoryID, req.IDs); err != nil {
		h.logger.Error("Batch train failed", "error", err)
		utils.RespondInternalError(c, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Questions trained successfully", "count": len(req.IDs)})
}

func (h *TrainingHandler) BatchIgnoreUnknown(c *gin.Context) {
	userID, _ := c.Get("userID")
	var req struct {
		IDs []string `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, "IDs are required")
		return
	}

	if err := h.service.BatchIgnoreUnknown(c.Request.Context(), userID.(string), req.IDs); err != nil {
		h.logger.Error("Batch ignore failed", "error", err)
		utils.RespondInternalError(c, "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Questions ignored successfully", "count": len(req.IDs)})
}

func (h *TrainingHandler) TrainUnknown(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Answer     string `json:"answer" binding:"required"`
		CategoryID string `json:"category_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	if err := h.service.TrainUnknown(c.Request.Context(), userID.(string), id, req.Answer, req.CategoryID); err != nil {
		if errors.Is(err, apperrors.ErrUnknownQuestion) || errors.Is(err, apperrors.ErrNotFound) {
			utils.RespondNotFound(c, "Unknown question")
			return
		}
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question trained successfully"})
}

func (h *TrainingHandler) IgnoreUnknown(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	if err := h.service.IgnoreUnknown(c.Request.Context(), userID.(string), id); err != nil {
		if errors.Is(err, apperrors.ErrUnknownQuestion) || errors.Is(err, apperrors.ErrNotFound) {
			utils.RespondNotFound(c, "Unknown question")
			return
		}
		h.logger.Error("Ignore unknown question failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question ignored successfully"})
}

func (h *TrainingHandler) ClearUnknown(c *gin.Context) {
	userID, _ := c.Get("userID")

	if err := h.service.ClearUnknownQuestions(c.Request.Context(), userID.(string)); err != nil {
		h.logger.Error("Clear unknown questions failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unknown questions cleared successfully"})
}

func (h *TrainingHandler) ListQAPairs(c *gin.Context) {
	categoryID := c.Param("id")
	userID, _ := c.Get("userID")

	qaPairs, err := h.service.ListQAPairs(c.Request.Context(), userID.(string), categoryID)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"qa_pairs": qaPairs})
}

// CreateQAPair adds a new question-answer pair to the training data.
// The pair is used by the AI Brain for intent matching and response generation.
func (h *TrainingHandler) CreateQAPair(c *gin.Context) {
	var req struct {
		CategoryID string `json:"category_id" binding:"required"`
		Question   string `json:"question" binding:"required"`
		Answer     string `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	qa, err := h.service.CreateQAPair(c.Request.Context(), userID.(string), req.CategoryID, req.Question, req.Answer)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Q&A pair created successfully", "qa_pair": qa})
}

func (h *TrainingHandler) UpdateQAPair(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		CategoryID string `json:"category_id" binding:"required"`
		Question   string `json:"question" binding:"required"`
		Answer     string `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	userID, _ := c.Get("userID")
	err := h.service.UpdateQAPair(c.Request.Context(), userID.(string), id, req.CategoryID, req.Question, req.Answer)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Q&A pair updated successfully"})
}

func (h *TrainingHandler) DeleteQAPair(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	err := h.service.DeleteQAPair(c.Request.Context(), userID.(string), id)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Q&A pair deleted successfully"})
}

func (h *TrainingHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	err := h.service.DeleteCategory(c.Request.Context(), userID.(string), id)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category and all associated Q&A pairs deleted successfully"})
}

func (h *TrainingHandler) SearchQAPairs(c *gin.Context) {
	query := c.Query("q")
	userID, _ := c.Get("userID")

	qaPairs, err := h.service.SearchQAPairs(c.Request.Context(), userID.(string), query)
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"qa_pairs": qaPairs})
}
