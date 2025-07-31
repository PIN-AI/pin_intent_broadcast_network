package monitoring

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"pin_intent_broadcast_network/internal/conf"
	"pin_intent_broadcast_network/internal/transport"

	"go.uber.org/zap"
)

// TopicSubscriptionManager manages topic subscriptions for intent monitoring
type TopicSubscriptionManager interface {
	// Start starts the subscription manager
	Start(ctx context.Context) error
	// Stop stops the subscription manager
	Stop() error
	// IsRunning returns whether the manager is running
	IsRunning() bool

	// SubscribeByConfig subscribes to topics based on configuration
	SubscribeByConfig(config *conf.IntentMonitoring) error
	// SubscribeWildcard subscribes to topics matching wildcard patterns
	SubscribeWildcard(pattern string) error
	// SubscribeExplicit subscribes to explicit topic list
	SubscribeExplicit(topics []string) error
	// SubscribeAll subscribes to all known intent topics
	SubscribeAll() error

	// DiscoverAndSubscribe discovers and subscribes to new topics
	DiscoverAndSubscribe() error
	// Unsubscribe unsubscribes from a topic
	Unsubscribe(topic string) error
	// UnsubscribeAll unsubscribes from all topics
	UnsubscribeAll() error

	// GetSubscriptions returns current subscription list
	GetSubscriptions() []string
	// GetTopicStats returns statistics for a topic
	GetTopicStats(topic string) *TopicStatistics
	// GetAllTopicStats returns statistics for all subscribed topics
	GetAllTopicStats() map[string]*TopicStatistics

	// UpdateConfig updates the subscription configuration
	UpdateConfig(config *conf.IntentMonitoring) error
}

// TopicStatistics holds statistics for a topic
type TopicStatistics struct {
	Topic         string    `json:"topic"`
	SubscribedAt  time.Time `json:"subscribed_at"`
	MessageCount  int64     `json:"message_count"`
	LastMessage   time.Time `json:"last_message"`
	PeerCount     int       `json:"peer_count"`
	IsActive      bool      `json:"is_active"`
	ErrorCount    int64     `json:"error_count"`
	LastError     string    `json:"last_error,omitempty"`
	LastErrorTime time.Time `json:"last_error_time,omitempty"`
}

// topicSubscriptionManager implements TopicSubscriptionManager
type topicSubscriptionManager struct {
	transportMgr   transport.TransportManager
	configMgr      *ConfigManager
	messageHandler transport.MessageHandler
	logger         *zap.Logger
	ctx            context.Context
	cancel         context.CancelFunc
	isRunning      bool

	// Subscription tracking
	subscriptions map[string]transport.Subscription
	topicStats    map[string]*TopicStatistics
	mu            sync.RWMutex

	// Configuration
	currentConfig   *conf.IntentMonitoring
	discoveryTicker *time.Ticker
}

// NewTopicSubscriptionManager creates a new topic subscription manager
func NewTopicSubscriptionManager(
	transportMgr transport.TransportManager,
	configMgr *ConfigManager,
	messageHandler transport.MessageHandler,
	logger *zap.Logger,
) TopicSubscriptionManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &topicSubscriptionManager{
		transportMgr:   transportMgr,
		configMgr:      configMgr,
		messageHandler: messageHandler,
		logger:         logger.Named("topic_subscription_manager"),
		subscriptions:  make(map[string]transport.Subscription),
		topicStats:     make(map[string]*TopicStatistics),
	}
}

// Start starts the subscription manager
func (tsm *topicSubscriptionManager) Start(ctx context.Context) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if tsm.isRunning {
		return fmt.Errorf("topic subscription manager already running")
	}

	tsm.ctx, tsm.cancel = context.WithCancel(ctx)
	tsm.isRunning = true

	// Load initial configuration
	config := tsm.configMgr.GetConfig()
	tsm.currentConfig = config

	// Subscribe based on initial configuration
	if err := tsm.subscribeByConfigLocked(config); err != nil {
		tsm.logger.Error("Failed to subscribe with initial config", zap.Error(err))
		// Don't fail startup, continue with empty subscriptions
	}

	// Start discovery ticker if needed
	if config.SubscriptionMode == "wildcard" || config.SubscriptionMode == "all" {
		tsm.discoveryTicker = time.NewTicker(30 * time.Second) // Discovery every 30 seconds
		go tsm.discoveryLoop()
	}

	tsm.logger.Info("Topic subscription manager started",
		zap.String("subscription_mode", config.SubscriptionMode),
		zap.Int("initial_subscriptions", len(tsm.subscriptions)),
	)

	return nil
}

// Stop stops the subscription manager
func (tsm *topicSubscriptionManager) Stop() error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if !tsm.isRunning {
		return fmt.Errorf("topic subscription manager not running")
	}

	// Stop discovery ticker
	if tsm.discoveryTicker != nil {
		tsm.discoveryTicker.Stop()
		tsm.discoveryTicker = nil
	}

	// Cancel all subscriptions
	for topic, subscription := range tsm.subscriptions {
		if err := subscription.Cancel(); err != nil {
			tsm.logger.Warn("Failed to cancel subscription",
				zap.String("topic", topic),
				zap.Error(err),
			)
		}
	}

	// Clear subscriptions and stats
	tsm.subscriptions = make(map[string]transport.Subscription)
	tsm.topicStats = make(map[string]*TopicStatistics)

	// Cancel context
	if tsm.cancel != nil {
		tsm.cancel()
	}

	tsm.isRunning = false
	tsm.logger.Info("Topic subscription manager stopped")

	return nil
}

// IsRunning returns whether the manager is running
func (tsm *topicSubscriptionManager) IsRunning() bool {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()
	return tsm.isRunning
}

// SubscribeByConfig subscribes to topics based on configuration
func (tsm *topicSubscriptionManager) SubscribeByConfig(config *conf.IntentMonitoring) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if !tsm.isRunning {
		return fmt.Errorf("subscription manager not running")
	}

	return tsm.subscribeByConfigLocked(config)
}

// subscribeByConfigLocked implements subscription logic (must be called with lock held)
func (tsm *topicSubscriptionManager) subscribeByConfigLocked(config *conf.IntentMonitoring) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	switch config.SubscriptionMode {
	case "disabled":
		tsm.logger.Info("Subscription mode disabled, no topics will be subscribed")
		return nil

	case "explicit":
		topics := config.ExplicitTopics
		if len(topics) == 0 {
			// Use default explicit topics
			defaultTopics := tsm.configMgr.getDefaultExplicitTopics()
			topics = defaultTopics
		}
		return tsm.subscribeExplicitLocked(topics)

	case "wildcard":
		patterns := config.WildcardPatterns
		if len(patterns) == 0 {
			patterns = []string{"intent-broadcast.*"}
		}
		return tsm.subscribeWildcardPatternsLocked(patterns)

	case "all":
		return tsm.subscribeAllLocked()

	default:
		return fmt.Errorf("unknown subscription mode: %s", config.SubscriptionMode)
	}
}

// SubscribeWildcard subscribes to topics matching wildcard patterns
func (tsm *topicSubscriptionManager) SubscribeWildcard(pattern string) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if !tsm.isRunning {
		return fmt.Errorf("subscription manager not running")
	}

	return tsm.subscribeWildcardPatternLocked(pattern)
}

// subscribeWildcardPatternsLocked subscribes to multiple wildcard patterns
func (tsm *topicSubscriptionManager) subscribeWildcardPatternsLocked(patterns []string) error {
	for _, pattern := range patterns {
		if err := tsm.subscribeWildcardPatternLocked(pattern); err != nil {
			tsm.logger.Error("Failed to subscribe to wildcard pattern",
				zap.String("pattern", pattern),
				zap.Error(err),
			)
			// Continue with other patterns
		}
	}
	return nil
}

// subscribeWildcardPatternLocked subscribes to topics matching a wildcard pattern
func (tsm *topicSubscriptionManager) subscribeWildcardPatternLocked(pattern string) error {
	// For wildcard patterns, we need to discover existing topics that match
	// and also set up monitoring for new topics

	// Get all known topics that match the pattern
	matchingTopics := tsm.getTopicsMatchingPattern(pattern)

	tsm.logger.Info("Found topics matching pattern",
		zap.String("pattern", pattern),
		zap.Int("count", len(matchingTopics)),
		zap.Strings("topics", matchingTopics),
	)

	// Subscribe to matching topics
	for _, topic := range matchingTopics {
		if err := tsm.subscribeToTopicLocked(topic); err != nil {
			tsm.logger.Error("Failed to subscribe to matching topic",
				zap.String("topic", topic),
				zap.String("pattern", pattern),
				zap.Error(err),
			)
		}
	}

	return nil
}

// getTopicsMatchingPattern returns topics that match the given pattern
func (tsm *topicSubscriptionManager) getTopicsMatchingPattern(pattern string) []string {
	// For now, we'll use the known intent broadcast topics
	// In a real implementation, this could query the P2P network for active topics
	allKnownTopics := tsm.configMgr.getAllKnownTopics()

	var matchingTopics []string
	for _, topic := range allKnownTopics {
		if matched, _ := filepath.Match(pattern, topic); matched {
			matchingTopics = append(matchingTopics, topic)
		}
	}

	return matchingTopics
}

// SubscribeExplicit subscribes to explicit topic list
func (tsm *topicSubscriptionManager) SubscribeExplicit(topics []string) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if !tsm.isRunning {
		return fmt.Errorf("subscription manager not running")
	}

	return tsm.subscribeExplicitLocked(topics)
}

// subscribeExplicitLocked subscribes to explicit topics
func (tsm *topicSubscriptionManager) subscribeExplicitLocked(topics []string) error {
	for _, topic := range topics {
		if err := tsm.subscribeToTopicLocked(topic); err != nil {
			tsm.logger.Error("Failed to subscribe to explicit topic",
				zap.String("topic", topic),
				zap.Error(err),
			)
			// Continue with other topics
		}
	}
	return nil
}

// SubscribeAll subscribes to all known intent topics
func (tsm *topicSubscriptionManager) SubscribeAll() error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if !tsm.isRunning {
		return fmt.Errorf("subscription manager not running")
	}

	return tsm.subscribeAllLocked()
}

// subscribeAllLocked subscribes to all known topics
func (tsm *topicSubscriptionManager) subscribeAllLocked() error {
	allTopics := tsm.configMgr.getAllKnownTopics()

	tsm.logger.Info("Subscribing to all known topics",
		zap.Int("count", len(allTopics)),
		zap.Strings("topics", allTopics),
	)

	for _, topic := range allTopics {
		if err := tsm.subscribeToTopicLocked(topic); err != nil {
			tsm.logger.Error("Failed to subscribe to topic in all mode",
				zap.String("topic", topic),
				zap.Error(err),
			)
			// Continue with other topics
		}
	}

	return nil
}

// subscribeToTopicLocked subscribes to a single topic (must be called with lock held)
func (tsm *topicSubscriptionManager) subscribeToTopicLocked(topic string) error {
	// Check if already subscribed
	if _, exists := tsm.subscriptions[topic]; exists {
		tsm.logger.Debug("Already subscribed to topic", zap.String("topic", topic))
		return nil
	}

	// Create a wrapped message handler that updates statistics
	wrappedHandler := tsm.createWrappedHandler(topic)

	// Subscribe through transport manager
	subscription, err := tsm.transportMgr.SubscribeToTopic(topic, wrappedHandler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	// Store subscription
	tsm.subscriptions[topic] = subscription

	// Initialize topic statistics
	tsm.topicStats[topic] = &TopicStatistics{
		Topic:        topic,
		SubscribedAt: time.Now(),
		IsActive:     true,
		PeerCount:    tsm.getPeerCountForTopic(topic),
	}

	tsm.logger.Info("Successfully subscribed to topic", zap.String("topic", topic))
	return nil
}

// createWrappedHandler creates a message handler that updates statistics
func (tsm *topicSubscriptionManager) createWrappedHandler(topic string) func(*transport.TransportMessage) error {
	return func(msg *transport.TransportMessage) error {
		// Update statistics
		tsm.updateTopicStats(topic, msg, nil)

		// Call the original handler
		if tsm.messageHandler != nil {
			if err := tsm.messageHandler(msg); err != nil {
				// Update error statistics
				tsm.updateTopicStats(topic, msg, err)
				return err
			}
		}

		return nil
	}
}

// updateTopicStats updates statistics for a topic
func (tsm *topicSubscriptionManager) updateTopicStats(topic string, msg *transport.TransportMessage, err error) {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	stats, exists := tsm.topicStats[topic]
	if !exists {
		return
	}

	stats.MessageCount++
	stats.LastMessage = time.Now()
	stats.PeerCount = tsm.getPeerCountForTopic(topic)

	if err != nil {
		stats.ErrorCount++
		stats.LastError = err.Error()
		stats.LastErrorTime = time.Now()
	}
}

// getPeerCountForTopic gets the peer count for a topic
func (tsm *topicSubscriptionManager) getPeerCountForTopic(topic string) int {
	if tsm.transportMgr == nil {
		return 0
	}

	pubsubMgr := tsm.transportMgr.GetPubSubManager()
	if pubsubMgr == nil {
		return 0
	}

	return pubsubMgr.GetPeerCount(topic)
}

// Unsubscribe unsubscribes from a topic
func (tsm *topicSubscriptionManager) Unsubscribe(topic string) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if !tsm.isRunning {
		return fmt.Errorf("subscription manager not running")
	}

	subscription, exists := tsm.subscriptions[topic]
	if !exists {
		return fmt.Errorf("not subscribed to topic: %s", topic)
	}

	// Cancel subscription
	if err := subscription.Cancel(); err != nil {
		return fmt.Errorf("failed to cancel subscription for topic %s: %w", topic, err)
	}

	// Remove from tracking
	delete(tsm.subscriptions, topic)

	// Mark stats as inactive
	if stats, exists := tsm.topicStats[topic]; exists {
		stats.IsActive = false
	}

	tsm.logger.Info("Successfully unsubscribed from topic", zap.String("topic", topic))
	return nil
}

// UnsubscribeAll unsubscribes from all topics
func (tsm *topicSubscriptionManager) UnsubscribeAll() error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if !tsm.isRunning {
		return fmt.Errorf("subscription manager not running")
	}

	var errors []string

	for topic, subscription := range tsm.subscriptions {
		if err := subscription.Cancel(); err != nil {
			errors = append(errors, fmt.Sprintf("topic %s: %v", topic, err))
		}

		// Mark stats as inactive
		if stats, exists := tsm.topicStats[topic]; exists {
			stats.IsActive = false
		}
	}

	// Clear all subscriptions
	tsm.subscriptions = make(map[string]transport.Subscription)

	if len(errors) > 0 {
		return fmt.Errorf("failed to unsubscribe from some topics: %s", strings.Join(errors, "; "))
	}

	tsm.logger.Info("Successfully unsubscribed from all topics")
	return nil
}

// GetSubscriptions returns current subscription list
func (tsm *topicSubscriptionManager) GetSubscriptions() []string {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()

	topics := make([]string, 0, len(tsm.subscriptions))
	for topic := range tsm.subscriptions {
		topics = append(topics, topic)
	}

	return topics
}

// GetTopicStats returns statistics for a topic
func (tsm *topicSubscriptionManager) GetTopicStats(topic string) *TopicStatistics {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()

	stats, exists := tsm.topicStats[topic]
	if !exists {
		return nil
	}

	// Return a copy to prevent modification
	statsCopy := *stats
	return &statsCopy
}

// GetAllTopicStats returns statistics for all subscribed topics
func (tsm *topicSubscriptionManager) GetAllTopicStats() map[string]*TopicStatistics {
	tsm.mu.RLock()
	defer tsm.mu.RUnlock()

	result := make(map[string]*TopicStatistics)
	for topic, stats := range tsm.topicStats {
		// Return copies to prevent modification
		statsCopy := *stats
		result[topic] = &statsCopy
	}

	return result
}

// UpdateConfig updates the subscription configuration
func (tsm *topicSubscriptionManager) UpdateConfig(config *conf.IntentMonitoring) error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if !tsm.isRunning {
		return fmt.Errorf("subscription manager not running")
	}

	// Check if configuration actually changed
	if tsm.configsEqual(tsm.currentConfig, config) {
		tsm.logger.Debug("Configuration unchanged, skipping update")
		return nil
	}

	tsm.logger.Info("Updating subscription configuration",
		zap.String("old_mode", tsm.currentConfig.SubscriptionMode),
		zap.String("new_mode", config.SubscriptionMode),
	)

	// Unsubscribe from all current topics
	if err := tsm.unsubscribeAllLocked(); err != nil {
		tsm.logger.Error("Failed to unsubscribe from existing topics", zap.Error(err))
		// Continue anyway
	}

	// Subscribe with new configuration
	if err := tsm.subscribeByConfigLocked(config); err != nil {
		return fmt.Errorf("failed to subscribe with new configuration: %w", err)
	}

	// Update current configuration
	tsm.currentConfig = config

	tsm.logger.Info("Successfully updated subscription configuration",
		zap.String("subscription_mode", config.SubscriptionMode),
		zap.Int("active_subscriptions", len(tsm.subscriptions)),
	)

	return nil
}

// unsubscribeAllLocked unsubscribes from all topics (must be called with lock held)
func (tsm *topicSubscriptionManager) unsubscribeAllLocked() error {
	var errors []string

	for topic, subscription := range tsm.subscriptions {
		if err := subscription.Cancel(); err != nil {
			errors = append(errors, fmt.Sprintf("topic %s: %v", topic, err))
		}

		// Mark stats as inactive
		if stats, exists := tsm.topicStats[topic]; exists {
			stats.IsActive = false
		}
	}

	// Clear all subscriptions
	tsm.subscriptions = make(map[string]transport.Subscription)

	if len(errors) > 0 {
		return fmt.Errorf("failed to unsubscribe from some topics: %s", strings.Join(errors, "; "))
	}

	return nil
}

// configsEqual checks if two configurations are equal
func (tsm *topicSubscriptionManager) configsEqual(a, b *conf.IntentMonitoring) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare subscription mode
	if a.SubscriptionMode != b.SubscriptionMode {
		return false
	}

	// Compare explicit topics
	if !tsm.stringSlicesEqual(a.ExplicitTopics, b.ExplicitTopics) {
		return false
	}

	// Compare wildcard patterns
	if !tsm.stringSlicesEqual(a.WildcardPatterns, b.WildcardPatterns) {
		return false
	}

	return true
}

// stringSlicesEqual checks if two string slices are equal
func (tsm *topicSubscriptionManager) stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// DiscoverAndSubscribe discovers and subscribes to new topics
func (tsm *topicSubscriptionManager) DiscoverAndSubscribe() error {
	tsm.mu.Lock()
	defer tsm.mu.Unlock()

	if !tsm.isRunning {
		return fmt.Errorf("subscription manager not running")
	}

	// This is a placeholder for topic discovery logic
	// In a real implementation, this could:
	// 1. Query the P2P network for active topics
	// 2. Check with peers for new intent broadcast topics
	// 3. Monitor network traffic for new topic patterns

	tsm.logger.Debug("Topic discovery not yet implemented")
	return nil
}

// discoveryLoop runs the topic discovery loop
func (tsm *topicSubscriptionManager) discoveryLoop() {
	defer func() {
		if tsm.discoveryTicker != nil {
			tsm.discoveryTicker.Stop()
		}
	}()

	if tsm.discoveryTicker == nil {
		return
	}

	for {
		select {
		case <-tsm.ctx.Done():
			return
		case <-tsm.discoveryTicker.C:
			if err := tsm.DiscoverAndSubscribe(); err != nil {
				tsm.logger.Error("Topic discovery failed", zap.Error(err))
			}
		}
	}
}
