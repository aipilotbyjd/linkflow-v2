package webhook

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"
)

type SignatureVerifier struct {
	algorithm string
	secret    string
	encoding  string
	prefix    string
}

func NewSignatureVerifier(algorithm, secret string) *SignatureVerifier {
	return &SignatureVerifier{
		algorithm: algorithm,
		secret:    secret,
		encoding:  "hex",
		prefix:    "",
	}
}

func (v *SignatureVerifier) WithEncoding(encoding string) *SignatureVerifier {
	v.encoding = encoding
	return v
}

func (v *SignatureVerifier) WithPrefix(prefix string) *SignatureVerifier {
	v.prefix = prefix
	return v
}

func (v *SignatureVerifier) Verify(payload []byte, signature string) bool {
	expected := v.Sign(payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (v *SignatureVerifier) Sign(payload []byte) string {
	var h hash.Hash

	switch strings.ToLower(v.algorithm) {
	case "md5":
		h = hmac.New(md5.New, []byte(v.secret))
	case "sha1":
		h = hmac.New(sha1.New, []byte(v.secret))
	case "sha256":
		h = hmac.New(sha256.New, []byte(v.secret))
	case "sha512":
		h = hmac.New(sha512.New, []byte(v.secret))
	default:
		h = hmac.New(sha256.New, []byte(v.secret))
	}

	h.Write(payload)
	sum := h.Sum(nil)

	var encoded string
	switch strings.ToLower(v.encoding) {
	case "base64":
		encoded = base64.StdEncoding.EncodeToString(sum)
	case "hex":
		encoded = hex.EncodeToString(sum)
	default:
		encoded = hex.EncodeToString(sum)
	}

	if v.prefix != "" {
		return v.prefix + encoded
	}
	return encoded
}

type GitHubSignatureVerifier struct {
	*SignatureVerifier
}

func NewGitHubSignatureVerifier(secret string) *GitHubSignatureVerifier {
	return &GitHubSignatureVerifier{
		SignatureVerifier: NewSignatureVerifier("sha256", secret).WithPrefix("sha256="),
	}
}

func (v *GitHubSignatureVerifier) Verify(payload []byte, signature string) bool {
	signature = strings.TrimPrefix(signature, "sha256=")
	expected := strings.TrimPrefix(v.Sign(payload), "sha256=")
	return hmac.Equal([]byte(expected), []byte(signature))
}

type StripeSignatureVerifier struct {
	secret string
}

func NewStripeSignatureVerifier(secret string) *StripeSignatureVerifier {
	return &StripeSignatureVerifier{secret: secret}
}

func (v *StripeSignatureVerifier) Verify(payload []byte, header string) bool {
	parts := strings.Split(header, ",")
	var timestamp, signature string

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signature = kv[1]
		}
	}

	if timestamp == "" || signature == "" {
		return false
	}

	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(payload))
	h := hmac.New(sha256.New, []byte(v.secret))
	h.Write([]byte(signedPayload))
	expected := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

type SlackSignatureVerifier struct {
	signingSecret string
	version       string
}

func NewSlackSignatureVerifier(signingSecret string) *SlackSignatureVerifier {
	return &SlackSignatureVerifier{
		signingSecret: signingSecret,
		version:       "v0",
	}
}

func (v *SlackSignatureVerifier) Verify(payload []byte, timestamp, signature string) bool {
	baseString := fmt.Sprintf("%s:%s:%s", v.version, timestamp, string(payload))
	h := hmac.New(sha256.New, []byte(v.signingSecret))
	h.Write([]byte(baseString))
	expected := fmt.Sprintf("v0=%s", hex.EncodeToString(h.Sum(nil)))

	return hmac.Equal([]byte(expected), []byte(signature))
}

type TwilioSignatureVerifier struct {
	authToken string
}

func NewTwilioSignatureVerifier(authToken string) *TwilioSignatureVerifier {
	return &TwilioSignatureVerifier{authToken: authToken}
}

func (v *TwilioSignatureVerifier) Verify(url string, params map[string]string, signature string) bool {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}

	data := url
	for _, k := range keys {
		data += k + params[k]
	}

	h := hmac.New(sha1.New, []byte(v.authToken))
	h.Write([]byte(data))
	expected := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

type GenericSignatureVerifier struct {
	config SignatureConfig
}

type SignatureConfig struct {
	Algorithm     string
	Secret        string
	Encoding      string
	HeaderName    string
	Prefix        string
	TimestampName string
	Version       string
}

func NewGenericSignatureVerifier(config SignatureConfig) *GenericSignatureVerifier {
	if config.Algorithm == "" {
		config.Algorithm = "sha256"
	}
	if config.Encoding == "" {
		config.Encoding = "hex"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-Signature"
	}
	return &GenericSignatureVerifier{config: config}
}

func (v *GenericSignatureVerifier) Verify(payload []byte, signature string) bool {
	signature = strings.TrimPrefix(signature, v.config.Prefix)

	var h hash.Hash
	switch strings.ToLower(v.config.Algorithm) {
	case "md5":
		h = hmac.New(md5.New, []byte(v.config.Secret))
	case "sha1":
		h = hmac.New(sha1.New, []byte(v.config.Secret))
	case "sha256":
		h = hmac.New(sha256.New, []byte(v.config.Secret))
	case "sha512":
		h = hmac.New(sha512.New, []byte(v.config.Secret))
	default:
		h = hmac.New(sha256.New, []byte(v.config.Secret))
	}

	h.Write(payload)
	sum := h.Sum(nil)

	var expected string
	switch strings.ToLower(v.config.Encoding) {
	case "base64":
		expected = base64.StdEncoding.EncodeToString(sum)
	default:
		expected = hex.EncodeToString(sum)
	}

	return hmac.Equal([]byte(expected), []byte(signature))
}
