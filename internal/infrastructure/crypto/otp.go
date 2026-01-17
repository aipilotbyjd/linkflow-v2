package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	otpDigits = 6
	otpPeriod = 30 // seconds
	otpSkew   = 1  // allow 1 period before/after
)

// OTP provides TOTP (Time-based One-Time Password) functionality
type OTP struct {
	issuer string
}

// NewOTP creates a new OTP instance
func NewOTP(issuer string) *OTP {
	return &OTP{issuer: issuer}
}

// GenerateSecret generates a new random secret for TOTP
func (o *OTP) GenerateSecret() (string, error) {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// GenerateURI generates an otpauth:// URI for QR code generation
func (o *OTP) GenerateURI(secret, accountName string) string {
	return fmt.Sprintf(
		"otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		o.issuer,
		accountName,
		secret,
		o.issuer,
		otpDigits,
		otpPeriod,
	)
}

// Validate validates a TOTP code against the secret
func (o *OTP) Validate(secret, code string) bool {
	// Clean up the secret
	secret = strings.ToUpper(strings.TrimSpace(secret))

	// Decode the secret
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return false
	}

	now := time.Now().Unix()
	counter := now / otpPeriod

	// Check current and adjacent time periods
	for i := -otpSkew; i <= otpSkew; i++ {
		expected := generateTOTP(key, counter+int64(i))
		if code == expected {
			return true
		}
	}

	return false
}

// Generate generates the current TOTP code for a secret
func (o *OTP) Generate(secret string) (string, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))

	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("invalid secret: %w", err)
	}

	counter := time.Now().Unix() / otpPeriod
	return generateTOTP(key, counter), nil
}

// RemainingSeconds returns seconds until the current code expires
func (o *OTP) RemainingSeconds() int {
	now := time.Now().Unix()
	return otpPeriod - int(now%otpPeriod)
}

func generateTOTP(key []byte, counter int64) string {
	// Convert counter to bytes
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(counter))

	// Generate HMAC-SHA1
	h := hmac.New(sha1.New, key)
	h.Write(msg)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff
	code = code % 1000000

	return fmt.Sprintf("%06d", code)
}

// RecoveryCode generates a set of recovery codes
func GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code := make([]byte, 10)
		if _, err := rand.Read(code); err != nil {
			return nil, fmt.Errorf("failed to generate recovery code: %w", err)
		}
		codes[i] = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(code)[:10]
	}
	return codes, nil
}

// HashRecoveryCode hashes a recovery code for storage
func HashRecoveryCode(code string) string {
	hash, _ := HashPassword(code)
	return hash
}

// ValidateRecoveryCode checks if a recovery code matches
func ValidateRecoveryCode(code, hash string) bool {
	return CheckPassword(code, hash)
}
