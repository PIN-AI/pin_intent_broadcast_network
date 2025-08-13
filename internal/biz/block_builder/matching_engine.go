package block_builder

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"
	
	"pin_intent_broadcast_network/internal/biz/common"
	"pin_intent_broadcast_network/internal/transport"
	"go.uber.org/zap"
)

type MatchingEngine struct {
	config *BuilderConfig
	logger *zap.Logger
}

func NewMatchingEngine(config *BuilderConfig, logger *zap.Logger) *MatchingEngine {
	return &MatchingEngine{
		config: config,
		logger: logger.Named("matching_engine"),
	}
}

func (me *MatchingEngine) FindBestMatch(intent *common.Intent, bids []*transport.BidMessage) (*transport.MatchResult, error) {
	if len(bids) == 0 {
		me.logger.Error("No bids available for matching", zap.String("intent_id", intent.ID))
		return nil, fmt.Errorf("no bids available for matching")
	}
	
	me.logger.Info("Finding best match",
		zap.String("intent_id", intent.ID),
		zap.String("intent_type", intent.Type),
		zap.Int("bid_count", len(bids)),
		zap.String("algorithm", me.config.MatchingAlgorithm),
	)
	
	var winningBid *transport.BidMessage
	var err error
	var algorithmUsed string
	
	switch me.config.MatchingAlgorithm {
	case "highest_bid":
		algorithmUsed = "highest_bid"
		winningBid, err = me.findHighestBid(bids)
	case "reputation_weighted":
		algorithmUsed = "reputation_weighted"
		winningBid, err = me.findReputationWeightedBid(bids)
	case "random":
		algorithmUsed = "random"
		winningBid, err = me.findRandomBid(bids)
	default:
		me.logger.Warn("Unknown matching algorithm, falling back to highest_bid",
			zap.String("requested_algorithm", me.config.MatchingAlgorithm),
		)
		algorithmUsed = "highest_bid"
		winningBid, err = me.findHighestBid(bids) // 默认使用最高出价
	}
	
	if err != nil {
		me.logger.Error("Matching algorithm execution failed",
			zap.String("intent_id", intent.ID),
			zap.String("algorithm", algorithmUsed),
			zap.Error(err),
		)
		return nil, fmt.Errorf("matching algorithm '%s' failed: %w", algorithmUsed, err)
	}
	
	if winningBid == nil {
		me.logger.Error("Matching algorithm returned nil winning bid",
			zap.String("intent_id", intent.ID),
			zap.String("algorithm", algorithmUsed),
		)
		return nil, fmt.Errorf("matching algorithm '%s' returned no winner", algorithmUsed)
	}
	
	result := &transport.MatchResult{
		IntentID:       intent.ID,
		WinningAgent:   winningBid.AgentID,
		WinningBid:     winningBid.BidAmount,
		TotalBids:      len(bids),
		MatchedAt:      time.Now().UnixMilli(),
		Status:         MatchStatusMatched,
		BlockBuilderID: me.config.BuilderID,
		Metadata: map[string]string{
			"algorithm":   algorithmUsed,
			"agent_type":  winningBid.AgentType,
			"intent_type": intent.Type,
		},
	}
	
	me.logger.Info("Match found successfully",
		zap.String("intent_id", intent.ID),
		zap.String("algorithm", algorithmUsed),
		zap.String("winning_agent", result.WinningAgent),
		zap.String("winning_bid", result.WinningBid),
		zap.String("winning_agent_type", winningBid.AgentType),
	)
	
	return result, nil
}

func (me *MatchingEngine) findHighestBid(bids []*transport.BidMessage) (*transport.BidMessage, error) {
	var highestBid *transport.BidMessage
	var highestAmount float64 = -1
	
	for _, bid := range bids {
		amount, err := strconv.ParseFloat(bid.BidAmount, 64)
		if err != nil {
			me.logger.Warn("Invalid bid amount format",
				zap.String("agent_id", bid.AgentID),
				zap.String("bid_amount", bid.BidAmount),
			)
			continue
		}
		
		if amount > highestAmount {
			highestAmount = amount
			highestBid = bid
		}
	}
	
	if highestBid == nil {
		return nil, fmt.Errorf("no valid bids found")
	}
	
	return highestBid, nil
}

func (me *MatchingEngine) findReputationWeightedBid(bids []*transport.BidMessage) (*transport.BidMessage, error) {
	// 简化的声誉加权算法
	type WeightedBid struct {
		Bid    *transport.BidMessage
		Score  float64
		Amount float64
	}
	
	weightedBids := make([]WeightedBid, 0, len(bids))
	
	for _, bid := range bids {
		amount, err := strconv.ParseFloat(bid.BidAmount, 64)
		if err != nil {
			continue
		}
		
		// 简化的声誉计算（实际应该从声誉系统获取）
		reputation := me.calculateSimpleReputation(bid.AgentID, bid.AgentType)
		
		// 综合评分：出价金额 * 声誉权重
		score := amount * reputation
		
		weightedBids = append(weightedBids, WeightedBid{
			Bid:    bid,
			Score:  score,
			Amount: amount,
		})
	}
	
	if len(weightedBids) == 0 {
		return nil, fmt.Errorf("no valid weighted bids found")
	}
	
	// 按评分排序
	sort.Slice(weightedBids, func(i, j int) bool {
		return weightedBids[i].Score > weightedBids[j].Score
	})
	
	return weightedBids[0].Bid, nil
}

func (me *MatchingEngine) findRandomBid(bids []*transport.BidMessage) (*transport.BidMessage, error) {
	validBids := make([]*transport.BidMessage, 0, len(bids))
	
	for _, bid := range bids {
		if _, err := strconv.ParseFloat(bid.BidAmount, 64); err == nil {
			validBids = append(validBids, bid)
		}
	}
	
	if len(validBids) == 0 {
		return nil, fmt.Errorf("no valid bids found")
	}
	
	// 使用 Go 1.20+ 推荐的随机数生成方式
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rng.Intn(len(validBids))
	
	me.logger.Debug("Random bid selected",
		zap.Int("total_valid_bids", len(validBids)),
		zap.Int("selected_index", randomIndex),
		zap.String("selected_agent", validBids[randomIndex].AgentID),
	)
	
	return validBids[randomIndex], nil
}

func (me *MatchingEngine) calculateSimpleReputation(agentID, agentType string) float64 {
	// 简化的声誉计算
	// 实际实现应该查询声誉数据库或区块链
	
	baseReputation := 1.0
	
	// 基于代理类型的基础声誉
	switch agentType {
	case "trading":
		baseReputation = 1.2
	case "data_access":
		baseReputation = 1.1
	case "computation":
		baseReputation = 1.0
	default:
		baseReputation = 0.9
	}
	
	// 使用 Go 1.20+ 推荐的随机数生成方式添加变化
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	variation := (rng.Float64() - 0.5) * 0.2 // ±10%的变化
	
	reputation := baseReputation + variation
	
	// 确保声誉在合理范围内
	if reputation < 0.1 {
		reputation = 0.1
	}
	if reputation > 2.0 {
		reputation = 2.0
	}
	
	if me.logger.Core().Enabled(zap.DebugLevel) {
		me.logger.Debug("Calculated reputation",
			zap.String("agent_id", agentID),
			zap.String("agent_type", agentType),
			zap.Float64("base_reputation", baseReputation),
			zap.Float64("final_reputation", reputation),
		)
	}
	
	return reputation
}