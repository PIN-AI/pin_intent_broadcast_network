package transport

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pin_intent_broadcast_network/internal/p2p"
	"go.uber.org/zap"
)

// LazyTransportManager wraps a TransportManager and initializes it lazily
// when the network manager is ready
type LazyTransportManager struct {
	networkManager p2p.NetworkManager
	logger         *zap.Logger
	
	mu            sync.RWMutex
	transport     TransportManager
	initialized   bool
	lastInitTry   time.Time
	retryInterval time.Duration
}

// NewLazyTransportManager creates a new lazy transport manager
func NewLazyTransportManager(networkManager p2p.NetworkManager, logger *zap.Logger) TransportManager {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	return &LazyTransportManager{
		networkManager: networkManager,
		logger:         logger.Named("lazy_transport"),
		retryInterval:  time.Second * 2,
	}
}

// SetActualTransportManager directly sets the actual transport manager
// This is used when the transport manager is created externally
func (ltm *LazyTransportManager) SetActualTransportManager(transport TransportManager) {
	ltm.mu.Lock()
	defer ltm.mu.Unlock()
	
	ltm.transport = transport
	ltm.initialized = true
	ltm.logger.Info("Actual transport manager set directly")
}

// tryInitialize attempts to initialize the underlying transport manager
func (ltm *LazyTransportManager) tryInitialize() error {
	ltm.mu.Lock()
	defer ltm.mu.Unlock()
	
	if ltm.initialized && ltm.transport != nil {
		return nil
	}
	
	// Avoid frequent retry attempts
	if time.Since(ltm.lastInitTry) < ltm.retryInterval {
		return fmt.Errorf("transport manager not ready, waiting for retry")
	}
	
	ltm.lastInitTry = time.Now()
	
	// Check if network manager is ready
	hostManager := ltm.networkManager.GetHostManager()
	if hostManager == nil {
		return fmt.Errorf("host manager not available yet")
	}
	
	host := hostManager.GetHost()
	if host == nil {
		return fmt.Errorf("libp2p host not available yet")
	}
	
	// Initialize transport manager
	transport := NewTransportManager(host, ltm.logger)
	if transport == nil {
		return fmt.Errorf("failed to create transport manager")
	}
	
	ltm.transport = transport
	ltm.initialized = true
	ltm.logger.Info("Transport manager initialized successfully")
	
	return nil
}

// ensureReady ensures the transport manager is ready, or returns an error
func (ltm *LazyTransportManager) ensureReady() (TransportManager, error) {
	// Fast path - already initialized
	ltm.mu.RLock()
	if ltm.initialized && ltm.transport != nil {
		transport := ltm.transport
		ltm.mu.RUnlock()
		return transport, nil
	}
	ltm.mu.RUnlock()
	
	// Slow path - need to initialize
	if err := ltm.tryInitialize(); err != nil {
		return nil, fmt.Errorf("transport manager not ready: %w", err)
	}
	
	ltm.mu.RLock()
	transport := ltm.transport
	ltm.mu.RUnlock()
	
	return transport, nil
}

// Implementation of TransportManager interface

func (ltm *LazyTransportManager) Start(ctx context.Context, config *TransportConfig) error {
	transport, err := ltm.ensureReady()
	if err != nil {
		return err
	}
	return transport.Start(ctx, config)
}

func (ltm *LazyTransportManager) Stop() error {
	ltm.mu.RLock()
	transport := ltm.transport
	ltm.mu.RUnlock()
	
	if transport != nil {
		return transport.Stop()
	}
	return nil
}

func (ltm *LazyTransportManager) IsRunning() bool {
	ltm.mu.RLock()
	transport := ltm.transport
	ltm.mu.RUnlock()
	
	if transport != nil {
		return transport.IsRunning()
	}
	return false
}

func (ltm *LazyTransportManager) GetPubSubManager() PubSubManager {
	transport, err := ltm.ensureReady()
	if err != nil {
		ltm.logger.Error("Failed to get pubsub manager", zap.Error(err))
		return nil
	}
	return transport.GetPubSubManager()
}

func (ltm *LazyTransportManager) GetTopicManager() TopicManager {
	transport, err := ltm.ensureReady()
	if err != nil {
		ltm.logger.Error("Failed to get topic manager", zap.Error(err))
		return nil
	}
	return transport.GetTopicManager()
}

func (ltm *LazyTransportManager) GetMessageSerializer() MessageSerializer {
	transport, err := ltm.ensureReady()
	if err != nil {
		ltm.logger.Error("Failed to get message serializer", zap.Error(err))
		return nil
	}
	return transport.GetMessageSerializer()
}

func (ltm *LazyTransportManager) GetMessageRouter() MessageRouter {
	transport, err := ltm.ensureReady()
	if err != nil {
		ltm.logger.Error("Failed to get message router", zap.Error(err))
		return nil
	}
	return transport.GetMessageRouter()
}

func (ltm *LazyTransportManager) PublishMessage(ctx context.Context, topic string, msg *TransportMessage) error {
	transport, err := ltm.ensureReady()
	if err != nil {
		return err
	}
	return transport.PublishMessage(ctx, topic, msg)
}

func (ltm *LazyTransportManager) SubscribeToTopic(topic string, handler func(*TransportMessage) error) (Subscription, error) {
	transport, err := ltm.ensureReady()
	if err != nil {
		return nil, err
	}
	return transport.SubscribeToTopic(topic, handler)
}

func (ltm *LazyTransportManager) PublishBidMessage(ctx context.Context, bid *BidMessage) error {
	transport, err := ltm.ensureReady()
	if err != nil {
		return err
	}
	return transport.PublishBidMessage(ctx, bid)
}

func (ltm *LazyTransportManager) PublishMatchResult(ctx context.Context, result *MatchResult) error {
	transport, err := ltm.ensureReady()
	if err != nil {
		return err
	}
	return transport.PublishMatchResult(ctx, result)
}

func (ltm *LazyTransportManager) SubscribeToBids(handler func(*BidMessage) error) (Subscription, error) {
	transport, err := ltm.ensureReady()
	if err != nil {
		return nil, err
	}
	return transport.SubscribeToBids(handler)
}

func (ltm *LazyTransportManager) SubscribeToMatches(handler func(*MatchResult) error) (Subscription, error) {
	transport, err := ltm.ensureReady()
	if err != nil {
		return nil, err
	}
	return transport.SubscribeToMatches(handler)
}

func (ltm *LazyTransportManager) GetTransportMetrics() *TransportMetrics {
	transport, err := ltm.ensureReady()
	if err != nil {
		ltm.logger.Error("Failed to get transport metrics", zap.Error(err))
		return &TransportMetrics{} // Return empty metrics instead of nil
	}
	return transport.GetTransportMetrics()
}