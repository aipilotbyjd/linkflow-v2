package logic

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"

	"github.com/linkflow-ai/linkflow/internal/worker/core"
)

// CryptoNode performs cryptographic operations
type CryptoNode struct{}

func (n *CryptoNode) Type() string {
	return "logic.crypto"
}

func (n *CryptoNode) Execute(ctx context.Context, execCtx *core.ExecutionContext) (map[string]interface{}, error) {
	config := execCtx.Config
	input := execCtx.Input

	operation := core.GetString(config, "operation", "hash")

	switch operation {
	case "hash":
		return n.hash(config, input)
	case "hmac":
		return n.hmacSign(config, input)
	case "encrypt":
		return n.encrypt(config, input)
	case "decrypt":
		return n.decrypt(config, input)
	case "base64encode":
		return n.base64Encode(config, input)
	case "base64decode":
		return n.base64Decode(config, input)
	case "generateKey":
		return n.generateKey(config)
	case "generateIV":
		return n.generateIV(config)
	case "randomBytes":
		return n.randomBytes(config)
	default:
		return n.hash(config, input)
	}
}

func (n *CryptoNode) hash(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getDataInput(config, input)
	algorithm := core.GetString(config, "algorithm", "sha256")
	encoding := core.GetString(config, "encoding", "hex")

	var h hash.Hash
	switch algorithm {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha384":
		h = sha512.New384()
	case "sha512":
		h = sha512.New()
	default:
		h = sha256.New()
	}

	h.Write([]byte(data))
	hashBytes := h.Sum(nil)

	var result string
	switch encoding {
	case "base64":
		result = base64.StdEncoding.EncodeToString(hashBytes)
	default:
		result = hex.EncodeToString(hashBytes)
	}

	return map[string]interface{}{
		"hash":      result,
		"algorithm": algorithm,
		"encoding":  encoding,
	}, nil
}

func (n *CryptoNode) hmacSign(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getDataInput(config, input)
	secret := core.GetString(config, "secret", "")
	algorithm := core.GetString(config, "algorithm", "sha256")
	encoding := core.GetString(config, "encoding", "hex")

	if secret == "" {
		return nil, fmt.Errorf("secret is required for HMAC")
	}

	var h func() hash.Hash
	switch algorithm {
	case "md5":
		h = md5.New
	case "sha1":
		h = sha1.New
	case "sha256":
		h = sha256.New
	case "sha384":
		h = sha512.New384
	case "sha512":
		h = sha512.New
	default:
		h = sha256.New
	}

	mac := hmac.New(h, []byte(secret))
	mac.Write([]byte(data))
	hashBytes := mac.Sum(nil)

	var result string
	switch encoding {
	case "base64":
		result = base64.StdEncoding.EncodeToString(hashBytes)
	default:
		result = hex.EncodeToString(hashBytes)
	}

	return map[string]interface{}{
		"signature": result,
		"algorithm": "hmac-" + algorithm,
		"encoding":  encoding,
	}, nil
}

func (n *CryptoNode) encrypt(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	plaintext := getDataInput(config, input)
	key := core.GetString(config, "key", "")
	algorithm := core.GetString(config, "algorithm", "aes-256-gcm")

	if key == "" {
		return nil, fmt.Errorf("encryption key is required")
	}

	// Derive key to correct length
	keyBytes := deriveKey(key, 32) // AES-256

	switch algorithm {
	case "aes-256-gcm", "aes-gcm":
		return n.encryptAESGCM([]byte(plaintext), keyBytes)
	case "aes-256-cbc", "aes-cbc":
		return n.encryptAESCBC([]byte(plaintext), keyBytes)
	default:
		return n.encryptAESGCM([]byte(plaintext), keyBytes)
	}
}

func (n *CryptoNode) encryptAESGCM(plaintext, key []byte) (map[string]interface{}, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)

	return map[string]interface{}{
		"ciphertext": base64.StdEncoding.EncodeToString(ciphertext),
		"nonce":      base64.StdEncoding.EncodeToString(nonce),
		"algorithm":  "aes-256-gcm",
	}, nil
}

func (n *CryptoNode) encryptAESCBC(plaintext, key []byte) (map[string]interface{}, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Pad plaintext to block size
	blockSize := block.BlockSize()
	padding := blockSize - (len(plaintext) % blockSize)
	paddedPlaintext := make([]byte, len(plaintext)+padding)
	copy(paddedPlaintext, plaintext)
	for i := len(plaintext); i < len(paddedPlaintext); i++ {
		paddedPlaintext[i] = byte(padding)
	}

	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedPlaintext))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	return map[string]interface{}{
		"ciphertext": base64.StdEncoding.EncodeToString(ciphertext),
		"iv":         base64.StdEncoding.EncodeToString(iv),
		"algorithm":  "aes-256-cbc",
	}, nil
}

func (n *CryptoNode) decrypt(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	ciphertextB64 := core.GetString(config, "ciphertext", "")
	if ciphertextB64 == "" {
		if ct, ok := input["ciphertext"].(string); ok {
			ciphertextB64 = ct
		}
	}

	key := core.GetString(config, "key", "")
	algorithm := core.GetString(config, "algorithm", "aes-256-gcm")

	if key == "" {
		return nil, fmt.Errorf("decryption key is required")
	}
	if ciphertextB64 == "" {
		return nil, fmt.Errorf("ciphertext is required")
	}

	keyBytes := deriveKey(key, 32)

	switch algorithm {
	case "aes-256-gcm", "aes-gcm":
		nonceB64 := core.GetString(config, "nonce", "")
		if nonceB64 == "" {
			if n, ok := input["nonce"].(string); ok {
				nonceB64 = n
			}
		}
		return n.decryptAESGCM(ciphertextB64, nonceB64, keyBytes)
	case "aes-256-cbc", "aes-cbc":
		ivB64 := core.GetString(config, "iv", "")
		if ivB64 == "" {
			if iv, ok := input["iv"].(string); ok {
				ivB64 = iv
			}
		}
		return n.decryptAESCBC(ciphertextB64, ivB64, keyBytes)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

func (n *CryptoNode) decryptAESGCM(ciphertextB64, nonceB64 string, key []byte) (map[string]interface{}, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return map[string]interface{}{
		"plaintext": string(plaintext),
		"algorithm": "aes-256-gcm",
	}, nil
}

func (n *CryptoNode) decryptAESCBC(ciphertextB64, ivB64 string, key []byte) (map[string]interface{}, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(ivB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	if len(plaintext) > 0 {
		padding := int(plaintext[len(plaintext)-1])
		if padding > 0 && padding <= block.BlockSize() {
			plaintext = plaintext[:len(plaintext)-padding]
		}
	}

	return map[string]interface{}{
		"plaintext": string(plaintext),
		"algorithm": "aes-256-cbc",
	}, nil
}

func (n *CryptoNode) base64Encode(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getDataInput(config, input)
	urlSafe := core.GetBool(config, "urlSafe", false)

	var encoded string
	if urlSafe {
		encoded = base64.URLEncoding.EncodeToString([]byte(data))
	} else {
		encoded = base64.StdEncoding.EncodeToString([]byte(data))
	}

	return map[string]interface{}{
		"encoded": encoded,
	}, nil
}

func (n *CryptoNode) base64Decode(config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data := getDataInput(config, input)
	urlSafe := core.GetBool(config, "urlSafe", false)

	var decoded []byte
	var err error
	if urlSafe {
		decoded, err = base64.URLEncoding.DecodeString(data)
	} else {
		decoded, err = base64.StdEncoding.DecodeString(data)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return map[string]interface{}{
		"decoded": string(decoded),
	}, nil
}

func (n *CryptoNode) generateKey(config map[string]interface{}) (map[string]interface{}, error) {
	length := core.GetInt(config, "length", 32)
	encoding := core.GetString(config, "encoding", "hex")

	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	var result string
	switch encoding {
	case "base64":
		result = base64.StdEncoding.EncodeToString(key)
	default:
		result = hex.EncodeToString(key)
	}

	return map[string]interface{}{
		"key":      result,
		"length":   length,
		"encoding": encoding,
	}, nil
}

func (n *CryptoNode) generateIV(config map[string]interface{}) (map[string]interface{}, error) {
	length := core.GetInt(config, "length", 16) // Default 16 bytes for AES
	encoding := core.GetString(config, "encoding", "base64")

	iv := make([]byte, length)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	var result string
	switch encoding {
	case "hex":
		result = hex.EncodeToString(iv)
	default:
		result = base64.StdEncoding.EncodeToString(iv)
	}

	return map[string]interface{}{
		"iv":       result,
		"length":   length,
		"encoding": encoding,
	}, nil
}

func (n *CryptoNode) randomBytes(config map[string]interface{}) (map[string]interface{}, error) {
	length := core.GetInt(config, "length", 32)
	encoding := core.GetString(config, "encoding", "hex")

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	var result string
	switch encoding {
	case "base64":
		result = base64.StdEncoding.EncodeToString(bytes)
	default:
		result = hex.EncodeToString(bytes)
	}

	return map[string]interface{}{
		"bytes":    result,
		"length":   length,
		"encoding": encoding,
	}, nil
}

// Helper functions

func getDataInput(config map[string]interface{}, input map[string]interface{}) string {
	if data := core.GetString(config, "data", ""); data != "" {
		return data
	}
	if data := core.GetString(config, "text", ""); data != "" {
		return data
	}
	if data, ok := input["data"].(string); ok {
		return data
	}
	if data, ok := input["text"].(string); ok {
		return data
	}
	if data, ok := input["body"].(string); ok {
		return data
	}
	return ""
}

func deriveKey(key string, length int) []byte {
	keyBytes := []byte(key)
	if len(keyBytes) >= length {
		return keyBytes[:length]
	}

	// Pad key with SHA256 hash
	h := sha256.New()
	h.Write(keyBytes)
	hash := h.Sum(nil)

	if len(hash) >= length {
		return hash[:length]
	}

	// For longer keys, concatenate hashes
	result := make([]byte, 0, length)
	counter := byte(0)
	for len(result) < length {
		h.Reset()
		h.Write(keyBytes)
		h.Write([]byte{counter})
		result = append(result, h.Sum(nil)...)
		counter++
	}
	return result[:length]
}

// Note: CryptoNode is registered in logic/init.go
