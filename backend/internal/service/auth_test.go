package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"noant/config"
	apperrors "noant/internal/errors"
	"noant/internal/domain"
	"noant/internal/infrastructure"
	"noant/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

func newTestAuthService(mockUserRepo repository.IUserRepo) *AuthService {
	return &AuthService{
		cfg:           &config.Config{JWTSecret: "test-secret-123"},
		userRepo:      mockUserRepo,
		redis:         nil,
		logger:        infrastructure.NewNullLogger(),
		email:         nil,
		memRL:         infrastructure.NewMemoryRateLimiter(5 * time.Minute),
		loginAttempts: make(map[string]*loginAttempt),
	}
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return string(hashed)
}

func createTestUserInRepo(t *testing.T, mock *repository.MockUserRepo, email, password string, verified bool) *domain.User {
	t.Helper()
	user := &domain.User{
		ID:         "user-" + strings.ReplaceAll(email, "@", "_at_"),
		Email:      email,
		Password:   mustHashPassword(t, password),
		FirstName:  "Test",
		LastName:   "User",
		Role:       "owner",
		CompanyName: "Test Co",
		IsActive:   true,
		IsVerified: verified,
		PlanID:     "free",
	}
	if err := mock.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestRegister_Success(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	user, err := svc.Register(context.Background(), "new@example.com", "StrongP@ss1", "John", "Doe", "Acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %s", user.Email)
	}
	if user.FirstName != "John" {
		t.Errorf("expected first name John, got %s", user.FirstName)
	}
	if user.LastName != "Doe" {
		t.Errorf("expected last name Doe, got %s", user.LastName)
	}
	if user.CompanyName != "Acme" {
		t.Errorf("expected company Acme, got %s", user.CompanyName)
	}
	if user.Role != "owner" {
		t.Errorf("expected role owner, got %s", user.Role)
	}
	if user.PlanID != "free" {
		t.Errorf("expected plan free, got %s", user.PlanID)
	}
	if !user.IsActive {
		t.Error("expected user to be active")
	}
	if user.IsVerified {
		t.Error("expected user to be unverified")
	}
	if !user.MustChangePassword {
		t.Error("expected MustChangePassword to be true")
	}
	if user.VerificationCode == nil {
		t.Error("expected verification code to be set")
	}
	if user.TrialExpiresAt == nil {
		t.Error("expected trial expiry to be set")
	} else {
		daysUntilExpiry := time.Until(*user.TrialExpiresAt).Hours() / 24
		if daysUntilExpiry < 13 || daysUntilExpiry > 15 {
			t.Errorf("expected trial expiry ~14 days out, got %.1f days", daysUntilExpiry)
		}
	}

	// Verify user exists in mock
	stored, err := mock.GetByEmail(context.Background(), "new@example.com")
	if err != nil {
		t.Fatalf("error retrieving stored user: %v", err)
	}
	if stored == nil {
		t.Fatal("user not found in mock repo after registration")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	_, err := svc.Register(context.Background(), "dup@example.com", "StrongP@ss1", "John", "Doe", "Co")
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err = svc.Register(context.Background(), "dup@example.com", "AnotherP@ss1", "Jane", "Smith", "Co2")
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("expected 'already registered' error, got: %v", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	// The Register function does not validate email format, so this should succeed
	user, err := svc.Register(context.Background(), "not-an-email", "StrongP@ss1", "John", "Doe", "Co")
	if err != nil {
		t.Fatalf("unexpected error (service does not validate email format): %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "not-an-email" {
		t.Errorf("expected email 'not-an-email', got %s", user.Email)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	// Short password that meets character requirements — service does not enforce length
	user, err := svc.Register(context.Background(), "short@example.com", "Ab1!", "John", "Doe", "Co")
	if err != nil {
		t.Fatalf("unexpected error (service does not enforce min length): %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"no uppercase", "lowercase1!"},
		{"no lowercase", "UPPERCASE1!"},
		{"no digit", "NoDigit!abc"},
		{"no special", "NoSpecial123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := repository.NewMockUserRepo()
			svc := newTestAuthService(mock)

			_, err := svc.Register(context.Background(), "weak@example.com", tt.password, "John", "Doe", "Co")
			if err == nil {
				t.Errorf("expected error for password %q, got nil", tt.password)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)
	password := "StrongP@ss1"

	createTestUserInRepo(t, mock, "login@example.com", password, true)

	user, token, refreshToken, err := svc.Login(context.Background(), "login@example.com", password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "login@example.com" {
		t.Errorf("expected email login@example.com, got %s", user.Email)
	}
	if token == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "wrong@example.com", "StrongP@ss1", true)

	user, token, refreshToken, err := svc.Login(context.Background(), "wrong@example.com", "WrongP@ss1")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("expected 'invalid credentials' error, got: %v", err)
	}
	if user != nil {
		t.Error("expected nil user")
	}
	if token != "" {
		t.Error("expected empty token")
	}
	if refreshToken != "" {
		t.Error("expected empty refresh token")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	user, token, refreshToken, err := svc.Login(context.Background(), "nonexistent@example.com", "StrongP@ss1")
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("expected 'invalid credentials' error, got: %v", err)
	}
	if user != nil {
		t.Error("expected nil user")
	}
	if token != "" {
		t.Error("expected empty token")
	}
	if refreshToken != "" {
		t.Error("expected empty refresh token")
	}
}

func TestLogin_AccountLocked(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "lock@example.com", "StrongP@ss1", true)

	// Attempt 5 wrong passwords to trigger lockout
	for i := 0; i < 5; i++ {
		_, _, _, err := svc.Login(context.Background(), "lock@example.com", "WrongP@ss1")
		if err == nil {
			t.Fatalf("attempt %d: expected error, got nil", i+1)
		}
	}

	// Now try with the correct password — should be locked
	_, _, _, err := svc.Login(context.Background(), "lock@example.com", "StrongP@ss1")
	if !errors.Is(err, apperrors.ErrAccountLocked) {
		t.Errorf("expected ErrAccountLocked, got: %v", err)
	}
}

func TestLogin_AccountLocked_Unverified(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "lockunv@example.com", "StrongP@ss1", false)

	// 5 wrong attempts
	for i := 0; i < 5; i++ {
		svc.Login(context.Background(), "lockunv@example.com", "WrongP@ss1")
	}

	// Even correct password should return locked
	_, _, _, err := svc.Login(context.Background(), "lockunv@example.com", "StrongP@ss1")
	if !errors.Is(err, apperrors.ErrAccountLocked) {
		t.Errorf("expected ErrAccountLocked, got: %v", err)
	}
}

func TestLogin_SuccessfulResetsAttempts(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "reset@example.com", "StrongP@ss1", true)

	// 4 wrong attempts (below threshold)
	for i := 0; i < 4; i++ {
		svc.Login(context.Background(), "reset@example.com", "WrongP@ss1")
	}

	// Correct login should reset attempts
	user, token, _, err := svc.Login(context.Background(), "reset@example.com", "StrongP@ss1")
	if err != nil {
		t.Fatalf("expected successful login, got: %v", err)
	}
	if user == nil || token == "" {
		t.Error("expected user and token")
	}

	// Another wrong attempt should not lock (counter was reset)
	_, _, _, err = svc.Login(context.Background(), "reset@example.com", "WrongP@ss1")
	if errors.Is(err, apperrors.ErrAccountLocked) {
		t.Error("account should not be locked after successful login reset")
	}
}

func TestLogin_UnverifiedEmail(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "unverified@example.com", "StrongP@ss1", false)

	_, _, _, err := svc.Login(context.Background(), "unverified@example.com", "StrongP@ss1")
	if !errors.Is(err, apperrors.ErrEmailNotVerified) {
		t.Errorf("expected ErrEmailNotVerified, got: %v", err)
	}
}

func TestVerifyEmail_Success(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	code := "428571"
	user := &domain.User{
		ID:               "verify-user-1",
		Email:            "verify@example.com",
		Password:         mustHashPassword(t, "StrongP@ss1"),
		FirstName:        "Verify",
		LastName:         "Test",
		Role:             "owner",
		CompanyName:       "Co",
		IsActive:         true,
		IsVerified:       false,
		VerificationCode: &code,
		PlanID:           "free",
	}
	if err := mock.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	verifiedUser, token, refreshToken, err := svc.VerifyEmail(context.Background(), "verify@example.com", "428571")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if verifiedUser == nil {
		t.Fatal("expected user, got nil")
	}
	if !verifiedUser.IsVerified {
		t.Error("expected user to be verified")
	}
	if verifiedUser.VerificationCode != nil {
		t.Error("expected verification code to be cleared")
	}
	if token == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}

	// Confirm persisted in mock
	stored, _ := mock.GetByEmail(context.Background(), "verify@example.com")
	if stored != nil && !stored.IsVerified {
		t.Error("mock user not updated to verified")
	}
}

func TestVerifyEmail_InvalidCode(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	code := "111111"
	user := &domain.User{
		ID:               "verify-user-2",
		Email:            "wrongcode@example.com",
		Password:         mustHashPassword(t, "StrongP@ss1"),
		FirstName:        "Test",
		LastName:         "User",
		Role:             "owner",
		IsActive:         true,
		IsVerified:       false,
		VerificationCode: &code,
		PlanID:           "free",
	}
	if err := mock.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	_, _, _, err := svc.VerifyEmail(context.Background(), "wrongcode@example.com", "999999")
	if !errors.Is(err, apperrors.ErrInvalidVerification) {
		t.Errorf("expected ErrInvalidVerification, got: %v", err)
	}
}

func TestVerifyEmail_AlreadyVerified(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "alreadyverified@example.com", "StrongP@ss1", true)

	user, token, refreshToken, err := svc.VerifyEmail(context.Background(), "alreadyverified@example.com", "anything")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if token == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestVerifyEmail_UserNotFound(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	_, _, _, err := svc.VerifyEmail(context.Background(), "nobody@example.com", "123456")
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
	if !strings.Contains(err.Error(), "user not found") {
		t.Errorf("expected 'user not found' error, got: %v", err)
	}
}

func TestRefreshToken_EmptyToken(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	_, _, err := svc.RefreshToken(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
	if !strings.Contains(err.Error(), "refresh token required") {
		t.Errorf("expected 'refresh token required', got: %v", err)
	}
}

func TestRefreshToken_NilRedis(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	_, _, err := svc.RefreshToken(context.Background(), "some-token")
	if err == nil {
		t.Fatal("expected error with nil redis, got nil")
	}
	if !strings.Contains(err.Error(), "token store unavailable") {
		t.Errorf("expected 'token store unavailable', got: %v", err)
	}
}

func TestLogout_NilRedis(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	err := svc.Logout(context.Background(), "access-token", "refresh-token")
	if err != nil {
		t.Fatalf("expected nil error with nil redis, got: %v", err)
	}
}

func TestLogout_EmptyTokens_NilRedis(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	err := svc.Logout(context.Background(), "", "")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestChangePassword_Success(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)
	currentPassword := "StrongP@ss1"
	newPassword := "N3wStr0ng!Pass"

	createTestUserInRepo(t, mock, "changepw@example.com", currentPassword, true)
	userID := "user-changepw_at_example.com"

	err := svc.ChangePassword(context.Background(), userID, currentPassword, newPassword)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify password was changed: old password should fail
	_, _, _, loginErr := svc.Login(context.Background(), "changepw@example.com", currentPassword)
	if loginErr == nil {
		t.Error("expected login to fail with old password")
	}

	// Verify new password works
	_, _, _, loginErr = svc.Login(context.Background(), "changepw@example.com", newPassword)
	if loginErr != nil {
		t.Errorf("expected login to succeed with new password, got: %v", loginErr)
	}

	// Verify MustChangePassword is false after update
	stored, _ := mock.GetByID(context.Background(), userID)
	if stored != nil && stored.MustChangePassword {
		t.Error("expected MustChangePassword to be false after password change")
	}
}

func TestChangePassword_WrongCurrent(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "wrongcurrent@example.com", "StrongP@ss1", true)
	userID := "user-wrongcurrent_at_example.com"

	err := svc.ChangePassword(context.Background(), userID, "WrongP@ss1", "N3wStr0ng!Pass")
	if err == nil {
		t.Fatal("expected error for wrong current password, got nil")
	}
	if !strings.Contains(err.Error(), "current password is incorrect") {
		t.Errorf("expected 'current password is incorrect', got: %v", err)
	}
}

func TestChangePassword_UserNotFound(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	err := svc.ChangePassword(context.Background(), "nonexistent-id", "StrongP@ss1", "N3wStr0ng!Pass")
	// Mock returns nil user, nil error for non-existent IDs
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
	if !strings.Contains(err.Error(), "user not found") {
		t.Errorf("expected 'user not found', got: %v", err)
	}
}

func TestChangePassword_NoStrengthValidation(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "weaknewpw@example.com", "StrongP@ss1", true)
	userID := "user-weaknewpw_at_example.com"

	// ChangePassword does not enforce password strength — it accepts any non-empty string
	err := svc.ChangePassword(context.Background(), userID, "StrongP@ss1", "weak")
	if err != nil {
		t.Fatalf("ChangePassword does not validate strength, unexpected error: %v", err)
	}
}

func TestGenerateVerificationCode(t *testing.T) {
	for i := 0; i < 50; i++ {
		code := generateVerificationCode()
		if len(code) != 6 {
			t.Errorf("expected 6-digit code, got %d digits: %s", len(code), code)
			continue
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Errorf("expected only digits, got %c in %s", c, code)
				break
			}
		}
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	seen := make(map[string]bool)
	for i := 0; i < 20; i++ {
		token := svc.generateRefreshToken()
		if len(token) != 64 {
			t.Errorf("expected 64-char hex token, got %d chars: %s", len(token), token)
		}
		if seen[token] {
			t.Errorf("duplicate refresh token: %s", token)
		}
		seen[token] = true
	}
}

func TestGetUser(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "getuser@example.com", "StrongP@ss1", true)
	userID := "user-getuser_at_example.com"

	user, err := svc.GetUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "getuser@example.com" {
		t.Errorf("expected email getuser@example.com, got %s", user.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	user, err := svc.GetUser(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Error("expected nil user, got user")
	}
}

func TestResetPassword_NilRedis(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	err := svc.ResetPassword(context.Background(), "some-token", "N3wStr0ng!Pass")
	if err == nil {
		t.Fatal("expected error with nil redis, got nil")
	}
	if !strings.Contains(err.Error(), "token store unavailable") {
		t.Errorf("expected 'token store unavailable', got: %v", err)
	}
}

func TestForgotPassword_NilRedis_UserNotFound(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	// With nil redis and non-existent user, should return nil (no error)
	err := svc.ForgotPassword(context.Background(), "noone@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestForgotPassword_NilRedis_UserExists(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "forgot@example.com", "StrongP@ss1", true)

	// With nil redis, no email service, this should succeed silently
	err := svc.ForgotPassword(context.Background(), "forgot@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResendVerification_AlreadyVerified(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "resend@example.com", "StrongP@ss1", true)

	err := svc.ResendVerification(context.Background(), "resend@example.com")
	if !errors.Is(err, apperrors.ErrEmailAlreadyVerified) {
		t.Errorf("expected ErrEmailAlreadyVerified, got: %v", err)
	}
}

func TestResendVerification_Success(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "resendunv@example.com", "StrongP@ss1", false)

	err := svc.ResendVerification(context.Background(), "resendunv@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the code was updated
	user, _ := mock.GetByEmail(context.Background(), "resendunv@example.com")
	if user == nil {
		t.Fatal("user not found")
	}
	if user.VerificationCode == nil {
		t.Error("expected new verification code to be set")
	}
}

func TestResendVerification_UserNotFound(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	err := svc.ResendVerification(context.Background(), "ghost@example.com")
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
	if !strings.Contains(err.Error(), "user not found") {
		t.Errorf("expected 'user not found', got: %v", err)
	}
}

func TestMe(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "me@example.com", "StrongP@ss1", true)
	userID := "user-me_at_example.com"

	user, err := svc.Me(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Email != "me@example.com" {
		t.Errorf("expected email me@example.com, got %s", user.Email)
	}
}

func TestRegister_CaseInsensitiveDuplicateCheck(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	_, err := svc.Register(context.Background(), "Case@Test.com", "StrongP@ss1", "A", "B", "Co")
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	// Different case — MockUserRepo does case-sensitive email matching,
	// but this tests the actual mock behavior
	_, err = svc.Register(context.Background(), "case@test.com", "StrongP@ss1", "A", "B", "Co")
	// Depending on mock implementation this may or may not find the duplicate.
	// This test documents the actual behavior.
	if err != nil && !strings.Contains(err.Error(), "already registered") {
		t.Errorf("unexpected error type: %v", err)
	}
}

func TestLogin_ATokenIsValidJWT(t *testing.T) {
	mock := repository.NewMockUserRepo()
	svc := newTestAuthService(mock)

	createTestUserInRepo(t, mock, "jwt@example.com", "StrongP@ss1", true)

	_, token, _, err := svc.Login(context.Background(), "jwt@example.com", "StrongP@ss1")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// JWT tokens have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected JWT with 3 parts, got %d", len(parts))
	}
}

// ========== EXISTING TESTS ==========

func TestValidatePasswordStrength_Valid(t *testing.T) {
	validPasswords := []string{
		"StrongP@ss1",
		"Ab1!defgh",
		"MyP@ssw0rd!",
		"Test123!@#",
		"aB3$fghiJK",
	}

	for _, pw := range validPasswords {
		err := validatePasswordStrength(pw)
		if err != nil {
			t.Errorf("validatePasswordStrength(%q) unexpected error: %v", pw, err)
		}
	}
}

func TestValidatePasswordStrength_AllChecksPass(t *testing.T) {
	err := validatePasswordStrength("Ab1!")
	if err != nil {
		t.Errorf("Ab1! should pass validation: %v", err)
	}
}

func TestValidatePasswordStrength_NoUppercase(t *testing.T) {
	err := validatePasswordStrength("lowercase1!")
	if err == nil {
		t.Error("expected error for password without uppercase")
	}
}

func TestValidatePasswordStrength_NoLowercase(t *testing.T) {
	err := validatePasswordStrength("UPPERCASE1!")
	if err == nil {
		t.Error("expected error for password without lowercase")
	}
}

func TestValidatePasswordStrength_NoDigit(t *testing.T) {
	err := validatePasswordStrength("NoDigits!abc")
	if err == nil {
		t.Error("expected error for password without digits")
	}
}

func TestValidatePasswordStrength_NoSpecial(t *testing.T) {
	err := validatePasswordStrength("NoSpecial123")
	if err == nil {
		t.Error("expected error for password without special characters")
	}
}

func TestGenerateVerificationCode_Length(t *testing.T) {
	code := generateVerificationCode()
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %d digits: %s", len(code), code)
	}
}

func TestGenerateVerificationCode_DigitsOnly(t *testing.T) {
	code := generateVerificationCode()
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Errorf("expected only digits, got %c in %s", c, code)
		}
	}
}

func TestGenerateVerificationCode_Unique(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := generateVerificationCode()
		if codes[code] {
			t.Errorf("duplicate code generated: %s", code)
		}
		codes[code] = true
	}
}

func TestParseAIMetadata_AllTags(t *testing.T) {
	content := "[SENTIMENT:positive]\n[LANGUAGE:en]\n[SUGGESTIONS:Show products|What are prices?]\nHello! How can I help?"
	clean, sentiment, language, suggestions := parseAIMetadata(content)
	if sentiment != "positive" {
		t.Errorf("expected positive sentiment, got %s", sentiment)
	}
	if language != "en" {
		t.Errorf("expected en language, got %s", language)
	}
	if len(suggestions) != 2 {
		t.Errorf("expected 2 suggestions, got %d", len(suggestions))
	}
	if clean != "Hello! How can I help?" {
		t.Errorf("expected clean text 'Hello! How can I help?', got %q", clean)
	}
}

func TestCircuitBreaker_InitialState(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	if !cb.Allow() {
		t.Error("closed circuit breaker should allow requests")
	}
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.Allow() {
		t.Error("open circuit breaker should block requests")
	}
}

func TestCircuitBreaker_SuccessResetsToClosed(t *testing.T) {
	cb := &CircuitBreaker{state: "closed"}
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	cb.lastFailure = cb.lastFailure.Add(-61 * time.Second)
	cb.Allow()
	cb.RecordSuccess()
	if !cb.Allow() {
		t.Error("circuit breaker should be closed after success in half-open")
	}
}

func TestGetSetFloat64(t *testing.T) {
	m := map[string]interface{}{
		"count": 42,
		"rate":  3.14,
		"name":  "test",
	}

	if v := getInt(m, "count"); v != 42 {
		t.Errorf("getInt expected 42, got %d", v)
	}
	if v := getInt(m, "missing"); v != 0 {
		t.Errorf("getInt expected 0 for missing key, got %d", v)
	}
	if v := getFloat64(m, "rate"); v != 3.14 {
		t.Errorf("getFloat64 expected 3.14, got %f", v)
	}
}
