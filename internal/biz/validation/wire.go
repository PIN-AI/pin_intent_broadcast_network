//go:build wireinject
// +build wireinject

package validation

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"pin_intent_broadcast_network/internal/biz/common"
)

// ProviderSet is validation providers.
var ProviderSet = wire.NewSet(
	NewValidator,
	NewFormatValidator,
	NewBusinessValidator,
	NewPermissionValidator,
	NewValidatorConfig,
	NewBusinessConfig,
	NewPermissionConfig,
	wire.Bind(new(common.IntentValidator), new(*Validator)),
)

// NewValidator creates a new validator with Wire injection
func NewValidator(config *Config, logger log.Logger) *Validator {
	validator := &Validator{
		formatRules:   make([]common.FormatRule, 0),
		businessRules: make([]common.BusinessRule, 0),
		config:        config,
		logger:        log.NewHelper(logger),
	}

	// Initialize with default rules
	validator.addDefaultFormatRules()
	validator.addDefaultBusinessRules()

	return validator
}

// NewValidatorConfig creates a default validator configuration
func NewValidatorConfig() *Config {
	return &Config{
		EnableStrict:   true,
		MaxPayloadSize: common.DefaultMaxPayloadSize,
		MaxTTL:         int64(common.DefaultMaxTTL.Seconds()),
		AllowedTypes: []string{
			common.IntentTypeTrade,
			common.IntentTypeTransfer,
			common.IntentTypeLending,
			common.IntentTypeSwap,
		},
		ValidationCache: true,
	}
}

// NewBusinessConfig creates a default business validation configuration
func NewBusinessConfig() *BusinessConfig {
	return &BusinessConfig{
		MaxPayloadSize: common.DefaultMaxPayloadSize,
		MaxTTL:         common.DefaultMaxTTL,
		AllowedTypes: []string{
			common.IntentTypeTrade,
			common.IntentTypeTransfer,
			common.IntentTypeLending,
			common.IntentTypeSwap,
		},
	}
}

// NewPermissionConfig creates a default permission configuration
func NewPermissionConfig() *PermissionConfig {
	return &PermissionConfig{
		EnablePermissionCheck: false, // Disabled by default
		DefaultPermissions:    []string{},
		AdminPeers:            []string{},
	}
}
