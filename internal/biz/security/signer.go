package security

import (
	"crypto"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"pin_intent_broadcast_network/internal/biz/common"
)

// Signer implements the IntentSigner interface
// It handles digital signing and verification of intents
type Signer struct {
	keyStore  common.KeyStore
	algorithm common.SignatureAlgorithm
	config    *SignerConfig
}

// SignerConfig holds configuration for the signer
type SignerConfig struct {
	Algorithm         string `yaml:"algorithm"`
	KeyRotationPeriod int64  `yaml:"key_rotation_period"`
	EnableEncryption  bool   `yaml:"enable_encryption"`
}

// NewSigner creates a new intent signer
func NewSigner(keyStore common.KeyStore, algorithm common.SignatureAlgorithm, config *SignerConfig) *Signer {
	return &Signer{
		keyStore:  keyStore,
		algorithm: algorithm,
		config:    config,
	}
}

// SignIntent signs an intent with the provided private key
func (s *Signer) SignIntent(intent *common.Intent, privateKey crypto.PrivateKey) error {
	// TODO: Implement in task 4.1

	// Create signature data
	signData, err := s.createSignatureData(intent)
	if err != nil {
		return fmt.Errorf("failed to create signature data: %w", err)
	}

	// Generate signature
	signature, err := s.algorithm.Sign(signData, privateKey)
	if err != nil {
		return fmt.Errorf("signing failed: %w", err)
	}

	// Set signature information
	intent.Signature = signature
	intent.SignatureAlgorithm = s.algorithm.GetAlgorithmName()

	return nil
}

// VerifySignature verifies the signature of an intent
func (s *Signer) VerifySignature(intent *common.Intent) error {
	// TODO: Implement in task 4.1

	if len(intent.Signature) == 0 {
		return errors.New("intent has no signature")
	}

	// Get sender's public key
	publicKey, err := s.keyStore.GetPublicKey(peer.ID(intent.SenderID))
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Create signature data
	signData, err := s.createSignatureData(intent)
	if err != nil {
		return fmt.Errorf("failed to create signature data: %w", err)
	}

	// Verify signature
	if err := s.algorithm.Verify(signData, intent.Signature, publicKey); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// GenerateKeyPair generates a new key pair
func (s *Signer) GenerateKeyPair() (crypto.PrivateKey, crypto.PublicKey, error) {
	// TODO: Implement in task 4.1
	return s.keyStore.GenerateKeyPair()
}

// GetPublicKey retrieves a public key for a peer
func (s *Signer) GetPublicKey(peerID peer.ID) (crypto.PublicKey, error) {
	// TODO: Implement in task 4.1
	return s.keyStore.GetPublicKey(peerID)
}

// createSignatureData creates standardized data for signing
func (s *Signer) createSignatureData(intent *common.Intent) ([]byte, error) {
	// TODO: Implement signature data creation
	// Create standardized signature data structure
	signData := &common.IntentSignatureData{
		ID:        intent.ID,
		Type:      intent.Type,
		Payload:   intent.Payload,
		Timestamp: intent.Timestamp,
		SenderID:  intent.SenderID,
		Metadata:  intent.Metadata,
		Priority:  intent.Priority,
		TTL:       intent.TTL,
	}

	// Serialize to bytes for signing
	return serializeSignatureData(signData)
}

// serializeSignatureData serializes signature data to bytes
func serializeSignatureData(data *common.IntentSignatureData) ([]byte, error) {
	// TODO: Implement serialization logic
	return nil, nil
}
