package service_agent

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	
	"pin_intent_broadcast_network/internal/biz/common"
	"go.uber.org/zap"
)

type BidDecisionManager struct {
	config *AgentConfig
	logger *zap.Logger
}

func NewBidDecisionManager(config *AgentConfig, logger *zap.Logger) *BidDecisionManager {
	return &BidDecisionManager{
		config: config,
		logger: logger.Named("bid_decision"),
	}
}

func (bdm *BidDecisionManager) MakeBidDecision(ctx context.Context, intent *common.Intent) (*BidDecision, error) {
	bdm.logger.Debug("Making bid decision",
		zap.String("intent_id", intent.ID),
		zap.String("intent_type", intent.Type),
	)
	
	// 1. 计算基础成本
	baseCost, err := bdm.calculateBaseCost(intent)
	if err != nil {
		return &BidDecision{
			ShouldBid: false,
			Reason:    fmt.Sprintf("Failed to calculate base cost: %v", err),
		}, nil
	}
	
	// 2. 评估能力匹配度
	capabilityScore := bdm.evaluateCapabilityMatch(intent)
	
	// 对于激进策略，即使能力匹配度低也要出价（POC需求）
	if capabilityScore < 0.3 && bdm.config.BidStrategy.Type != "aggressive" {
		return &BidDecision{
			ShouldBid: false,
			Reason:    "Insufficient capability match",
			Metadata:  map[string]string{"capability_score": fmt.Sprintf("%.2f", capabilityScore)},
		}, nil
	}
	
	// 3. 计算竞争性出价
	bidAmount, confidence := bdm.calculateCompetitiveBid(baseCost, capabilityScore, intent)
	
	// 4. 应用出价策略
	finalBidAmount := bdm.applyBidStrategy(bidAmount, intent)
	
	// 5. 验证出价范围
	if !bdm.isValidBidAmount(finalBidAmount) {
		return &BidDecision{
			ShouldBid: false,
			Reason:    "Bid amount outside acceptable range",
			Metadata:  map[string]string{"calculated_bid": finalBidAmount},
		}, nil
	}
	
	return &BidDecision{
		ShouldBid:  true,
		BidAmount:  finalBidAmount,
		Confidence: confidence,
		Reason:     "Competitive bid calculated",
		Metadata: map[string]string{
			"base_cost":        baseCost,
			"capability_score": fmt.Sprintf("%.2f", capabilityScore),
			"strategy":         bdm.config.BidStrategy.Type,
		},
	}, nil
}

func (bdm *BidDecisionManager) calculateBaseCost(intent *common.Intent) (string, error) {
	totalCost := 0.0
	
	// 计算数据标签成本
	for _, tag := range intent.RelevantTags {
		if tag.IsTradable {
			tagFee, err := strconv.ParseFloat(tag.TagFee, 64)
			if err != nil {
				bdm.logger.Warn("Invalid tag fee format",
					zap.String("tag_name", tag.TagName),
					zap.String("tag_fee", tag.TagFee),
				)
				continue
			}
			totalCost += tagFee
		}
	}
	
	// 添加基础服务费
	baseFee, err := strconv.ParseFloat(bdm.config.BidStrategy.BaseFee, 64)
	if err != nil {
		baseFee = 1000.0 // 默认基础费用
	}
	totalCost += baseFee
	
	// 根据意图优先级调整
	priorityMultiplier := 1.0 + float64(intent.Priority-1)*0.1
	totalCost *= priorityMultiplier
	
	return fmt.Sprintf("%.0f", totalCost), nil
}

func (bdm *BidDecisionManager) evaluateCapabilityMatch(intent *common.Intent) float64 {
	if len(bdm.config.Capabilities) == 0 {
		return 0.5 // 默认匹配度
	}
	
	// 基于意图类型的匹配
	typeScore := 0.0
	intentType := strings.ToLower(intent.Type)
	
	for _, capability := range bdm.config.Capabilities {
		capLower := strings.ToLower(capability)
		if strings.Contains(intentType, capLower) || strings.Contains(capLower, intentType) {
			typeScore += 0.3
		}
	}
	
	// 基于专业化的匹配
	specializationScore := 0.0
	for _, spec := range bdm.config.Specializations {
		specLower := strings.ToLower(spec)
		if strings.Contains(intentType, specLower) || strings.Contains(specLower, intentType) {
			specializationScore += 0.4
		}
	}
	
	// 基于代理类型的匹配
	agentTypeScore := 0.0
	agentTypeLower := strings.ToLower(string(bdm.config.AgentType))
	if strings.Contains(intentType, agentTypeLower) || strings.Contains(agentTypeLower, intentType) {
		agentTypeScore = 0.3
	}
	
	totalScore := typeScore + specializationScore + agentTypeScore
	return math.Min(totalScore, 1.0)
}

func (bdm *BidDecisionManager) calculateCompetitiveBid(baseCost string, capabilityScore float64, intent *common.Intent) (string, float64) {
	cost, _ := strconv.ParseFloat(baseCost, 64)
	
	// 基于能力匹配度的调整
	capabilityAdjustment := 1.0 + (capabilityScore-0.5)*0.5
	
	// 基于意图紧急程度的调整
	urgencyAdjustment := 1.0
	if intent.MaxDuration > 0 && intent.MaxDuration < 3600 { // 小于1小时
		urgencyAdjustment = 1.2
	}
	
	// 计算利润边际
	profitMargin := bdm.config.BidStrategy.ProfitMargin
	if profitMargin <= 0 {
		profitMargin = 0.15 // 默认15%利润
	}
	
	// 最终出价计算
	finalBid := cost * capabilityAdjustment * urgencyAdjustment * (1 + profitMargin)
	
	// 置信度计算
	confidence := capabilityScore * 0.7 + (urgencyAdjustment-1)*0.3
	confidence = math.Max(0.1, math.Min(confidence, 1.0))
	
	return fmt.Sprintf("%.0f", finalBid), confidence
}

func (bdm *BidDecisionManager) applyBidStrategy(bidAmount string, intent *common.Intent) string {
	bid, _ := strconv.ParseFloat(bidAmount, 64)
	
	switch bdm.config.BidStrategy.Type {
	case "conservative":
		// 保守策略：降低出价以提高获胜概率
		bid *= 0.9
	case "aggressive":
		// 激进策略：提高出价以确保服务质量
		bid *= 1.1
	case "balanced":
		// 平衡策略：保持原出价
		// bid = bid
	default:
		// 默认平衡策略
	}
	
	// 应用风险因子
	riskFactor := bdm.config.BidStrategy.RiskFactor
	if riskFactor > 0 {
		bid *= (1 + riskFactor*0.1)
	}
	
	return fmt.Sprintf("%.0f", bid)
}

func (bdm *BidDecisionManager) isValidBidAmount(bidAmount string) bool {
	bid, err := strconv.ParseFloat(bidAmount, 64)
	if err != nil {
		return false
	}
	
	minBid, _ := strconv.ParseFloat(bdm.config.MinBidAmount, 64)
	maxBid, _ := strconv.ParseFloat(bdm.config.MaxBidAmount, 64)
	
	if minBid > 0 && bid < minBid {
		return false
	}
	if maxBid > 0 && bid > maxBid {
		return false
	}
	
	return true
}