package common

import (
	"context"
	"crypto"

	"github.com/libp2p/go-libp2p/core/peer"
)

// IntentManager Intent 管理器接口
type IntentManager interface {
	CreateIntent(ctx context.Context, req *CreateIntentRequest) (*CreateIntentResponse, error)
	BroadcastIntent(ctx context.Context, req *BroadcastIntentRequest) (*BroadcastIntentResponse, error)
	QueryIntents(ctx context.Context, req *QueryIntentsRequest) (*QueryIntentsResponse, error)
	SubscribeIntents(ctx context.Context, req *SubscribeIntentsRequest) (<-chan *Intent, error)
	ProcessIntent(ctx context.Context, intent *Intent) error
	GetIntentStatus(ctx context.Context, id string) (*Intent, error)
	CancelIntent(ctx context.Context, id string) error
}

// IntentProcessor Intent 处理器接口
type IntentProcessor interface {
	ProcessIncomingIntent(ctx context.Context, intent *Intent) error
	ProcessOutgoingIntent(ctx context.Context, intent *Intent) error
	RegisterHandler(intentType string, handler IntentHandler) error
	UnregisterHandler(intentType string) error
	GetProcessingStatus() *ProcessingStatus
}

// IntentHandler Intent 处理器
type IntentHandler interface {
	Handle(ctx context.Context, intent *Intent) error
	GetSupportedTypes() []string
	GetPriority() int
}

// IntentValidator Intent 验证器接口
type IntentValidator interface {
	ValidateIntent(ctx context.Context, intent *Intent) error
	ValidateFormat(intent *Intent) error
	ValidateBusinessRules(ctx context.Context, intent *Intent) error
	ValidatePermissions(intent *Intent, sender peer.ID) error
	RegisterRule(rule ValidationRule) error
}

// ValidationRule 验证规则接口
type ValidationRule interface {
	Name() string
	Validate(ctx context.Context, intent *Intent) error
	GetPriority() int
}

// FormatRule 格式验证规则接口
type FormatRule interface {
	ValidationRule
	ValidateFormat(intent *Intent) error
}

// BusinessRule 业务规则验证接口
type BusinessRule interface {
	ValidationRule
	ValidateBusinessLogic(ctx context.Context, intent *Intent) error
}

// IntentSigner Intent 签名器接口
type IntentSigner interface {
	SignIntent(intent *Intent, privateKey crypto.PrivateKey) error
	VerifySignature(intent *Intent) error
	GenerateKeyPair() (crypto.PrivateKey, crypto.PublicKey, error)
	GetPublicKey(peerID peer.ID) (crypto.PublicKey, error)
}

// KeyStore 密钥存储接口
type KeyStore interface {
	GetPrivateKey(peerID peer.ID) (crypto.PrivateKey, error)
	GetPublicKey(peerID peer.ID) (crypto.PublicKey, error)
	StoreKeyPair(peerID peer.ID, priv crypto.PrivateKey, pub crypto.PublicKey) error
	GenerateKeyPair() (crypto.PrivateKey, crypto.PublicKey, error)
}

// SignatureAlgorithm 签名算法接口
type SignatureAlgorithm interface {
	Sign(data []byte, privateKey crypto.PrivateKey) ([]byte, error)
	Verify(data []byte, signature []byte, publicKey crypto.PublicKey) error
	GetAlgorithmName() string
}

// IntentMatcher Intent 匹配器接口
type IntentMatcher interface {
	FindMatches(ctx context.Context, intent *Intent, candidates []*Intent) ([]*MatchResult, error)
	AddMatchingRule(rule MatchingRule) error
	RemoveMatchingRule(ruleName string) error
	GetMatchingRules() []MatchingRule
}

// MatchingRule 匹配规则接口
type MatchingRule interface {
	Match(intent1, intent2 *Intent) (MatchResult, error)
	GetRuleName() string
	GetPriority() int
}

// ProcessingPipeline 处理管道接口
type ProcessingPipeline interface {
	AddStage(stage ProcessingStage) error
	Process(ctx context.Context, intent *Intent) error
	GetStages() []ProcessingStage
}

// ProcessingStage 处理阶段接口
type ProcessingStage interface {
	Name() string
	Process(ctx context.Context, intent *Intent) error
	ShouldProcess(intent *Intent) bool
	GetPriority() int
}

// HandlerRegistry 处理器注册表接口
type HandlerRegistry interface {
	RegisterHandler(intentType string, handler IntentHandler) error
	UnregisterHandler(intentType string) error
	GetHandler(intentType string) (IntentHandler, error)
	ListHandlers() map[string][]IntentHandler
}

// LifecycleManager 生命周期管理器接口
type LifecycleManager interface {
	StartTracking(intent *Intent)
	StopTracking(intentID string)
	UpdateStatus(intentID string, status IntentStatus) error
	GetTracker(intentID string) (*IntentTracker, error)
	RegisterCallback(callback LifecycleCallback)
}

// LifecycleCallback 生命周期回调接口
type LifecycleCallback interface {
	OnStatusChange(intent *Intent, oldStatus, newStatus IntentStatus) error
	OnExpired(intent *Intent) error
}

// NetworkManager 网络管理器接口
type NetworkManager interface {
	GetNetworkStatus(ctx context.Context) (*NetworkStatusResponse, error)
	GetConnectedPeers() []peer.ID
	GetNetworkTopology() *NetworkTopology
	HandleNetworkEvent(event NetworkEvent) error
}

// NetworkTopology 网络拓扑结构
type NetworkTopology struct {
	Nodes []peer.ID              `json:"nodes"`
	Edges map[string][]string    `json:"edges"`
	Stats map[string]interface{} `json:"stats"`
}

// NetworkEvent 网络事件
type NetworkEvent struct {
	Type      string                 `json:"type"`
	PeerID    peer.ID                `json:"peer_id"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// ProcessingStatus 处理状态
type ProcessingStatus struct {
	ActiveIntents  int                    `json:"active_intents"`
	ProcessedCount int64                  `json:"processed_count"`
	FailedCount    int64                  `json:"failed_count"`
	AverageLatency int64                  `json:"average_latency"`
	HandlerStatus  map[string]interface{} `json:"handler_status"`
	PipelineStatus map[string]interface{} `json:"pipeline_status"`
}
