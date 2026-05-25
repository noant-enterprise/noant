package service

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"

	"github.com/pquerna/otp/totp"
)

type TFAService struct{}

func NewTFAService() *TFAService {
	return &TFAService{}
}

func (s *TFAService) GenerateSecret(userID string) (string, string, error) {
	key := make([]byte, 20)
	if _, err := rand.Read(key); err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	secret := base32.StdEncoding.EncodeToString(key)

	// Generate proper otpauth URI using the library
	uri, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Noant",
		AccountName: userID,
		Secret:      key,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP URI: %w", err)
	}

	return secret, uri.URL(), nil
}

func (s *TFAService) ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}