package service_agent

import (
	"time"
)

// AgentType 代理类型
type AgentType string

const (
	AgentTypeTrading     AgentType = "trading"
	AgentTypeDataAccess  AgentType = "data_access"
	AgentTypeComputation AgentType = "computation"
	AgentTypeGeneral     AgentType = "general"
)

// AgentConfig 代理配置
type AgentConfig struct {
	// 基础信息
	AgentID      string    `yaml:"agent_id"`
	AgentType    AgentType `yaml:"agent_type"`
	Name         string    `yaml:"name"`
	Description  string    `yaml:"description"`
	
	// 能力配置
	Capabilities []string `yaml:"capabilities"`
	Specializations []string `yaml:"specializations"`
	
	// 出价策略
	BidStrategy BidStrategy `yaml:"bid_strategy"`
	
	// 资源限制
	MaxConcurrentIntents int     `yaml:"max_concurrent_intents"`
	MinBidAmount        string   `yaml:"min_bid_amount"`
	MaxBidAmount        string   `yaml:"max_bid_amount"`
	
	// 过滤规则
	IntentFilter IntentFilter `yaml:"intent_filter"`
}

// BidStrategy 出价策略
type BidStrategy struct {
	Type        string  `yaml:"type"` // "conservative", "aggressive", "balanced"
	BaseFee     string  `yaml:"base_fee"`
	ProfitMargin float64 `yaml:"profit_margin"`
	RiskFactor  float64 `yaml:"risk_factor"`
}

// IntentFilter 意图过滤器
type IntentFilter struct {
	AllowedTypes    []string `yaml:"allowed_types"`
	BlockedTypes    []string `yaml:"blocked_types"`
	AllowedSenders  []string `yaml:"allowed_senders"`
	BlockedSenders  []string `yaml:"blocked_senders"`
	MinPriority     int32    `yaml:"min_priority"`
	MaxPriority     int32    `yaml:"max_priority"`
	RequiredTags    []string `yaml:"required_tags"`
}

// AgentStatus 代理状态
type AgentStatus struct {
	AgentID           string    `json:"agent_id"`
	Status            string    `json:"status"` // "active", "busy", "offline"
	ActiveIntents     int       `json:"active_intents"`
	ProcessedIntents  int64     `json:"processed_intents"`
	SuccessfulBids    int64     `json:"successful_bids"`
	TotalEarnings     string    `json:"total_earnings"`
	LastActivity      time.Time `json:"last_activity"`
	ConnectedPeers    int       `json:"connected_peers"`
}

// BidDecision 出价决策结果
type BidDecision struct {
	ShouldBid   bool              `json:"should_bid"`
	BidAmount   string            `json:"bid_amount"`
	Confidence  float64           `json:"confidence"`
	Reason      string            `json:"reason"`
	Metadata    map[string]string `json:"metadata"`
}

// IntentEvent 意图事件
type IntentEvent struct {
	Intent    interface{} `json:"intent"`  // Will be *common.Intent 
	Topic     string      `json:"topic"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source"`
}

// AgentMetrics 代理性能指标
type AgentMetrics struct {
	IntentsReceived     int64     `json:"intents_received"`
	IntentsFiltered     int64     `json:"intents_filtered"`
	BidsSubmitted       int64     `json:"bids_submitted"`
	BidsWon             int64     `json:"bids_won"`
	TotalEarnings       string    `json:"total_earnings"`
	AverageConfidence   float64   `json:"average_confidence"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastUpdated         time.Time `json:"last_updated"`
}

// DefaultAgentConfig 返回默认代理配置
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		AgentType:            AgentTypeGeneral,
		MaxConcurrentIntents: 10,
		MinBidAmount:        "1000",
		MaxBidAmount:        "100000",
		BidStrategy: BidStrategy{
			Type:         "balanced",
			BaseFee:      "5000",
			ProfitMargin: 0.15,
			RiskFactor:   0.1,
		},
		IntentFilter: IntentFilter{
			MinPriority: 1,
			MaxPriority: 10,
		},
	}
}