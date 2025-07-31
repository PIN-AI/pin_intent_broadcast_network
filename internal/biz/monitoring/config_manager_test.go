package monitoring

import (
	"testing"
	"time"

	"pin_intent_broadcast_network/internal/conf"

	"google.golang.org/protobuf/types/known/durationpb"
)

func TestConfigManager_LoadConfig(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name            string
		transportConfig *conf.Transport
		expectedMode    string
		expectError     bool
	}{
		{
			name:            "nil transport config",
			transportConfig: nil,
			expectedMode:    "all",
			expectError:     false,
		},
		{
			name: "nil intent monitoring config",
			transportConfig: &conf.Transport{
				IntentMonitoring: nil,
			},
			expectedMode: "all",
			expectError:  false,
		},
		{
			name: "valid config with explicit mode",
			transportConfig: &conf.Transport{
				IntentMonitoring: &conf.IntentMonitoring{
					SubscriptionMode: "explicit",
					ExplicitTopics:   []string{"intent-broadcast.trade"},
				},
			},
			expectedMode: "explicit",
			expectError:  false,
		},
		{
			name: "invalid subscription mode",
			transportConfig: &conf.Transport{
				IntentMonitoring: &conf.IntentMonitoring{
					SubscriptionMode: "invalid",
				},
			},
			expectedMode: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := cm.LoadConfig(tt.transportConfig)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config.SubscriptionMode != tt.expectedMode {
				t.Errorf("expected subscription mode %s, got %s", tt.expectedMode, config.SubscriptionMode)
			}
		})
	}
}

func TestConfigManager_GetDefaultConfig(t *testing.T) {
	cm := NewConfigManager()
	config := cm.getDefaultConfig()

	// Test default subscription mode
	if config.SubscriptionMode != "all" {
		t.Errorf("expected default subscription mode 'all', got %s", config.SubscriptionMode)
	}

	// Test default wildcard patterns
	if len(config.WildcardPatterns) == 0 {
		t.Error("expected default wildcard patterns to be set")
	}

	// Test default explicit topics
	if len(config.ExplicitTopics) == 0 {
		t.Error("expected default explicit topics to be set")
	}

	// Test default filter
	if config.Filter == nil {
		t.Error("expected default filter to be set")
	}

	// Test default statistics
	if config.Statistics == nil {
		t.Error("expected default statistics to be set")
	}
	if !config.Statistics.Enabled {
		t.Error("expected statistics to be enabled by default")
	}

	// Test default performance
	if config.Performance == nil {
		t.Error("expected default performance to be set")
	}
	if config.Performance.MaxSubscriptions <= 0 {
		t.Error("expected default max subscriptions to be positive")
	}
}

func TestConfigManager_IsValidSubscriptionMode(t *testing.T) {
	cm := NewConfigManager()

	validModes := []string{"wildcard", "explicit", "all", "disabled"}
	invalidModes := []string{"invalid", "", "unknown", "test"}

	for _, mode := range validModes {
		if !cm.isValidSubscriptionMode(mode) {
			t.Errorf("expected %s to be valid subscription mode", mode)
		}
	}

	for _, mode := range invalidModes {
		if cm.isValidSubscriptionMode(mode) {
			t.Errorf("expected %s to be invalid subscription mode", mode)
		}
	}
}

func TestConfigManager_GetSubscriptionTopics(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name        string
		config      *conf.IntentMonitoring
		expectError bool
		expectEmpty bool
		expectCount int
	}{
		{
			name: "disabled mode",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "disabled",
			},
			expectError: false,
			expectEmpty: true,
		},
		{
			name: "explicit mode with topics",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "explicit",
				ExplicitTopics:   []string{"topic1", "topic2"},
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name: "explicit mode without topics (uses defaults)",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "explicit",
			},
			expectError: false,
			expectCount: 13, // Default explicit topics count
		},
		{
			name: "wildcard mode",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "wildcard",
				WildcardPatterns: []string{"intent-broadcast.*"},
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "all mode",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "all",
			},
			expectError: false,
			expectCount: 16, // All known topics count
		},
		{
			name: "invalid mode",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm.config = tt.config
			topics, err := cm.GetSubscriptionTopics()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectEmpty && len(topics) != 0 {
				t.Errorf("expected empty topics list, got %d topics", len(topics))
				return
			}

			if tt.expectCount > 0 && len(topics) != tt.expectCount {
				t.Errorf("expected %d topics, got %d", tt.expectCount, len(topics))
			}
		})
	}
}

func TestConfigManager_ValidateFilter(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name     string
		filter   *conf.IntentFilter
		expected *conf.IntentFilter
	}{
		{
			name:   "nil filter",
			filter: nil,
			expected: &conf.IntentFilter{
				MinPriority: 0,
				MaxPriority: 10,
			},
		},
		{
			name: "filter with inverted priorities",
			filter: &conf.IntentFilter{
				MinPriority: 8,
				MaxPriority: 3,
			},
			expected: &conf.IntentFilter{
				MinPriority: 3,
				MaxPriority: 8,
			},
		},
		{
			name: "filter with zero priorities",
			filter: &conf.IntentFilter{
				MinPriority: 0,
				MaxPriority: 0,
			},
			expected: &conf.IntentFilter{
				MinPriority: 0,
				MaxPriority: 10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.validateFilter(tt.filter)

			if result.MinPriority != tt.expected.MinPriority {
				t.Errorf("expected min priority %d, got %d", tt.expected.MinPriority, result.MinPriority)
			}

			if result.MaxPriority != tt.expected.MaxPriority {
				t.Errorf("expected max priority %d, got %d", tt.expected.MaxPriority, result.MaxPriority)
			}
		})
	}
}

func TestConfigManager_ValidateStatistics(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name     string
		stats    *conf.StatisticsConfig
		expected *conf.StatisticsConfig
	}{
		{
			name:  "nil statistics",
			stats: nil,
			expected: &conf.StatisticsConfig{
				Enabled:             true,
				RetentionPeriod:     durationpb.New(24 * time.Hour),
				AggregationInterval: durationpb.New(time.Minute),
			},
		},
		{
			name: "statistics with nil durations",
			stats: &conf.StatisticsConfig{
				Enabled:             false,
				RetentionPeriod:     nil,
				AggregationInterval: nil,
			},
			expected: &conf.StatisticsConfig{
				Enabled:             false,
				RetentionPeriod:     durationpb.New(24 * time.Hour),
				AggregationInterval: durationpb.New(time.Minute),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.validateStatistics(tt.stats)

			if result.Enabled != tt.expected.Enabled {
				t.Errorf("expected enabled %v, got %v", tt.expected.Enabled, result.Enabled)
			}

			if result.RetentionPeriod.AsDuration() != tt.expected.RetentionPeriod.AsDuration() {
				t.Errorf("expected retention period %v, got %v",
					tt.expected.RetentionPeriod.AsDuration(),
					result.RetentionPeriod.AsDuration())
			}

			if result.AggregationInterval.AsDuration() != tt.expected.AggregationInterval.AsDuration() {
				t.Errorf("expected aggregation interval %v, got %v",
					tt.expected.AggregationInterval.AsDuration(),
					result.AggregationInterval.AsDuration())
			}
		})
	}
}

func TestConfigManager_ValidatePerformance(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name     string
		perf     *conf.PerformanceConfig
		expected *conf.PerformanceConfig
	}{
		{
			name: "nil performance",
			perf: nil,
			expected: &conf.PerformanceConfig{
				MaxSubscriptions:  100,
				MessageBufferSize: 1000,
				BatchSize:         10,
			},
		},
		{
			name: "performance with zero values",
			perf: &conf.PerformanceConfig{
				MaxSubscriptions:  0,
				MessageBufferSize: 0,
				BatchSize:         0,
			},
			expected: &conf.PerformanceConfig{
				MaxSubscriptions:  100,
				MessageBufferSize: 1000,
				BatchSize:         10,
			},
		},
		{
			name: "performance with negative values",
			perf: &conf.PerformanceConfig{
				MaxSubscriptions:  -10,
				MessageBufferSize: -100,
				BatchSize:         -5,
			},
			expected: &conf.PerformanceConfig{
				MaxSubscriptions:  100,
				MessageBufferSize: 1000,
				BatchSize:         10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.validatePerformance(tt.perf)

			if result.MaxSubscriptions != tt.expected.MaxSubscriptions {
				t.Errorf("expected max subscriptions %d, got %d",
					tt.expected.MaxSubscriptions, result.MaxSubscriptions)
			}

			if result.MessageBufferSize != tt.expected.MessageBufferSize {
				t.Errorf("expected message buffer size %d, got %d",
					tt.expected.MessageBufferSize, result.MessageBufferSize)
			}

			if result.BatchSize != tt.expected.BatchSize {
				t.Errorf("expected batch size %d, got %d",
					tt.expected.BatchSize, result.BatchSize)
			}
		})
	}
}

func TestConfigManager_IsFilterEnabled(t *testing.T) {
	cm := NewConfigManager()

	tests := []struct {
		name     string
		config   *conf.IntentMonitoring
		expected bool
	}{
		{
			name: "no filter",
			config: &conf.IntentMonitoring{
				Filter: nil,
			},
			expected: false,
		},
		{
			name: "empty filter",
			config: &conf.IntentMonitoring{
				Filter: &conf.IntentFilter{},
			},
			expected: false,
		},
		{
			name: "filter with allowed types",
			config: &conf.IntentMonitoring{
				Filter: &conf.IntentFilter{
					AllowedTypes: []string{"trade"},
				},
			},
			expected: true,
		},
		{
			name: "filter with blocked types",
			config: &conf.IntentMonitoring{
				Filter: &conf.IntentFilter{
					BlockedTypes: []string{"spam"},
				},
			},
			expected: true,
		},
		{
			name: "filter with priority range only",
			config: &conf.IntentMonitoring{
				Filter: &conf.IntentFilter{
					MinPriority: 5,
					MaxPriority: 8,
				},
			},
			expected: false, // Priority-only filter is not considered "enabled"
		},
		{
			name: "filter with allowed types and priority",
			config: &conf.IntentMonitoring{
				Filter: &conf.IntentFilter{
					AllowedTypes: []string{"trade"},
					MinPriority:  5,
					MaxPriority:  8,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate the config first to apply defaults
			if tt.config.Filter != nil {
				tt.config.Filter = cm.validateFilter(tt.config.Filter)
			}
			cm.config = tt.config
			result := cm.IsFilterEnabled()

			if result != tt.expected {
				t.Errorf("expected filter enabled %v, got %v", tt.expected, result)
			}
		})
	}
}
