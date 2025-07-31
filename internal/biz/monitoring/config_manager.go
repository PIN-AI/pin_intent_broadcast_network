package monitoring

import (
	"fmt"
	"time"

	"pin_intent_broadcast_network/internal/conf"

	"google.golang.org/protobuf/types/known/durationpb"
)

// ConfigManager manages intent monitoring configuration
type ConfigManager struct {
	config *conf.IntentMonitoring
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

// LoadConfig loads and validates the intent monitoring configuration
func (cm *ConfigManager) LoadConfig(transportConfig *conf.Transport) (*conf.IntentMonitoring, error) {
	if transportConfig == nil {
		return cm.getDefaultConfig(), nil
	}

	config := transportConfig.GetIntentMonitoring()
	if config == nil {
		return cm.getDefaultConfig(), nil
	}

	// Validate and apply defaults
	validatedConfig, err := cm.validateAndApplyDefaults(config)
	if err != nil {
		return nil, fmt.Errorf("failed to validate intent monitoring config: %w", err)
	}

	cm.config = validatedConfig
	return validatedConfig, nil
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *conf.IntentMonitoring {
	if cm.config == nil {
		return cm.getDefaultConfig()
	}
	return cm.config
}

// validateAndApplyDefaults validates configuration and applies default values
func (cm *ConfigManager) validateAndApplyDefaults(config *conf.IntentMonitoring) (*conf.IntentMonitoring, error) {
	// Create a copy to avoid modifying the original
	validatedConfig := &conf.IntentMonitoring{
		SubscriptionMode: config.SubscriptionMode,
		ExplicitTopics:   config.ExplicitTopics,
		WildcardPatterns: config.WildcardPatterns,
		Filter:           config.Filter,
		Statistics:       config.Statistics,
		Performance:      config.Performance,
	}

	// Apply default subscription mode
	if validatedConfig.SubscriptionMode == "" {
		validatedConfig.SubscriptionMode = "all" // Default to listen all topics
	}

	// Validate subscription mode
	if !cm.isValidSubscriptionMode(validatedConfig.SubscriptionMode) {
		return nil, fmt.Errorf("invalid subscription mode: %s", validatedConfig.SubscriptionMode)
	}

	// Apply default wildcard patterns if in wildcard mode
	if validatedConfig.SubscriptionMode == "wildcard" && len(validatedConfig.WildcardPatterns) == 0 {
		validatedConfig.WildcardPatterns = []string{"intent-broadcast.*"}
	}

	// Apply default explicit topics if in explicit mode
	if validatedConfig.SubscriptionMode == "explicit" && len(validatedConfig.ExplicitTopics) == 0 {
		validatedConfig.ExplicitTopics = cm.getDefaultExplicitTopics()
	}

	// Validate and apply filter defaults
	if validatedConfig.Filter == nil {
		validatedConfig.Filter = cm.getDefaultFilter()
	} else {
		validatedConfig.Filter = cm.validateFilter(validatedConfig.Filter)
	}

	// Validate and apply statistics defaults
	if validatedConfig.Statistics == nil {
		validatedConfig.Statistics = cm.getDefaultStatistics()
	} else {
		validatedConfig.Statistics = cm.validateStatistics(validatedConfig.Statistics)
	}

	// Validate and apply performance defaults
	if validatedConfig.Performance == nil {
		validatedConfig.Performance = cm.getDefaultPerformance()
	} else {
		validatedConfig.Performance = cm.validatePerformance(validatedConfig.Performance)
	}

	return validatedConfig, nil
}

// getDefaultConfig returns the default intent monitoring configuration
func (cm *ConfigManager) getDefaultConfig() *conf.IntentMonitoring {
	return &conf.IntentMonitoring{
		SubscriptionMode: "all", // Default to listen all topics
		WildcardPatterns: []string{"intent-broadcast.*"},
		ExplicitTopics:   cm.getDefaultExplicitTopics(),
		Filter:           cm.getDefaultFilter(),
		Statistics:       cm.getDefaultStatistics(),
		Performance:      cm.getDefaultPerformance(),
	}
}

// getDefaultExplicitTopics returns the default list of explicit topics
func (cm *ConfigManager) getDefaultExplicitTopics() []string {
	return []string{
		"intent-broadcast.trade",
		"intent-broadcast.swap",
		"intent-broadcast.exchange",
		"intent-broadcast.transfer",
		"intent-broadcast.send",
		"intent-broadcast.payment",
		"intent-broadcast.lending",
		"intent-broadcast.borrow",
		"intent-broadcast.loan",
		"intent-broadcast.investment",
		"intent-broadcast.staking",
		"intent-broadcast.yield",
		"intent-broadcast.general",
	}
}

// getDefaultFilter returns the default filter configuration
func (cm *ConfigManager) getDefaultFilter() *conf.IntentFilter {
	return &conf.IntentFilter{
		AllowedTypes:   []string{}, // Empty means allow all
		BlockedTypes:   []string{},
		AllowedSenders: []string{},
		BlockedSenders: []string{},
		MinPriority:    0,
		MaxPriority:    10,
	}
}

// getDefaultStatistics returns the default statistics configuration
func (cm *ConfigManager) getDefaultStatistics() *conf.StatisticsConfig {
	return &conf.StatisticsConfig{
		Enabled:             true,
		RetentionPeriod:     durationpb.New(24 * time.Hour), // 24 hours
		AggregationInterval: durationpb.New(time.Minute),    // 1 minute
	}
}

// getDefaultPerformance returns the default performance configuration
func (cm *ConfigManager) getDefaultPerformance() *conf.PerformanceConfig {
	return &conf.PerformanceConfig{
		MaxSubscriptions:  100,
		MessageBufferSize: 1000,
		BatchSize:         10,
	}
}

// isValidSubscriptionMode checks if the subscription mode is valid
func (cm *ConfigManager) isValidSubscriptionMode(mode string) bool {
	validModes := []string{"wildcard", "explicit", "all", "disabled"}
	for _, validMode := range validModes {
		if mode == validMode {
			return true
		}
	}
	return false
}

// validateFilter validates and applies defaults to filter configuration
func (cm *ConfigManager) validateFilter(filter *conf.IntentFilter) *conf.IntentFilter {
	if filter == nil {
		return cm.getDefaultFilter()
	}

	// Apply defaults for priority if not set
	if filter.MinPriority == 0 && filter.MaxPriority == 0 {
		filter.MinPriority = 0
		filter.MaxPriority = 10
	}

	// Ensure min <= max priority
	if filter.MinPriority > filter.MaxPriority {
		filter.MinPriority, filter.MaxPriority = filter.MaxPriority, filter.MinPriority
	}

	return filter
}

// validateStatistics validates and applies defaults to statistics configuration
func (cm *ConfigManager) validateStatistics(stats *conf.StatisticsConfig) *conf.StatisticsConfig {
	if stats == nil {
		return cm.getDefaultStatistics()
	}

	// Apply default retention period if not set
	if stats.RetentionPeriod == nil {
		stats.RetentionPeriod = durationpb.New(24 * time.Hour)
	}

	// Apply default aggregation interval if not set
	if stats.AggregationInterval == nil {
		stats.AggregationInterval = durationpb.New(time.Minute)
	}

	return stats
}

// validatePerformance validates and applies defaults to performance configuration
func (cm *ConfigManager) validatePerformance(perf *conf.PerformanceConfig) *conf.PerformanceConfig {
	if perf == nil {
		return cm.getDefaultPerformance()
	}

	// Apply defaults if values are not set or invalid
	if perf.MaxSubscriptions <= 0 {
		perf.MaxSubscriptions = 100
	}

	if perf.MessageBufferSize <= 0 {
		perf.MessageBufferSize = 1000
	}

	if perf.BatchSize <= 0 {
		perf.BatchSize = 10
	}

	return perf
}

// IsFilterEnabled checks if intent filtering is enabled
func (cm *ConfigManager) IsFilterEnabled() bool {
	config := cm.GetConfig()
	if config.Filter == nil {
		return false
	}

	// Filter is considered enabled if any explicit filter rules are set
	// (not just the default priority range)
	return len(config.Filter.AllowedTypes) > 0 ||
		len(config.Filter.BlockedTypes) > 0 ||
		len(config.Filter.AllowedSenders) > 0 ||
		len(config.Filter.BlockedSenders) > 0
}

// GetSubscriptionTopics returns the list of topics to subscribe based on configuration
func (cm *ConfigManager) GetSubscriptionTopics() ([]string, error) {
	config := cm.GetConfig()

	switch config.SubscriptionMode {
	case "disabled":
		return []string{}, nil

	case "explicit":
		if len(config.ExplicitTopics) == 0 {
			return cm.getDefaultExplicitTopics(), nil
		}
		return config.ExplicitTopics, nil

	case "wildcard":
		// For wildcard mode, we'll return the patterns
		// The actual subscription logic will handle pattern matching
		if len(config.WildcardPatterns) == 0 {
			return []string{"intent-broadcast.*"}, nil
		}
		return config.WildcardPatterns, nil

	case "all":
		// For "all" mode, subscribe to all known intent broadcast topics
		return cm.getAllKnownTopics(), nil

	default:
		return nil, fmt.Errorf("unknown subscription mode: %s", config.SubscriptionMode)
	}
}

// getAllKnownTopics returns all known intent broadcast topics
func (cm *ConfigManager) getAllKnownTopics() []string {
	// This includes all the explicit topics plus any additional known topics
	topics := cm.getDefaultExplicitTopics()

	// Add any additional known topics that might not be in the explicit list
	additionalTopics := []string{
		"intent-broadcast.matching",
		"intent-broadcast.notification",
		"intent-broadcast.status",
	}

	topics = append(topics, additionalTopics...)
	return topics
}
