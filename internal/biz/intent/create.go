package intent

import (
	"context"
	"fmt"
	"time"

	pb "pin_intent_broadcast_network/api/pinai_intent/v1"
	"pin_intent_broadcast_network/internal/biz/common"
)

// CreateIntent implements the intent creation logic for API layer
func CreateIntent(ctx context.Context, req *pb.CreateIntentRequest) (*pb.CreateIntentResponse, error) {
	startTime := time.Now()

	// Validate create request
	if err := validateCreateRequest(req); err != nil {
		return &pb.CreateIntentResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// 1. Generate Intent ID and basic information
	intent := &pb.Intent{
		Id:        common.GenerateIntentID(),
		Type:      req.Type,
		Payload:   req.Payload,
		Timestamp: time.Now().Unix(),
		SenderId:  req.SenderId,
		Metadata:  req.Metadata,
		Status:    pb.IntentStatus_INTENT_STATUS_CREATED,
		Priority:  req.Priority,
		Ttl:       req.Ttl,
	}

	// Set default values if not provided
	if intent.Priority == 0 {
		intent.Priority = common.PriorityNormal
	}
	if intent.Ttl == 0 {
		intent.Ttl = int64(common.DefaultTTL.Seconds())
	}

	// TODO: Add validation, signing, and lifecycle management
	// For now, return the created intent
	_ = startTime // Use startTime to avoid unused variable error

	return &pb.CreateIntentResponse{
		Intent:  intent,
		Success: true,
		Message: "Intent created successfully",
	}, nil
}

// validateCreateRequest validates the create intent request
func validateCreateRequest(req *pb.CreateIntentRequest) error {
	if req == nil {
		return common.NewIntentError(common.ErrorCodeInvalidFormat, "Request cannot be nil", "")
	}

	// Validate required fields
	if common.Strings.IsEmpty(req.Type) {
		return common.NewValidationError("type", req.Type, "Intent type cannot be empty")
	}

	// mark for not validate
	// if !common.Validation.IsValidIntentType(req.Type) {
	// 	return common.NewValidationError("type", req.Type, "Invalid intent type")
	// }

	if len(req.Payload) == 0 {
		return common.NewValidationError("payload", "", "Intent payload cannot be empty")
	}

	if !common.Validation.IsValidPayloadSize(len(req.Payload)) {
		return common.NewValidationError("payload", "", fmt.Sprintf("Payload size %d exceeds maximum allowed", len(req.Payload)))
	}

	if common.Strings.IsEmpty(req.SenderId) {
		return common.NewValidationError("sender_id", req.SenderId, "Sender ID cannot be empty")
	}

	if req.Priority != 0 && !common.Validation.IsValidPriority(req.Priority) {
		return common.NewValidationError("priority", fmt.Sprintf("%d", req.Priority), "Invalid priority value")
	}

	if req.Ttl != 0 && !common.Validation.IsValidTTL(req.Ttl) {
		return common.NewValidationError("ttl", fmt.Sprintf("%d", req.Ttl), "Invalid TTL value")
	}

	return nil
}
