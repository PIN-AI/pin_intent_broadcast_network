package common

import "time"

// Intent type constants
const (
	// Trading intent types
	IntentTypeTrade    = "trade"
	IntentTypeSwap     = "swap"
	IntentTypeExchange = "exchange"

	// Transfer intent types
	IntentTypeTransfer = "transfer"
	IntentTypeSend     = "send"
	IntentTypePayment  = "payment"

	// Lending intent types
	IntentTypeLending = "lending"
	IntentTypeBorrow  = "borrow"
	IntentTypeLoan    = "loan"

	// Investment intent types
	IntentTypeInvestment = "investment"
	IntentTypeStaking    = "staking"
	IntentTypeYield      = "yield"
)

// Intent status constants (already defined in types.go, but listed here for reference)
const (
	StatusCreated     = "created"
	StatusValidated   = "validated"
	StatusBroadcasted = "broadcasted"
	StatusProcessed   = "processed"
	StatusMatched     = "matched"
	StatusCompleted   = "completed"
	StatusFailed      = "failed"
	StatusExpired     = "expired"
)

// Intent priority constants
const (
	PriorityLow    int32 = 1
	PriorityNormal int32 = 5
	PriorityHigh   int32 = 10
	PriorityUrgent int32 = 20
)

// Network topic constants
const (
	TopicIntentBroadcast = "intent.broadcast"
	TopicIntentMatching  = "intent.matching"
	TopicIntentStatus    = "intent.status"
	TopicNetworkStatus   = "network.status"
	TopicPeerDiscovery   = "peer.discovery"
)

// Configuration defaults
const (
	// Intent configuration defaults
	DefaultMaxConcurrentIntents = 1000
	DefaultProcessingTimeout    = 30 * time.Second
	DefaultRetryAttempts        = 3
	DefaultTTL                  = 3600 * time.Second // 1 hour
	DefaultIntentPriority       = PriorityNormal

	// Validation configuration defaults
	DefaultMaxPayloadSize = 1024 * 1024 // 1MB
	DefaultMaxTTL         = 24 * time.Hour

	// Security configuration defaults
	DefaultSignatureAlgorithm = "Ed25519"
	DefaultKeyRotationPeriod  = 30 * 24 * time.Hour // 30 days

	// Processing configuration defaults
	DefaultPipelineTimeout = 60 * time.Second
	DefaultStageTimeout    = 10 * time.Second
	DefaultMaxRetries      = 3

	// Matching configuration defaults
	DefaultConfidenceThreshold = 0.8
	DefaultMaxMatchesPerIntent = 10
	DefaultMatchingTimeout     = 5 * time.Second

	// Network configuration defaults
	DefaultMaxPeers               = 100
	DefaultStatusUpdateInterval   = 30 * time.Second
	DefaultTopologyUpdateInterval = 60 * time.Second

	// Cache configuration defaults
	DefaultCacheSize = 1000
	DefaultCacheTTL  = 5 * time.Minute
)

// Signature algorithm constants
const (
	SignatureAlgorithmEd25519   = "Ed25519"
	SignatureAlgorithmECDSA     = "ECDSA"
	SignatureAlgorithmRSA       = "RSA"
	SignatureAlgorithmSecp256k1 = "secp256k1"
)

// Hash algorithm constants
const (
	HashAlgorithmSHA256    = "SHA256"
	HashAlgorithmSHA512    = "SHA512"
	HashAlgorithmBlake2b   = "BLAKE2b"
	HashAlgorithmKeccak256 = "Keccak256"
)

// Encryption algorithm constants
const (
	EncryptionAlgorithmAES256   = "AES256"
	EncryptionAlgorithmChaCha20 = "ChaCha20"
	EncryptionAlgorithmXSalsa20 = "XSalsa20"
)

// Network health status constants
const (
	NetworkHealthUnknown      = "unknown"
	NetworkHealthDisconnected = "disconnected"
	NetworkHealthPoor         = "poor"
	NetworkHealthGood         = "good"
	NetworkHealthExcellent    = "excellent"
	NetworkHealthStale        = "stale"
)

// Match type constants (already defined in types.go, but listed here for reference)
const (
	MatchTypeExactStr    = "exact"
	MatchTypePartialStr  = "partial"
	MatchTypeSemanticStr = "semantic"
	MatchTypePatternStr  = "pattern"
)

// Validation rule priority constants
const (
	ValidationPriorityHigh   = 100
	ValidationPriorityMedium = 50
	ValidationPriorityLow    = 10
)

// Processing stage priority constants
const (
	ProcessingStagePriorityValidation = 100
	ProcessingStagePrioritySignature  = 90
	ProcessingStagePriorityMatching   = 80
	ProcessingStagePriorityRouting    = 70
	ProcessingStagePriorityStorage    = 60
)

// Error code constants (already defined in errors.go, but listed here for reference)
const (
	ErrorCodeIntentNotFound       = "INTENT_NOT_FOUND"
	ErrorCodeInvalidFormat        = "INVALID_FORMAT"
	ErrorCodeValidationFailed     = "VALIDATION_FAILED"
	ErrorCodeSignatureFailed      = "SIGNATURE_FAILED"
	ErrorCodePermissionDenied     = "PERMISSION_DENIED"
	ErrorCodeProcessingFailed     = "PROCESSING_FAILED"
	ErrorCodeHandlerNotFound      = "HANDLER_NOT_FOUND"
	ErrorCodeIntentExpired        = "INTENT_EXPIRED"
	ErrorCodeAlreadyProcessed     = "ALREADY_PROCESSED"
	ErrorCodeNetworkUnavailable   = "NETWORK_UNAVAILABLE"
	ErrorCodeBroadcastFailed      = "BROADCAST_FAILED"
	ErrorCodeInvalidConfiguration = "INVALID_CONFIGURATION"
	ErrorCodeMatchingFailed       = "MATCHING_FAILED"
	ErrorCodeStorageUnavailable   = "STORAGE_UNAVAILABLE"
	ErrorCodeProcessingTimeout    = "PROCESSING_TIMEOUT"
	ErrorCodeRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"
)

// Metric names constants
const (
	MetricIntentsCreated    = "intents_created_total"
	MetricIntentsProcessed  = "intents_processed_total"
	MetricIntentsMatched    = "intents_matched_total"
	MetricIntentsFailed     = "intents_failed_total"
	MetricIntentsExpired    = "intents_expired_total"
	MetricProcessingLatency = "intent_processing_duration_seconds"
	MetricValidationErrors  = "validation_errors_total"
	MetricSignatureFailures = "signature_failures_total"
	MetricMatchingAccuracy  = "matching_accuracy_ratio"
	MetricNetworkPeers      = "network_peers_connected"
	MetricMessagesSent      = "messages_sent_total"
	MetricMessagesReceived  = "messages_received_total"
)

// Log level constants
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// Environment constants
const (
	EnvironmentDevelopment = "development"
	EnvironmentTesting     = "testing"
	EnvironmentStaging     = "staging"
	EnvironmentProduction  = "production"
)

// Version information
const (
	Version     = "0.1.0"
	APIVersion  = "v1"
	BuildCommit = "unknown" // Will be set during build
	BuildTime   = "unknown" // Will be set during build
)

// File and directory constants
const (
	DefaultConfigDir   = ".kiro"
	DefaultKeyStoreDir = "keystore"
	DefaultDataDir     = "data"
	DefaultLogDir      = "logs"
	DefaultCacheDir    = "cache"
	DefaultBackupDir   = "backup"

	ConfigFileName     = "config.yaml"
	KeyStoreFileName   = "keystore.json"
	NetworkConfigFile  = "network.yaml"
	SecurityConfigFile = "security.yaml"
)

// Network protocol constants
const (
	ProtocolVersion = "1.0.0"
	ProtocolName    = "pin-intent-broadcast"
	UserAgent       = "pin-intent-broadcast-network/" + Version
)

// Rate limiting constants
const (
	DefaultRateLimit       = 100 // requests per minute
	DefaultBurstLimit      = 10  // burst capacity
	DefaultRateLimitWindow = 1 * time.Minute
)

// Timeout constants
const (
	DefaultConnectTimeout     = 10 * time.Second
	DefaultReadTimeout        = 30 * time.Second
	DefaultWriteTimeout       = 30 * time.Second
	DefaultShutdownTimeout    = 30 * time.Second
	DefaultHealthCheckTimeout = 5 * time.Second
)

// Buffer size constants
const (
	DefaultChannelBufferSize = 100
	DefaultMessageBufferSize = 1024
	DefaultEventBufferSize   = 50
)
