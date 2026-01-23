package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WebhookValidator validates incoming webhook requests
type WebhookValidator struct {
	nonceCache      map[string]time.Time
	nonceMu         sync.RWMutex
	maxTimestampAge time.Duration
	cleanupInterval time.Duration
}

// WebhookSecurityConfig holds security configuration for a webhook
type WebhookSecurityConfig struct {
	// HMAC signature verification
	Secret          string `json:"secret,omitempty"`
	SignatureHeader string `json:"signature_header,omitempty"` // Default: X-Webhook-Signature

	// IP allowlisting
	AllowedIPs   []string `json:"allowed_ips,omitempty"`   // CIDR notation supported
	AllowedCIDRs []string `json:"allowed_cidrs,omitempty"` // Explicit CIDRs

	// Replay attack prevention
	RequireTimestamp   bool   `json:"require_timestamp,omitempty"`
	TimestampHeader    string `json:"timestamp_header,omitempty"`      // Default: X-Webhook-Timestamp
	TimestampMaxAgeSec int64  `json:"timestamp_max_age_sec,omitempty"` // Default: 300 (5 min)
	RequireNonce       bool   `json:"require_nonce,omitempty"`
	NonceHeader        string `json:"nonce_header,omitempty"` // Default: X-Webhook-Nonce
}

// ValidationResult holds the result of webhook validation
type ValidationResult struct {
	Valid     bool   `json:"valid"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
}

// Error codes
const (
	ErrCodeInvalidSignature = "INVALID_SIGNATURE"
	ErrCodeMissingSignature = "MISSING_SIGNATURE"
	ErrCodeIPNotAllowed     = "IP_NOT_ALLOWED"
	ErrCodeMissingTimestamp = "MISSING_TIMESTAMP"
	ErrCodeTimestampExpired = "TIMESTAMP_EXPIRED"
	ErrCodeInvalidTimestamp = "INVALID_TIMESTAMP"
	ErrCodeMissingNonce     = "MISSING_NONCE"
	ErrCodeNonceReused      = "NONCE_REUSED"
)

// NewWebhookValidator creates a new webhook validator
func NewWebhookValidator() *WebhookValidator {
	v := &WebhookValidator{
		nonceCache:      make(map[string]time.Time),
		maxTimestampAge: 5 * time.Minute,
		cleanupInterval: 1 * time.Minute,
	}

	// Start background cleanup
	go v.cleanupLoop()

	return v
}

// Validate validates a webhook request
func (v *WebhookValidator) Validate(config *WebhookSecurityConfig, req *WebhookRequest) *ValidationResult {
	result := &ValidationResult{Valid: true, IPAddress: req.IPAddress}

	// 1. IP allowlist check
	if len(config.AllowedIPs) > 0 || len(config.AllowedCIDRs) > 0 {
		if !v.validateIP(req.IPAddress, config.AllowedIPs, config.AllowedCIDRs) {
			return &ValidationResult{
				Valid:     false,
				Error:     fmt.Sprintf("IP address %s is not allowed", req.IPAddress),
				ErrorCode: ErrCodeIPNotAllowed,
				IPAddress: req.IPAddress,
			}
		}
	}

	// 2. Timestamp validation (replay attack prevention)
	if config.RequireTimestamp {
		timestampHeader := config.TimestampHeader
		if timestampHeader == "" {
			timestampHeader = "X-Webhook-Timestamp"
		}

		timestampStr := req.Headers[timestampHeader]
		if timestampStr == "" {
			// Try lowercase
			timestampStr = req.Headers[strings.ToLower(timestampHeader)]
		}

		if timestampStr == "" {
			return &ValidationResult{
				Valid:     false,
				Error:     "missing timestamp header",
				ErrorCode: ErrCodeMissingTimestamp,
			}
		}

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return &ValidationResult{
				Valid:     false,
				Error:     "invalid timestamp format",
				ErrorCode: ErrCodeInvalidTimestamp,
			}
		}

		maxAge := config.TimestampMaxAgeSec
		if maxAge == 0 {
			maxAge = 300 // 5 minutes default
		}

		now := time.Now().Unix()
		if now-timestamp > maxAge || timestamp-now > 60 { // Allow 60s clock skew into future
			return &ValidationResult{
				Valid:     false,
				Error:     fmt.Sprintf("timestamp expired or too far in future (age: %ds, max: %ds)", now-timestamp, maxAge),
				ErrorCode: ErrCodeTimestampExpired,
				Timestamp: timestamp,
			}
		}

		result.Timestamp = timestamp
	}

	// 3. Nonce validation (replay attack prevention)
	if config.RequireNonce {
		nonceHeader := config.NonceHeader
		if nonceHeader == "" {
			nonceHeader = "X-Webhook-Nonce"
		}

		nonce := req.Headers[nonceHeader]
		if nonce == "" {
			nonce = req.Headers[strings.ToLower(nonceHeader)]
		}

		if nonce == "" {
			return &ValidationResult{
				Valid:     false,
				Error:     "missing nonce header",
				ErrorCode: ErrCodeMissingNonce,
			}
		}

		if v.isNonceUsed(nonce) {
			return &ValidationResult{
				Valid:     false,
				Error:     "nonce has already been used",
				ErrorCode: ErrCodeNonceReused,
				Nonce:     nonce,
			}
		}

		v.recordNonce(nonce)
		result.Nonce = nonce
	}

	// 4. Signature verification
	if config.Secret != "" {
		sigHeader := config.SignatureHeader
		if sigHeader == "" {
			sigHeader = "X-Webhook-Signature"
		}

		signature := req.Headers[sigHeader]
		if signature == "" {
			signature = req.Headers[strings.ToLower(sigHeader)]
		}

		if signature == "" {
			return &ValidationResult{
				Valid:     false,
				Error:     "missing signature header",
				ErrorCode: ErrCodeMissingSignature,
			}
		}

		// Build signing payload (timestamp + body for extra security)
		var signingPayload string
		if result.Timestamp > 0 {
			signingPayload = fmt.Sprintf("%d.%s", result.Timestamp, req.RawBody)
		} else {
			signingPayload = req.RawBody
		}

		if !v.verifySignature(config.Secret, signingPayload, signature) {
			return &ValidationResult{
				Valid:     false,
				Error:     "invalid signature",
				ErrorCode: ErrCodeInvalidSignature,
			}
		}
	}

	return result
}

// WebhookRequest represents an incoming webhook request
type WebhookRequest struct {
	Headers   map[string]string
	RawBody   string
	IPAddress string
}

// verifySignature verifies HMAC-SHA256 signature
func (v *WebhookValidator) verifySignature(secret, payload, signature string) bool {
	// Remove common prefixes
	signature = strings.TrimPrefix(signature, "sha256=")
	signature = strings.TrimPrefix(signature, "v1=")

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	expected := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

// validateIP checks if IP is in allowlist
func (v *WebhookValidator) validateIP(ipStr string, allowedIPs, allowedCIDRs []string) bool {
	// Parse the IP (handle port if present)
	host, _, err := net.SplitHostPort(ipStr)
	if err != nil {
		host = ipStr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	// Check exact IPs
	for _, allowed := range allowedIPs {
		allowedIP := net.ParseIP(allowed)
		if allowedIP != nil && ip.Equal(allowedIP) {
			return true
		}

		// Try as CIDR if not a plain IP
		_, cidr, err := net.ParseCIDR(allowed)
		if err == nil && cidr.Contains(ip) {
			return true
		}
	}

	// Check CIDRs
	for _, cidrStr := range allowedCIDRs {
		_, cidr, err := net.ParseCIDR(cidrStr)
		if err == nil && cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// isNonceUsed checks if a nonce has been used
func (v *WebhookValidator) isNonceUsed(nonce string) bool {
	v.nonceMu.RLock()
	defer v.nonceMu.RUnlock()
	_, exists := v.nonceCache[nonce]
	return exists
}

// recordNonce records a nonce as used
func (v *WebhookValidator) recordNonce(nonce string) {
	v.nonceMu.Lock()
	defer v.nonceMu.Unlock()
	v.nonceCache[nonce] = time.Now()
}

// cleanupLoop periodically removes old nonces
func (v *WebhookValidator) cleanupLoop() {
	ticker := time.NewTicker(v.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		v.cleanupNonces()
	}
}

// cleanupNonces removes nonces older than maxTimestampAge
func (v *WebhookValidator) cleanupNonces() {
	v.nonceMu.Lock()
	defer v.nonceMu.Unlock()

	cutoff := time.Now().Add(-v.maxTimestampAge)
	for nonce, timestamp := range v.nonceCache {
		if timestamp.Before(cutoff) {
			delete(v.nonceCache, nonce)
		}
	}
}

// GenerateSignature generates an HMAC-SHA256 signature for outgoing webhooks
func GenerateSignature(secret, payload string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// GenerateSignatureWithTimestamp generates a signature with timestamp
func GenerateSignatureWithTimestamp(secret, payload string, timestamp int64) string {
	signingPayload := fmt.Sprintf("%d.%s", timestamp, payload)
	return GenerateSignature(secret, signingPayload)
}
