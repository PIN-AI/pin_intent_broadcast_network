package transport

import (
	"crypto/sha256"
	"fmt"
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

// MessageType constants for different message types
const (
	MessageTypeIntentBroadcast = "intent_broadcast"
	MessageTypeBidSubmission   = "bid_submission"
	MessageTypeMatchResult     = "match_result"
	MessageTypeBidCommitment   = "bid_commitment"
	MessageTypeBidReveal       = "bid_reveal"
)

// Topic constants for bidding and matching
const (
	TopicIntentBroadcast = "/intent-network/intents/1.0.0"
	TopicBidSubmission   = "/intent-network/bids/1.0.0"
	TopicMatchResults    = "/intent-network/matches/1.0.0"
	TopicBidCommitments  = "/intent-network/commitments/1.0.0"
	TopicBidReveals      = "/intent-network/reveals/1.0.0"
)

// BidMessage represents a bid submission for an intent
type BidMessage struct {
	IntentID     string            `json:"intent_id"`
	AgentID      string            `json:"agent_id"`
	BidAmount    string            `json:"bid_amount"`
	Capabilities []string          `json:"capabilities"`
	Timestamp    int64             `json:"timestamp"`
	AgentType    string            `json:"agent_type"`
	Metadata     map[string]string `json:"metadata"`
	Signature    []byte            `json:"signature,omitempty"`
}

// BidCommitment represents a commitment hash for a bid
type BidCommitment struct {
	IntentID      string `json:"intent_id"`
	AgentID       string `json:"agent_id"`
	CommitmentHash string `json:"commitment_hash"`
	Timestamp     int64  `json:"timestamp"`
	Signature     []byte `json:"signature,omitempty"`
}

// BidReveal represents the reveal of a committed bid
type BidReveal struct {
	IntentID  string     `json:"intent_id"`
	AgentID   string     `json:"agent_id"`
	BidData   *BidMessage `json:"bid_data"`
	Nonce     string     `json:"nonce"`
	Timestamp int64      `json:"timestamp"`
	Signature []byte     `json:"signature,omitempty"`
}

// MatchResult represents the result of intent matching
type MatchResult struct {
	IntentID      string            `json:"intent_id"`
	WinningAgent  string            `json:"winning_agent"`
	WinningBid    string            `json:"winning_bid"`
	TotalBids     int               `json:"total_bids"`
	MatchedAt     int64             `json:"matched_at"`
	Status        string            `json:"status"`
	Metadata      map[string]string `json:"metadata"`
	BlockBuilderID string           `json:"block_builder_id"`
}

// IntentSession represents an active intent bidding session
type IntentSession struct {
	IntentID        string          `json:"intent_id"`
	StartTime       int64           `json:"start_time"`
	EndTime         int64           `json:"end_time"`
	Status          string          `json:"status"` // "collecting", "matching", "completed", "expired"
	BidCount        int             `json:"bid_count"`
	CommitmentCount int             `json:"commitment_count"`
	MatchResult     *MatchResult    `json:"match_result,omitempty"`
}

// BidStatus represents the status of a bid
type BidStatus struct {
	IntentID  string `json:"intent_id"`
	AgentID   string `json:"agent_id"`
	Status    string `json:"status"` // "submitted", "committed", "revealed", "matched", "rejected"
	Timestamp int64  `json:"timestamp"`
	Reason    string `json:"reason,omitempty"`
}

// CalculateBidCommitment calculates SHA2-256 commitment hash for a bid
func CalculateBidCommitment(bid *BidMessage, nonce string) string {
	// Create commitment payload
	commitmentData := fmt.Sprintf("%s:%s:%s:%s:%d:%s", 
		bid.IntentID, 
		bid.AgentID, 
		bid.BidAmount, 
		bid.AgentType,
		bid.Timestamp,
		nonce,
	)
	
	// Calculate SHA2-256 hash
	hash := sha256.Sum256([]byte(commitmentData))
	return fmt.Sprintf("%x", hash)
}

// ValidateBidReveal validates that a bid reveal matches its commitment
func ValidateBidReveal(reveal *BidReveal, commitment *BidCommitment) bool {
	calculatedHash := CalculateBidCommitment(reveal.BidData, reveal.Nonce)
	return calculatedHash == commitment.CommitmentHash
}

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