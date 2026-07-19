package errors

import "errors"

// Sentinel errors for domain-level error handling.
// Services return these; handlers compare with errors.Is() instead of string matching.

var (
	// Auth errors
	ErrEmailNotVerified     = errors.New("email_not_verified")
	ErrAccountLocked        = errors.New("account_locked")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidVerification  = errors.New("invalid verification code")
	ErrTooManyVerifications = errors.New("too many verification attempts")
	ErrEmailAlreadyVerified = errors.New("email already verified")

	// Not found errors
	ErrNotFound        = errors.New("resource not found")
	ErrUnknownQuestion = errors.New("unknown question not found")
	ErrCampaign        = errors.New("campaign not found or access denied")

	// Business errors
	ErrInsufficientCredit = errors.New("insufficient credit balance")
	ErrCreditExpired      = errors.New("credit balance has expired")
	ErrCircuitBreakerOpen = errors.New("circuit breaker open")
)
