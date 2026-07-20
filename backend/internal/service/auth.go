package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"sync"
	"time"

	"noant/config"
	apperrors "noant/internal/errors"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ========== AUTH SERVICE ==========

type AuthService struct {
	cfg           *config.Config
	userRepo      *repository.UserRepository
	redis         *infrastructure.RedisClient
	logger        *infrastructure.Logger
	email         *EmailService
	memRL         *infrastructure.MemoryRateLimiter
	loginAttempts map[string]*loginAttempt
	attemptMu     sync.Mutex
}

type loginAttempt struct {
	count     int
	lockedUntil time.Time
}

// NewAuthService creates an AuthService with JWT generation, login attempt tracking,
// and background cleanup of expired lockouts. Requires a valid EmailService for
// verification code delivery.
func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository, redis *infrastructure.RedisClient, logger *infrastructure.Logger, email *EmailService) *AuthService {
	s := &AuthService{cfg: cfg, userRepo: userRepo, redis: redis, logger: logger, email: email, memRL: infrastructure.NewMemoryRateLimiter(5 * time.Minute), loginAttempts: make(map[string]*loginAttempt)}
	// Periodic cleanup of expired lockouts
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			s.attemptMu.Lock()
			now := time.Now()
			for k, v := range s.loginAttempts {
				if now.After(v.lockedUntil) {
					delete(s.loginAttempts, k)
				}
			}
			s.attemptMu.Unlock()
		}
	}()
	return s
}

func generateVerificationCode() string {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, 6)
	// We import "io" in service.go. Let's make sure rand.Reader is used from crypto/rand (which is already imported).
	n, err := rand.Read(b)
	if n != 6 || err != nil {
		for i := 0; i < 6; i++ {
			b[i] = table[time.Now().UnixNano()%10]
		}
		return string(b)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

func validatePasswordStrength(password string) error {
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) ||
		!regexp.MustCompile(`[a-z]`).MatchString(password) ||
		!regexp.MustCompile(`\d`).MatchString(password) ||
		!regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;':",./<>?]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character")
	}
	return nil
}

// Register creates a new user account with a 14-day trial, hashed password,
// and a 6-digit verification code. Returns the created user. Returns an error
// if the email is already registered or the password doesn't meet strength requirements.
func (s *AuthService) Register(ctx context.Context, email, password, firstName, lastName, companyName string) (*domain.User, error) {
	if err := validatePasswordStrength(password); err != nil {
		return nil, err
	}
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set 14-day trial period
	now := time.Now()
	trialExpires := now.AddDate(0, 0, 14)

	code := generateVerificationCode()
	user := &domain.User{
		Email:              email,
		Password:           string(hashedPassword),
		FirstName:          firstName,
		LastName:           lastName,
		CompanyName:        companyName,
		Role:               "owner",
		PlanID:             "free",
		IsActive:           true,
		MustChangePassword: true,
		TrialExpiresAt:     &trialExpires,
		IsVerified:         false,
		VerificationCode:   &code,
	}
	if err := s.userRepo.RunInTx(ctx, func(tx *sql.Tx) error {
		return s.userRepo.CreateTx(ctx, tx, user)
	}); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if s.email != nil {
		if _, err := s.email.SendVerificationEmail(ctx, email, code); err != nil {
			s.logger.Error("Failed to send verification email on registration", "error", err)
		}
	}

	created, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created user: %w", err)
	}
	return created, nil
}

func (s *AuthService) generateRefreshToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Login authenticates a user by email and password. Returns a JWT access token,
// refresh token, and the authenticated user on success. Returns ErrInvalidCredentials
// on wrong password, ErrAccountLocked after 5 failed attempts (15-minute lockout),
// or ErrEmailNotVerified if the account hasn't been verified.
func (s *AuthService) Login(ctx context.Context, email, password string) (user *domain.User, token, refreshToken string, err error) {
	user, err = s.userRepo.GetByEmail(ctx, email)
	if user == nil {
		return nil, "", "", fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return nil, "", "", fmt.Errorf("database error: %w", err)
	}

	// Check account lockout before password comparison
	s.attemptMu.Lock()
	attempt := s.loginAttempts[email]
	if attempt != nil && time.Now().Before(attempt.lockedUntil) {
		s.attemptMu.Unlock()
		return nil, "", "", apperrors.ErrAccountLocked
	}
	s.attemptMu.Unlock()

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.attemptMu.Lock()
		if s.loginAttempts[email] == nil {
			s.loginAttempts[email] = &loginAttempt{}
		}
		s.loginAttempts[email].count++
		if s.loginAttempts[email].count >= 5 {
			s.loginAttempts[email].lockedUntil = time.Now().Add(15 * time.Minute)
			s.logger.Warn("Account locked due to failed login attempts", "email", email, "lockout_minutes", 15)
		}
		s.attemptMu.Unlock()
		return nil, "", "", fmt.Errorf("invalid credentials")
	}

	// Successful login: reset attempt counter
	s.attemptMu.Lock()
	delete(s.loginAttempts, email)
	s.attemptMu.Unlock()

	if !user.IsVerified {
		return nil, "", "", apperrors.ErrEmailNotVerified
	}
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)
	token, err = s.generateToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken = s.generateRefreshToken()
	if s.redis != nil {
		_ = s.redis.Set(ctx, "refresh:"+refreshToken, user.ID, 7*24*time.Hour)
	}

	return user, token, refreshToken, nil
}

// VerifyEmail validates the 6-digit code sent to the user's email. On success,
// marks the account as verified and returns JWT tokens. Rate-limited to 5 attempts
// per minute. Returns ErrInvalidVerification on code mismatch or
// ErrTooManyVerifications when rate-limited.
func (s *AuthService) VerifyEmail(ctx context.Context, email, code string) (user *domain.User, token, refreshToken string, err error) {
	if !s.memRL.Allow("verify:"+email, 5, time.Minute) {
		return nil, "", "", apperrors.ErrTooManyVerifications
	}

	user, err = s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		return nil, "", "", fmt.Errorf("user not found")
	}
	if user.IsVerified {
		token, err = s.generateToken(user)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
		}
		refreshToken = s.generateRefreshToken()
		if s.redis != nil {
			_ = s.redis.Set(ctx, "refresh:"+refreshToken, user.ID, 7*24*time.Hour)
		}
		return user, token, refreshToken, nil
	}
	if user.VerificationCode == nil || *user.VerificationCode != code {
		return nil, "", "", apperrors.ErrInvalidVerification
	}

	if err := s.userRepo.UpdateVerificationStatus(ctx, user.ID, true); err != nil {
		return nil, "", "", fmt.Errorf("failed to update verification status: %w", err)
	}

	user.IsVerified = true
	user.VerificationCode = nil

	token, err = s.generateToken(user)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate token: %w", err)
	}
	refreshToken = s.generateRefreshToken()
	if s.redis != nil {
		_ = s.redis.Set(ctx, "refresh:"+refreshToken, user.ID, 7*24*time.Hour)
	}

	return user, token, refreshToken, nil
}

// ResendVerification generates a new 6-digit verification code and sends it
// via email. Returns ErrEmailAlreadyVerified if the account is already verified.
func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if user.IsVerified {
		return apperrors.ErrEmailAlreadyVerified
	}

	code := generateVerificationCode()
	if err := s.userRepo.UpdateVerificationCode(ctx, user.ID, code); err != nil {
		return fmt.Errorf("failed to update verification code: %w", err)
	}

	if s.email != nil {
		if _, err := s.email.SendVerificationEmail(ctx, user.Email, code); err != nil {
			s.logger.Error("Failed to resend verification email", "error", err)
			return fmt.Errorf("failed to send verification email: %w", err)
		}
	}

	return nil
}

func (s *AuthService) generateToken(user *domain.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"type":    "access",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"iss":     "noant",
		"aud":     "noant-api",
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error) {
	if refreshToken == "" {
		return "", "", fmt.Errorf("refresh token required")
	}
	if s.redis == nil {
		return "", "", fmt.Errorf("token store unavailable")
	}

	userID, err := s.redis.Get(ctx, "refresh:"+refreshToken)
	if err != nil || userID == "" {
		return "", "", fmt.Errorf("invalid or expired refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return "", "", fmt.Errorf("user not found")
	}

	accessToken, err = s.generateToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshToken = s.generateRefreshToken()
	_ = s.redis.Delete(ctx, "refresh:"+refreshToken)
	if err := s.redis.Set(ctx, "refresh:"+newRefreshToken, user.ID, 7*24*time.Hour); err != nil {
		return "", "", err
	}
	return accessToken, newRefreshToken, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	return s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	key := "forgot-password:" + email
	if s.redis != nil {
		countStr, err := s.redis.Get(ctx, key)
		var count int
		if err == nil {
			_, _ = fmt.Sscanf(countStr, "%d", &count)
		}
		if count >= 3 {
			s.logger.Warn("Forgot password request rate limited", "email", email)
			return fmt.Errorf("too many forgot password requests, please try again in an hour")
		}
	} else if !s.memRL.Allow(key, 3, time.Hour) {
		s.logger.Warn("Forgot password request rate limited (memory)", "email", email)
		return fmt.Errorf("too many forgot password requests, please try again in an hour")
	}

	if s.redis != nil {
		newVal, err := s.redis.Incr(ctx, key)
		if err == nil && newVal == 1 {
			_ = s.redis.Expire(ctx, key, time.Hour)
		}
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	resetToken := make([]byte, 32)
	if _, err := rand.Read(resetToken); err != nil {
		return err
	}
	token := hex.EncodeToString(resetToken)
	if s.redis != nil {
		_ = s.redis.Set(ctx, "reset:"+token, user.ID, time.Hour)
	}
	if s.email != nil {
		if _, err := s.email.SendPasswordReset(ctx, user.Email, token); err != nil {
			s.logger.Error("Failed to send password reset email", "error", err)
		}
	}
	return nil
}

func (s *AuthService) Logout(ctx context.Context, token, refreshToken string) error {
	if s.redis == nil {
		return nil
	}
	if token != "" {
		if err := s.redis.Set(ctx, "blacklist:"+token, "true", 24*time.Hour); err != nil {
			return err
		}
	}
	if refreshToken != "" {
		_ = s.redis.Delete(ctx, "refresh:"+refreshToken)
	}
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if s.redis == nil {
		return fmt.Errorf("token store unavailable")
	}
	userID, err := s.redis.Get(ctx, "reset:"+token)
	if err != nil || userID == "" {
		return fmt.Errorf("invalid or expired reset token")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return err
	}
	_ = s.redis.Delete(ctx, "reset:"+token)
	return nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
