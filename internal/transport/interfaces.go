package transport

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

// TransportManager manages all transport layer operations
type TransportManager interface {
	// Start starts the transport manager
	Start(ctx context.Context, config *TransportConfig) error
	// Stop stops the transport manager
	Stop() error
	// IsRunning returns whether the transport manager is running
	IsRunning() bool
	
	// GetPubSubManager returns the pubsub manager
	GetPubSubManager() PubSubManager
	// GetTopicManager returns the topic manager
	GetTopicManager() TopicManager
	// GetMessageSerializer returns the message serializer
	GetMessageSerializer() MessageSerializer
	// GetMessageRouter returns the message router
	GetMessageRouter() MessageRouter
	
	// PublishMessage publishes a message to a topic (high-level API)
	PublishMessage(ctx context.Context, topic string, msg *TransportMessage) error
	// SubscribeToTopic subscribes to a topic with automatic message deserialization
	SubscribeToTopic(topic string, handler func(*TransportMessage) error) (Subscription, error)
}

// PubSubManager manages GossipSub publish/subscribe operations
type PubSubManager interface {
	// Start starts the pubsub manager
	Start(ctx context.Context, config *PubSubConfig) error
	// Stop stops the pubsub manager
	Stop() error
	
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, data []byte) error
	// Subscribe subscribes to a topic
	Subscribe(topic string, handler MessageHandler) (Subscription, error)
	// Unsubscribe unsubscribes from a topic
	Unsubscribe(topic string) error
	
	// GetConnectedPeers returns connected peers
	GetConnectedPeers() []peer.ID
	// GetTopics returns subscribed topics
	GetTopics() []string
	// GetPeerCount returns peer count for a topic
	GetPeerCount(topic string) int
}

// TopicManager manages topic registration and validation
type TopicManager interface {
	// RegisterTopic registers a new topic
	RegisterTopic(topic *TopicConfig) error
	// UnregisterTopic unregisters a topic
	UnregisterTopic(topicName string) error
	// ValidateTopic validates topic access
	ValidateTopic(topicName string, peerID peer.ID) bool
	// GetTopicConfig returns topic configuration
	GetTopicConfig(topicName string) (*TopicConfig, error)
	// ListTopics returns all registered topics
	ListTopics() []string
}

// MessageSerializer handles message serialization and validation
type MessageSerializer interface {
	// Serialize serializes a transport message
	Serialize(msg *TransportMessage) ([]byte, error)
	// Deserialize deserializes data to transport message
	Deserialize(data []byte) (*TransportMessage, error)
	// ValidateMessage validates message integrity
	ValidateMessage(msg *TransportMessage) error
	// SignMessage signs a message
	SignMessage(msg *TransportMessage) error
	// VerifySignature verifies message signature
	VerifySignature(msg *TransportMessage) error
}

// MessageHandler handles incoming messages
type MessageHandler func(msg *TransportMessage) error

// Subscription represents a topic subscription
type Subscription interface {
	// Topic returns the subscribed topic
	Topic() string
	// Cancel cancels the subscription
	Cancel() error
	// IsActive returns whether subscription is active
	IsActive() bool
}