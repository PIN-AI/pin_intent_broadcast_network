//go:build wireinject
// +build wireinject

package matching

import (
	"github.com/google/wire"

	"pin_intent_broadcast_network/internal/biz/common"
)

// ProviderSet is matching providers.
var ProviderSet = wire.NewSet(
	NewEngine,
	NewMatcher,
	NewEngineConfig,
	NewMatcherConfig,
	wire.Bind(new(common.IntentMatcher), new(*Matcher)),
)

// NewEngineConfig creates a default engine configuration
func NewEngineConfig() *EngineConfig {
	return &EngineConfig{
		ConfidenceThreshold: 0.8,
		MaxMatchesPerIntent: 10,
		MatchingTimeout:     5,
		EnableCaching:       true,
		CacheSize:           1000,
	}
}

// NewMatcherConfig creates a default matcher configuration
func NewMatcherConfig() *MatcherConfig {
	return &MatcherConfig{
		EnableContentMatching:  true,
		EnableMetadataMatching: true,
		ContentWeight:          0.4,
		MetadataWeight:         0.3,
		TypeWeight:             0.3,
	}
}
