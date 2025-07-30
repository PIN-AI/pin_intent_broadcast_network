package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"pin_intent_broadcast_network/internal/biz/common"
)

// CryptoUtils provides cryptographic utility functions
// This file will contain the implementation for task 4.3
type CryptoUtils struct {
	config *CryptoConfig
}

// CryptoConfig holds configuration for crypto utilities
type CryptoConfig struct {
	HashAlgorithm     string `yaml:"hash_algorithm"`
	EncryptionEnabled bool   `yaml:"encryption_enabled"`
	RandomSeedEnabled bool   `yaml:"random_seed_enabled"`
}

// NewCryptoUtils creates a new crypto utilities instance
func NewCryptoUtils(config *CryptoConfig) *CryptoUtils {
	return &CryptoUtils{
		config: config,
	}
}

// GenerateIntentID generates a unique ID for an intent
func GenerateIntentID() string {
	// TODO: Implement in task 4.3
	// Generate based on timestamp and random bytes
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)

	data := append([]byte(fmt.Sprintf("%d", timestamp)), randomBytes...)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes as ID
}

// HashIntent creates a hash of an intent for integrity verification
func HashIntent(intent *common.Intent) ([]byte, error) {
	// TODO: Implement in task 4.3
	// Create standardized hash data
	hashData := &IntentHashData{
		ID:        intent.ID,
		Type:      intent.Type,
		Payload:   intent.Payload,
		Timestamp: intent.Timestamp,
		SenderID:  intent.SenderID,
		Metadata:  intent.Metadata,
	}

	data, err := json.Marshal(hashData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal intent for hashing: %w", err)
	}

	hash := sha256.Sum256(data)
	return hash[:], nil
}

// IntentHashData represents data used for intent hashing
type IntentHashData struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Payload   []byte            `json:"payload"`
	Timestamp int64             `json:"timestamp"`
	SenderID  string            `json:"sender_id"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(length int) ([]byte, error) {
	// TODO: Implement in task 4.3
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// EncryptData encrypts data using configured encryption algorithm
func (cu *CryptoUtils) EncryptData(data []byte, key []byte) ([]byte, error) {
	// TODO: Implement in task 4.3
	if !cu.config.EncryptionEnabled {
		return data, nil // Encryption disabled
	}

	// Implement encryption logic
	return nil, fmt.Errorf("encryption not implemented")
}

// DecryptData decrypts data using configured encryption algorithm
func (cu *CryptoUtils) DecryptData(encryptedData []byte, key []byte) ([]byte, error) {
	// TODO: Implement in task 4.3
	if !cu.config.EncryptionEnabled {
		return encryptedData, nil // Encryption disabled
	}

	// Implement decryption logic
	return nil, fmt.Errorf("decryption not implemented")
}

// ComputeHash computes hash of data using configured algorithm
func (cu *CryptoUtils) ComputeHash(data []byte) []byte {
	// TODO: Implement in task 4.3
	switch cu.config.HashAlgorithm {
	case "sha256":
		hash := sha256.Sum256(data)
		return hash[:]
	default:
		// Default to SHA256
		hash := sha256.Sum256(data)
		return hash[:]
	}
}

// VerifyHash verifies if data matches the provided hash
func (cu *CryptoUtils) VerifyHash(data []byte, expectedHash []byte) bool {
	// TODO: Implement in task 4.3
	actualHash := cu.ComputeHash(data)
	return compareBytes(actualHash, expectedHash)
}

// compareBytes securely compares two byte slices
func compareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}
