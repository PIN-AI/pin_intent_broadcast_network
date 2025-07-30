package intent

import (
	"context"
	"time"

	"pin_intent_broadcast_network/internal/biz/common"
)

// SubscribeIntents implements the intent subscription logic
// This file contains the implementation for task 2.5
func (m *Manager) subscribeIntents(ctx context.Context, req *common.SubscribeIntentsRequest) (<-chan *common.Intent, error) {
	if req == nil {
		return nil, common.NewIntentError(common.ErrorCodeInvalidFormat, "Subscribe request cannot be nil", "")
	}

	m.logger.Infof("Creating intent subscription for types: %v", req.Types)

	// 1. Create subscription channel
	bufferSize := common.DefaultChannelBufferSize
	intentChan := make(chan *common.Intent, bufferSize)

	// 2. Determine subscription topics
	topics := determineSubscriptionTopics(req.Types)
	if len(topics) == 0 {
		topics = []string{common.TopicIntentBroadcast} // Subscribe to all if no specific types
	}

	m.logger.Debugf("Subscribing to topics: %v", topics)

	// 3. Create subscription filter
	filter := &SubscriptionFilter{
		Types: req.Types,
	}

	// TODO: Integrate with transport layer in Phase 2
	// For now, we simulate subscription by monitoring in-memory intents
	go m.simulateSubscription(ctx, intentChan, filter)

	// 4. Start cleanup goroutine
	go func() {
		<-ctx.Done()
		close(intentChan)
		m.logger.Debug("Intent subscription channel closed")
	}()

	m.logger.Info("Intent subscription created successfully")
	return intentChan, nil
}

// simulateSubscription simulates intent subscription until transport layer is integrated
func (m *Manager) simulateSubscription(ctx context.Context, intentChan chan<- *common.Intent, filter *SubscriptionFilter) {
	// This is a temporary implementation that monitors in-memory intents
	// In Phase 2, this will be replaced with actual transport layer subscription

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	seenIntents := make(map[string]bool)

	for {
		select {
		case <-ticker.C:
			m.mu.RLock()
			for _, intent := range m.intents {
				// Skip if already seen
				if seenIntents[intent.ID] {
					continue
				}

				// Apply filter
				if matchesFilter(intent, filter) {
					select {
					case intentChan <- intent:
						seenIntents[intent.ID] = true
						m.logger.Debugf("Sent intent to subscriber: %s", intent.ID)
					case <-ctx.Done():
						m.mu.RUnlock()
						return
					default:
						// Channel is full, skip this intent
						m.logger.Warnf("Subscription channel full, skipping intent: %s", intent.ID)
					}
				}
			}
			m.mu.RUnlock()
		case <-ctx.Done():
			return
		}
	}
}

// determineSubscriptionTopics determines topics to subscribe based on intent types
func determineSubscriptionTopics(intentTypes []string) []string {
	if len(intentTypes) == 0 {
		return []string{common.TopicIntentBroadcast} // Subscribe to all
	}

	topics := make([]string, 0, len(intentTypes))
	for _, intentType := range intentTypes {
		topic := determineTopicByType(intentType)
		if !common.Slices.ContainsString(topics, topic) {
			topics = append(topics, topic)
		}
	}

	return topics
}

// matchesFilter checks if an intent matches subscription filter
func matchesFilter(intent *common.Intent, filter *SubscriptionFilter) bool {
	if filter == nil {
		return true
	}

	// Filter by types
	if len(filter.Types) > 0 {
		if !common.Slices.ContainsString(filter.Types, intent.Type) {
			return false
		}
	}

	// Filter by sender ID if specified
	if filter.SenderID != "" && intent.SenderID != filter.SenderID {
		return false
	}

	// Skip expired intents
	if common.Times.IsExpired(intent.Timestamp, intent.TTL) {
		return false
	}

	// Only include intents that have been broadcasted or processed
	if intent.Status != common.IntentStatusBroadcasted &&
		intent.Status != common.IntentStatusProcessed &&
		intent.Status != common.IntentStatusMatched {
		return false
	}

	return true
}

// SubscriptionFilter represents subscription filter criteria
type SubscriptionFilter struct {
	Types    []string
	SenderID string
	// Add more filter criteria as needed
}

// deserializeIntent deserializes an intent from network data
func deserializeIntent(data []byte) (*common.Intent, error) {
	var intent common.Intent
	if err := common.JSON.Unmarshal(data, &intent); err != nil {
		return nil, common.WrapError(err, common.ErrorCodeProcessingFailed, "Failed to deserialize intent")
	}
	return &intent, nil
}
