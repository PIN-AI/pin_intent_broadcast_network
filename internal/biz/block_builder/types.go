package block_builder

import (
	"time"
	
	"pin_intent_broadcast_network/internal/biz/common"
	"pin_intent_broadcast_network/internal/transport"
)

// BuilderConfig Block Builder配置
type BuilderConfig struct {
	BuilderID         string        `yaml:"builder_id"`
	MatchingAlgorithm string        `yaml:"matching_algorithm"` // "highest_bid", "reputation_weighted", "random"
	SettlementMode    string        `yaml:"settlement_mode"`    // "simulated", "blockchain"
	BidCollectionWindow time.Duration `yaml:"bid_collection_window"`
	MaxConcurrentIntents int         `yaml:"max_concurrent_intents"`
	MinBidsRequired   int           `yaml:"min_bids_required"`
}

// IntentSession 意图会话
type IntentSession struct {
	Intent      *common.Intent            `json:"intent"`
	Bids        []*transport.BidMessage   `json:"bids"`
	StartTime   time.Time                 `json:"start_time"`
	EndTime     time.Time                 `json:"end_time"`
	Status      string                    `json:"status"` // "collecting", "matching", "completed", "expired"
	MatchResult *transport.MatchResult    `json:"match_result,omitempty"`
}

// BlockBuilderStatus Block Builder状态
type BlockBuilderStatus struct {
	BuilderID         string    `json:"builder_id"`
	Status            string    `json:"status"` // "active", "busy", "offline"
	ActiveSessions    int       `json:"active_sessions"`
	CompletedMatches  int64     `json:"completed_matches"`
	TotalBidsReceived int64     `json:"total_bids_received"`
	LastActivity      time.Time `json:"last_activity"`
	ConnectedPeers    int       `json:"connected_peers"`
}

// BlockBuilderMetrics Block Builder性能指标
type BlockBuilderMetrics struct {
	SessionsCreated      int64         `json:"sessions_created"`
	SessionsCompleted    int64         `json:"sessions_completed"`
	SessionsExpired      int64         `json:"sessions_expired"`
	BidsReceived         int64         `json:"bids_received"`
	MatchesCompleted     int64         `json:"matches_completed"`
	AverageSessionTime   time.Duration `json:"average_session_time"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
	LastUpdated          time.Time     `json:"last_updated"`
}

// MatchingRequest 匹配请求
type MatchingRequest struct {
	Intent *common.Intent            `json:"intent"`
	Bids   []*transport.BidMessage   `json:"bids"`
	Config *BuilderConfig            `json:"config"`
}

// MatchingResponse 匹配响应
type MatchingResponse struct {
	Winner     *transport.BidMessage  `json:"winner"`
	Result     *transport.MatchResult `json:"result"`
	Algorithm  string                 `json:"algorithm"`
	Reason     string                 `json:"reason"`
	Metadata   map[string]string      `json:"metadata"`
}

// SessionState 会话状态常量
const (
	SessionStateCollecting = "collecting"
	SessionStateMatching   = "matching"
	SessionStateCompleted  = "completed"
	SessionStateExpired    = "expired"
)

// MatchStatus 匹配状态常量
const (
	MatchStatusMatched     = "matched"
	MatchStatusNoMatch     = "no_match"
	MatchStatusMatchFailed = "match_failed"
)

// BuilderStatus Block Builder状态常量
const (
	BuilderStatusActive  = "active"
	BuilderStatusBusy    = "busy"
	BuilderStatusOffline = "offline"
)

// DefaultBuilderConfig 返回默认Block Builder配置
func DefaultBuilderConfig() *BuilderConfig {
	return &BuilderConfig{
		MatchingAlgorithm:    "highest_bid",
		SettlementMode:       "simulated",
		BidCollectionWindow:  30 * time.Second,
		MaxConcurrentIntents: 100,
		MinBidsRequired:      1,
	}
}