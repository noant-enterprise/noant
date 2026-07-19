// Package errors defines typed sentinel errors for the NOANT domain layer.
//
// Services return these errors to communicate domain-level conditions to handlers.
// Handlers use errors.Is() to match them, replacing the fragile pattern of
// string comparison (e.g. err.Error() == "email_not_verified").
//
// Usage:
//
//	// In a service method:
//	if !user.IsVerified {
//	    return nil, apperrors.ErrEmailNotVerified
//	}
//
//	// In a handler:
//	if errors.Is(err, apperrors.ErrEmailNotVerified) {
//	    c.JSON(403, gin.H{"error": "email_not_verified"})
//	    return
//	}
package errors

import "errors"

// Sentinel errors for domain-level error handling.
// Services return these; handlers compare with errors.Is() instead of string matching.
var (
	// Auth errors — returned by AuthService methods.
	ErrEmailNotVerified     = errors.New("email_not_verified")      // Login succeeds but email not yet verified
	ErrAccountLocked        = errors.New("account_locked")           // Too many failed login attempts (15min lockout)
	ErrInvalidCredentials   = errors.New("invalid credentials")     // Wrong email or password
	ErrInvalidVerification  = errors.New("invalid verification code") // Email verification code mismatch
	ErrTooManyVerifications = errors.New("too many verification attempts") // Rate limit on verification endpoint
	ErrEmailAlreadyVerified = errors.New("email already verified")  // ResendVerification called on verified account

	// Not-found errors — returned when a resource lookup fails.
	ErrNotFound        = errors.New("resource not found")            // Generic not-found fallback
	ErrUnknownQuestion = errors.New("unknown question not found")    // TrainingService.TrainUnknown/IgnoreUnknown
	ErrCampaign        = errors.New("campaign not found or access denied") // CampaignService.Cancel

	// Business-rule errors — returned when a business constraint is violated.
	ErrInsufficientCredit = errors.New("insufficient credit balance") // CreditService.Deduct
	ErrCreditExpired      = errors.New("credit balance has expired")  // CreditService.Deduct
	ErrCircuitBreakerOpen = errors.New("circuit breaker open")        // AIBrain when Groq API is down
)
