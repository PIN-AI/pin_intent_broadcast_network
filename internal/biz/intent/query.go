package intent

import (
	"context"
	pb "pin_intent_broadcast_network/api/pinai_intent/v1"
	"pin_intent_broadcast_network/internal/biz/common"
	"time"
)

// QueryIntents implements the intent querying logic for API layer
func QueryIntents(ctx context.Context, req *pb.QueryIntentsRequest) (*pb.QueryIntentsResponse, error) {
	if req == nil {
		return &pb.QueryIntentsResponse{
			Intents: []*pb.Intent{},
			Total:   0,
		}, common.NewIntentError(common.ErrorCodeInvalidFormat, "Query request cannot be nil", "")
	}

	// TODO: Add actual query logic
	// For now, return mock data

	mockIntents := []*pb.Intent{
		{
			Id:        "mock-intent-1",
			Type:      req.Type,
			Status:    req.Status,
			SenderId:  "mock-sender",
			Timestamp: time.Now().Unix(),
		},
	}

	return &pb.QueryIntentsResponse{
		Intents: mockIntents,
		Total:   int32(len(mockIntents)),
	}, nil
}

// QueryFilter represents query filter criteria
type QueryFilter struct {
	Type      string
	StartTime int64
	EndTime   int64
	Limit     int32
	Status    common.IntentStatus
	SenderID  string
}

// buildQueryFilter builds a query filter from request
func buildQueryFilter(req *common.QueryIntentsRequest) *QueryFilter {
	filter := &QueryFilter{
		Type:      req.Type,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Limit:     req.Limit,
	}

	// Set default end time if not provided
	if filter.EndTime == 0 {
		filter.EndTime = time.Now().Unix()
	}

	return filter
}

// applyFilters applies filters to intent list
func applyFilters(intents []*common.Intent, filter *QueryFilter) []*common.Intent {
	if filter == nil {
		return intents
	}

	filtered := make([]*common.Intent, 0)

	for _, intent := range intents {
		// Filter by type
		if filter.Type != "" && intent.Type != filter.Type {
			continue
		}

		// Filter by time range
		if filter.StartTime > 0 && intent.Timestamp < filter.StartTime {
			continue
		}

		if filter.EndTime > 0 && intent.Timestamp > filter.EndTime {
			continue
		}

		// Filter by status
		if filter.Status != 0 && intent.Status != filter.Status {
			continue
		}

		// Filter by sender ID
		if filter.SenderID != "" && intent.SenderID != filter.SenderID {
			continue
		}

		// Skip expired intents unless specifically requested
		if common.Times.IsExpired(intent.Timestamp, intent.TTL) && filter.Status != common.IntentStatusExpired {
			continue
		}

		filtered = append(filtered, intent)
	}

	// Sort by timestamp (newest first)
	sortIntentsByTimestamp(filtered)

	return filtered
}

// sortIntentsByTimestamp sorts intents by timestamp in descending order
func sortIntentsByTimestamp(intents []*common.Intent) {
	// Simple bubble sort for now - could be optimized with sort.Slice
	for i := 0; i < len(intents)-1; i++ {
		for j := 0; j < len(intents)-i-1; j++ {
			if intents[j].Timestamp < intents[j+1].Timestamp {
				intents[j], intents[j+1] = intents[j+1], intents[j]
			}
		}
	}
}
