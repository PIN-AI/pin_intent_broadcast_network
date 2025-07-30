package common

import (
	"time"
)

// BusinessConfig holds the complete configuration for the business logic layer
type BusinessConfig struct {
	Intent     IntentConfig     `yaml:"intent"`
	Validation ValidationConfig `yaml:"validation"`
	Security   SecurityConfig   `yaml:"security"`
	Processing ProcessingConfig `yaml:"processing"`
	Matching   MatchingConfig   `yaml:"matching"`
	Network    NetworkConfig    `yaml:"network"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

// IntentConfig holds configuration for intent management
type IntentConfig struct {
	MaxConcurrentIntents int           `yaml:"max_concurrent_intents"`
	ProcessingTimeout    time.Duration `yaml:"processing_timeout"`
	RetryAttempts        int           `yaml:"retry_attempts"`
	DefaultTTL           time.Duration `yaml:"default_ttl"`
	EnableMatching       bool          `yaml:"enable_matching"`
	EnableLifecycle      bool          `yaml:"enable_lifecycle"`
	CleanupInterval      time.Duration `yaml:"cleanup_interval"`
}

// ValidationConfig holds configuration for intent validation
type ValidationConfig struct {
	MaxPayloadSize int           `yaml:"max_payload_size"`
	MaxTTL         time.Duration `yaml:"max_ttl"`
	AllowedTypes   []string      `yaml:"allowed_types"`
	EnableStrict   bool          `yaml:"enable_strict"`
	EnableCache    bool          `yaml:"enable_cache"`
	CacheTTL       time.Duration `yaml:"cache_ttl"`
}

// SecurityConfig holds configuration for security operations
type SecurityConfig struct {
	SignatureAlgorithm  string        `yaml:"signature_algorithm"`
	KeyStoreType        string        `yaml:"keystore_type"`
	KeyStoreDir         string        `yaml:"keystore_dir"`
	KeyRotationPeriod   time.Duration `yaml:"key_rotation_period"`
	EnableEncryption    bool          `yaml:"enable_encryption"`
	EncryptionAlgorithm string        `yaml:"encryption_algorithm"`
	HashAlgorithm       string        `yaml:"hash_algorithm"`
	EnableBackup        bool          `yaml:"enable_backup"`
}

// ProcessingConfig holds configuration for intent processing
type ProcessingConfig struct {
	PipelineTimeout     time.Duration `yaml:"pipeline_timeout"`
	StageTimeout        time.Duration `yaml:"stage_timeout"`
	MaxRetries          int           `yaml:"max_retries"`
	EnableAsync         bool          `yaml:"enable_async"`
	MaxConcurrentStages int           `yaml:"max_concurrent_stages"`
	EnableLoadBalancing bool          `yaml:"enable_load_balancing"`
}

// MatchingConfig holds configuration for intent matching
type MatchingConfig struct {
	ConfidenceThreshold    float64       `yaml:"confidence_threshold"`
	MaxMatchesPerIntent    int           `yaml:"max_matches_per_intent"`
	MatchingTimeout        time.Duration `yaml:"matching_timeout"`
	EnableCaching          bool          `yaml:"enable_caching"`
	CacheSize              int           `yaml:"cache_size"`
	EnableContentMatching  bool          `yaml:"enable_content_matching"`
	EnableMetadataMatching bool          `yaml:"enable_metadata_matching"`
	ContentWeight          float64       `yaml:"content_weight"`
	MetadataWeight         float64       `yaml:"metadata_weight"`
	TypeWeight             float64       `yaml:"type_weight"`
}

// NetworkConfig holds configuration for network management
type NetworkConfig struct {
	MaxPeers               int           `yaml:"max_peers"`
	StatusUpdateInterval   time.Duration `yaml:"status_update_interval"`
	TopologyUpdateInterval time.Duration `yaml:"topology_update_interval"`
	EnableTopologyTracking bool          `yaml:"enable_topology_tracking"`
	EnableMetrics          bool          `yaml:"enable_metrics"`
	ConnectionTimeout      time.Duration `yaml:"connection_timeout"`
	HeartbeatInterval      time.Duration `yaml:"heartbeat_interval"`
}

// MonitoringConfig holds configuration for monitoring and metrics
type MonitoringConfig struct {
	EnableMetrics   bool          `yaml:"enable_metrics"`
	MetricsPort     int           `yaml:"metrics_port"`
	MetricsPath     string        `yaml:"metrics_path"`
	EnableProfiling bool          `yaml:"enable_profiling"`
	ProfilingPort   int           `yaml:"profiling_port"`
	LogLevel        string        `yaml:"log_level"`
	LogFormat       string        `yaml:"log_format"`
	EnableTracing   bool          `yaml:"enable_tracing"`
	TracingEndpoint string        `yaml:"tracing_endpoint"`
	SampleRate      float64       `yaml:"sample_rate"`
	MetricsInterval time.Duration `yaml:"metrics_interval"`
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *BusinessConfig {
	return &BusinessConfig{
		Intent: IntentConfig{
			MaxConcurrentIntents: DefaultMaxConcurrentIntents,
			ProcessingTimeout:    DefaultProcessingTimeout,
			RetryAttempts:        DefaultRetryAttempts,
			DefaultTTL:           DefaultTTL,
			EnableMatching:       true,
			EnableLifecycle:      true,
			CleanupInterval:      5 * time.Minute,
		},
		Validation: ValidationConfig{
			MaxPayloadSize: DefaultMaxPayloadSize,
			MaxTTL:         DefaultMaxTTL,
			AllowedTypes: []string{
				IntentTypeTrade,
				IntentTypeTransfer,
				IntentTypeLending,
				IntentTypeSwap,
			},
			EnableStrict: true,
			EnableCache:  true,
			CacheTTL:     DefaultCacheTTL,
		},
		Security: SecurityConfig{
			SignatureAlgorithm:  DefaultSignatureAlgorithm,
			KeyStoreType:        "file",
			KeyStoreDir:         DefaultKeyStoreDir,
			KeyRotationPeriod:   DefaultKeyRotationPeriod,
			EnableEncryption:    true,
			EncryptionAlgorithm: EncryptionAlgorithmAES256,
			HashAlgorithm:       HashAlgorithmSHA256,
			EnableBackup:        true,
		},
		Processing: ProcessingConfig{
			PipelineTimeout:     DefaultPipelineTimeout,
			StageTimeout:        DefaultStageTimeout,
			MaxRetries:          DefaultMaxRetries,
			EnableAsync:         true,
			MaxConcurrentStages: 10,
			EnableLoadBalancing: true,
		},
		Matching: MatchingConfig{
			ConfidenceThreshold:    DefaultConfidenceThreshold,
			MaxMatchesPerIntent:    DefaultMaxMatchesPerIntent,
			MatchingTimeout:        DefaultMatchingTimeout,
			EnableCaching:          true,
			CacheSize:              DefaultCacheSize,
			EnableContentMatching:  true,
			EnableMetadataMatching: true,
			ContentWeight:          0.4,
			MetadataWeight:         0.3,
			TypeWeight:             0.3,
		},
		Network: NetworkConfig{
			MaxPeers:               DefaultMaxPeers,
			StatusUpdateInterval:   DefaultStatusUpdateInterval,
			TopologyUpdateInterval: DefaultTopologyUpdateInterval,
			EnableTopologyTracking: true,
			EnableMetrics:          true,
			ConnectionTimeout:      DefaultConnectTimeout,
			HeartbeatInterval:      30 * time.Second,
		},
		Monitoring: MonitoringConfig{
			EnableMetrics:   true,
			MetricsPort:     9090,
			MetricsPath:     "/metrics",
			EnableProfiling: false,
			ProfilingPort:   6060,
			LogLevel:        LogLevelInfo,
			LogFormat:       "json",
			EnableTracing:   false,
			TracingEndpoint: "",
			SampleRate:      0.1,
			MetricsInterval: 30 * time.Second,
		},
	}
}

// Validate validates the configuration
func (c *BusinessConfig) Validate() error {
	// Validate Intent configuration
	if c.Intent.MaxConcurrentIntents <= 0 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "MaxConcurrentIntents must be positive", "")
	}

	if c.Intent.ProcessingTimeout <= 0 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "ProcessingTimeout must be positive", "")
	}

	if c.Intent.RetryAttempts < 0 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "RetryAttempts cannot be negative", "")
	}

	// Validate Validation configuration
	if c.Validation.MaxPayloadSize <= 0 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "MaxPayloadSize must be positive", "")
	}

	if c.Validation.MaxTTL <= 0 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "MaxTTL must be positive", "")
	}

	// Validate Security configuration
	if c.Security.SignatureAlgorithm == "" {
		return NewIntentError(ErrorCodeInvalidConfiguration, "SignatureAlgorithm cannot be empty", "")
	}

	if c.Security.KeyStoreType == "" {
		return NewIntentError(ErrorCodeInvalidConfiguration, "KeyStoreType cannot be empty", "")
	}

	// Validate Processing configuration
	if c.Processing.PipelineTimeout <= 0 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "PipelineTimeout must be positive", "")
	}

	if c.Processing.StageTimeout <= 0 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "StageTimeout must be positive", "")
	}

	// Validate Matching configuration
	if c.Matching.ConfidenceThreshold < 0 || c.Matching.ConfidenceThreshold > 1 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "ConfidenceThreshold must be between 0 and 1", "")
	}

	if c.Matching.MaxMatchesPerIntent <= 0 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "MaxMatchesPerIntent must be positive", "")
	}

	// Validate Network configuration
	if c.Network.MaxPeers <= 0 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "MaxPeers must be positive", "")
	}

	// Validate Monitoring configuration
	if c.Monitoring.MetricsPort <= 0 || c.Monitoring.MetricsPort > 65535 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "MetricsPort must be between 1 and 65535", "")
	}

	if c.Monitoring.SampleRate < 0 || c.Monitoring.SampleRate > 1 {
		return NewIntentError(ErrorCodeInvalidConfiguration, "SampleRate must be between 0 and 1", "")
	}

	return nil
}

// Merge merges another configuration into this one
func (c *BusinessConfig) Merge(other *BusinessConfig) {
	if other == nil {
		return
	}

	// Merge Intent config
	if other.Intent.MaxConcurrentIntents > 0 {
		c.Intent.MaxConcurrentIntents = other.Intent.MaxConcurrentIntents
	}
	if other.Intent.ProcessingTimeout > 0 {
		c.Intent.ProcessingTimeout = other.Intent.ProcessingTimeout
	}
	if other.Intent.RetryAttempts >= 0 {
		c.Intent.RetryAttempts = other.Intent.RetryAttempts
	}
	if other.Intent.DefaultTTL > 0 {
		c.Intent.DefaultTTL = other.Intent.DefaultTTL
	}

	// Merge Validation config
	if other.Validation.MaxPayloadSize > 0 {
		c.Validation.MaxPayloadSize = other.Validation.MaxPayloadSize
	}
	if other.Validation.MaxTTL > 0 {
		c.Validation.MaxTTL = other.Validation.MaxTTL
	}
	if len(other.Validation.AllowedTypes) > 0 {
		c.Validation.AllowedTypes = other.Validation.AllowedTypes
	}

	// Merge Security config
	if other.Security.SignatureAlgorithm != "" {
		c.Security.SignatureAlgorithm = other.Security.SignatureAlgorithm
	}
	if other.Security.KeyStoreType != "" {
		c.Security.KeyStoreType = other.Security.KeyStoreType
	}
	if other.Security.KeyStoreDir != "" {
		c.Security.KeyStoreDir = other.Security.KeyStoreDir
	}

	// Continue merging other configurations as needed...
}

// Clone creates a deep copy of the configuration
func (c *BusinessConfig) Clone() *BusinessConfig {
	clone := *c

	// Deep copy slices
	clone.Validation.AllowedTypes = make([]string, len(c.Validation.AllowedTypes))
	copy(clone.Validation.AllowedTypes, c.Validation.AllowedTypes)

	return &clone
}
