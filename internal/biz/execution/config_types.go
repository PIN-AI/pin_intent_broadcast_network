package execution

import (
	"time"
	
	"pin_intent_broadcast_network/internal/biz/service_agent"
	"pin_intent_broadcast_network/internal/biz/block_builder"
)

// YAML Configuration Structures
// These structures match the YAML configuration files

// AgentsConfig represents the entire agents configuration file (now includes builders)
type AgentsConfig struct {
	Automation     AutomationConfig    `yaml:"automation"`
	Agents         []YAMLAgentConfig   `yaml:"agents"`
	Builders       YAMLBuildersSection `yaml:"builders"`
	GlobalSettings GlobalSettings      `yaml:"global_settings"`
}

// YAMLBuildersSection represents the builders section in unified config
type YAMLBuildersSection struct {
	Enabled   bool                  `yaml:"enabled"`
	AutoStart bool                  `yaml:"auto_start"`
	Configs   []YAMLBuilderConfig   `yaml:"configs"`
}

// BuildersConfig represents the entire builders configuration file
type BuildersConfig struct {
	Automation     AutomationConfig      `yaml:"automation"`
	Builders       []YAMLBuilderConfig   `yaml:"builders"`
	GlobalSettings GlobalSettings        `yaml:"global_settings"`
}

// AutomationConfig contains automation-related settings
type AutomationConfig struct {
	Enabled   bool   `yaml:"enabled"`
	AutoStart bool   `yaml:"auto_start"`
	LogLevel  string `yaml:"log_level"`
	
	// Async initialization configuration
	AsyncInit *AsyncInitConfig `yaml:"async_init,omitempty"`
}

// AsyncInitConfig contains async initialization settings
type AsyncInitConfig struct {
	Enabled                    bool   `yaml:"enabled"`
	TransportReadinessTimeout  string `yaml:"transport_readiness_timeout"`
	ComponentStartTimeout      string `yaml:"component_start_timeout"`
	MaxInitRetries             int    `yaml:"max_init_retries"`
	RetryBackoffInterval       string `yaml:"retry_backoff_interval"`
	
	// Component lifecycle settings
	ComponentLifecycle *YAMLComponentLifecycleConfig `yaml:"component_lifecycle,omitempty"`
}

// YAMLComponentLifecycleConfig contains component lifecycle settings
type YAMLComponentLifecycleConfig struct {
	ServiceAgents  *ComponentPhaseConfig `yaml:"service_agents,omitempty"`
	BlockBuilders  *ComponentPhaseConfig `yaml:"block_builders,omitempty"`
	StartTimeout   string                `yaml:"start_timeout"`
	StopTimeout    string                `yaml:"stop_timeout"`
	MaxRetries     int                   `yaml:"max_retries"`
	RetryInterval  string                `yaml:"retry_interval"`
	DependencyWait string                `yaml:"dependency_wait"`
	ParallelStart  bool                  `yaml:"parallel_start"`
}

// ComponentPhaseConfig contains phase-specific component settings
type ComponentPhaseConfig struct {
	StartPriority         int    `yaml:"start_priority"`
	HealthCheckInterval   string `yaml:"health_check_interval"`
	RestartOnFailure      bool   `yaml:"restart_on_failure"`
	Dependencies          []string `yaml:"dependencies,omitempty"`
}

// YAMLAgentConfig represents an agent configuration in YAML
type YAMLAgentConfig struct {
	AgentID              string             `yaml:"agent_id"`
	AgentType            string             `yaml:"agent_type"`
	Name                 string             `yaml:"name"`
	Description          string             `yaml:"description"`
	AutoStart            bool               `yaml:"auto_start"`
	Capabilities         []string           `yaml:"capabilities"`
	Specializations      []string           `yaml:"specializations"`
	BidStrategy          YAMLBidStrategy    `yaml:"bid_strategy"`
	MaxConcurrentIntents int                `yaml:"max_concurrent_intents"`
	MinBidAmount         string             `yaml:"min_bid_amount"`
	MaxBidAmount         string             `yaml:"max_bid_amount"`
	IntentFilter         YAMLIntentFilter   `yaml:"intent_filter"`
}

// YAMLBidStrategy represents bid strategy configuration in YAML
type YAMLBidStrategy struct {
	Type         string  `yaml:"type"`
	BaseFee      string  `yaml:"base_fee"`
	ProfitMargin float64 `yaml:"profit_margin"`
	RiskFactor   float64 `yaml:"risk_factor"`
}

// YAMLIntentFilter represents intent filter configuration in YAML
type YAMLIntentFilter struct {
	AllowedTypes    []string `yaml:"allowed_types"`
	BlockedTypes    []string `yaml:"blocked_types"`
	AllowedSenders  []string `yaml:"allowed_senders"`
	BlockedSenders  []string `yaml:"blocked_senders"`
	MinPriority     int32    `yaml:"min_priority"`
	MaxPriority     int32    `yaml:"max_priority"`
	RequiredTags    []string `yaml:"required_tags"`
}

// YAMLBuilderConfig represents a builder configuration in YAML
type YAMLBuilderConfig struct {
	BuilderID            string                 `yaml:"builder_id"`
	Name                 string                 `yaml:"name"`
	Description          string                 `yaml:"description"`
	AutoStart            bool                   `yaml:"auto_start"`
	MatchingAlgorithm    string                 `yaml:"matching_algorithm"`
	SettlementMode       string                 `yaml:"settlement_mode"`
	BidCollectionWindow  string                 `yaml:"bid_collection_window"`
	MatchingInterval     string                 `yaml:"matching_interval"`
	CleanupInterval      string                 `yaml:"cleanup_interval"`
	MaxConcurrentIntents int                    `yaml:"max_concurrent_intents"`
	MinBidsRequired      int                    `yaml:"min_bids_required"`
	MaxSessionLifetime   string                 `yaml:"max_session_lifetime"`
	IntentFilter         *YAMLBuilderFilter     `yaml:"intent_filter,omitempty"`
	Subscriptions        YAMLSubscriptions      `yaml:"subscriptions"`
}

// YAMLBuilderFilter represents builder-specific intent filter
type YAMLBuilderFilter struct {
	MinPriority    int32    `yaml:"min_priority"`
	MinBidAmount   string   `yaml:"min_bid_amount"`
	RequiredTags   []string `yaml:"required_tags"`
}

// YAMLSubscriptions represents P2P topic subscriptions
type YAMLSubscriptions struct {
	IntentTopics      []string `yaml:"intent_topics"`
	BidTopics         []string `yaml:"bid_topics"`
	MatchResultTopics []string `yaml:"match_result_topics"`
}

// GlobalSettings represents global configuration settings
type GlobalSettings struct {
	P2PNetwork  P2PNetworkConfig  `yaml:"p2p_network"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	Logging     LoggingConfig     `yaml:"logging"`
	Bidding     BiddingConfig     `yaml:"bidding,omitempty"`
	Matching    MatchingConfig    `yaml:"matching,omitempty"`
	Performance PerformanceConfig `yaml:"performance,omitempty"`
	Security    SecurityConfig    `yaml:"security,omitempty"`
	Storage     StorageConfig     `yaml:"storage,omitempty"`
}

// P2PNetworkConfig represents P2P network configuration
type P2PNetworkConfig struct {
	Enabled           bool          `yaml:"enabled"`
	BootstrapPeers    []string      `yaml:"bootstrap_peers"`
	ConnectionTimeout string        `yaml:"connection_timeout,omitempty"`
	HeartbeatInterval string        `yaml:"heartbeat_interval,omitempty"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	MetricsEnabled        bool `yaml:"metrics_enabled"`
	MetricsInterval       int  `yaml:"metrics_interval"`
	StatusCheckInterval   int  `yaml:"status_check_interval"`
	SessionMonitoring     bool `yaml:"session_monitoring,omitempty"`
	PerformanceTracking   bool `yaml:"performance_tracking,omitempty"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level             string `yaml:"level"`
	FilePath          string `yaml:"file_path"`
	MaxFileSizeMB     int    `yaml:"max_file_size_mb"`
	MaxBackups        int    `yaml:"max_backups"`
	StructuredLogging bool   `yaml:"structured_logging,omitempty"`
}

// BiddingConfig represents bidding system configuration
type BiddingConfig struct {
	TimeoutSeconds       int    `yaml:"timeout_seconds"`
	RetryAttempts        int    `yaml:"retry_attempts"`
	CommitmentEnabled    bool   `yaml:"commitment_enabled"`
	CommitmentAlgorithm  string `yaml:"commitment_algorithm"`
}

// MatchingConfig represents matching system configuration
type MatchingConfig struct {
	AlgorithmWeights   AlgorithmWeights   `yaml:"algorithm_weights"`
	ReputationSystem   ReputationSystem   `yaml:"reputation_system"`
}

// AlgorithmWeights represents matching algorithm weight configuration
type AlgorithmWeights struct {
	HighestBid         float64                  `yaml:"highest_bid"`
	ReputationWeighted ReputationWeightedConfig `yaml:"reputation_weighted"`
	Random             RandomConfig             `yaml:"random"`
}

// ReputationWeightedConfig represents reputation-weighted algorithm config
type ReputationWeightedConfig struct {
	BidWeight        float64 `yaml:"bid_weight"`
	ReputationWeight float64 `yaml:"reputation_weight"`
}

// RandomConfig represents random algorithm config
type RandomConfig struct {
	SeedRotationInterval string `yaml:"seed_rotation_interval"`
}

// ReputationSystem represents reputation system configuration
type ReputationSystem struct {
	Enabled       bool    `yaml:"enabled"`
	BaseScore     float64 `yaml:"base_score"`
	MaxScore      float64 `yaml:"max_score"`
	MinScore      float64 `yaml:"min_score"`
	SuccessBonus  float64 `yaml:"success_bonus"`
	FailurePenalty float64 `yaml:"failure_penalty"`
}

// PerformanceConfig represents performance optimization configuration
type PerformanceConfig struct {
	WorkerPoolSize      int    `yaml:"worker_pool_size"`
	MessageBufferSize   int    `yaml:"message_buffer_size"`
	BatchProcessing     bool   `yaml:"batch_processing"`
	BatchSize           int    `yaml:"batch_size"`
	ProcessingTimeout   string `yaml:"processing_timeout"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	SignatureVerification bool         `yaml:"signature_verification"`
	MessageValidation     bool         `yaml:"message_validation"`
	RateLimiting          RateLimiting `yaml:"rate_limiting"`
}

// RateLimiting represents rate limiting configuration
type RateLimiting struct {
	Enabled              bool `yaml:"enabled"`
	MaxRequestsPerSecond int  `yaml:"max_requests_per_second"`
	BurstCapacity        int  `yaml:"burst_capacity"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	SessionPersistence     bool   `yaml:"session_persistence"`
	MatchHistoryRetention  string `yaml:"match_history_retention"`
	MetricsRetention       string `yaml:"metrics_retention"`
	CleanupEnabled         bool   `yaml:"cleanup_enabled"`
}

// Configuration conversion functions

// convertToServiceAgentConfig converts YAML config to service_agent.AgentConfig
func (am *AutomationManager) convertToServiceAgentConfig(yamlConfig YAMLAgentConfig) *service_agent.AgentConfig {
	// Convert agent type
	var agentType service_agent.AgentType
	switch yamlConfig.AgentType {
	case "trading":
		agentType = service_agent.AgentTypeTrading
	case "data_access":
		agentType = service_agent.AgentTypeDataAccess
	case "computation":
		agentType = service_agent.AgentTypeComputation
	case "general":
		agentType = service_agent.AgentTypeGeneral
	default:
		agentType = service_agent.AgentTypeGeneral
	}

	return &service_agent.AgentConfig{
		AgentID:              yamlConfig.AgentID,
		AgentType:            agentType,
		Name:                 yamlConfig.Name,
		Description:          yamlConfig.Description,
		Capabilities:         yamlConfig.Capabilities,
		Specializations:      yamlConfig.Specializations,
		BidStrategy: service_agent.BidStrategy{
			Type:         yamlConfig.BidStrategy.Type,
			BaseFee:      yamlConfig.BidStrategy.BaseFee,
			ProfitMargin: yamlConfig.BidStrategy.ProfitMargin,
			RiskFactor:   yamlConfig.BidStrategy.RiskFactor,
		},
		MaxConcurrentIntents: yamlConfig.MaxConcurrentIntents,
		MinBidAmount:         yamlConfig.MinBidAmount,
		MaxBidAmount:         yamlConfig.MaxBidAmount,
		IntentFilter: service_agent.IntentFilter{
			AllowedTypes:   yamlConfig.IntentFilter.AllowedTypes,
			BlockedTypes:   yamlConfig.IntentFilter.BlockedTypes,
			AllowedSenders: yamlConfig.IntentFilter.AllowedSenders,
			BlockedSenders: yamlConfig.IntentFilter.BlockedSenders,
			MinPriority:    yamlConfig.IntentFilter.MinPriority,
			MaxPriority:    yamlConfig.IntentFilter.MaxPriority,
			RequiredTags:   yamlConfig.IntentFilter.RequiredTags,
		},
	}
}

// convertToBlockBuilderConfig converts YAML config to block_builder.BuilderConfig
func (am *AutomationManager) convertToBlockBuilderConfig(yamlConfig YAMLBuilderConfig) *block_builder.BuilderConfig {
	// Parse durations
	bidCollectionWindow, _ := time.ParseDuration(yamlConfig.BidCollectionWindow)

	return &block_builder.BuilderConfig{
		BuilderID:            yamlConfig.BuilderID,
		MatchingAlgorithm:    yamlConfig.MatchingAlgorithm,
		SettlementMode:       yamlConfig.SettlementMode,
		BidCollectionWindow:  bidCollectionWindow,
		MaxConcurrentIntents: yamlConfig.MaxConcurrentIntents,
		MinBidsRequired:      yamlConfig.MinBidsRequired,
	}
}