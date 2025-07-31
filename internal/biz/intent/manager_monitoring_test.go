package intent

import (
	"context"
	"os"
	"testing"

	"pin_intent_broadcast_network/internal/biz/monitoring"
	"pin_intent_broadcast_network/internal/conf"
	"pin_intent_broadcast_network/internal/transport"

	"github.com/go-kratos/kratos/v2/log"
)

// mockIntentMonitoringManager implements monitoring.IntentMonitoringManager for testing
type mockIntentMonitoringManager struct {
	started bool
	stopped bool
}

func (m *mockIntentMonitoringManager) Start(ctx context.Context) error {
	m.started = true
	return nil
}

func (m *mockIntentMonitoringManager) Stop() error {
	m.stopped = true
	return nil
}

func (m *mockIntentMonitoringManager) IsRunning() bool {
	return m.started && !m.stopped
}

func (m *mockIntentMonitoringManager) UpdateConfig(config *conf.IntentMonitoring) error {
	return nil
}

func (m *mockIntentMonitoringManager) GetSubscriptionStatus() *monitoring.SubscriptionStatus {
	return &monitoring.SubscriptionStatus{
		Mode:                "all",
		ActiveSubscriptions: []string{"intent-broadcast.trade", "intent-broadcast.swap"},
	}
}

func (m *mockIntentMonitoringManager) GetStatistics() *monitoring.MonitoringStatistics {
	return &monitoring.MonitoringStatistics{
		TotalReceived: 10,
	}
}

// mockTransportManager implements transport.TransportManager for testing
type mockTransportManager struct {
	subscriptions map[string]func(*transport.TransportMessage) error
}

func newMockTransportManager() *mockTransportManager {
	return &mockTransportManager{
		subscriptions: make(map[string]func(*transport.TransportMessage) error),
	}
}

func (m *mockTransportManager) Start(ctx context.Context, config *transport.TransportConfig) error {
	return nil
}

func (m *mockTransportManager) Stop() error {
	return nil
}

func (m *mockTransportManager) IsRunning() bool {
	return true
}

func (m *mockTransportManager) GetPubSubManager() transport.PubSubManager {
	return nil
}

func (m *mockTransportManager) GetTopicManager() transport.TopicManager {
	return nil
}

func (m *mockTransportManager) GetMessageSerializer() transport.MessageSerializer {
	return nil
}

func (m *mockTransportManager) GetMessageRouter() transport.MessageRouter {
	return nil
}

func (m *mockTransportManager) PublishMessage(ctx context.Context, topic string, msg *transport.TransportMessage) error {
	return nil
}

func (m *mockTransportManager) SubscribeToTopic(topic string, handler func(*transport.TransportMessage) error) (transport.Subscription, error) {
	m.subscriptions[topic] = handler
	return &mockSubscription{topic: topic, active: true}, nil
}

// mockSubscription implements transport.Subscription for testing
type mockSubscription struct {
	topic  string
	active bool
}

func (m *mockSubscription) Topic() string {
	return m.topic
}

func (m *mockSubscription) Cancel() error {
	m.active = false
	return nil
}

func (m *mockSubscription) IsActive() bool {
	return m.active
}

func TestManager_SetIntentMonitoringManager(t *testing.T) {
	// Create a basic manager
	manager := NewManager(
		nil, nil, nil, nil, nil,
		newMockTransportManager(),
		&Config{},
		log.NewStdLogger(os.Stdout),
	)

	// Create mock monitoring manager
	mockMonitoringMgr := &mockIntentMonitoringManager{}

	// Set the monitoring manager
	manager.SetIntentMonitoringManager(mockMonitoringMgr)

	// Verify it was set
	if manager.intentMonitoringMgr == nil {
		t.Error("expected intent monitoring manager to be set")
	}

	if manager.intentMonitoringMgr != mockMonitoringMgr {
		t.Error("expected intent monitoring manager to match the one set")
	}
}

func TestManager_StartIntentSubscription_WithMonitoring(t *testing.T) {
	// Create a basic manager
	transportMgr := newMockTransportManager()
	manager := NewManager(
		nil, nil, nil, nil, nil,
		transportMgr,
		&Config{},
		log.NewStdLogger(os.Stdout),
	)

	// Create mock monitoring manager
	mockMonitoringMgr := &mockIntentMonitoringManager{}
	manager.SetIntentMonitoringManager(mockMonitoringMgr)

	// Start intent subscription
	ctx := context.Background()
	err := manager.StartIntentSubscription(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify monitoring manager was started
	if !mockMonitoringMgr.started {
		t.Error("expected monitoring manager to be started")
	}
}

func TestManager_StartIntentSubscription_WithoutMonitoring(t *testing.T) {
	// Create a basic manager without monitoring
	transportMgr := newMockTransportManager()
	manager := NewManager(
		nil, nil, nil, nil, nil,
		transportMgr,
		&Config{},
		log.NewStdLogger(os.Stdout),
	)

	// Start intent subscription (should use legacy method)
	ctx := context.Background()
	err := manager.StartIntentSubscription(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify that topics were subscribed using legacy method
	expectedTopics := []string{
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
		"intent-broadcast.matching",
		"intent-broadcast.notification",
		"intent-broadcast.status",
	}

	for _, topic := range expectedTopics {
		if _, exists := transportMgr.subscriptions[topic]; !exists {
			t.Errorf("expected to be subscribed to topic: %s", topic)
		}
	}

	// Verify we subscribed to more topics than the original 5
	if len(transportMgr.subscriptions) < 10 {
		t.Errorf("expected at least 10 subscriptions, got %d", len(transportMgr.subscriptions))
	}
}

func TestManager_StartIntentSubscription_NoTransportManager(t *testing.T) {
	// Create a manager without transport manager
	manager := NewManager(
		nil, nil, nil, nil, nil,
		nil, // No transport manager
		&Config{},
		log.NewStdLogger(os.Stdout),
	)

	// Start intent subscription (should not fail)
	ctx := context.Background()
	err := manager.StartIntentSubscription(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
