package monitoring

import (
	"context"

	"pin_intent_broadcast_network/internal/conf"
)

// IntentMonitoringManager manages intent monitoring functionality
type IntentMonitoringManager interface {
	// Start starts the intent monitoring manager
	Start(ctx context.Context) error
	// Stop stops the intent monitoring manager
	Stop() error
	// IsRunning returns whether the manager is running
	IsRunning() bool

	// UpdateConfig updates the monitoring configuration
	UpdateConfig(config *conf.IntentMonitoring) error
	// GetSubscriptionStatus returns current subscription status
	GetSubscriptionStatus() *SubscriptionStatus
	// GetStatistics returns monitoring statistics
	GetStatistics() *MonitoringStatistics
}

// SubscriptionStatus holds information about current subscriptions
type SubscriptionStatus struct {
	Mode                string                      `json:"mode"`
	ActiveSubscriptions []string                    `json:"active_subscriptions"`
	TopicStats          map[string]*TopicStatistics `json:"topic_stats"`
	TotalMessages       int64                       `json:"total_messages"`
	TotalErrors         int64                       `json:"total_errors"`
}

// MonitoringStatistics holds monitoring statistics
type MonitoringStatistics struct {
	TotalReceived   int64                        `json:"total_received"`
	TotalFiltered   int64                        `json:"total_filtered"`
	TotalDuplicates int64                        `json:"total_duplicates"`
	ByType          map[string]*TypeStatistics   `json:"by_type"`
	BySender        map[string]*SenderStatistics `json:"by_sender"`
	ByTopic         map[string]*TopicStatistics  `json:"by_topic"`
}

// TypeStatistics holds statistics for intent types
type TypeStatistics struct {
	Count        int64  `json:"count"`
	LastReceived string `json:"last_received"`
	AverageSize  int64  `json:"average_size"`
	ErrorCount   int64  `json:"error_count"`
}

// SenderStatistics holds statistics for senders
type SenderStatistics struct {
	Count        int64    `json:"count"`
	LastReceived string   `json:"last_received"`
	IntentTypes  []string `json:"intent_types"`
	ErrorCount   int64    `json:"error_count"`
}

// intentMonitoringManager implements IntentMonitoringManager
type intentMonitoringManager struct {
	configMgr       *ConfigManager
	subscriptionMgr TopicSubscriptionManager
	isRunning       bool
}

// NewIntentMonitoringManager creates a new intent monitoring manager
func NewIntentMonitoringManager(
	configMgr *ConfigManager,
	subscriptionMgr TopicSubscriptionManager,
) IntentMonitoringManager {
	return &intentMonitoringManager{
		configMgr:       configMgr,
		subscriptionMgr: subscriptionMgr,
	}
}

// Start starts the intent monitoring manager
func (imm *intentMonitoringManager) Start(ctx context.Context) error {
	if imm.isRunning {
		return nil
	}

	// Start the subscription manager
	if err := imm.subscriptionMgr.Start(ctx); err != nil {
		return err
	}

	imm.isRunning = true
	return nil
}

// Stop stops the intent monitoring manager
func (imm *intentMonitoringManager) Stop() error {
	if !imm.isRunning {
		return nil
	}

	// Stop the subscription manager
	if err := imm.subscriptionMgr.Stop(); err != nil {
		return err
	}

	imm.isRunning = false
	return nil
}

// IsRunning returns whether the manager is running
func (imm *intentMonitoringManager) IsRunning() bool {
	return imm.isRunning
}

// UpdateConfig updates the monitoring configuration
func (imm *intentMonitoringManager) UpdateConfig(config *conf.IntentMonitoring) error {
	// Update config manager
	imm.configMgr.config = config

	// Update subscription manager
	return imm.subscriptionMgr.UpdateConfig(config)
}

// GetSubscriptionStatus returns current subscription status
func (imm *intentMonitoringManager) GetSubscriptionStatus() *SubscriptionStatus {
	subscriptions := imm.subscriptionMgr.GetSubscriptions()
	topicStats := imm.subscriptionMgr.GetAllTopicStats()

	var totalMessages, totalErrors int64
	for _, stats := range topicStats {
		totalMessages += stats.MessageCount
		totalErrors += stats.ErrorCount
	}

	return &SubscriptionStatus{
		Mode:                imm.configMgr.GetConfig().SubscriptionMode,
		ActiveSubscriptions: subscriptions,
		TopicStats:          topicStats,
		TotalMessages:       totalMessages,
		TotalErrors:         totalErrors,
	}
}

// GetStatistics returns monitoring statistics
func (imm *intentMonitoringManager) GetStatistics() *MonitoringStatistics {
	topicStats := imm.subscriptionMgr.GetAllTopicStats()

	var totalReceived int64
	byTopic := make(map[string]*TopicStatistics)

	for topic, stats := range topicStats {
		totalReceived += stats.MessageCount
		byTopic[topic] = stats
	}

	return &MonitoringStatistics{
		TotalReceived: totalReceived,
		ByTopic:       byTopic,
		// Other statistics would be populated by actual usage
		ByType:   make(map[string]*TypeStatistics),
		BySender: make(map[string]*SenderStatistics),
	}
}
