package transport

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// GenerateMessageID generates unique message ID
func GenerateMessageID(msg *TransportMessage) string {
	data := fmt.Sprintf("%s_%s_%d_%s", msg.Type, msg.Sender, msg.Timestamp, string(msg.Payload))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for ID
}

// ValidateMessageFormat validates basic message format
func ValidateMessageFormat(msg *TransportMessage) error {
	if msg == nil {
		return ErrInvalidMessage
	}
	
	if msg.ID == "" {
		return &TransportError{"INVALID_MESSAGE_ID", "Message ID cannot be empty", ""}
	}
	
	if msg.Type == "" {
		return &TransportError{"INVALID_MESSAGE_TYPE", "Message type cannot be empty", ""}
	}
	
	if msg.Sender == "" {
		return &TransportError{"INVALID_SENDER", "Message sender cannot be empty", ""}
	}
	
	if msg.Timestamp <= 0 {
		return &TransportError{"INVALID_TIMESTAMP", "Message timestamp must be positive", ""}
	}
	
	if len(msg.Payload) == 0 {
		return &TransportError{"EMPTY_PAYLOAD", "Message payload cannot be empty", ""}
	}
	
	return nil
}

// IsMessageExpired checks if message has expired
func IsMessageExpired(msg *TransportMessage) bool {
	if msg.TTL <= 0 {
		return false // No TTL means never expires
	}
	
	expiryTime := time.Unix(msg.Timestamp/1000, 0).Add(time.Duration(msg.TTL) * time.Millisecond)
	return time.Now().After(expiryTime)
}

// FormatPeerID formats peer ID for logging
func FormatPeerID(peerID peer.ID) string {
	s := peerID.String()
	if len(s) > 16 {
		return s[:8] + "..." + s[len(s)-8:]
	}
	return s
}

// FormatTopic formats topic name for logging
func FormatTopic(topic string) string {
	// Remove any sensitive information or clean up topic name
	return strings.TrimSpace(topic)
}

// ValidateTopicName validates topic name format
func ValidateTopicName(topic string) error {
	if topic == "" {
		return &TransportError{"INVALID_TOPIC_NAME", "Topic name cannot be empty", ""}
	}
	
	if len(topic) > 256 {
		return &TransportError{"TOPIC_NAME_TOO_LONG", "Topic name exceeds maximum length", ""}
	}
	
	// Check for invalid characters
	if strings.ContainsAny(topic, " \t\n\r") {
		return &TransportError{"INVALID_TOPIC_CHARACTERS", "Topic name contains invalid characters", ""}
	}
	
	return nil
}

// GetMessageSize calculates approximate message size in bytes
func GetMessageSize(msg *TransportMessage) int {
	size := len(msg.ID) + len(msg.Type) + len(msg.Payload) + len(msg.Sender)
	size += 8 + 4 + 8 // timestamp, priority, ttl
	
	if msg.Signature != nil {
		size += len(msg.Signature)
	}
	
	for k, v := range msg.Metadata {
		size += len(k) + len(v)
	}
	
	return size
}

// CreateTransportMessage creates a new transport message
func CreateTransportMessage(msgType string, payload []byte, sender peer.ID) *TransportMessage {
	now := time.Now().UnixMilli()
	msg := &TransportMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: now,
		Sender:    sender.String(),
		Priority:  PriorityNormal,
		TTL:       int64(5 * time.Minute / time.Millisecond), // 5 minutes default TTL
		Metadata:  make(map[string]string),
	}
	msg.ID = GenerateMessageID(msg)
	return msg
}