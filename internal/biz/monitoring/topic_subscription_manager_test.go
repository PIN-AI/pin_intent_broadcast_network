package monitoring

import (
	"context"
	"strings"
	"testing"
	"time"

	"pin_intent_broadcast_network/internal/conf"
	"pin_intent_broadcast_network/internal/transport"

	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

// mockTransportManager implements transport.TransportManager for testing
type mockTransportManager struct {
	subscriptions map[string]func(*transport.TransportMessage) error
	pubsubMgr     *mockPubSubManager
}

func newMockTransportManager() *mockTransportManager {
	return &mockTransportManager{
		subscriptions: make(map[string]func(*transport.TransportMessage) error),
		pubsubMgr:     newMockPubSubManager(),
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
	return m.pubsubMgr
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

// mockPubSubManager implements transport.PubSubManager for testing
type mockPubSubManager struct {
	peerCounts map[string]int
}

func newMockPubSubManager() *mockPubSubManager {
	return &mockPubSubManager{
		peerCounts: make(map[string]int),
	}
}

func (m *mockPubSubManager) Start(ctx context.Context, config *transport.PubSubConfig) error {
	return nil
}

func (m *mockPubSubManager) Stop() error {
	return nil
}

func (m *mockPubSubManager) Publish(ctx context.Context, topic string, data []byte) error {
	return nil
}

func (m *mockPubSubManager) Subscribe(topic string, handler transport.MessageHandler) (transport.Subscription, error) {
	return &mockSubscription{topic: topic, active: true}, nil
}

func (m *mockPubSubManager) Unsubscribe(topic string) error {
	return nil
}

func (m *mockPubSubManager) GetConnectedPeers() []peer.ID {
	return []peer.ID{}
}

func (m *mockPubSubManager) GetTopics() []string {
	return []string{}
}

func (m *mockPubSubManager) GetPeerCount(topic string) int {
	if count, exists := m.peerCounts[topic]; exists {
		return count
	}
	return 3 // Default peer count for testing
}

func (m *mockPubSubManager) SetPeerCount(topic string, count int) {
	m.peerCounts[topic] = count
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

// setupTestManager creates a test manager with disabled initial config
func setupTestManager() (TopicSubscriptionManager, *mockTransportManager, *ConfigManager) {
	transportMgr := newMockTransportManager()
	configMgr := NewConfigManager()
	logger := zap.NewNop()

	// Set initial config to disabled to avoid auto-subscription
	configMgr.config = &conf.IntentMonitoring{
		SubscriptionMode: "disabled",
	}

	tsm := NewTopicSubscriptionManager(transportMgr, configMgr, nil, logger)
	return tsm, transportMgr, configMgr
}

func TestTopicSubscriptionManager_Start(t *testing.T) {
	transportMgr := newMockTransportManager()
	configMgr := NewConfigManager()
	logger := zap.NewNop()

	tsm := NewTopicSubscriptionManager(transportMgr, configMgr, nil, logger)

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Errorf("unexpected error starting manager: %v", err)
	}

	if !tsm.IsRunning() {
		t.Error("expected manager to be running")
	}

	// Clean up
	err = tsm.Stop()
	if err != nil {
		t.Errorf("unexpected error stopping manager: %v", err)
	}
}

func TestTopicSubscriptionManager_SubscribeByConfig(t *testing.T) {
	tsm, _, _ := setupTestManager()

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer tsm.Stop()

	tests := []struct {
		name           string
		config         *conf.IntentMonitoring
		expectedTopics int
		expectError    bool
	}{
		{
			name: "disabled mode",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "disabled",
			},
			expectedTopics: 0,
			expectError:    false,
		},
		{
			name: "explicit mode with topics",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "explicit",
				ExplicitTopics:   []string{"topic1", "topic2"},
			},
			expectedTopics: 2,
			expectError:    false,
		},
		{
			name: "explicit mode without topics (uses defaults)",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "explicit",
			},
			expectedTopics: 13, // Default explicit topics count
			expectError:    false,
		},
		{
			name: "wildcard mode",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "wildcard",
				WildcardPatterns: []string{"intent-broadcast.*"},
			},
			expectedTopics: 16, // Should match all intent broadcast topics (including additional ones)
			expectError:    false,
		},
		{
			name: "all mode",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "all",
			},
			expectedTopics: 16, // All known topics
			expectError:    false,
		},
		{
			name: "invalid mode",
			config: &conf.IntentMonitoring{
				SubscriptionMode: "invalid",
			},
			expectedTopics: 0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous subscriptions
			err := tsm.UnsubscribeAll()
			if err != nil {
				t.Logf("Warning: failed to clear subscriptions: %v", err)
			}

			err = tsm.SubscribeByConfig(tt.config)

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

			subscriptions := tsm.GetSubscriptions()
			if len(subscriptions) != tt.expectedTopics {
				t.Errorf("expected %d subscriptions, got %d", tt.expectedTopics, len(subscriptions))
			}
		})
	}
}

func TestTopicSubscriptionManager_SubscribeExplicit(t *testing.T) {
	tsm, _, _ := setupTestManager()

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer tsm.Stop()

	topics := []string{"topic1", "topic2", "topic3"}
	err = tsm.SubscribeExplicit(topics)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	subscriptions := tsm.GetSubscriptions()
	if len(subscriptions) != len(topics) {
		t.Errorf("expected %d subscriptions, got %d", len(topics), len(subscriptions))
	}

	// Check that all topics are subscribed
	subscriptionMap := make(map[string]bool)
	for _, topic := range subscriptions {
		subscriptionMap[topic] = true
	}

	for _, topic := range topics {
		if !subscriptionMap[topic] {
			t.Errorf("expected to be subscribed to topic %s", topic)
		}
	}
}

func TestTopicSubscriptionManager_SubscribeWildcard(t *testing.T) {
	tsm, _, _ := setupTestManager()

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer tsm.Stop()

	pattern := "intent-broadcast.*"
	err = tsm.SubscribeWildcard(pattern)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	subscriptions := tsm.GetSubscriptions()
	if len(subscriptions) == 0 {
		t.Error("expected some subscriptions for wildcard pattern")
	}

	// All subscriptions should match the pattern
	for _, topic := range subscriptions {
		if !strings.HasPrefix(topic, "intent-broadcast.") {
			t.Errorf("topic %s does not match pattern %s", topic, pattern)
		}
	}
}

func TestTopicSubscriptionManager_SubscribeAll(t *testing.T) {
	tsm, _, _ := setupTestManager()

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer tsm.Stop()

	err = tsm.SubscribeAll()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	subscriptions := tsm.GetSubscriptions()
	expectedCount := 16 // All known topics
	if len(subscriptions) != expectedCount {
		t.Errorf("expected %d subscriptions, got %d", expectedCount, len(subscriptions))
	}
}

func TestTopicSubscriptionManager_Unsubscribe(t *testing.T) {
	tsm, _, _ := setupTestManager()

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer tsm.Stop()

	// Subscribe to some topics first
	topics := []string{"topic1", "topic2", "topic3"}
	err = tsm.SubscribeExplicit(topics)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Unsubscribe from one topic
	err = tsm.Unsubscribe("topic2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	subscriptions := tsm.GetSubscriptions()
	if len(subscriptions) != 2 {
		t.Errorf("expected 2 subscriptions after unsubscribe, got %d", len(subscriptions))
	}

	// Check that topic2 is not in subscriptions
	for _, topic := range subscriptions {
		if topic == "topic2" {
			t.Error("topic2 should not be in subscriptions after unsubscribe")
		}
	}

	// Try to unsubscribe from non-existent topic
	err = tsm.Unsubscribe("non-existent")
	if err == nil {
		t.Error("expected error when unsubscribing from non-existent topic")
	}
}

func TestTopicSubscriptionManager_UnsubscribeAll(t *testing.T) {
	tsm, _, _ := setupTestManager()

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer tsm.Stop()

	// Subscribe to some topics first
	topics := []string{"topic1", "topic2", "topic3"}
	err = tsm.SubscribeExplicit(topics)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Unsubscribe from all topics
	err = tsm.UnsubscribeAll()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	subscriptions := tsm.GetSubscriptions()
	if len(subscriptions) != 0 {
		t.Errorf("expected 0 subscriptions after unsubscribe all, got %d", len(subscriptions))
	}
}

func TestTopicSubscriptionManager_GetTopicStats(t *testing.T) {
	tsm, transportMgr, _ := setupTestManager()

	// Set up peer count for testing
	transportMgr.pubsubMgr.SetPeerCount("topic1", 5)

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer tsm.Stop()

	// Subscribe to a topic
	err = tsm.SubscribeExplicit([]string{"topic1"})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Get stats for the topic
	stats := tsm.GetTopicStats("topic1")
	if stats == nil {
		t.Error("expected stats for topic1, got nil")
		return
	}

	if stats.Topic != "topic1" {
		t.Errorf("expected topic name 'topic1', got %s", stats.Topic)
	}

	if !stats.IsActive {
		t.Error("expected topic to be active")
	}

	if stats.PeerCount != 5 {
		t.Errorf("expected peer count 5, got %d", stats.PeerCount)
	}

	// Get stats for non-existent topic
	stats = tsm.GetTopicStats("non-existent")
	if stats != nil {
		t.Error("expected nil stats for non-existent topic")
	}
}

func TestTopicSubscriptionManager_UpdateConfig(t *testing.T) {
	tsm, _, _ := setupTestManager()

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer tsm.Stop()

	// Initial configuration - explicit mode
	initialConfig := &conf.IntentMonitoring{
		SubscriptionMode: "explicit",
		ExplicitTopics:   []string{"topic1", "topic2"},
	}

	err = tsm.SubscribeByConfig(initialConfig)
	if err != nil {
		t.Fatalf("failed to subscribe with initial config: %v", err)
	}

	if len(tsm.GetSubscriptions()) != 2 {
		t.Errorf("expected 2 initial subscriptions, got %d", len(tsm.GetSubscriptions()))
	}

	// Update to different configuration - different explicit topics
	newConfig := &conf.IntentMonitoring{
		SubscriptionMode: "explicit",
		ExplicitTopics:   []string{"topic3", "topic4", "topic5"},
	}

	err = tsm.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("unexpected error updating config: %v", err)
	}

	subscriptions := tsm.GetSubscriptions()
	if len(subscriptions) != 3 {
		t.Errorf("expected 3 subscriptions after config update, got %d", len(subscriptions))
	}

	// Check that new topics are subscribed
	subscriptionMap := make(map[string]bool)
	for _, topic := range subscriptions {
		subscriptionMap[topic] = true
	}

	expectedTopics := []string{"topic3", "topic4", "topic5"}
	for _, topic := range expectedTopics {
		if !subscriptionMap[topic] {
			t.Errorf("expected to be subscribed to topic %s after config update", topic)
		}
	}

	// Update to same configuration - should be no-op
	err = tsm.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("unexpected error updating to same config: %v", err)
	}

	if len(tsm.GetSubscriptions()) != 3 {
		t.Errorf("expected 3 subscriptions after no-op config update, got %d", len(tsm.GetSubscriptions()))
	}
}

func TestTopicSubscriptionManager_MessageHandling(t *testing.T) {
	// Track messages received
	var receivedMessages []*transport.TransportMessage
	messageHandler := func(msg *transport.TransportMessage) error {
		receivedMessages = append(receivedMessages, msg)
		return nil
	}

	transportMgr := newMockTransportManager()
	configMgr := NewConfigManager()
	logger := zap.NewNop()

	// Set initial config to disabled to avoid auto-subscription
	configMgr.config = &conf.IntentMonitoring{
		SubscriptionMode: "disabled",
	}

	tsm := NewTopicSubscriptionManager(transportMgr, configMgr, messageHandler, logger)

	ctx := context.Background()
	err := tsm.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	defer tsm.Stop()

	// Subscribe to a topic
	err = tsm.SubscribeExplicit([]string{"topic1"})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Simulate receiving a message
	testMsg := &transport.TransportMessage{
		ID:        "test-msg-1",
		Type:      "intent_broadcast",
		Payload:   []byte("test payload"),
		Timestamp: time.Now().UnixMilli(),
		Sender:    "test-sender",
	}

	// Get the handler for topic1 and call it
	handler := transportMgr.subscriptions["topic1"]
	if handler == nil {
		t.Fatal("no handler found for topic1")
	}

	err = handler(testMsg)
	if err != nil {
		t.Errorf("unexpected error handling message: %v", err)
	}

	// Check that message was received
	if len(receivedMessages) != 1 {
		t.Errorf("expected 1 received message, got %d", len(receivedMessages))
	}

	if len(receivedMessages) > 0 && receivedMessages[0].ID != testMsg.ID {
		t.Errorf("expected message ID %s, got %s", testMsg.ID, receivedMessages[0].ID)
	}

	// Check that statistics were updated
	stats := tsm.GetTopicStats("topic1")
	if stats == nil {
		t.Error("expected stats for topic1")
		return
	}

	if stats.MessageCount != 1 {
		t.Errorf("expected message count 1, got %d", stats.MessageCount)
	}
}
