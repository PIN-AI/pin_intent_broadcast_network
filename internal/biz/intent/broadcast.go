package intent

import (
	"context"
	pb "pin_intent_broadcast_network/api/pinai_intent/v1"
	"pin_intent_broadcast_network/internal/biz/common"
)

// intentManager is a global reference to the intent manager
// This will be set during application startup
var intentManager common.IntentManager

// SetIntentManager sets the global intent manager reference
func SetIntentManager(manager common.IntentManager) {
	intentManager = manager
}

// BroadcastIntent implements the intent broadcasting logic for API layer
func BroadcastIntent(ctx context.Context, req *pb.BroadcastIntentRequest) (*pb.BroadcastIntentResponse, error) {
	if req == nil || req.IntentId == "" {
		return &pb.BroadcastIntentResponse{
			Success: false,
			Message: "Invalid broadcast request",
		}, common.NewIntentError(common.ErrorCodeInvalidFormat, "Invalid broadcast request", "Request or intent ID is empty")
	}

	// Check if intent manager is available
	if intentManager == nil {
		// Fallback to simulation mode
		topic := req.Topic
		if topic == "" {
			topic = "intent-broadcast.general"
		}

		return &pb.BroadcastIntentResponse{
			Success:  true,
			IntentId: req.IntentId,
			Topic:    topic,
			Message:  "Intent broadcast initiated successfully (simulation mode)",
		}, nil
	}

	// Get the intent first
	intent, err := intentManager.GetIntentStatus(ctx, req.IntentId)
	if err != nil {
		return &pb.BroadcastIntentResponse{
			Success:  false,
			IntentId: req.IntentId,
			Message:  "Intent not found: " + err.Error(),
		}, err
	}

	// Determine topic if not provided
	topic := req.Topic
	if topic == "" {
		topic = determineTopicByType(intent.Type)
	}

	// Create broadcast request
	broadcastReq := &common.BroadcastIntentRequest{
		Intent: intent,
		Topic:  topic,
	}

	// Call the intent manager's broadcast method
	response, err := intentManager.BroadcastIntent(ctx, broadcastReq)
	if err != nil {
		return &pb.BroadcastIntentResponse{
			Success:  false,
			IntentId: req.IntentId,
			Topic:    topic,
			Message:  "Broadcast failed: " + err.Error(),
		}, err
	}

	return &pb.BroadcastIntentResponse{
		Success:  response.Success,
		IntentId: response.IntentID,
		Topic:    response.Topic,
		Message:  response.Message,
	}, nil
}

// determineTopicByType determines the appropriate topic for an intent type
func determineTopicByType(intentType string) string {
	// Map intent types to broadcast topics
	topicMap := map[string]string{
		common.IntentTypeTrade:      common.TopicIntentBroadcast + ".trade",
		common.IntentTypeSwap:       common.TopicIntentBroadcast + ".swap",
		common.IntentTypeExchange:   common.TopicIntentBroadcast + ".exchange",
		common.IntentTypeTransfer:   common.TopicIntentBroadcast + ".transfer",
		common.IntentTypeSend:       common.TopicIntentBroadcast + ".send",
		common.IntentTypePayment:    common.TopicIntentBroadcast + ".payment",
		common.IntentTypeLending:    common.TopicIntentBroadcast + ".lending",
		common.IntentTypeBorrow:     common.TopicIntentBroadcast + ".borrow",
		common.IntentTypeLoan:       common.TopicIntentBroadcast + ".loan",
		common.IntentTypeInvestment: common.TopicIntentBroadcast + ".investment",
		common.IntentTypeStaking:    common.TopicIntentBroadcast + ".staking",
		common.IntentTypeYield:      common.TopicIntentBroadcast + ".yield",
	}

	if topic, exists := topicMap[intentType]; exists {
		return topic
	}

	// Default topic for unknown types
	return common.TopicIntentBroadcast + ".general"
}
