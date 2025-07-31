package common

import (
	"time"
)

// Intent 数据结构定义
type Intent struct {
	ID                 string            `json:"id"`
	Type               string            `json:"type"`
	Payload            []byte            `json:"payload"`
	Timestamp          int64             `json:"timestamp"`
	SenderID           string            `json:"sender_id"`
	Signature          []byte            `json:"signature,omitempty"`
	SignatureAlgorithm string            `json:"signature_algorithm,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Status             IntentStatus      `json:"status"`
	Priority           int32             `json:"priority"`
	TTL                int64             `json:"ttl"`
	ProcessedAt        int64             `json:"processed_at,omitempty"`
	Error              string            `json:"error,omitempty"`
	MatchedIntents     []string          `json:"matched_intents,omitempty"`
}

// IntentStatus Intent 状态枚举
type IntentStatus int32

const (
	IntentStatusCreated IntentStatus = iota
	IntentStatusValidated
	IntentStatusBroadcasted
	IntentStatusReceived
	IntentStatusProcessed
	IntentStatusMatched
	IntentStatusCompleted
	IntentStatusFailed
	IntentStatusExpired
)

// String 返回状态的字符串表示
func (s IntentStatus) String() string {
	switch s {
	case IntentStatusCreated:
		return "created"
	case IntentStatusValidated:
		return "validated"
	case IntentStatusBroadcasted:
		return "broadcasted"
	case IntentStatusProcessed:
		return "processed"
	case IntentStatusMatched:
		return "matched"
	case IntentStatusCompleted:
		return "completed"
	case IntentStatusFailed:
		return "failed"
	case IntentStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// CreateIntentRequest Intent 创建请求
type CreateIntentRequest struct {
	Type       string            `json:"type"`
	Payload    []byte            `json:"payload"`
	SenderID   string            `json:"sender_id"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Priority   int32             `json:"priority"`
	TTL        int64             `json:"ttl"`
	PrivateKey interface{}       `json:"-"` // 不序列化私钥
}

// CreateIntentResponse Intent 创建响应
type CreateIntentResponse struct {
	Intent  *Intent `json:"intent"`
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
}

// BroadcastIntentRequest Intent 广播请求
type BroadcastIntentRequest struct {
	Intent *Intent `json:"intent"`
	Topic  string  `json:"topic,omitempty"`
}

// BroadcastIntentResponse Intent 广播响应
type BroadcastIntentResponse struct {
	Success  bool   `json:"success"`
	IntentID string `json:"intent_id"`
	Topic    string `json:"topic"`
	Message  string `json:"message,omitempty"`
}

// QueryIntentsRequest Intent 查询请求
type QueryIntentsRequest struct {
	Type      string `json:"type,omitempty"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
	Limit     int32  `json:"limit,omitempty"`
	Offset    int32  `json:"offset,omitempty"`
}

// QueryIntentsResponse Intent 查询响应
type QueryIntentsResponse struct {
	Intents []*Intent `json:"intents"`
	Total   int32     `json:"total"`
}

// SubscribeIntentsRequest Intent 订阅请求
type SubscribeIntentsRequest struct {
	Types  []string `json:"types,omitempty"`
	Topics []string `json:"topics,omitempty"`
}

// IntentTracker Intent 跟踪器
type IntentTracker struct {
	Intent      *Intent
	CreatedAt   time.Time
	LastUpdated time.Time
	Status      IntentStatus
	Callbacks   []LifecycleCallback
}

// MatchResult 匹配结果
type MatchResult struct {
	IsMatch    bool                   `json:"is_match"`
	Confidence float64                `json:"confidence"`
	MatchType  MatchType              `json:"match_type"`
	Details    map[string]interface{} `json:"details"`
}

// MatchType 匹配类型
type MatchType int32

const (
	MatchTypeExact MatchType = iota
	MatchTypePartial
	MatchTypeSemantic
	MatchTypePattern
)

// String 返回匹配类型的字符串表示
func (mt MatchType) String() string {
	switch mt {
	case MatchTypeExact:
		return "exact"
	case MatchTypePartial:
		return "partial"
	case MatchTypeSemantic:
		return "semantic"
	case MatchTypePattern:
		return "pattern"
	default:
		return "unknown"
	}
}

// IntentSignatureData 用于签名的数据结构
type IntentSignatureData struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Payload   []byte            `json:"payload"`
	Timestamp int64             `json:"timestamp"`
	SenderID  string            `json:"sender_id"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Priority  int32             `json:"priority"`
	TTL       int64             `json:"ttl"`
}

// NetworkStatusResponse 网络状态响应
type NetworkStatusResponse struct {
	PeerCount        int                    `json:"peer_count"`
	ConnectedPeers   []string               `json:"connected_peers"`
	NetworkHealth    string                 `json:"network_health"`
	TopicCount       int                    `json:"topic_count"`
	MessagesSent     int64                  `json:"messages_sent"`
	MessagesReceived int64                  `json:"messages_received"`
	Metrics          map[string]interface{} `json:"metrics"`
}
