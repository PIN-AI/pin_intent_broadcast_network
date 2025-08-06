//go:build wireinject
// +build wireinject

package security

import (
	"github.com/google/wire"

	"pin_intent_broadcast_network/internal/biz/common"
)

// ProviderSet is security providers.
var ProviderSet = wire.NewSet(
	NewSigner,
	NewKeyStore,
	NewCryptoUtils,
	NewSignerConfig,
	NewKeyStoreConfig,
	NewCryptoConfig,
	wire.Bind(new(common.IntentSigner), new(*Signer)),
	wire.Bind(new(common.KeyStore), new(*KeyStore)),
)

// NewSignerConfig creates a default signer configuration
func NewSignerConfig() *SignerConfig {
	return &SignerConfig{
		Algorithm:         "Ed25519",
		KeyRotationPeriod: 30 * 24 * 3600, // 30 days
		EnableEncryption:  true,
	}
}

// NewKeyStoreConfig creates a default keystore configuration
func NewKeyStoreConfig() *KeyStoreConfig {
	return &KeyStoreConfig{
		KeyDir:            "./keystore",
		CacheEnabled:      true,
		KeyRotationPeriod: 30 * 24 * 3600, // 30 days
		BackupEnabled:     true,
	}
}

// NewCryptoConfig creates a default crypto configuration
func NewCryptoConfig() *CryptoConfig {
	return &CryptoConfig{
		HashAlgorithm:     "SHA256",
		EncryptionEnabled: true,
		RandomSeedEnabled: true,
	}
}
