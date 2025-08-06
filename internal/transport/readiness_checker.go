package transport

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"pin_intent_broadcast_network/internal/p2p"
	"go.uber.org/zap"
)

// TransportHealthStatus represents the current health status of transport
type TransportHealthStatus struct {
	IsReady          bool      `json:"is_ready"`
	P2PConnected     bool      `json:"p2p_connected"`
	GossipSubStarted bool      `json:"gossipsub_started"`
	TopicsRegistered bool      `json:"topics_registered"`
	ConnectedPeers   int       `json:"connected_peers"`
	LastChecked      time.Time `json:"last_checked"`
	ErrorMessage     string    `json:"error_message,omitempty"`
}

// TransportReadinessConfig holds configuration for transport readiness checker
type TransportReadinessConfig struct {
	MaxRetries     int           `yaml:"max_retries"`        
	RetryInterval  time.Duration `yaml:"retry_interval"`     
	HealthTimeout  time.Duration `yaml:"health_timeout"`     
	MaxWaitTime    time.Duration `yaml:"max_wait_time"`      
}

// DefaultTransportReadinessConfig returns default configuration
func DefaultTransportReadinessConfig() *TransportReadinessConfig {
	return &TransportReadinessConfig{
		MaxRetries:    5,
		RetryInterval: 2 * time.Second,
		HealthTimeout: 5 * time.Second,
		MaxWaitTime:   60 * time.Second,
	}
}

// TransportReadinessChecker checks and waits for transport manager readiness
type TransportReadinessChecker struct {
	networkManager p2p.NetworkManager
	transportMgr   TransportManager
	logger         *zap.Logger
	config         *TransportReadinessConfig
	
	// State tracking
	mu             sync.RWMutex
	lastHealth     *TransportHealthStatus
	readyCallbacks []func(TransportManager)
	isChecking     atomic.Bool
}

// NewTransportReadinessChecker creates a new transport readiness checker
func NewTransportReadinessChecker(
	networkManager p2p.NetworkManager,
	transportMgr TransportManager,
	logger *zap.Logger,
) *TransportReadinessChecker {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &TransportReadinessChecker{
		networkManager: networkManager,
		transportMgr:   transportMgr,
		logger:         logger.Named("transport_readiness"),
		config:         DefaultTransportReadinessConfig(),
		readyCallbacks: make([]func(TransportManager), 0),
	}
}

// SetConfig updates the checker configuration
func (trc *TransportReadinessChecker) SetConfig(config *TransportReadinessConfig) {
	trc.mu.Lock()
	defer trc.mu.Unlock()
	trc.config = config
}

// WaitForTransportReady waits for transport manager to be ready with retries
func (trc *TransportReadinessChecker) WaitForTransportReady(ctx context.Context) error {
	if !trc.isChecking.CompareAndSwap(false, true) {
		return fmt.Errorf("readiness check already in progress")
	}
	defer trc.isChecking.Store(false)

	trc.logger.Info("Starting transport readiness check",
		zap.Int("max_retries", trc.config.MaxRetries),
		zap.Duration("retry_interval", trc.config.RetryInterval),
		zap.Duration("max_wait_time", trc.config.MaxWaitTime),
	)

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, trc.config.MaxWaitTime)
	defer cancel()

	retryCount := 0
	lastErr := fmt.Errorf("transport readiness check not started")

	for retryCount < trc.config.MaxRetries {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("transport readiness check timeout after %v: %w", trc.config.MaxWaitTime, timeoutCtx.Err())
		default:
		}

		// Perform health check
		health := trc.performHealthCheck()
		trc.updateHealthStatus(health)

		if health.IsReady {
			trc.logger.Info("Transport is ready",
				zap.Int("retry_count", retryCount),
				zap.Int("connected_peers", health.ConnectedPeers),
			)

			// Trigger ready callbacks
			trc.triggerReadyCallbacks()
			return nil
		}

		retryCount++
		lastErr = fmt.Errorf("transport not ready: %s", health.ErrorMessage)

		trc.logger.Warn("Transport not ready, retrying",
			zap.Int("retry", retryCount),
			zap.Int("max_retries", trc.config.MaxRetries),
			zap.String("error", health.ErrorMessage),
			zap.Duration("next_retry_in", trc.config.RetryInterval),
		)

		if retryCount < trc.config.MaxRetries {
			select {
			case <-timeoutCtx.Done():
				return fmt.Errorf("transport readiness check timeout during retry: %w", timeoutCtx.Err())
			case <-time.After(trc.config.RetryInterval):
				continue
			}
		}
	}

	return fmt.Errorf("transport not ready after %d retries: %w", trc.config.MaxRetries, lastErr)
}

// CheckTransportHealth performs a single health check and returns status
func (trc *TransportReadinessChecker) CheckTransportHealth() *TransportHealthStatus {
	health := trc.performHealthCheck()
	trc.updateHealthStatus(health)
	return health
}

// GetLastHealthStatus returns the last known health status
func (trc *TransportReadinessChecker) GetLastHealthStatus() *TransportHealthStatus {
	trc.mu.RLock()
	defer trc.mu.RUnlock()
	
	if trc.lastHealth == nil {
		return &TransportHealthStatus{
			IsReady:     false,
			LastChecked: time.Now(),
			ErrorMessage: "Health check not performed yet",
		}
	}

	// Return a copy
	health := *trc.lastHealth
	return &health
}

// RegisterReadinessCallback registers a callback to be called when transport becomes ready
func (trc *TransportReadinessChecker) RegisterReadinessCallback(callback func(TransportManager)) {
	trc.mu.Lock()
	defer trc.mu.Unlock()
	trc.readyCallbacks = append(trc.readyCallbacks, callback)
}

// IsChecking returns whether a readiness check is currently in progress
func (trc *TransportReadinessChecker) IsChecking() bool {
	return trc.isChecking.Load()
}

// performHealthCheck performs the actual health check
func (trc *TransportReadinessChecker) performHealthCheck() *TransportHealthStatus {
	health := &TransportHealthStatus{
		IsReady:     false,
		LastChecked: time.Now(),
	}

	// Check if transport manager exists and is running
	if trc.transportMgr == nil {
		health.ErrorMessage = "transport manager is nil"
		return health
	}

	if !trc.transportMgr.IsRunning() {
		health.ErrorMessage = "transport manager not running"
		return health
	}

	// Check P2P network connectivity
	if trc.networkManager == nil {
		health.ErrorMessage = "network manager is nil"
		return health
	}

	hostManager := trc.networkManager.GetHostManager()
	if hostManager == nil {
		health.ErrorMessage = "host manager not available"
		return health
	}

	host := hostManager.GetHost()
	if host == nil {
		health.ErrorMessage = "libp2p host not available"
		return health
	}

	health.P2PConnected = true

	// Check GossipSub availability
	pubsubMgr := trc.transportMgr.GetPubSubManager()
	if pubsubMgr == nil {
		health.ErrorMessage = "pubsub manager not available"
		return health
	}

	health.GossipSubStarted = true

	// Check topic registration
	topicMgr := trc.transportMgr.GetTopicManager()
	if topicMgr == nil {
		health.ErrorMessage = "topic manager not available"
		return health
	}

	health.TopicsRegistered = true

	// Get transport metrics for peer count
	metrics := trc.transportMgr.GetTransportMetrics()
	if metrics != nil {
		health.ConnectedPeers = int(metrics.ConnectedPeerCount)
	}

	// All checks passed
	health.IsReady = true
	return health
}

// updateHealthStatus updates the last known health status
func (trc *TransportReadinessChecker) updateHealthStatus(health *TransportHealthStatus) {
	trc.mu.Lock()
	defer trc.mu.Unlock()
	trc.lastHealth = health
}

// triggerReadyCallbacks calls all registered callbacks when transport becomes ready
func (trc *TransportReadinessChecker) triggerReadyCallbacks() {
	trc.mu.RLock()
	callbacks := make([]func(TransportManager), len(trc.readyCallbacks))
	copy(callbacks, trc.readyCallbacks)
	trc.mu.RUnlock()

	for _, callback := range callbacks {
		go func(cb func(TransportManager)) {
			defer func() {
				if r := recover(); r != nil {
					trc.logger.Error("Transport ready callback panicked",
						zap.Any("panic", r),
					)
				}
			}()
			cb(trc.transportMgr)
		}(callback)
	}
}