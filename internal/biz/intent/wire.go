//go:build wireinject
// +build wireinject

package intent

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"pin_intent_broadcast_network/internal/biz/common"
)

// ProviderSet is intent providers.
var ProviderSet = wire.NewSet(
	NewManager,
	NewLifecycleManager,
	wire.Bind(new(common.IntentManager), new(*Manager)),
	wire.Bind(new(common.LifecycleManager), new(*LifecycleManager)),
)

// NewManager creates a new Intent Manager instance with Wire injection
func NewManager(
	validator common.IntentValidator,
	signer common.IntentSigner,
	processor common.IntentProcessor,
	matcher common.IntentMatcher,
	lifecycle common.LifecycleManager,
	config *Config,
	logger log.Logger,
) *Manager {
	return &Manager{
		validator: validator,
		signer:    signer,
		processor: processor,
		matcher:   matcher,
		lifecycle: lifecycle,
		config:    config,
		metrics:   &Metrics{},
		logger:    log.NewHelper(logger),
		intents:   make(map[string]*common.Intent),
	}
}

// NewLifecycleManager creates a new lifecycle manager with Wire injection
func NewLifecycleManager(config *LifecycleConfig) *LifecycleManager {
	if config == nil {
		config = &LifecycleConfig{
			CleanupInterval: 5 * time.Minute,
			MaxAge:          24 * time.Hour,
		}
	}

	lm := &LifecycleManager{
		trackedIntents: make(map[string]*common.IntentTracker),
		callbacks:      make(map[common.IntentStatus][]common.LifecycleCallback),
		stopCh:         make(chan struct{}),
	}

	// Start cleanup ticker
	lm.cleanupTicker = time.NewTicker(config.CleanupInterval)
	go lm.cleanupLoop()

	return lm
}

// NewConfig creates a default intent configuration
func NewConfig() *Config {
	return &Config{
		MaxConcurrentIntents: common.DefaultMaxConcurrentIntents,
		ProcessingTimeout:    common.DefaultProcessingTimeout,
		RetryAttempts:        common.DefaultRetryAttempts,
		EnableMatching:       true,
		DefaultTTL:           common.DefaultTTL,
	}
}

// NewLifecycleConfig creates a default lifecycle configuration
func NewLifecycleConfig() *LifecycleConfig {
	return &LifecycleConfig{
		CleanupInterval: 5 * time.Minute,
		MaxAge:          24 * time.Hour,
	}
}
