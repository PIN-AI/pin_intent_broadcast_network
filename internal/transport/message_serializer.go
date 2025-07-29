package transport

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// messageSerializer basic message serializer implementation
type messageSerializer struct {
	logger *zap.Logger
}

// NewMessageSerializer creates new message serializer
func NewMessageSerializer(logger *zap.Logger) MessageSerializer {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	return &messageSerializer{
		logger: logger.Named("message_serializer"),
	}
}

// Serialize serializes a transport message to bytes
func (ms *messageSerializer) Serialize(msg *TransportMessage) ([]byte, error) {
	if msg == nil {
		return nil, ErrInvalidMessage
	}
	
	// Validate message format first
	if err := ValidateMessageFormat(msg); err != nil {
		return nil, err
	}
	
	// Use JSON serialization for simplicity (can be replaced with protobuf later)
	data, err := json.Marshal(msg)
	if err != nil {
		ms.logger.Error("Failed to serialize message",
			zap.String("message_id", msg.ID),
			zap.String("message_type", msg.Type),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	
	ms.logger.Debug("Message serialized successfully",
		zap.String("message_id", msg.ID),
		zap.String("message_type", msg.Type),
		zap.Int("serialized_size", len(data)),
	)
	
	return data, nil
}

// Deserialize deserializes bytes to transport message
func (ms *messageSerializer) Deserialize(data []byte) (*TransportMessage, error) {
	if len(data) == 0 {
		return nil, &TransportError{"EMPTY_DATA", "Cannot deserialize empty data", ""}
	}
	
	var msg TransportMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		ms.logger.Error("Failed to deserialize message",
			zap.Int("data_size", len(data)),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to deserialize message: %w", err)
	}
	
	// Validate deserialized message
	if err := ValidateMessageFormat(&msg); err != nil {
		ms.logger.Error("Deserialized message validation failed",
			zap.String("message_id", msg.ID),
			zap.Error(err),
		)
		return nil, err
	}
	
	ms.logger.Debug("Message deserialized successfully",
		zap.String("message_id", msg.ID),
		zap.String("message_type", msg.Type),
		zap.Int("data_size", len(data)),
	)
	
	return &msg, nil
}

// ValidateMessage validates message integrity
func (ms *messageSerializer) ValidateMessage(msg *TransportMessage) error {
	if msg == nil {
		return ErrInvalidMessage
	}
	
	// Basic format validation
	if err := ValidateMessageFormat(msg); err != nil {
		return err
	}
	
	// Check message expiration
	if IsMessageExpired(msg) {
		return &TransportError{
			Code:    "MESSAGE_EXPIRED",
			Message: "Message has expired",
			Details: fmt.Sprintf("message_id: %s, ttl: %d", msg.ID, msg.TTL),
		}
	}
	
	// Validate message ID consistency
	expectedID := GenerateMessageID(msg)
	if msg.ID != expectedID {
		ms.logger.Warn("Message ID mismatch",
			zap.String("provided_id", msg.ID),
			zap.String("expected_id", expectedID),
		)
		// For now, just log warning instead of failing
		// In production, this might be more strict
	}
	
	ms.logger.Debug("Message validation passed",
		zap.String("message_id", msg.ID),
		zap.String("message_type", msg.Type),
	)
	
	return nil
}

// SignMessage signs a message (basic implementation)
func (ms *messageSerializer) SignMessage(msg *TransportMessage) error {
	if msg == nil {
		return ErrInvalidMessage
	}
	
	// Create signature data from message content
	signatureData := fmt.Sprintf("%s|%s|%d|%s|%s",
		msg.ID, msg.Type, msg.Timestamp, msg.Sender, string(msg.Payload))
	
	// Simple hash-based signature (in production, use proper cryptographic signing)
	hash := sha256.Sum256([]byte(signatureData))
	msg.Signature = hash[:]
	
	ms.logger.Debug("Message signed successfully",
		zap.String("message_id", msg.ID),
		zap.Int("signature_size", len(msg.Signature)),
	)
	
	return nil
}

// VerifySignature verifies message signature
func (ms *messageSerializer) VerifySignature(msg *TransportMessage) error {
	if msg == nil {
		return ErrInvalidMessage
	}
	
	if len(msg.Signature) == 0 {
		return &TransportError{
			Code:    "MISSING_SIGNATURE",
			Message: "Message signature is missing",
			Details: fmt.Sprintf("message_id: %s", msg.ID),
		}
	}
	
	// Recreate signature data
	signatureData := fmt.Sprintf("%s|%s|%d|%s|%s",
		msg.ID, msg.Type, msg.Timestamp, msg.Sender, string(msg.Payload))
	
	// Calculate expected signature
	expectedHash := sha256.Sum256([]byte(signatureData))
	
	// Compare signatures
	if len(msg.Signature) != len(expectedHash) {
		return ErrSignatureVerificationFailed
	}
	
	for i, b := range expectedHash {
		if msg.Signature[i] != b {
			ms.logger.Error("Signature verification failed",
				zap.String("message_id", msg.ID),
				zap.String("sender", msg.Sender),
			)
			return ErrSignatureVerificationFailed
		}
	}
	
	ms.logger.Debug("Signature verification passed",
		zap.String("message_id", msg.ID),
		zap.String("sender", msg.Sender),
	)
	
	return nil
}

// SerializeWithCompression serializes message with optional compression
func (ms *messageSerializer) SerializeWithCompression(msg *TransportMessage, compress bool) ([]byte, error) {
	// For now, just use regular serialization
	// TODO: Add compression support (gzip) when needed
	data, err := ms.Serialize(msg)
	if err != nil {
		return nil, err
	}
	
	if compress && len(data) > 1024 { // Only compress larger messages
		ms.logger.Debug("Compression requested but not implemented yet",
			zap.String("message_id", msg.ID),
			zap.Int("original_size", len(data)),
		)
		// TODO: Implement gzip compression
	}
	
	return data, nil
}

// CreateSignedMessage creates a signed transport message
func CreateSignedMessage(msgType string, payload []byte, sender string, serializer MessageSerializer) (*TransportMessage, error) {
	msg := &TransportMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: GenerateTimestamp(),
		Sender:    sender,
		Priority:  PriorityNormal,
		TTL:       int64(DefaultMessageTTL().Milliseconds()),
		Metadata:  make(map[string]string),
	}
	
	// Generate message ID
	msg.ID = GenerateMessageID(msg)
	
	// Sign the message
	if err := serializer.SignMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}
	
	return msg, nil
}

// Helper functions for message creation
func GenerateTimestamp() int64 {
	return GetCurrentTimestamp()
}

func DefaultMessageTTL() time.Duration {
	return 5 * time.Minute
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}