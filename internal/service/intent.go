package service

import (
	"context"

	pb "pin_intent_broadcast_network/api/pinai_intent/v1"
	"pin_intent_broadcast_network/internal/biz/intent"

	"github.com/go-kratos/kratos/v2/log"
)

// IntentService is the intent service implementation
type IntentService struct {
	pb.UnimplementedIntentServiceServer

	log *log.Helper
}

// NewIntentService creates a new intent service
func NewIntentService(logger log.Logger) *IntentService {
	return &IntentService{
		log: log.NewHelper(logger),
	}
}

// CreateIntent creates a new intent
func (s *IntentService) CreateIntent(ctx context.Context, req *pb.CreateIntentRequest) (*pb.CreateIntentResponse, error) {
	s.log.WithContext(ctx).Infof("CreateIntent called with type: %s", req.Type)

	// Forward to business logic layer
	return intent.CreateIntent(ctx, req)
}

// BroadcastIntent broadcasts an intent
func (s *IntentService) BroadcastIntent(ctx context.Context, req *pb.BroadcastIntentRequest) (*pb.BroadcastIntentResponse, error) {
	s.log.WithContext(ctx).Infof("BroadcastIntent called for intent: %s", req.IntentId)

	// Forward to business logic layer
	return intent.BroadcastIntent(ctx, req)
}

// QueryIntents queries intents with filters
func (s *IntentService) QueryIntents(ctx context.Context, req *pb.QueryIntentsRequest) (*pb.QueryIntentsResponse, error) {
	s.log.WithContext(ctx).Infof("QueryIntents called with type: %s", req.Type)

	// Forward to business logic layer
	return intent.QueryIntents(ctx, req)
}

// GetIntentStatus gets intent status by ID
func (s *IntentService) GetIntentStatus(ctx context.Context, req *pb.GetIntentStatusRequest) (*pb.GetIntentStatusResponse, error) {
	s.log.WithContext(ctx).Infof("GetIntentStatus called for intent: %s", req.IntentId)

	// Forward to business logic layer
	return intent.GetIntentStatus(ctx, req)
}
