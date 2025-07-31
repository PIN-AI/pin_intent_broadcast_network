package intent

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"pin_intent_broadcast_network/internal/biz/common"
	"pin_intent_broadcast_network/internal/transport"
)

// ProviderSet is intent providers.
var ProviderSet = wire.NewSet(
	NewManagerWithDefaults,
	NewConfig,
	wire.Bind(new(common.IntentManager), new(*Manager)),
)

// NewManagerWithDefaults creates a new Intent Manager instance with minimal dependencies
func NewManagerWithDefaults(
	transportMgr transport.TransportManager,
	config *Config,
	logger log.Logger,
) *Manager {
	// Create a minimal manager for now
	// TODO: Add proper dependency injection for all components
	return &Manager{
		validator:     nil, // Will be set later
		signer:        nil, // Will be set later
		processor:     nil, // Will be set later
		matcher:       nil, // Will be set later
		lifecycle:     nil, // Will be set later
		transportMgr:  transportMgr,
		config:        config,
		metrics:       &Metrics{},
		logger:        log.NewHelper(logger),
		intents:       make(map[string]*common.Intent),
		subscriptions: make(map[string]chan *common.Intent),
	}
}

// NewLifecycleManagerWithDefaults creates a new lifecycle manager with default config
func NewLifecycleManagerWithDefaults() *LifecycleManager {
	config := &LifecycleConfig{
		CleanupInterval: 5 * time.Minute,
		MaxAge:          24 * time.Hour,
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
