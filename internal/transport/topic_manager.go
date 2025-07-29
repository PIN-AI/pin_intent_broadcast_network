package transport

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

// topicManager simple topic manager implementation
type topicManager struct {
	topics  map[string]*TopicConfig
	mu      sync.RWMutex
	logger  *zap.Logger
}

// NewTopicManager creates new topic manager
func NewTopicManager(logger *zap.Logger) TopicManager {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	return &topicManager{
		topics: make(map[string]*TopicConfig),
		logger: logger.Named("topic_manager"),
	}
}

// RegisterTopic registers a new topic
func (tm *topicManager) RegisterTopic(topic *TopicConfig) error {
	if topic == nil {
		return &TransportError{"INVALID_TOPIC_CONFIG", "Topic configuration cannot be nil", ""}
	}
	
	if err := ValidateTopicName(topic.Name); err != nil {
		return err
	}
	
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	// Check if topic already exists
	if _, exists := tm.topics[topic.Name]; exists {
		return ErrTopicAlreadyExists
	}
	
	// Set default values if not provided
	if topic.MaxMessageSize <= 0 {
		topic.MaxMessageSize = 1024 * 1024 // 1MB default
	}
	
	if topic.RateLimit <= 0 {
		topic.RateLimit = 100 // 100 messages per second default
	}
	
	// Set timestamps
	now := time.Now()
	topic.CreatedAt = now
	topic.UpdatedAt = now
	
	// Store topic configuration
	tm.topics[topic.Name] = topic
	
	tm.logger.Info("Topic registered successfully",
		zap.String("topic", FormatTopic(topic.Name)),
		zap.String("access_control", getAccessControlString(topic.AccessControl)),
		zap.Int("max_message_size", topic.MaxMessageSize),
		zap.Int("rate_limit", topic.RateLimit),
	)
	
	return nil
}

// UnregisterTopic unregisters a topic
func (tm *topicManager) UnregisterTopic(topicName string) error {
	if err := ValidateTopicName(topicName); err != nil {
		return err
	}
	
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if _, exists := tm.topics[topicName]; !exists {
		return ErrTopicNotFound
	}
	
	delete(tm.topics, topicName)
	
	tm.logger.Info("Topic unregistered successfully",
		zap.String("topic", FormatTopic(topicName)),
	)
	
	return nil
}

// ValidateTopic validates topic access for a peer
func (tm *topicManager) ValidateTopic(topicName string, peerID peer.ID) bool {
	if err := ValidateTopicName(topicName); err != nil {
		tm.logger.Debug("Invalid topic name",
			zap.String("topic", topicName),
			zap.Error(err),
		)
		return false
	}
	
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	topic, exists := tm.topics[topicName]
	if !exists {
		// If topic is not registered, allow access (simple policy)
		tm.logger.Debug("Topic not registered, allowing access",
			zap.String("topic", FormatTopic(topicName)),
			zap.String("peer_id", FormatPeerID(peerID)),
		)
		return true
	}
	
	// Check access control
	switch topic.AccessControl {
	case TopicAccessPublic:
		return true
		
	case TopicAccessWhitelist:
		// Check if peer is in allowed list
		for _, allowedPeer := range topic.AllowedPeers {
			if allowedPeer == peerID {
				return true
			}
		}
		tm.logger.Debug("Peer not in whitelist",
			zap.String("topic", FormatTopic(topicName)),
			zap.String("peer_id", FormatPeerID(peerID)),
		)
		return false
		
	case TopicAccessBlacklist:
		// Check if peer is in denied list
		for _, deniedPeer := range topic.DeniedPeers {
			if deniedPeer == peerID {
				tm.logger.Debug("Peer in blacklist",
					zap.String("topic", FormatTopic(topicName)),
					zap.String("peer_id", FormatPeerID(peerID)),
				)
				return false
			}
		}
		return true
		
	case TopicAccessPrivate:
		tm.logger.Debug("Topic is private",
			zap.String("topic", FormatTopic(topicName)),
			zap.String("peer_id", FormatPeerID(peerID)),
		)
		return false
		
	default:
		tm.logger.Warn("Unknown access control type",
			zap.String("topic", FormatTopic(topicName)),
			zap.Int("access_control", int(topic.AccessControl)),
		)
		return false
	}
}

// GetTopicConfig returns topic configuration
func (tm *topicManager) GetTopicConfig(topicName string) (*TopicConfig, error) {
	if err := ValidateTopicName(topicName); err != nil {
		return nil, err
	}
	
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	topic, exists := tm.topics[topicName]
	if !exists {
		return nil, ErrTopicNotFound
	}
	
	// Return a copy to prevent modification
	topicCopy := *topic
	return &topicCopy, nil
}

// ListTopics returns all registered topics
func (tm *topicManager) ListTopics() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	topics := make([]string, 0, len(tm.topics))
	for topicName := range tm.topics {
		topics = append(topics, topicName)
	}
	
	return topics
}

// ValidateMessage validates message against topic configuration
func (tm *topicManager) ValidateMessage(topicName string, msg *TransportMessage) error {
	if err := ValidateTopicName(topicName); err != nil {
		return err
	}
	
	if err := ValidateMessageFormat(msg); err != nil {
		return err
	}
	
	tm.mu.RLock()
	topic, exists := tm.topics[topicName]
	tm.mu.RUnlock()
	
	if !exists {
		// If topic is not registered, use basic validation
		return ValidateMessageFormat(msg)
	}
	
	// Check message size
	msgSize := GetMessageSize(msg)
	if msgSize > topic.MaxMessageSize {
		return &TransportError{
			Code:    "MESSAGE_TOO_LARGE",
			Message: fmt.Sprintf("Message size %d exceeds topic limit %d", msgSize, topic.MaxMessageSize),
			Details: topicName,
		}
	}
	
	// Check TTL
	if IsMessageExpired(msg) {
		return &TransportError{
			Code:    "MESSAGE_EXPIRED",
			Message: "Message has expired",
			Details: fmt.Sprintf("topic: %s, ttl: %d", topicName, msg.TTL),
		}
	}
	
	// Run custom validator if provided
	if topic.ValidatorFunc != nil {
		if err := topic.ValidatorFunc(msg); err != nil {
			tm.logger.Debug("Custom validator failed",
				zap.String("topic", FormatTopic(topicName)),
				zap.Error(err),
			)
			return err
		}
	}
	
	return nil
}

// CreateDefaultTopic creates a topic with default public access
func CreateDefaultTopic(name, description string) *TopicConfig {
	return &TopicConfig{
		Name:            name,
		Description:     description,
		AccessControl:   TopicAccessPublic,
		MaxMessageSize:  1024 * 1024, // 1MB
		RateLimit:       100,          // 100 messages/second
		EnableSigning:   true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// CreateWhitelistTopic creates a topic with whitelist access control
func CreateWhitelistTopic(name, description string, allowedPeers []peer.ID) *TopicConfig {
	return &TopicConfig{
		Name:            name,
		Description:     description,
		AccessControl:   TopicAccessWhitelist,
		AllowedPeers:    allowedPeers,
		MaxMessageSize:  1024 * 1024, // 1MB
		RateLimit:       100,          // 100 messages/second
		EnableSigning:   true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// CreateBlacklistTopic creates a topic with blacklist access control
func CreateBlacklistTopic(name, description string, deniedPeers []peer.ID) *TopicConfig {
	return &TopicConfig{
		Name:            name,
		Description:     description,
		AccessControl:   TopicAccessBlacklist,
		DeniedPeers:     deniedPeers,
		MaxMessageSize:  1024 * 1024, // 1MB
		RateLimit:       100,          // 100 messages/second
		EnableSigning:   true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// getAccessControlString returns string representation of access control type
func getAccessControlString(accessType TopicAccessType) string {
	switch accessType {
	case TopicAccessPublic:
		return "public"
	case TopicAccessWhitelist:
		return "whitelist"
	case TopicAccessBlacklist:
		return "blacklist"
	case TopicAccessPrivate:
		return "private"
	default:
		return "unknown"
	}
}