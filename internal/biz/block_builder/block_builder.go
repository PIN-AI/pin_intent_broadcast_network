package block_builder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pin_intent_broadcast_network/internal/biz/common"
	"pin_intent_broadcast_network/internal/transport"

	"go.uber.org/zap"
)

type BlockBuilder struct {
	config         *BuilderConfig
	transportMgr   transport.TransportManager
	matchingEngine *MatchingEngine
	logger         *zap.Logger
	isRunning      bool

	// 状态管理
	mu               sync.RWMutex
	activeIntents    map[string]*IntentSession
	completedMatches map[string]*transport.MatchResult
	status           *BlockBuilderStatus
	metrics          *BlockBuilderMetrics
}

func NewBlockBuilder(config *BuilderConfig, transportMgr transport.TransportManager, logger *zap.Logger) *BlockBuilder {
	if config == nil {
		config = DefaultBuilderConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &BlockBuilder{
		config:           config,
		transportMgr:     transportMgr,
		matchingEngine:   NewMatchingEngine(config, logger),
		logger:           logger.Named("block_builder"),
		activeIntents:    make(map[string]*IntentSession),
		completedMatches: make(map[string]*transport.MatchResult),
		status: &BlockBuilderStatus{
			BuilderID:         config.BuilderID,
			Status:            BuilderStatusOffline,
			ActiveSessions:    0,
			CompletedMatches:  0,
			TotalBidsReceived: 0,
			LastActivity:      time.Now(),
			ConnectedPeers:    0,
		},
		metrics: &BlockBuilderMetrics{
			SessionsCreated:     0,
			SessionsCompleted:   0,
			SessionsExpired:     0,
			BidsReceived:        0,
			MatchesCompleted:    0,
			AverageSessionTime:  0,
			AverageResponseTime: 0,
			LastUpdated:         time.Now(),
		},
	}
}

func (bb *BlockBuilder) Start(ctx context.Context) error {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	if bb.isRunning {
		return fmt.Errorf("block builder already running")
	}

	// 订阅意图广播 - 订阅所有Intent类型的广播topics
	intentTopics := []string{
		"intent.broadcast",
		"intent.broadcast.trade",
		"intent.broadcast.swap",
		"intent.broadcast.exchange",
		"intent.broadcast.transfer",
		"intent.broadcast.send",
		"intent.broadcast.payment",
		"intent.broadcast.lending",
		"intent.broadcast.borrow",
		"intent.broadcast.loan",
		"intent.broadcast.investment",
		"intent.broadcast.staking",
		"intent.broadcast.yield",
		"intent.broadcast.general",
		"intent.broadcast.matching",
		"intent.broadcast.notification",
		"intent.broadcast.status",
	}

	for _, topic := range intentTopics {
		_, err := bb.transportMgr.SubscribeToTopic(topic, bb.handleIntentBroadcast)
		if err != nil {
			bb.logger.Warn("Failed to subscribe to intent topic", zap.String("topic", topic), zap.Error(err))
		} else {
			bb.logger.Info("Subscribed to intent topic", zap.String("topic", topic))
		}
	}

	// 订阅出价提交
	var err error
	_, err = bb.transportMgr.SubscribeToBids(bb.handleBidSubmission)
	if err != nil {
		return fmt.Errorf("failed to subscribe to bid submissions: %w", err)
	}

	// 启动匹配处理协程
	go bb.processMatching(ctx)
	go bb.cleanupExpiredSessions(ctx)
	go bb.metricsUpdater(ctx)

	bb.isRunning = true
	bb.status.Status = BuilderStatusActive
	bb.status.LastActivity = time.Now()

	bb.logger.Info("Block builder started",
		zap.String("builder_id", bb.config.BuilderID),
		zap.String("matching_algorithm", bb.config.MatchingAlgorithm),
	)

	return nil
}

func (bb *BlockBuilder) Stop() error {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	if !bb.isRunning {
		return fmt.Errorf("block builder not running")
	}

	bb.isRunning = false
	bb.status.Status = BuilderStatusOffline
	bb.status.LastActivity = time.Now()

	bb.logger.Info("Block builder stopped")
	return nil
}

func (bb *BlockBuilder) IsRunning() bool {
	bb.mu.RLock()
	defer bb.mu.RUnlock()
	return bb.isRunning
}

func (bb *BlockBuilder) GetStatus() *BlockBuilderStatus {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	// Update connected peers count
	if bb.transportMgr != nil {
		metrics := bb.transportMgr.GetTransportMetrics()
		bb.status.ConnectedPeers = metrics.ConnectedPeerCount
	}

	// Create a copy to avoid concurrent access issues
	status := *bb.status
	return &status
}

func (bb *BlockBuilder) GetMetrics() *BlockBuilderMetrics {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	metrics := *bb.metrics
	return &metrics
}

func (bb *BlockBuilder) handleIntentBroadcast(msg *transport.TransportMessage) error {
	if msg.Type != transport.MessageTypeIntentBroadcast {
		return nil
	}

	// 反序列化意图
	intent, err := bb.deserializeIntent(msg.Payload)
	if err != nil {
		bb.logger.Error("Failed to deserialize intent",
			zap.Error(err),
		)
		return err
	}

	bb.logger.Info("Received intent broadcast",
		zap.String("intent_id", intent.ID),
		zap.String("intent_type", intent.Type),
		zap.String("agent_id", intent.SenderID),
	)

	// 创建意图会话 - 使用原子操作避免重复创建
	bb.mu.Lock()
	defer bb.mu.Unlock()

	if existingSession, exists := bb.activeIntents[intent.ID]; exists {
		bb.logger.Info("Intent session already exists, skipping creation",
			zap.String("intent_id", intent.ID),
			zap.String("existing_status", existingSession.Status),
			zap.Int("existing_bids", len(existingSession.Bids)),
		)
		return nil
	}

	session := &IntentSession{
		Intent:    intent,
		Bids:      make([]*transport.BidMessage, 0),
		StartTime: time.Now(),
		EndTime:   time.Now().Add(bb.config.BidCollectionWindow),
		Status:    SessionStateCollecting,
	}

	bb.activeIntents[intent.ID] = session
	bb.metrics.SessionsCreated++
	bb.status.ActiveSessions = len(bb.activeIntents)
	bb.status.LastActivity = time.Now()

	bb.logger.Info("Intent session created successfully",
		zap.String("intent_id", intent.ID),
		zap.Duration("collection_window", bb.config.BidCollectionWindow),
		zap.Time("end_time", session.EndTime),
	)

	return nil
}

func (bb *BlockBuilder) handleBidSubmission(bid *transport.BidMessage) error {
	bb.logger.Info("Received bid submission",
		zap.String("intent_id", bid.IntentID),
		zap.String("agent_id", bid.AgentID),
		zap.String("bid_amount", bid.BidAmount),
	)

	// 添加出价到对应的意图会话
	bb.mu.Lock()
	defer bb.mu.Unlock()

	session, exists := bb.activeIntents[bid.IntentID]
	if !exists {
		bb.logger.Warn("Received bid for unknown intent",
			zap.String("intent_id", bid.IntentID),
			zap.String("agent_id", bid.AgentID),
		)
		return nil
	}

	if session.Status != SessionStateCollecting {
		bb.logger.Warn("Received bid for non-collecting session",
			zap.String("intent_id", bid.IntentID),
			zap.String("status", session.Status),
		)
		return nil
	}

	// 检查重复出价
	for _, existingBid := range session.Bids {
		if existingBid.AgentID == bid.AgentID {
			bb.logger.Debug("Duplicate bid from agent, updating",
				zap.String("intent_id", bid.IntentID),
				zap.String("agent_id", bid.AgentID),
			)
			existingBid.BidAmount = bid.BidAmount
			existingBid.Timestamp = bid.Timestamp
			return nil
		}
	}

	// 添加新出价
	session.Bids = append(session.Bids, bid)
	bb.metrics.BidsReceived++
	bb.status.TotalBidsReceived++
	bb.status.LastActivity = time.Now()

	bb.logger.Debug("Bid added to session",
		zap.String("intent_id", bid.IntentID),
		zap.String("agent_id", bid.AgentID),
		zap.Int("total_bids", len(session.Bids)),
		zap.Int("min_required", bb.config.MinBidsRequired),
	)

	// 检查是否达到最小出价数量，如果是则触发立即匹配检查
	if len(session.Bids) >= bb.config.MinBidsRequired {
		bb.logger.Info("Minimum bids reached, triggering immediate matching",
			zap.String("intent_id", bid.IntentID),
			zap.Int("current_bids", len(session.Bids)),
		)
		// 异步触发匹配检查，避免阻塞当前处理
		go bb.checkSessionForMatching(session)
	}

	return nil
}

func (bb *BlockBuilder) processMatching(ctx context.Context) {
	// 减少检查间隔到2秒以提高响应速度
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	bb.logger.Info("Matching processor started", zap.Duration("check_interval", 2*time.Second))

	for {
		select {
		case <-ctx.Done():
			bb.logger.Info("Matching processor stopped")
			return
		case <-ticker.C:
			bb.processReadySessions()
		}
	}
}

func (bb *BlockBuilder) processReadySessions() {
	bb.mu.Lock()
	readySessions := make([]*IntentSession, 0)

	now := time.Now()
	for intentID, session := range bb.activeIntents {
		if session.Status == SessionStateCollecting && now.After(session.EndTime) {
			session.Status = SessionStateMatching
			readySessions = append(readySessions, session)
			bb.logger.Info("Session ready for matching (collection window expired)",
				zap.String("intent_id", intentID),
				zap.Int("bid_count", len(session.Bids)),
				zap.Time("end_time", session.EndTime),
			)
		}
	}
	bb.mu.Unlock()

	// 处理准备好的会话
	for _, session := range readySessions {
		bb.processSessionMatching(session)
	}
}

// checkSessionForMatching 检查单个会话是否可以立即进行匹配
func (bb *BlockBuilder) checkSessionForMatching(session *IntentSession) {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	// 从活跃会话中重新获取会话，确保状态一致性
	activeSession, exists := bb.activeIntents[session.Intent.ID]
	if !exists {
		bb.logger.Debug("Session no longer exists, skipping immediate match check",
			zap.String("intent_id", session.Intent.ID),
		)
		return
	}

	// 再次检查会话状态，确保仍在收集阶段
	if activeSession.Status != SessionStateCollecting {
		bb.logger.Debug("Session no longer collecting, skipping immediate match check",
			zap.String("intent_id", activeSession.Intent.ID),
			zap.String("status", activeSession.Status),
		)
		return
	}

	// 检查是否达到最小出价要求
	if len(activeSession.Bids) >= bb.config.MinBidsRequired {
		bb.logger.Info("Triggering immediate matching for session",
			zap.String("intent_id", activeSession.Intent.ID),
			zap.Int("bid_count", len(activeSession.Bids)),
		)
		
		// 原子性地更新状态，防止重复处理
		activeSession.Status = SessionStateMatching
		// 异步处理匹配，避免持有锁太久
		go bb.processSessionMatching(activeSession)
	}
}

func (bb *BlockBuilder) processSessionMatching(session *IntentSession) {
	// 防护检查：确保会话状态正确
	bb.mu.RLock()
	if session.Status != SessionStateMatching {
		bb.logger.Warn("Session not in matching state, skipping processing",
			zap.String("intent_id", session.Intent.ID),
			zap.String("current_status", session.Status),
		)
		bb.mu.RUnlock()
		return
	}
	bb.mu.RUnlock()

	bb.logger.Info("Processing session matching",
		zap.String("intent_id", session.Intent.ID),
		zap.String("intent_type", session.Intent.Type),
		zap.Int("bid_count", len(session.Bids)),
		zap.String("algorithm", bb.config.MatchingAlgorithm),
		zap.Duration("session_duration", time.Since(session.StartTime)),
	)

	// 记录所有出价详情（仅在 Debug 级别）
	if bb.logger.Core().Enabled(zap.DebugLevel) {
		for i, bid := range session.Bids {
			bb.logger.Debug("Bid details",
				zap.String("intent_id", session.Intent.ID),
				zap.Int("bid_index", i),
				zap.String("agent_id", bid.AgentID),
				zap.String("agent_type", bid.AgentType),
				zap.String("bid_amount", bid.BidAmount),
			)
		}
	}

	// 检查最小出价数量
	if len(session.Bids) < bb.config.MinBidsRequired {
		bb.logger.Warn("Insufficient bids for matching",
			zap.String("intent_id", session.Intent.ID),
			zap.Int("bid_count", len(session.Bids)),
			zap.Int("min_required", bb.config.MinBidsRequired),
		)

		session.Status = SessionStateExpired
		session.MatchResult = &transport.MatchResult{
			IntentID:       session.Intent.ID,
			Status:         MatchStatusNoMatch,
			TotalBids:      len(session.Bids),
			MatchedAt:      time.Now().UnixMilli(),
			BlockBuilderID: bb.config.BuilderID,
			Metadata: map[string]string{
				"reason": "insufficient_bids",
			},
		}
		bb.mu.Lock()
		bb.metrics.SessionsExpired++
		bb.mu.Unlock()
		
		bb.logger.Info("Session marked as expired due to insufficient bids",
			zap.String("intent_id", session.Intent.ID),
		)
		return
	}

	// 执行匹配算法
	bb.logger.Info("Starting matching algorithm execution",
		zap.String("intent_id", session.Intent.ID),
		zap.String("algorithm", bb.config.MatchingAlgorithm),
	)
	
	matchResult, err := bb.matchingEngine.FindBestMatch(session.Intent, session.Bids)
	if err != nil {
		bb.logger.Error("Matching algorithm failed",
			zap.String("intent_id", session.Intent.ID),
			zap.String("algorithm", bb.config.MatchingAlgorithm),
			zap.Error(err),
		)

		session.Status = SessionStateExpired
		session.MatchResult = &transport.MatchResult{
			IntentID:       session.Intent.ID,
			Status:         MatchStatusMatchFailed,
			TotalBids:      len(session.Bids),
			MatchedAt:      time.Now().UnixMilli(),
			BlockBuilderID: bb.config.BuilderID,
			Metadata: map[string]string{
				"error": err.Error(),
				"algorithm": bb.config.MatchingAlgorithm,
			},
		}
		bb.mu.Lock()
		bb.metrics.SessionsExpired++
		bb.mu.Unlock()
		
		bb.logger.Info("Session marked as expired due to matching failure",
			zap.String("intent_id", session.Intent.ID),
		)
		return
	}
	
	bb.logger.Info("Matching algorithm completed successfully",
		zap.String("intent_id", session.Intent.ID),
		zap.String("winning_agent", matchResult.WinningAgent),
		zap.String("winning_bid", matchResult.WinningBid),
	)

	// 更新会话状态
	session.Status = SessionStateCompleted
	session.MatchResult = matchResult

	// 存储完成的匹配
	bb.mu.Lock()
	bb.completedMatches[session.Intent.ID] = matchResult
	delete(bb.activeIntents, session.Intent.ID)
	bb.status.ActiveSessions = len(bb.activeIntents)
	bb.status.CompletedMatches++
	bb.metrics.SessionsCompleted++
	bb.metrics.MatchesCompleted++
	bb.mu.Unlock()

	// 广播匹配结果
	bb.broadcastMatchResult(matchResult)

	bb.logger.Info("Matching completed",
		zap.String("intent_id", session.Intent.ID),
		zap.String("winning_agent", matchResult.WinningAgent),
		zap.String("winning_bid", matchResult.WinningBid),
	)
}

func (bb *BlockBuilder) broadcastMatchResult(result *transport.MatchResult) {
	// 广播匹配结果
	ctx := context.Background()
	err := bb.transportMgr.PublishMatchResult(ctx, result)
	if err != nil {
		bb.logger.Error("Failed to broadcast match result",
			zap.String("intent_id", result.IntentID),
			zap.Error(err),
		)
	} else {
		bb.logger.Info("Match result broadcasted",
			zap.String("intent_id", result.IntentID),
		)
	}
}

func (bb *BlockBuilder) cleanupExpiredSessions(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bb.cleanupExpired()
		}
	}
}

func (bb *BlockBuilder) cleanupExpired() {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for intentID, session := range bb.activeIntents {
		// 清理超时的会话
		if session.Status == SessionStateCollecting && now.After(session.EndTime.Add(5*time.Minute)) {
			delete(bb.activeIntents, intentID)
			expiredCount++
		}
	}

	bb.status.ActiveSessions = len(bb.activeIntents)

	if expiredCount > 0 {
		bb.logger.Info("Cleaned up expired sessions",
			zap.Int("expired_count", expiredCount),
		)
	}
}

func (bb *BlockBuilder) metricsUpdater(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bb.updateMetrics()
		}
	}
}

func (bb *BlockBuilder) updateMetrics() {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	bb.metrics.LastUpdated = time.Now()

	// Calculate average session time
	if bb.metrics.SessionsCompleted > 0 {
		// This would need to be tracked during session completion
		// For now, use a placeholder calculation
	}

	// Update status based on activity
	if len(bb.activeIntents) >= bb.config.MaxConcurrentIntents {
		bb.status.Status = BuilderStatusBusy
	} else if bb.isRunning {
		bb.status.Status = BuilderStatusActive
	} else {
		bb.status.Status = BuilderStatusOffline
	}

	bb.logger.Debug("Metrics updated",
		zap.String("builder_id", bb.config.BuilderID),
		zap.Int64("sessions_completed", bb.metrics.SessionsCompleted),
		zap.Int64("matches_completed", bb.metrics.MatchesCompleted),
	)
}

func (bb *BlockBuilder) deserializeIntent(payload []byte) (*common.Intent, error) {
	var intent common.Intent
	if err := common.JSON.Unmarshal(payload, &intent); err != nil {
		return nil, fmt.Errorf("failed to deserialize intent: %w", err)
	}
	return &intent, nil
}

// GetActiveIntents 获取活跃意图会话
func (bb *BlockBuilder) GetActiveIntents() map[string]*IntentSession {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	result := make(map[string]*IntentSession)
	for k, v := range bb.activeIntents {
		result[k] = v
	}
	return result
}

// GetCompletedMatches 获取完成的匹配
func (bb *BlockBuilder) GetCompletedMatches() map[string]*transport.MatchResult {
	bb.mu.RLock()
	defer bb.mu.RUnlock()

	result := make(map[string]*transport.MatchResult)
	for k, v := range bb.completedMatches {
		result[k] = v
	}
	return result
}
