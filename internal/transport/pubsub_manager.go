package transport

import (
	"context"
	"fmt"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

// pubSubManager GossipSub pubsub manager implementation
type pubSubManager struct {
	host      host.Host
	pubsub    *pubsub.PubSub
	ctx       context.Context
	cancel    context.CancelFunc
	config    *PubSubConfig
	logger    *zap.Logger
	isRunning bool

	// Subscription management - support multiple handlers per topic
	subscriptions map[string]*pubSubSubscription  // topic -> subscription
	handlers map[string][]MessageHandler          // topic -> list of handlers
	// Topic management
	topics map[string]*pubsub.Topic
	mu     sync.RWMutex
}

// pubSubSubscription represents a topic subscription
type pubSubSubscription struct {
	topic        string
	subscription *pubsub.Subscription
	handler      MessageHandler
	isActive     bool
	cancel       context.CancelFunc
}

// NewPubSubManager creates new pubsub manager
func NewPubSubManager(h host.Host, logger *zap.Logger) PubSubManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &pubSubManager{
		host:          h,
		logger:        logger.Named("pubsub_manager"),
		subscriptions: make(map[string]*pubSubSubscription),
		handlers:      make(map[string][]MessageHandler),
		topics:        make(map[string]*pubsub.Topic),
	}
}

// Start starts the pubsub manager
func (pm *pubSubManager) Start(ctx context.Context, config *PubSubConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isRunning {
		return fmt.Errorf("pubsub manager already running")
	}

	pm.config = config
	pm.ctx, pm.cancel = context.WithCancel(ctx)

	// Create GossipSub options
	options := []pubsub.Option{
		pubsub.WithPeerExchange(true),
		pubsub.WithFloodPublish(true),
	}

	// Configure GossipSub parameters
	if config != nil {
		gossipSubConfig := pubsub.DefaultGossipSubParams()
		gossipSubConfig.D = config.D
		gossipSubConfig.Dlo = config.DLo
		gossipSubConfig.Dhi = config.DHi
		gossipSubConfig.HeartbeatInterval = config.HeartbeatInterval
		gossipSubConfig.FanoutTTL = config.FanoutTTL

		options = append(options, pubsub.WithGossipSubParams(gossipSubConfig))

		// Enable message signing if configured
		if config.EnableSigning {
			options = append(options, pubsub.WithMessageSigning(true))
		}

		// Enable strict signature verification if configured
		if config.EnableStrictVerification {
			options = append(options, pubsub.WithStrictSignatureVerification(true))
		}
	}

	// Create GossipSub instance
	ps, err := pubsub.NewGossipSub(pm.ctx, pm.host, options...)
	if err != nil {
		return fmt.Errorf("failed to create GossipSub: %w", err)
	}

	pm.pubsub = ps
	pm.isRunning = true

	pm.logger.Info("PubSub manager started successfully",
		zap.Int("d", config.D),
		zap.Int("d_lo", config.DLo),
		zap.Int("d_hi", config.DHi),
		zap.Duration("heartbeat_interval", config.HeartbeatInterval),
		zap.Bool("enable_signing", config.EnableSigning),
	)

	return nil
}

// Stop stops the pubsub manager
func (pm *pubSubManager) Stop() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.isRunning {
		return fmt.Errorf("pubsub manager not running")
	}

	// Cancel all subscriptions
	for topic, sub := range pm.subscriptions {
		if sub.cancel != nil {
			sub.cancel()
		}
		sub.isActive = false
		pm.logger.Debug("Cancelled subscription", zap.String("topic", topic))
	}

	// Clear subscriptions
	pm.subscriptions = make(map[string]*pubSubSubscription)

	// Close all topics
	for topic, topicHandle := range pm.topics {
		topicHandle.Close()
		pm.logger.Debug("Closed topic", zap.String("topic", topic))
	}
	pm.topics = make(map[string]*pubsub.Topic)

	// Cancel context
	if pm.cancel != nil {
		pm.cancel()
	}

	pm.isRunning = false
	pm.logger.Info("PubSub manager stopped")

	return nil
}

// Publish publishes a message to a topic
func (pm *pubSubManager) Publish(ctx context.Context, topic string, data []byte) error {
	if !pm.isRunning {
		return ErrTransportNotRunning
	}

	if err := ValidateTopicName(topic); err != nil {
		return err
	}

	if len(data) == 0 {
		return &TransportError{"EMPTY_DATA", "Cannot publish empty data", ""}
	}

	// Get or create topic handle
	pm.mu.Lock()
	topicHandle, exists := pm.topics[topic]
	if !exists {
		var err error
		topicHandle, err = pm.pubsub.Join(topic)
		if err != nil {
			pm.mu.Unlock()
			return fmt.Errorf("failed to join topic %s: %w", topic, err)
		}
		pm.topics[topic] = topicHandle
	}
	pm.mu.Unlock()

	// Publish message
	if err := topicHandle.Publish(ctx, data); err != nil {
		pm.logger.Error("Failed to publish message",
			zap.String("topic", FormatTopic(topic)),
			zap.Int("data_size", len(data)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}

	pm.logger.Debug("Message published successfully",
		zap.String("topic", FormatTopic(topic)),
		zap.Int("data_size", len(data)),
	)

	return nil
}

// Subscribe subscribes to a topic
func (pm *pubSubManager) Subscribe(topic string, handler MessageHandler) (Subscription, error) {
	if !pm.isRunning {
		return nil, ErrTransportNotRunning
	}

	if err := ValidateTopicName(topic); err != nil {
		return nil, err
	}

	if handler == nil {
		return nil, &TransportError{"INVALID_HANDLER", "Message handler cannot be nil", ""}
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Add handler to the list for this topic
	pm.handlers[topic] = append(pm.handlers[topic], handler)

	// Check if we already have a subscription for this topic
	if _, exists := pm.subscriptions[topic]; exists {
		// Just add the handler, subscription already exists
		pm.logger.Info("Added handler to existing topic subscription",
			zap.String("topic", FormatTopic(topic)),
			zap.Int("total_handlers", len(pm.handlers[topic])),
		)
		// Return a dummy subscription for consistency
		return &pubSubSubscription{topic: topic, isActive: true}, nil
	}

	// Get or create topic handle
	topicHandle, exists := pm.topics[topic]
	if !exists {
		var err error
		topicHandle, err = pm.pubsub.Join(topic)
		if err != nil {
			// Remove the handler we just added since subscription failed
			pm.handlers[topic] = pm.handlers[topic][:len(pm.handlers[topic])-1]
			return nil, fmt.Errorf("failed to join topic %s: %w", topic, err)
		}
		pm.topics[topic] = topicHandle
	}

	// Subscribe to topic
	subscription, err := topicHandle.Subscribe()
	if err != nil {
		// Remove the handler we just added since subscription failed
		pm.handlers[topic] = pm.handlers[topic][:len(pm.handlers[topic])-1]
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	// Create subscription context
	subCtx, subCancel := context.WithCancel(pm.ctx)

	// Create subscription wrapper - this will handle messages for ALL handlers
	pubsubSub := &pubSubSubscription{
		topic:        topic,
		subscription: subscription,
		handler:      nil, // We'll handle multiple handlers in handleMessages
		isActive:     true,
		cancel:       subCancel,
	}

	// Store subscription
	pm.subscriptions[topic] = pubsubSub

	// Start message handling goroutine with multi-handler support
	go pm.handleMessagesMultiHandler(subCtx, topicHandle, pubsubSub, topic)

	pm.logger.Info("Successfully subscribed to topic with multi-handler support",
		zap.String("topic", FormatTopic(topic)),
		zap.Int("handlers_count", len(pm.handlers[topic])),
	)

	return pubsubSub, nil
}

// Unsubscribe unsubscribes from a topic
func (pm *pubSubManager) Unsubscribe(topic string) error {
	if !pm.isRunning {
		return ErrTransportNotRunning
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	sub, exists := pm.subscriptions[topic]
	if !exists {
		return ErrTopicNotFound
	}

	// Cancel subscription
	if sub.cancel != nil {
		sub.cancel()
	}

	// Mark as inactive
	sub.isActive = false

	// Remove from subscriptions
	delete(pm.subscriptions, topic)

	pm.logger.Info("Successfully unsubscribed from topic",
		zap.String("topic", FormatTopic(topic)),
	)

	return nil
}

// GetConnectedPeers returns connected peers
func (pm *pubSubManager) GetConnectedPeers() []peer.ID {
	if !pm.isRunning || pm.pubsub == nil {
		return []peer.ID{}
	}

	return pm.host.Network().Peers()
}

// GetTopics returns subscribed topics
func (pm *pubSubManager) GetTopics() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	topics := make([]string, 0, len(pm.subscriptions))
	for topic := range pm.subscriptions {
		topics = append(topics, topic)
	}

	return topics
}

// GetPeerCount returns peer count for a topic
func (pm *pubSubManager) GetPeerCount(topic string) int {
	if !pm.isRunning || pm.pubsub == nil {
		return 0
	}

	return len(pm.pubsub.ListPeers(topic))
}

// handleMessagesMultiHandler handles messages with support for multiple handlers per topic
func (pm *pubSubManager) handleMessagesMultiHandler(ctx context.Context, topicHandle *pubsub.Topic, sub *pubSubSubscription, topic string) {
	defer func() {
		// Don't close topicHandle as it's shared, just cancel the subscription
		sub.subscription.Cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			pm.logger.Debug("Multi-handler message handler context cancelled",
				zap.String("topic", FormatTopic(sub.topic)),
			)
			return

		default:
			// Receive message with timeout
			msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			msg, err := sub.subscription.Next(msgCtx)
			cancel()

			if err != nil {
				if ctx.Err() != nil {
					return // Context cancelled
				}
				pm.logger.Debug("Failed to receive message",
					zap.String("topic", FormatTopic(sub.topic)),
					zap.Error(err),
				)
				continue
			}

			// Skip messages from self
			if msg.ReceivedFrom == pm.host.ID() {
				continue
			}

			// Create transport message
			transportMsg := &TransportMessage{
				ID: GenerateMessageID(&TransportMessage{
					Type:      "pubsub",
					Payload:   msg.Data,
					Timestamp: time.Now().UnixMilli(),
					Sender:    msg.ReceivedFrom.String(),
				}),
				Type:      "pubsub",
				Payload:   msg.Data,
				Timestamp: time.Now().UnixMilli(),
				Sender:    msg.ReceivedFrom.String(),
				Priority:  PriorityNormal,
				Metadata:  make(map[string]string),
			}

			// Add topic to metadata
			transportMsg.Metadata["topic"] = sub.topic
			transportMsg.Metadata["sequence"] = fmt.Sprintf("%d", msg.Message.GetSeqno())

			// Handle message with all registered handlers for this topic
			pm.mu.RLock()
			handlers := pm.handlers[topic]
			pm.mu.RUnlock()

			for i, handler := range handlers {
				if err := handler(transportMsg); err != nil {
					pm.logger.Error("Multi-handler message handler failed",
						zap.String("topic", FormatTopic(sub.topic)),
						zap.String("sender", FormatPeerID(msg.ReceivedFrom)),
						zap.Int("handler_index", i),
						zap.Error(err),
					)
				} else {
					pm.logger.Debug("Multi-handler message handled successfully",
						zap.String("topic", FormatTopic(sub.topic)),
						zap.String("sender", FormatPeerID(msg.ReceivedFrom)),
						zap.Int("handler_index", i),
						zap.Int("payload_size", len(msg.Data)),
					)
				}
			}
		}
	}
}

// handleMessages handles incoming messages for a subscription
func (pm *pubSubManager) handleMessages(ctx context.Context, topicHandle *pubsub.Topic, sub *pubSubSubscription) {
	defer func() {
		// Don't close topicHandle as it's shared, just cancel the subscription
		sub.subscription.Cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			pm.logger.Debug("Message handler context cancelled",
				zap.String("topic", FormatTopic(sub.topic)),
			)
			return

		default:
			// Receive message with timeout
			msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			msg, err := sub.subscription.Next(msgCtx)
			cancel()

			if err != nil {
				if ctx.Err() != nil {
					return // Context cancelled
				}
				pm.logger.Debug("Failed to receive message",
					zap.String("topic", FormatTopic(sub.topic)),
					zap.Error(err),
				)
				continue
			}

			// Skip messages from self
			if msg.ReceivedFrom == pm.host.ID() {
				continue
			}

			// Create transport message
			transportMsg := &TransportMessage{
				ID: GenerateMessageID(&TransportMessage{
					Type:      "pubsub",
					Payload:   msg.Data,
					Timestamp: time.Now().UnixMilli(),
					Sender:    msg.ReceivedFrom.String(),
				}),
				Type:      "pubsub",
				Payload:   msg.Data,
				Timestamp: time.Now().UnixMilli(),
				Sender:    msg.ReceivedFrom.String(),
				Priority:  PriorityNormal,
				Metadata:  make(map[string]string),
			}

			// Add topic to metadata
			transportMsg.Metadata["topic"] = sub.topic
			transportMsg.Metadata["sequence"] = fmt.Sprintf("%d", msg.Message.GetSeqno())

			// Handle message
			if err := sub.handler(transportMsg); err != nil {
				pm.logger.Error("Message handler failed",
					zap.String("topic", FormatTopic(sub.topic)),
					zap.String("sender", FormatPeerID(msg.ReceivedFrom)),
					zap.Error(err),
				)
			} else {
				pm.logger.Debug("Message handled successfully",
					zap.String("topic", FormatTopic(sub.topic)),
					zap.String("sender", FormatPeerID(msg.ReceivedFrom)),
					zap.Int("payload_size", len(msg.Data)),
				)
			}
		}
	}
}

// Subscription interface implementation
func (s *pubSubSubscription) Topic() string {
	return s.topic
}

func (s *pubSubSubscription) Cancel() error {
	if s.cancel != nil {
		s.cancel()
	}
	s.isActive = false
	return nil
}

func (s *pubSubSubscription) IsActive() bool {
	return s.isActive
}
