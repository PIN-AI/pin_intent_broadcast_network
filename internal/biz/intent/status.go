package intent

import (
	"context"
	"time"

	pb "pin_intent_broadcast_network/api/pinai_intent/v1"
	"pin_intent_broadcast_network/internal/biz/common"
)

// GetIntentStatus implements the intent status retrieval logic for API layer
func GetIntentStatus(ctx context.Context, req *pb.GetIntentStatusRequest) (*pb.GetIntentStatusResponse, error) {
	if req == nil || req.IntentId == "" {
		return nil, common.NewIntentError(common.ErrorCodeInvalidFormat, "Intent ID cannot be empty", "")
	}

	// TODO: Add actual status retrieval logic
	// For now, return mock data

	mockIntent := &pb.Intent{
		Id:        req.IntentId,
		Type:      "mock-type",
		Status:    pb.IntentStatus_INTENT_STATUS_CREATED,
		SenderId:  "mock-sender",
		Timestamp: time.Now().Unix(),
	}

	return &pb.GetIntentStatusResponse{
		Intent: mockIntent,
	}, nil
}
