package service_agent

import (
	"context"
	"fmt"
	"time"
	
	"pin_intent_broadcast_network/internal/biz/common"
	"pin_intent_broadcast_network/internal/transport"
	"go.uber.org/zap"
)

type IntentListener struct {
	config          *AgentConfig
	transportMgr    transport.TransportManager
	bidDecisionMgr  *BidDecisionManager
	logger          *zap.Logger
	isRunning       bool
	intentChan      chan *IntentEvent
	processingQueue chan *IntentEvent
}

func NewIntentListener(config *AgentConfig, transportMgr transport.TransportManager, logger *zap.Logger) *IntentListener {
	return &IntentListener{
		config:          config,
		transportMgr:    transportMgr,
		bidDecisionMgr:  NewBidDecisionManager(config, logger),
		logger:          logger.Named("intent_listener"),
		intentChan:      make(chan *IntentEvent, 1000),
		processingQueue: make(chan *IntentEvent, 100),
	}
}

func (il *IntentListener) Start(ctx context.Context) error {
	if il.isRunning {
		return fmt.Errorf("intent listener already running")
	}
	
	// 订阅意图广播主题
	topics := []string{
		transport.TopicIntentBroadcast,
		"intent-broadcast.trade",
		"intent-broadcast.swap",
		"intent-broadcast.exchange",
		"intent-broadcast.general",
	}
	
	for _, topic := range topics {
		_, err := il.transportMgr.SubscribeToTopic(topic, il.handleIntentMessage)
		if err != nil {
			il.logger.Error("Failed to subscribe to topic",
				zap.String("topic", topic),
				zap.Error(err),
			)
			continue
		}
		il.logger.Info("Subscribed to intent topic",
			zap.String("topic", topic),
		)
	}
	
	// 启动处理协程
	go il.processIntents(ctx)
	go il.handleProcessingQueue(ctx)
	
	il.isRunning = true
	il.logger.Info("Intent listener started",
		zap.String("agent_id", il.config.AgentID),
		zap.String("agent_type", string(il.config.AgentType)),
	)
	
	return nil
}

func (il *IntentListener) Stop() error {
	if !il.isRunning {
		return fmt.Errorf("intent listener not running")
	}
	
	il.isRunning = false
	close(il.intentChan)
	close(il.processingQueue)
	
	il.logger.Info("Intent listener stopped")
	return nil
}

func (il *IntentListener) handleIntentMessage(msg *transport.TransportMessage) error {
	if msg.Type != transport.MessageTypeIntentBroadcast {
		return nil
	}
	
	// 反序列化意图
	intent, err := il.deserializeIntent(msg.Payload)
	if err != nil {
		il.logger.Error("Failed to deserialize intent",
			zap.Error(err),
		)
		return err
	}
	
	// 创建意图事件
	event := &IntentEvent{
		Intent:    intent,
		Topic:     msg.Metadata["broadcast_topic"],
		Timestamp: time.Now(),
		Source:    msg.Sender,
	}
	
	// 发送到处理队列
	select {
	case il.intentChan <- event:
		il.logger.Debug("Intent event queued",
			zap.String("intent_id", intent.ID),
			zap.String("intent_type", intent.Type),
		)
	default:
		il.logger.Warn("Intent channel full, dropping event",
			zap.String("intent_id", intent.ID),
		)
	}
	
	return nil
}

func (il *IntentListener) processIntents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-il.intentChan:
			if !ok {
				return
			}
			
			// 应用过滤器
			intent := event.Intent.(*common.Intent)
			if il.shouldProcessIntent(intent) {
				select {
				case il.processingQueue <- event:
					il.logger.Debug("Intent passed filter, queued for processing",
						zap.String("intent_id", intent.ID),
					)
				default:
					il.logger.Warn("Processing queue full, dropping intent",
						zap.String("intent_id", intent.ID),
					)
				}
			} else {
				il.logger.Debug("Intent filtered out",
					zap.String("intent_id", intent.ID),
					zap.String("reason", "filter_rules"),
				)
			}
		}
	}
}

func (il *IntentListener) handleProcessingQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-il.processingQueue:
			if !ok {
				return
			}
			
			// 处理意图并生成出价
			go il.processIntentEvent(ctx, event)
		}
	}
}

func (il *IntentListener) processIntentEvent(ctx context.Context, event *IntentEvent) {
	intent := event.Intent.(*common.Intent)
	
	il.logger.Info("Processing intent event",
		zap.String("intent_id", intent.ID),
		zap.String("intent_type", intent.Type),
	)
	
	// 生成出价决策
	bidDecision, err := il.bidDecisionMgr.MakeBidDecision(ctx, intent)
	if err != nil {
		il.logger.Error("Failed to make bid decision",
			zap.String("intent_id", intent.ID),
			zap.Error(err),
		)
		return
	}
	
	if !bidDecision.ShouldBid {
		il.logger.Debug("Decision: do not bid",
			zap.String("intent_id", intent.ID),
			zap.String("reason", bidDecision.Reason),
		)
		return
	}
	
	// 提交出价
	err = il.submitBid(ctx, intent, bidDecision)
	if err != nil {
		il.logger.Error("Failed to submit bid",
			zap.String("intent_id", intent.ID),
			zap.Error(err),
		)
		return
	}
	
	il.logger.Info("Bid submitted successfully",
		zap.String("intent_id", intent.ID),
		zap.String("bid_amount", bidDecision.BidAmount),
	)
}

func (il *IntentListener) shouldProcessIntent(intent *common.Intent) bool {
	filter := il.config.IntentFilter
	
	// 检查类型过滤
	if len(filter.AllowedTypes) > 0 {
		allowed := false
		for _, allowedType := range filter.AllowedTypes {
			if intent.Type == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}
	
	// 检查阻止类型
	for _, blockedType := range filter.BlockedTypes {
		if intent.Type == blockedType {
			return false
		}
	}
	
	// 检查发送者过滤
	if len(filter.AllowedSenders) > 0 {
		allowed := false
		for _, allowedSender := range filter.AllowedSenders {
			if intent.SenderID == allowedSender {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}
	
	// 检查阻止发送者
	for _, blockedSender := range filter.BlockedSenders {
		if intent.SenderID == blockedSender {
			return false
		}
	}
	
	// 检查优先级范围
	if filter.MinPriority > 0 && intent.Priority < filter.MinPriority {
		return false
	}
	if filter.MaxPriority > 0 && intent.Priority > filter.MaxPriority {
		return false
	}
	
	// 检查必需标签
	if len(filter.RequiredTags) > 0 {
		intentTags := make(map[string]bool)
		for _, tag := range intent.RelevantTags {
			intentTags[tag.TagName] = true
		}
		
		for _, requiredTag := range filter.RequiredTags {
			if !intentTags[requiredTag] {
				return false
			}
		}
	}
	
	return true
}

func (il *IntentListener) deserializeIntent(payload []byte) (*common.Intent, error) {
	var intent common.Intent
	if err := common.JSON.Unmarshal(payload, &intent); err != nil {
		return nil, fmt.Errorf("failed to deserialize intent: %w", err)
	}
	return &intent, nil
}

func (il *IntentListener) submitBid(ctx context.Context, intent *common.Intent, decision *BidDecision) error {
	// 创建出价消息
	bidMsg := &transport.BidMessage{
		IntentID:     intent.ID,
		AgentID:      il.config.AgentID,
		BidAmount:    decision.BidAmount,
		Capabilities: il.config.Capabilities,
		Timestamp:    time.Now().UnixMilli(),
		AgentType:    string(il.config.AgentType),
		Metadata:     decision.Metadata,
	}
	
	// 发布出价消息
	return il.transportMgr.PublishBidMessage(ctx, bidMsg)
}