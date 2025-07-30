package intent

import (
	"context"
	pb "pin_intent_broadcast_network/api/pinai_intent/v1"
	"pin_intent_broadcast_network/internal/biz/common"
)

// BroadcastIntent implements the intent broadcasting logic for API layer
func BroadcastIntent(ctx context.Context, req *pb.BroadcastIntentRequest) (*pb.BroadcastIntentResponse, error) {
	if req == nil || req.IntentId == "" {
		return &pb.BroadcastIntentResponse{
			Success: false,
			Message: "Invalid broadcast request",
		}, common.NewIntentError(common.ErrorCodeInvalidFormat, "Invalid broadcast request", "Request or intent ID is empty")
	}

	// For now, simulate successful broadcast
	// TODO: Implement actual P2P broadcast functionality
	topic := req.Topic
	if topic == "" {
		// Use default topic based on intent type
		topic = "intent-broadcast.general"
	}

	return &pb.BroadcastIntentResponse{
		Success:  true,
		IntentId: req.IntentId,
		Topic:    topic,
		Message:  "Intent broadcast initiated successfully",
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
