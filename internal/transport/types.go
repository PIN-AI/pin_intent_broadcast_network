package transport

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// TransportConfig transport configuration
type TransportConfig struct {
	EnableGossipSub                 bool          `json:"enable_gossipsub"`
	GossipSubHeartbeatInterval      time.Duration `json:"gossipsub_heartbeat_interval"`
	GossipSubD                      int           `json:"gossipsub_d"`
	GossipSubDLo                    int           `json:"gossipsub_d_lo"`
	GossipSubDHi                    int           `json:"gossipsub_d_hi"`
	GossipSubFanoutTTL              time.Duration `json:"gossipsub_fanout_ttl"`
	EnableMessageSigning            bool          `json:"enable_message_signing"`
	EnableStrictSignatureVerification bool        `json:"enable_strict_signature_verification"`
	MessageIDCacheSize              int           `json:"message_id_cache_size"`
	MessageTTL                      time.Duration `json:"message_ttl"`
	MaxMessageSize                  int           `json:"max_message_size"`
}

// PubSubConfig pubsub configuration
type PubSubConfig struct {
	HeartbeatInterval      time.Duration `json:"heartbeat_interval"`
	D                      int           `json:"d"`
	DLo                    int           `json:"d_lo"`
	DHi                    int           `json:"d_hi"`
	FanoutTTL              time.Duration `json:"fanout_ttl"`
	EnableSigning          bool          `json:"enable_signing"`
	EnableStrictVerification bool        `json:"enable_strict_verification"`
}

// TopicConfig topic configuration
type TopicConfig struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	AccessControl   TopicAccessType   `json:"access_control"`
	AllowedPeers    []peer.ID         `json:"allowed_peers,omitempty"`
	DeniedPeers     []peer.ID         `json:"denied_peers,omitempty"`
	MaxMessageSize  int               `json:"max_message_size"`
	RateLimit       int               `json:"rate_limit"`       // messages per second
	EnableSigning   bool              `json:"enable_signing"`
	ValidatorFunc   MessageValidator  `json:"-"`                // custom validator function
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// TransportMessage transport layer message
type TransportMessage struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Payload   []byte            `json:"payload"`
	Timestamp int64             `json:"timestamp"`
	Sender    string            `json:"sender"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Signature []byte            `json:"signature,omitempty"`
	Priority  int32             `json:"priority"`
	TTL       int64             `json:"ttl"`
}

// TopicAccessType topic access control type
type TopicAccessType int

const (
	TopicAccessPublic TopicAccessType = iota
	TopicAccessWhitelist
	TopicAccessBlacklist
	TopicAccessPrivate
)

// MessageValidator custom message validator function
type MessageValidator func(msg *TransportMessage) error

// TransportError transport layer error
type TransportError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *TransportError) Error() string {
	if e.Details != "" {
		return e.Code + ": " + e.Message + " (" + e.Details + ")"
	}
	return e.Code + ": " + e.Message
}

// Common transport errors
var (
	ErrTransportNotRunning     = &TransportError{"TRANSPORT_NOT_RUNNING", "Transport manager is not running", ""}
	ErrTopicNotFound           = &TransportError{"TOPIC_NOT_FOUND", "Topic not found", ""}
	ErrTopicAlreadyExists      = &TransportError{"TOPIC_ALREADY_EXISTS", "Topic already exists", ""}
	ErrInvalidMessage          = &TransportError{"INVALID_MESSAGE", "Invalid message format", ""}
	ErrMessageTooLarge         = &TransportError{"MESSAGE_TOO_LARGE", "Message exceeds maximum size", ""}
	ErrSignatureVerificationFailed = &TransportError{"SIGNATURE_VERIFICATION_FAILED", "Message signature verification failed", ""}
	ErrRateLimitExceeded       = &TransportError{"RATE_LIMIT_EXCEEDED", "Message rate limit exceeded", ""}
	ErrAccessDenied            = &TransportError{"ACCESS_DENIED", "Access denied to topic", ""}
)

// TransportMetrics transport metrics
type TransportMetrics struct {
	MessagesPublished   int64     `json:"messages_published"`
	MessagesReceived    int64     `json:"messages_received"`
	MessagesSent        int64     `json:"messages_sent"`
	MessagesDropped     int64     `json:"messages_dropped"`
	SubscriptionCount   int       `json:"subscription_count"`
	ActiveTopicCount    int       `json:"active_topic_count"`
	ConnectedPeerCount  int       `json:"connected_peer_count"`
	LastUpdated         time.Time `json:"last_updated"`
}

// MessagePriority message priority levels
const (
	PriorityLow    int32 = 0
	PriorityNormal int32 = 50
	PriorityHigh   int32 = 100
	PriorityCritical int32 = 255
)

// DefaultTransportConfig returns default transport configuration
func DefaultTransportConfig() *TransportConfig {
	return &TransportConfig{
		EnableGossipSub:                   true,
		GossipSubHeartbeatInterval:        time.Second,
		GossipSubD:                        6,
		GossipSubDLo:                      4,
		GossipSubDHi:                      12,
		GossipSubFanoutTTL:                60 * time.Second,
		EnableMessageSigning:              true,
		EnableStrictSignatureVerification: true,
		MessageIDCacheSize:                1000,
		MessageTTL:                        5 * time.Minute,
		MaxMessageSize:                    1024 * 1024, // 1MB
	}
}