package crypto

import (
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type OTPConfig struct {
	Issuer      string
	AccountName string
}

type OTPManager struct {
	issuer string
}

func NewOTPManager(issuer string) *OTPManager {
	return &OTPManager{issuer: issuer}
}

func (m *OTPManager) GenerateSecret(email string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      m.issuer,
		AccountName: email,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate OTP secret: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

func (m *OTPManager) ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

func (m *OTPManager) GenerateCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}
