package handler

import (
	"errors"
	"net/http"
	"time"

	"noant/internal/infrastructure"
	apperrors "noant/internal/errors"
	"noant/internal/middleware"
	"noant/internal/service"
	"noant/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *service.AuthService
	logger  *infrastructure.Logger
}

func NewAuthHandler(svc *service.AuthService, logger *infrastructure.Logger) *AuthHandler {
	return &AuthHandler{service: svc, logger: logger}
}

// Register creates a new user account with email verification.
// Sends a verification email and returns JWT access + refresh tokens in httpOnly cookies.
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=8"`
		FirstName   string `json:"first_name" binding:"required"`
		LastName    string `json:"last_name" binding:"required"`
		CompanyName string `json:"company_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	user, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName, req.CompanyName)
	if err != nil {
		h.logger.Error("Registration failed", "error", err)
		utils.RespondConflict(c, "Registration failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login authenticates a user with email and password.
// Returns JWT tokens in httpOnly cookies. Supports 15-minute account lockout after 5 failed attempts.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	user, token, refreshToken, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrEmailNotVerified) {
			h.logger.Warn("Login failed: email not verified", "email", req.Email)
			c.JSON(http.StatusForbidden, gin.H{"error": "email_not_verified"})
			return
		}
		if errors.Is(err, apperrors.ErrAccountLocked) {
			h.logger.Warn("Login failed: account locked", "email", req.Email)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Account temporarily locked due to too many failed attempts. Try again in 15 minutes."})
			return
		}
		h.logger.Warn("Login failed", "email", req.Email)
		utils.RespondUnauthorized(c, "Invalid email or password")
		return
	}

	middleware.SetAuthCookies(c, token, refreshToken, 24*time.Hour, 7*24*time.Hour)
	c.Header("Cache-Control", "no-store")

	// Compute trial info for response
	var trialInfo map[string]interface{}
	if user.TrialExpiresAt != nil {
		trialInfo = map[string]interface{}{
			"trial_expires_at": user.TrialExpiresAt.Format(time.RFC3339),
			"trial_ended":      time.Now().After(*user.TrialExpiresAt),
			"trial_days_left":  int(time.Until(*user.TrialExpiresAt).Hours() / 24),
		}
		if trialInfo["trial_days_left"].(int) < 0 {
			trialInfo["trial_days_left"] = 0
		}
	} else {
		trialInfo = map[string]interface{}{
			"trial_ended": false,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       user,
		"trial_info": trialInfo,
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	user, token, refreshToken, err := h.service.VerifyEmail(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidVerification) {
			h.logger.Warn("Email verification: invalid code", "email", req.Email)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_code"})
			return
		}
		if errors.Is(err, apperrors.ErrTooManyVerifications) {
			h.logger.Warn("Email verification rate limit hit", "email", req.Email)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too_many_attempts"})
			return
		}
		h.logger.Error("Email verification failed", "error", err)
		utils.RespondInternalError(c, "")
		return
	}

	middleware.SetAuthCookies(c, token, refreshToken, 24*time.Hour, 7*24*time.Hour)
	c.Header("Cache-Control", "no-store")

	var trialInfo map[string]interface{}
	if user.TrialExpiresAt != nil {
		trialInfo = map[string]interface{}{
			"trial_expires_at": user.TrialExpiresAt.Format(time.RFC3339),
			"trial_ended":      time.Now().After(*user.TrialExpiresAt),
			"trial_days_left":  int(time.Until(*user.TrialExpiresAt).Hours() / 24),
		}
		if trialInfo["trial_days_left"].(int) < 0 {
			trialInfo["trial_days_left"] = 0
		}
	} else {
		trialInfo = map[string]interface{}{
			"trial_ended": false,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Email verified successfully",
		"user":       user,
		"trial_info": trialInfo,
	})
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	if err := h.service.ResendVerification(c.Request.Context(), req.Email); err != nil {
		h.logger.Error("Resend verification failed", "error", err)
		if errors.Is(err, apperrors.ErrEmailAlreadyVerified) {
			c.JSON(http.StatusOK, gin.H{"message": "Email is already verified"})
			return
		}
		utils.RespondInternalError(c, "Failed to resend verification email")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent successfully"})
}

// RefreshToken exchanges a valid refresh token for a new access token.
// The refresh token is rotated on each use to prevent replay attacks.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken := middleware.GetRefreshTokenFromRequest(c)
	if refreshToken == "" {
		utils.RespondUnauthorized(c, "refresh token required")
		return
	}

	token, newRefreshToken, err := h.service.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		utils.RespondUnauthorized(c, "Invalid or expired session")
		return
	}

	middleware.SetAuthCookies(c, token, newRefreshToken, 24*time.Hour, 7*24*time.Hour)
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"message": "Session refreshed"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := middleware.GetAccessTokenFromRequest(c)
	refreshToken := middleware.GetRefreshTokenFromRequest(c)
	if err := h.service.Logout(c.Request.Context(), token, refreshToken); err != nil {
		utils.RespondInternalError(c, "Failed to log out")
		return
	}
	if token != "" {
		middleware.BlacklistAccessToken(token)
	}
	middleware.ClearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

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
	if err := h.service.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		utils.RespondValidationError(c, "Failed to change password")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	if err := h.service.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		h.logger.Error("Forgot password failed", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}
	utils.SanitizeStruct(&req)

	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		h.logger.Error("Reset password failed", "error", err)
		utils.RespondInternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := getUserID(c)
	if userID == "" {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}
	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to retrieve user")
		return
	}
	if user == nil {
		utils.RespondUnauthorized(c, "User not found")
		return
	}

	c.Header("Cache-Control", "no-store")

	// Compute trial info for response
	var trialInfo map[string]interface{}
	if user.TrialExpiresAt != nil {
		trialInfo = map[string]interface{}{
			"trial_expires_at": user.TrialExpiresAt.Format(time.RFC3339),
			"trial_ended":      time.Now().After(*user.TrialExpiresAt),
			"trial_days_left":  int(time.Until(*user.TrialExpiresAt).Hours() / 24),
		}
		if trialInfo["trial_days_left"].(int) < 0 {
			trialInfo["trial_days_left"] = 0
		}
	} else {
		trialInfo = map[string]interface{}{
			"trial_ended": false,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":       user,
		"trial_info": trialInfo,
	})
}
