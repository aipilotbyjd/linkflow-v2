package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

// Signer provides cryptographic signing functionality
type Signer struct {
	algorithm string
	key       interface{}
}

// NewHMACSigner creates a new HMAC-SHA256 signer
func NewHMACSigner(secret []byte) *Signer {
	return &Signer{
		algorithm: "HMAC-SHA256",
		key:       secret,
	}
}

// NewRSASigner creates a new RSA-SHA256 signer
func NewRSASigner(privateKey *rsa.PrivateKey) *Signer {
	return &Signer{
		algorithm: "RSA-SHA256",
		key:       privateKey,
	}
}

// NewECDSASigner creates a new ECDSA-SHA256 signer
func NewECDSASigner(privateKey *ecdsa.PrivateKey) *Signer {
	return &Signer{
		algorithm: "ECDSA-SHA256",
		key:       privateKey,
	}
}

// NewEd25519Signer creates a new Ed25519 signer
func NewEd25519Signer(privateKey ed25519.PrivateKey) *Signer {
	return &Signer{
		algorithm: "Ed25519",
		key:       privateKey,
	}
}

// Sign signs the given data
func (s *Signer) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)

	switch key := s.key.(type) {
	case []byte: // HMAC
		h := hmac.New(sha256.New, key)
		h.Write(data)
		return h.Sum(nil), nil

	case *rsa.PrivateKey:
		return rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])

	case *ecdsa.PrivateKey:
		return ecdsa.SignASN1(rand.Reader, key, hash[:])

	case ed25519.PrivateKey:
		return ed25519.Sign(key, data), nil

	default:
		return nil, fmt.Errorf("unsupported key type")
	}
}

// SignHex returns the signature as a hex string
func (s *Signer) SignHex(data []byte) (string, error) {
	sig, err := s.Sign(data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

// SignBase64 returns the signature as a base64 string
func (s *Signer) SignBase64(data []byte) (string, error) {
	sig, err := s.Sign(data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// Verify verifies a signature
func (s *Signer) Verify(data, signature []byte) bool {
	hash := sha256.Sum256(data)

	switch key := s.key.(type) {
	case []byte: // HMAC
		expected, _ := s.Sign(data)
		return hmac.Equal(signature, expected)

	case *rsa.PrivateKey:
		return rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, hash[:], signature) == nil

	case *ecdsa.PrivateKey:
		return ecdsa.VerifyASN1(&key.PublicKey, hash[:], signature)

	case ed25519.PrivateKey:
		return ed25519.Verify(key.Public().(ed25519.PublicKey), data, signature)

	default:
		return false
	}
}

// GenerateRSAKeyPair generates a new RSA key pair
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

// GenerateECDSAKeyPair generates a new ECDSA key pair
func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// GenerateEd25519KeyPair generates a new Ed25519 key pair
func GenerateEd25519KeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// EncodePrivateKeyToPEM encodes a private key to PEM format
func EncodePrivateKeyToPEM(key interface{}) ([]byte, error) {
	var der []byte
	var keyType string
	var err error

	switch k := key.(type) {
	case *rsa.PrivateKey:
		der = x509.MarshalPKCS1PrivateKey(k)
		keyType = "RSA PRIVATE KEY"
	case *ecdsa.PrivateKey:
		der, err = x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		keyType = "EC PRIVATE KEY"
	case ed25519.PrivateKey:
		der, err = x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			return nil, err
		}
		keyType = "PRIVATE KEY"
	default:
		return nil, fmt.Errorf("unsupported key type")
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  keyType,
		Bytes: der,
	}), nil
}

// DecodePrivateKeyFromPEM decodes a private key from PEM format
func DecodePrivateKeyFromPEM(data []byte) (interface{}, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}
}

// WebhookSignature generates a webhook signature for payload verification
func WebhookSignature(secret, payload string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyWebhookSignature verifies a webhook signature
func VerifyWebhookSignature(secret, payload, signature string) bool {
	expected := WebhookSignature(secret, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}
