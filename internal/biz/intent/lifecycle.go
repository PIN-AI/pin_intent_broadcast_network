package intent

import (
	"sync"
	"time"

	"pin_intent_broadcast_network/internal/biz/common"
)

// LifecycleManager manages the lifecycle of intents
// This file will contain the implementation for task 2.6
type LifecycleManager struct {
	trackedIntents map[string]*common.IntentTracker
	cleanupTicker  *time.Ticker
	callbacks      map[common.IntentStatus][]common.LifecycleCallback
	mu             sync.RWMutex
	stopCh         chan struct{}
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(config *LifecycleConfig) *LifecycleManager {
	lm := &LifecycleManager{
		trackedIntents: make(map[string]*common.IntentTracker),
		callbacks:      make(map[common.IntentStatus][]common.LifecycleCallback),
		stopCh:         make(chan struct{}),
	}

	// Start cleanup ticker
	lm.cleanupTicker = time.NewTicker(config.CleanupInterval)
	go lm.cleanupLoop()

	return lm
}

// LifecycleConfig holds configuration for lifecycle management
type LifecycleConfig struct {
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
	MaxAge          time.Duration `yaml:"max_age"`
}

// StartTracking starts tracking an intent's lifecycle
func (lm *LifecycleManager) StartTracking(intent *common.Intent) {
	if intent == nil {
		return
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	tracker := &common.IntentTracker{
		Intent:      intent,
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
		Status:      intent.Status,
		Callbacks:   make([]common.LifecycleCallback, 0),
	}

	lm.trackedIntents[intent.ID] = tracker

	// Schedule expiration if TTL is set
	if intent.TTL > 0 {
		go lm.scheduleExpiration(intent)
	}
}

// StopTracking stops tracking an intent
func (lm *LifecycleManager) StopTracking(intentID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	delete(lm.trackedIntents, intentID)
}

// UpdateStatus updates the status of a tracked intent
func (lm *LifecycleManager) UpdateStatus(intentID string, status common.IntentStatus) error {
	lm.mu.Lock()
	tracker, exists := lm.trackedIntents[intentID]
	lm.mu.Unlock()

	if !exists {
		return common.NewIntentError(common.ErrorCodeIntentNotFound, "Intent not tracked", intentID)
	}

	oldStatus := tracker.Status
	tracker.Status = status
	tracker.LastUpdated = time.Now()
	tracker.Intent.Status = status

	// Trigger status change callbacks
	lm.triggerStatusChangeCallbacks(tracker.Intent, oldStatus, status)

	return nil
}

// GetTracker retrieves the tracker for an intent
func (lm *LifecycleManager) GetTracker(intentID string) (*common.IntentTracker, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	tracker, exists := lm.trackedIntents[intentID]
	if !exists {
		return nil, common.NewIntentError(common.ErrorCodeIntentNotFound, "Intent not tracked", intentID)
	}

	return tracker, nil
}

// RegisterCallback registers a lifecycle callback
func (lm *LifecycleManager) RegisterCallback(callback common.LifecycleCallback) {
	if callback == nil {
		return
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Register callback for all status changes
	for status := range lm.callbacks {
		lm.callbacks[status] = append(lm.callbacks[status], callback)
	}

	// Initialize if empty
	if len(lm.callbacks) == 0 {
		lm.callbacks[common.IntentStatusCreated] = []common.LifecycleCallback{callback}
		lm.callbacks[common.IntentStatusValidated] = []common.LifecycleCallback{callback}
		lm.callbacks[common.IntentStatusBroadcasted] = []common.LifecycleCallback{callback}
		lm.callbacks[common.IntentStatusProcessed] = []common.LifecycleCallback{callback}
		lm.callbacks[common.IntentStatusMatched] = []common.LifecycleCallback{callback}
		lm.callbacks[common.IntentStatusCompleted] = []common.LifecycleCallback{callback}
		lm.callbacks[common.IntentStatusFailed] = []common.LifecycleCallback{callback}
		lm.callbacks[common.IntentStatusExpired] = []common.LifecycleCallback{callback}
	}
}

// cleanupLoop runs the cleanup process for expired intents
func (lm *LifecycleManager) cleanupLoop() {
	for {
		select {
		case <-lm.cleanupTicker.C:
			lm.cleanupExpiredIntents()
		case <-lm.stopCh:
			return
		}
	}
}

// cleanupExpiredIntents removes expired intents from tracking
func (lm *LifecycleManager) cleanupExpiredIntents() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now()
	expiredIntents := make([]string, 0)

	for intentID, tracker := range lm.trackedIntents {
		// Check TTL expiration
		if common.Times.IsExpired(tracker.Intent.Timestamp, tracker.Intent.TTL) {
			expiredIntents = append(expiredIntents, intentID)
			continue
		}

		// Check if intent is too old (beyond max age)
		if now.Sub(tracker.CreatedAt) > 24*time.Hour {
			expiredIntents = append(expiredIntents, intentID)
		}
	}

	// Remove expired intents and trigger callbacks
	for _, intentID := range expiredIntents {
		if tracker, exists := lm.trackedIntents[intentID]; exists {
			// Trigger expiration callbacks
			for _, callback := range lm.callbacks[common.IntentStatusExpired] {
				go func(cb common.LifecycleCallback, intent *common.Intent) {
					cb.OnExpired(intent)
				}(callback, tracker.Intent)
			}

			delete(lm.trackedIntents, intentID)
		}
	}
}

// scheduleExpiration schedules expiration for an intent
func (lm *LifecycleManager) scheduleExpiration(intent *common.Intent) {
	if intent.TTL <= 0 {
		return
	}

	// Calculate expiration time
	createdAt := time.Unix(intent.Timestamp, 0)
	expirationTime := createdAt.Add(time.Duration(intent.TTL) * time.Second)
	waitDuration := time.Until(expirationTime)

	if waitDuration <= 0 {
		// Already expired
		lm.UpdateStatus(intent.ID, common.IntentStatusExpired)
		return
	}

	// Schedule expiration
	time.AfterFunc(waitDuration, func() {
		lm.UpdateStatus(intent.ID, common.IntentStatusExpired)
	})
}

// triggerStatusChangeCallbacks triggers callbacks for status changes
func (lm *LifecycleManager) triggerStatusChangeCallbacks(intent *common.Intent, oldStatus, newStatus common.IntentStatus) {
	lm.mu.RLock()
	callbacks := lm.callbacks[newStatus]
	lm.mu.RUnlock()

	// Trigger callbacks asynchronously
	for _, callback := range callbacks {
		go func(cb common.LifecycleCallback, i *common.Intent, old, new common.IntentStatus) {
			if err := cb.OnStatusChange(i, old, new); err != nil {
				// Log callback error but don't fail the status change
				// In production, this should be logged properly
			}
		}(callback, intent, oldStatus, newStatus)
	}
}

// Stop stops the lifecycle manager
func (lm *LifecycleManager) Stop() {
	close(lm.stopCh)
	lm.cleanupTicker.Stop()
}
