package service

import (
	"context"

	pb "pin_intent_broadcast_network/api/pinai_intent/v1"
	"pin_intent_broadcast_network/internal/biz/execution"
	"pin_intent_broadcast_network/internal/transport"

	"go.uber.org/zap"
)

// ExecutionService implements the IntentExecutionService
type ExecutionService struct {
	pb.UnimplementedIntentExecutionServiceServer
	
	automationMgr *execution.AutomationManager
	logger        *zap.Logger
}

// NewExecutionService creates a new execution service
func NewExecutionService(automationMgr *execution.AutomationManager, logger *zap.Logger) *ExecutionService {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ExecutionService{
		automationMgr: automationMgr,
		logger:        logger.Named("execution_service"),
	}
}

// GetAgentsStatus returns the status of all service agents
func (s *ExecutionService) GetAgentsStatus(ctx context.Context, req *pb.GetAgentsStatusRequest) (*pb.GetAgentsStatusResponse, error) {
	s.logger.Debug("GetAgentsStatus called")

	agents := s.automationMgr.GetAgents()
	agentStatuses := make([]*pb.AgentStatus, 0, len(agents))

	for _, agent := range agents {
		status := agent.GetStatus()
		config := agent.GetConfig()

		pbStatus := &pb.AgentStatus{
			AgentId:          status.AgentID,
			AgentType:        string(config.AgentType),
			Status:           status.Status,
			ActiveIntents:    int32(status.ActiveIntents),
			ProcessedIntents: status.ProcessedIntents,
			SuccessfulBids:   status.SuccessfulBids,
			TotalEarnings:    status.TotalEarnings,
			LastActivity:     status.LastActivity.Unix(),
			ConnectedPeers:   int32(status.ConnectedPeers),
		}

		agentStatuses = append(agentStatuses, pbStatus)
	}

	response := &pb.GetAgentsStatusResponse{
		Agents:      agentStatuses,
		TotalAgents: int32(len(agents)),
		Success:     true,
		Message:     "Successfully retrieved agents status",
	}

	s.logger.Info("GetAgentsStatus completed",
		zap.Int("total_agents", len(agents)),
	)

	return response, nil
}

// GetBuildersStatus returns the status of all block builders
func (s *ExecutionService) GetBuildersStatus(ctx context.Context, req *pb.GetBuildersStatusRequest) (*pb.GetBuildersStatusResponse, error) {
	s.logger.Debug("GetBuildersStatus called")

	builders := s.automationMgr.GetBuilders()
	builderStatuses := make([]*pb.BlockBuilderStatus, 0, len(builders))

	for _, builder := range builders {
		status := builder.GetStatus()

		pbStatus := &pb.BlockBuilderStatus{
			BuilderId:         status.BuilderID,
			Status:            status.Status,
			ActiveSessions:    int32(status.ActiveSessions),
			CompletedMatches:  status.CompletedMatches,
			TotalBidsReceived: status.TotalBidsReceived,
			LastActivity:      status.LastActivity.Unix(),
			ConnectedPeers:    int32(status.ConnectedPeers),
		}

		builderStatuses = append(builderStatuses, pbStatus)
	}

	response := &pb.GetBuildersStatusResponse{
		Builders:      builderStatuses,
		TotalBuilders: int32(len(builders)),
		Success:       true,
		Message:       "Successfully retrieved builders status",
	}

	s.logger.Info("GetBuildersStatus completed",
		zap.Int("total_builders", len(builders)),
	)

	return response, nil
}

// GetExecutionMetrics returns overall execution system metrics
func (s *ExecutionService) GetExecutionMetrics(ctx context.Context, req *pb.GetExecutionMetricsRequest) (*pb.GetExecutionMetricsResponse, error) {
	s.logger.Debug("GetExecutionMetrics called")

	metrics := s.automationMgr.GetMetrics()

	pbMetrics := &pb.ExecutionMetrics{
		TotalIntentsProcessed:  metrics.TotalIntentsReceived,
		TotalBidsSubmitted:     metrics.TotalBidsSubmitted,
		TotalMatchesCompleted:  metrics.TotalMatchesCompleted,
		SuccessRate:           metrics.AgentSuccessRate,
		AverageResponseTimeMs: metrics.AverageResponseTime,
		ActiveAgents:          metrics.ActiveAgents,
		ActiveBuilders:        metrics.ActiveBuilders,
		LastUpdated:           metrics.LastUpdated.Unix(),
	}

	response := &pb.GetExecutionMetricsResponse{
		Metrics: pbMetrics,
		Success: true,
		Message: "Successfully retrieved execution metrics",
	}

	s.logger.Info("GetExecutionMetrics completed")
	return response, nil
}

// StartAgent starts a specific service agent
func (s *ExecutionService) StartAgent(ctx context.Context, req *pb.StartAgentRequest) (*pb.StartAgentResponse, error) {
	s.logger.Info("StartAgent called", zap.String("agent_id", req.AgentId))

	if req.AgentId == "" {
		return &pb.StartAgentResponse{
			Success: false,
			Message: "Agent ID is required",
			AgentId: req.AgentId,
		}, nil
	}

	err := s.automationMgr.StartAgent(ctx, req.AgentId)
	if err != nil {
		s.logger.Error("Failed to start agent",
			zap.String("agent_id", req.AgentId),
			zap.Error(err),
		)
		return &pb.StartAgentResponse{
			Success: false,
			Message: err.Error(),
			AgentId: req.AgentId,
		}, nil
	}

	response := &pb.StartAgentResponse{
		Success: true,
		Message: "Agent started successfully",
		AgentId: req.AgentId,
	}

	s.logger.Info("Agent started successfully", zap.String("agent_id", req.AgentId))
	return response, nil
}

// StopAgent stops a specific service agent
func (s *ExecutionService) StopAgent(ctx context.Context, req *pb.StopAgentRequest) (*pb.StopAgentResponse, error) {
	s.logger.Info("StopAgent called", zap.String("agent_id", req.AgentId))

	if req.AgentId == "" {
		return &pb.StopAgentResponse{
			Success: false,
			Message: "Agent ID is required",
			AgentId: req.AgentId,
		}, nil
	}

	err := s.automationMgr.StopAgent(req.AgentId)
	if err != nil {
		s.logger.Error("Failed to stop agent",
			zap.String("agent_id", req.AgentId),
			zap.Error(err),
		)
		return &pb.StopAgentResponse{
			Success: false,
			Message: err.Error(),
			AgentId: req.AgentId,
		}, nil
	}

	response := &pb.StopAgentResponse{
		Success: true,
		Message: "Agent stopped successfully",
		AgentId: req.AgentId,
	}

	s.logger.Info("Agent stopped successfully", zap.String("agent_id", req.AgentId))
	return response, nil
}

// StartBuilder starts a specific block builder
func (s *ExecutionService) StartBuilder(ctx context.Context, req *pb.StartBuilderRequest) (*pb.StartBuilderResponse, error) {
	s.logger.Info("StartBuilder called", zap.String("builder_id", req.BuilderId))

	if req.BuilderId == "" {
		return &pb.StartBuilderResponse{
			Success:   false,
			Message:   "Builder ID is required",
			BuilderId: req.BuilderId,
		}, nil
	}

	err := s.automationMgr.StartBuilder(ctx, req.BuilderId)
	if err != nil {
		s.logger.Error("Failed to start builder",
			zap.String("builder_id", req.BuilderId),
			zap.Error(err),
		)
		return &pb.StartBuilderResponse{
			Success:   false,
			Message:   err.Error(),
			BuilderId: req.BuilderId,
		}, nil
	}

	response := &pb.StartBuilderResponse{
		Success:   true,
		Message:   "Builder started successfully",
		BuilderId: req.BuilderId,
	}

	s.logger.Info("Builder started successfully", zap.String("builder_id", req.BuilderId))
	return response, nil
}

// StopBuilder stops a specific block builder
func (s *ExecutionService) StopBuilder(ctx context.Context, req *pb.StopBuilderRequest) (*pb.StopBuilderResponse, error) {
	s.logger.Info("StopBuilder called", zap.String("builder_id", req.BuilderId))

	if req.BuilderId == "" {
		return &pb.StopBuilderResponse{
			Success:   false,
			Message:   "Builder ID is required",
			BuilderId: req.BuilderId,
		}, nil
	}

	err := s.automationMgr.StopBuilder(req.BuilderId)
	if err != nil {
		s.logger.Error("Failed to stop builder",
			zap.String("builder_id", req.BuilderId),
			zap.Error(err),
		)
		return &pb.StopBuilderResponse{
			Success:   false,
			Message:   err.Error(),
			BuilderId: req.BuilderId,
		}, nil
	}

	response := &pb.StopBuilderResponse{
		Success:   true,
		Message:   "Builder stopped successfully",
		BuilderId: req.BuilderId,
	}

	s.logger.Info("Builder stopped successfully", zap.String("builder_id", req.BuilderId))
	return response, nil
}

// GetActiveBids returns active bids for a specific intent
func (s *ExecutionService) GetActiveBids(ctx context.Context, req *pb.GetActiveBidsRequest) (*pb.GetActiveBidsResponse, error) {
	s.logger.Debug("GetActiveBids called", zap.String("intent_id", req.IntentId))

	if req.IntentId == "" {
		return &pb.GetActiveBidsResponse{
			Success: false,
			Message: "Intent ID is required",
		}, nil
	}

	// Get active bids from all builders
	builders := s.automationMgr.GetBuilders()
	allBids := make([]*pb.BidMessage, 0)

	for _, builder := range builders {
		activeIntents := builder.GetActiveIntents()
		if session, exists := activeIntents[req.IntentId]; exists {
			for _, bid := range session.Bids {
				pbBid := convertTransportBidToPbBid(bid)
				allBids = append(allBids, pbBid)
			}
		}
	}

	response := &pb.GetActiveBidsResponse{
		Bids:    allBids,
		Success: true,
		Message: "Successfully retrieved active bids",
	}

	s.logger.Info("GetActiveBids completed",
		zap.String("intent_id", req.IntentId),
		zap.Int("total_bids", len(allBids)),
	)

	return response, nil
}

// GetMatchHistory returns match history
func (s *ExecutionService) GetMatchHistory(ctx context.Context, req *pb.GetMatchHistoryRequest) (*pb.GetMatchHistoryResponse, error) {
	s.logger.Debug("GetMatchHistory called")

	// Get match history from all builders
	builders := s.automationMgr.GetBuilders()
	allMatches := make([]*pb.MatchResult, 0)

	for _, builder := range builders {
		completedMatches := builder.GetCompletedMatches()
		for _, match := range completedMatches {
			// Apply time filter if specified
			if req.StartTime > 0 && match.MatchedAt < req.StartTime {
				continue
			}
			if req.EndTime > 0 && match.MatchedAt > req.EndTime {
				continue
			}

			pbMatch := convertTransportMatchToPbMatch(match)
			allMatches = append(allMatches, pbMatch)
		}
	}

	// Apply limit and offset
	total := len(allMatches)
	start := int(req.Offset)
	end := start + int(req.Limit)

	if start >= total {
		allMatches = []*pb.MatchResult{}
	} else {
		if end > total {
			end = total
		}
		allMatches = allMatches[start:end]
	}

	response := &pb.GetMatchHistoryResponse{
		Matches: allMatches,
		Total:   int32(total),
		Success: true,
		Message: "Successfully retrieved match history",
	}

	s.logger.Info("GetMatchHistory completed",
		zap.Int("total_matches", total),
		zap.Int("returned_matches", len(allMatches)),
	)

	return response, nil
}

// Helper function to convert transport.BidMessage to pb.BidMessage
func convertTransportBidToPbBid(bid *transport.BidMessage) *pb.BidMessage {
	return &pb.BidMessage{
		IntentId:     bid.IntentID,
		AgentId:      bid.AgentID,
		BidAmount:    bid.BidAmount,
		Capabilities: bid.Capabilities,
		Timestamp:    bid.Timestamp,
		AgentType:    bid.AgentType,
		Metadata:     bid.Metadata,
	}
}

// Helper function to convert transport.MatchResult to pb.MatchResult
func convertTransportMatchToPbMatch(match *transport.MatchResult) *pb.MatchResult {
	return &pb.MatchResult{
		IntentId:       match.IntentID,
		WinningAgent:   match.WinningAgent,
		WinningBid:     match.WinningBid,
		TotalBids:      int32(match.TotalBids),
		MatchedAt:      match.MatchedAt,
		Status:         match.Status,
		BlockBuilderId: match.BlockBuilderID,
		Metadata:       match.Metadata,
	}
}