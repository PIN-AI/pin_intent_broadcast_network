package transport

import (
	"context"
	"fmt"
	"time"

	"pin_intent_broadcast_network/internal/conf"
	"pin_intent_broadcast_network/internal/biz/common"

	"github.com/libp2p/go-libp2p/core/host"
	"go.uber.org/zap"
)

// transportManager unified transport manager implementation
type transportManager struct {
	host              host.Host
	pubsubManager     PubSubManager
	topicManager      TopicManager
	messageSerializer MessageSerializer
	messageRouter     MessageRouter
	config            *TransportConfig
	logger            *zap.Logger
	isRunning         bool
}

// NewTransportManager creates new transport manager
func NewTransportManager(h host.Host, logger *zap.Logger) TransportManager {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	tm := &transportManager{
		host:   h,
		logger: logger.Named("transport_manager"),
	}
	
	// Initialize components
	tm.pubsubManager = NewPubSubManager(h, logger)
	tm.topicManager = NewTopicManager(logger)
	tm.messageSerializer = NewMessageSerializer(logger)
	tm.messageRouter = NewMessageRouter(1000, 10*time.Minute, logger)
	
	return tm
}

// Start starts the transport manager
func (tm *transportManager) Start(ctx context.Context, config *TransportConfig) error {
	if tm.isRunning {
		return fmt.Errorf("transport manager already running")
	}
	
	if config == nil {
		config = DefaultTransportConfig()
	}
	
	tm.config = config
	
	// Start PubSub manager if enabled
	if config.EnableGossipSub {
		pubsubConfig := &PubSubConfig{
			HeartbeatInterval:        config.GossipSubHeartbeatInterval,
			D:                        config.GossipSubD,
			DLo:                      config.GossipSubDLo,
			DHi:                      config.GossipSubDHi,
			FanoutTTL:                config.GossipSubFanoutTTL,
			EnableSigning:            config.EnableMessageSigning,
			EnableStrictVerification: config.EnableStrictSignatureVerification,
		}
		
		if err := tm.pubsubManager.Start(ctx, pubsubConfig); err != nil {
			return fmt.Errorf("failed to start pubsub manager: %w", err)
		}
	}
	
	// Start message router
	if err := tm.messageRouter.Start(ctx); err != nil {
		return fmt.Errorf("failed to start message router: %w", err)
	}
	
	// Add default filters
	tm.messageRouter.AddFilter(NewTTLFilter())
	tm.messageRouter.AddFilter(NewSizeFilter(config.MaxMessageSize))
	
	tm.isRunning = true
	
	tm.logger.Info("Transport manager started successfully",
		zap.Bool("gossipsub_enabled", config.EnableGossipSub),
		zap.Duration("heartbeat_interval", config.GossipSubHeartbeatInterval),
		zap.Int("gossipsub_d", config.GossipSubD),
		zap.Bool("message_signing", config.EnableMessageSigning),
		zap.Int("max_message_size", config.MaxMessageSize),
	)
	
	return nil
}

// Stop stops the transport manager
func (tm *transportManager) Stop() error {
	if !tm.isRunning {
		return fmt.Errorf("transport manager not running")
	}
	
	// Stop PubSub manager
	if tm.pubsubManager != nil {
		if err := tm.pubsubManager.Stop(); err != nil {
			tm.logger.Error("Failed to stop pubsub manager", zap.Error(err))
		}
	}
	
	// Stop message router
	if tm.messageRouter != nil {
		if err := tm.messageRouter.Stop(); err != nil {
			tm.logger.Error("Failed to stop message router", zap.Error(err))
		}
	}
	
	tm.isRunning = false
	tm.logger.Info("Transport manager stopped successfully")
	
	return nil
}

// IsRunning returns whether the transport manager is running
func (tm *transportManager) IsRunning() bool {
	return tm.isRunning
}

// GetPubSubManager returns the pubsub manager
func (tm *transportManager) GetPubSubManager() PubSubManager {
	return tm.pubsubManager
}

// GetTopicManager returns the topic manager
func (tm *transportManager) GetTopicManager() TopicManager {
	return tm.topicManager
}

// GetMessageSerializer returns the message serializer
func (tm *transportManager) GetMessageSerializer() MessageSerializer {
	return tm.messageSerializer
}

// GetMessageRouter returns the message router
func (tm *transportManager) GetMessageRouter() MessageRouter {
	return tm.messageRouter
}

// PublishMessage publishes a message to a topic (high-level API)
func (tm *transportManager) PublishMessage(ctx context.Context, topic string, msg *TransportMessage) error {
	if !tm.isRunning {
		return ErrTransportNotRunning
	}
	
	// Route message through router (includes deduplication and filtering)
	if err := tm.messageRouter.RouteMessage(ctx, topic, msg); err != nil {
		return err
	}
	
	// Validate topic access (if topic is registered)
	if config, err := tm.topicManager.GetTopicConfig(topic); err == nil {
		// Topic is registered, validate message against topic config
		if err := tm.topicManager.(*topicManager).ValidateMessage(topic, msg); err != nil {
			return err
		}
		
		// Check message size against topic limits
		msgSize := GetMessageSize(msg)
		if msgSize > config.MaxMessageSize {
			return ErrMessageTooLarge
		}
	}
	
	// Serialize message
	data, err := tm.messageSerializer.Serialize(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}
	
	// Check global message size limit
	if len(data) > tm.config.MaxMessageSize {
		return ErrMessageTooLarge
	}
	
	// Publish to topic
	if err := tm.pubsubManager.Publish(ctx, topic, data); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	
	tm.logger.Debug("Message published successfully",
		zap.String("topic", FormatTopic(topic)),
		zap.String("message_id", msg.ID),
		zap.String("message_type", msg.Type),
		zap.Int("serialized_size", len(data)),
	)
	
	return nil
}

// SubscribeToTopic subscribes to a topic with automatic message deserialization
func (tm *transportManager) SubscribeToTopic(topic string, handler func(*TransportMessage) error) (Subscription, error) {
	if !tm.isRunning {
		return nil, ErrTransportNotRunning
	}
	
	// Create message handler wrapper that deserializes messages
	wrappedHandler := func(rawMsg *TransportMessage) error {
		// For messages coming from PubSub, the payload contains the serialized TransportMessage
		if rawMsg.Type == "pubsub" && len(rawMsg.Payload) > 0 {
			// Deserialize the actual message from payload
			deserializedMsg, err := tm.messageSerializer.Deserialize(rawMsg.Payload)
			if err != nil {
				tm.logger.Error("Failed to deserialize received message",
					zap.String("topic", FormatTopic(topic)),
					zap.Error(err),
				)
				return err
			}
			
			// Call the original handler with deserialized message
			return handler(deserializedMsg)
		}
		
		// For other message types, pass through directly
		return handler(rawMsg)
	}
	
	// Subscribe to topic
	subscription, err := tm.pubsubManager.Subscribe(topic, wrappedHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}
	
	tm.logger.Info("Successfully subscribed to topic",
		zap.String("topic", FormatTopic(topic)),
	)
	
	return subscription, nil
}

// GetTransportMetrics returns transport metrics
func (tm *transportManager) GetTransportMetrics() *TransportMetrics {
	metrics := &TransportMetrics{
		LastUpdated: time.Now(),
	}
	
	if tm.pubsubManager != nil {
		metrics.SubscriptionCount = len(tm.pubsubManager.GetTopics())
		metrics.ConnectedPeerCount = len(tm.pubsubManager.GetConnectedPeers())
	}
	
	if tm.topicManager != nil {
		metrics.ActiveTopicCount = len(tm.topicManager.ListTopics())
	}
	
	// TODO: Add more detailed metrics collection
	// - MessagesPublished, MessagesReceived, MessagesSent, MessagesDropped
	// These would require counters in each component
	
	return metrics
}

// Helper function to create transport config from bootstrap config
func NewTransportConfigFromBootstrap(bc *conf.Bootstrap) *TransportConfig {
	if bc.Transport == nil {
		return DefaultTransportConfig()
	}
	
	config := &TransportConfig{
		EnableGossipSub:                   bc.Transport.EnableGossipsub,
		GossipSubHeartbeatInterval:        bc.Transport.GossipsubHeartbeatInterval.AsDuration(),
		GossipSubD:                        int(bc.Transport.GossipsubD),
		GossipSubDLo:                      int(bc.Transport.GossipsubDLo),
		GossipSubDHi:                      int(bc.Transport.GossipsubDHi),
		GossipSubFanoutTTL:                bc.Transport.GossipsubFanoutTtl.AsDuration(),
		EnableMessageSigning:              bc.Transport.EnableMessageSigning,
		EnableStrictSignatureVerification: bc.Transport.EnableStrictSignatureVerification,
		MessageIDCacheSize:                int(bc.Transport.MessageIdCacheSize),
		MessageTTL:                        bc.Transport.MessageTtl.AsDuration(),
		MaxMessageSize:                    int(bc.Transport.MaxMessageSize),
	}
	
	return config
}

// PublishBidMessage publishes a bid message
func (tm *transportManager) PublishBidMessage(ctx context.Context, bid *BidMessage) error {
	if !tm.isRunning {
		return ErrTransportNotRunning
	}
	
	// Serialize bid message
	bidData, err := common.JSON.Marshal(bid)
	if err != nil {
		return fmt.Errorf("failed to serialize bid message: %w", err)
	}
	
	// Create transport message
	transportMsg := &TransportMessage{
		Type:      MessageTypeBidSubmission,
		Payload:   bidData,
		Timestamp: time.Now().UnixMilli(),
		Sender:    bid.AgentID,
		Priority:  PriorityHigh,
		Metadata:  make(map[string]string),
	}
	
	// Generate message ID
	transportMsg.ID = GenerateMessageID(transportMsg)
	
	// Add bid metadata
	transportMsg.Metadata["intent_id"] = bid.IntentID
	transportMsg.Metadata["agent_id"] = bid.AgentID
	transportMsg.Metadata["bid_amount"] = bid.BidAmount
	transportMsg.Metadata["agent_type"] = bid.AgentType
	
	// Publish to bid submission topic
	return tm.PublishMessage(ctx, TopicBidSubmission, transportMsg)
}

// PublishMatchResult publishes a match result
func (tm *transportManager) PublishMatchResult(ctx context.Context, result *MatchResult) error {
	if !tm.isRunning {
		return ErrTransportNotRunning
	}
	
	// Serialize match result
	resultData, err := common.JSON.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to serialize match result: %w", err)
	}
	
	// Create transport message
	transportMsg := &TransportMessage{
		Type:      MessageTypeMatchResult,
		Payload:   resultData,
		Timestamp: time.Now().UnixMilli(),
		Sender:    result.BlockBuilderID,
		Priority:  PriorityCritical,
		Metadata:  make(map[string]string),
	}
	
	// Generate message ID
	transportMsg.ID = GenerateMessageID(transportMsg)
	
	// Add match metadata
	transportMsg.Metadata["intent_id"] = result.IntentID
	transportMsg.Metadata["winning_agent"] = result.WinningAgent
	transportMsg.Metadata["winning_bid"] = result.WinningBid
	transportMsg.Metadata["match_status"] = result.Status
	
	// Publish to match results topic
	return tm.PublishMessage(ctx, TopicMatchResults, transportMsg)
}

// SubscribeToBids subscribes to bid submissions
func (tm *transportManager) SubscribeToBids(handler func(*BidMessage) error) (Subscription, error) {
	if !tm.isRunning {
		return nil, ErrTransportNotRunning
	}
	
	// Create bid message handler wrapper
	bidHandler := func(msg *TransportMessage) error {
		if msg.Type != MessageTypeBidSubmission {
			return nil // Not a bid message
		}
		
		// Deserialize bid message
		var bid BidMessage
		if err := common.JSON.Unmarshal(msg.Payload, &bid); err != nil {
			tm.logger.Error("Failed to deserialize bid message",
				zap.String("message_id", msg.ID),
				zap.Error(err),
			)
			return err
		}
		
		// Call the bid handler
		return handler(&bid)
	}
	
	// Subscribe to bid submission topic
	subscription, err := tm.SubscribeToTopic(TopicBidSubmission, bidHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to bids: %w", err)
	}
	
	tm.logger.Info("Successfully subscribed to bid submissions")
	return subscription, nil
}

// SubscribeToMatches subscribes to match results
func (tm *transportManager) SubscribeToMatches(handler func(*MatchResult) error) (Subscription, error) {
	if !tm.isRunning {
		return nil, ErrTransportNotRunning
	}
	
	// Create match result handler wrapper
	matchHandler := func(msg *TransportMessage) error {
		if msg.Type != MessageTypeMatchResult {
			return nil // Not a match result message
		}
		
		// Deserialize match result
		var result MatchResult
		if err := common.JSON.Unmarshal(msg.Payload, &result); err != nil {
			tm.logger.Error("Failed to deserialize match result",
				zap.String("message_id", msg.ID),
				zap.Error(err),
			)
			return err
		}
		
		// Call the match handler
		return handler(&result)
	}
	
	// Subscribe to match results topic
	subscription, err := tm.SubscribeToTopic(TopicMatchResults, matchHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to matches: %w", err)
	}
	
	tm.logger.Info("Successfully subscribed to match results")
	return subscription, nil
}