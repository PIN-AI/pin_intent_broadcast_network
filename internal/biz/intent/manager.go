package intent

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"pin_intent_broadcast_network/internal/biz/common"
)

// Manager implements the IntentManager interface
// It coordinates all intent-related operations including creation, validation, processing, and lifecycle management
type Manager struct {
	validator common.IntentValidator
	signer    common.IntentSigner
	processor common.IntentProcessor
	matcher   common.IntentMatcher
	lifecycle common.LifecycleManager
	config    *Config
	metrics   *Metrics
	logger    *log.Helper
	mu        sync.RWMutex
	intents   map[string]*common.Intent // In-memory intent storage for quick access
}

// Config holds configuration for the Intent Manager
type Config struct {
	MaxConcurrentIntents int           `yaml:"max_concurrent_intents"`
	ProcessingTimeout    time.Duration `yaml:"processing_timeout"`
	RetryAttempts        int           `yaml:"retry_attempts"`
	EnableMatching       bool          `yaml:"enable_matching"`
	DefaultTTL           time.Duration `yaml:"default_ttl"`
}

// Metrics holds metrics for intent operations
type Metrics struct {
	IntentsCreated   int64
	IntentsProcessed int64
	IntentsFailed    int64
	ProcessingTime   time.Duration
}

// NewManager creates a new Intent Manager instance
func NewManager(
	validator common.IntentValidator,
	signer common.IntentSigner,
	processor common.IntentProcessor,
	matcher common.IntentMatcher,
	lifecycle common.LifecycleManager,
	config *Config,
	logger log.Logger,
) *Manager {
	return &Manager{
		validator: validator,
		signer:    signer,
		processor: processor,
		matcher:   matcher,
		lifecycle: lifecycle,
		config:    config,
		metrics:   &Metrics{},
		logger:    log.NewHelper(logger),
		intents:   make(map[string]*common.Intent),
	}
}

// ProcessIntent processes an incoming intent
func (m *Manager) ProcessIntent(ctx context.Context, intent *common.Intent) error {
	startTime := time.Now()
	defer func() {
		m.metrics.ProcessingTime = time.Since(startTime)
		m.updateMetrics(true, time.Since(startTime).Milliseconds())
	}()

	m.logger.Infof("Processing intent: %s, type: %s", intent.ID, intent.Type)

	// Validate intent
	if err := m.validator.ValidateIntent(ctx, intent); err != nil {
		m.logger.Errorf("Intent validation failed: %v", err)
		m.updateMetrics(false, time.Since(startTime).Milliseconds())
		return common.WrapError(err, common.ErrorCodeValidationFailed, "Intent validation failed")
	}

	// Process through pipeline
	if err := m.processor.ProcessIncomingIntent(ctx, intent); err != nil {
		m.logger.Errorf("Intent processing failed: %v", err)
		m.updateMetrics(false, time.Since(startTime).Milliseconds())
		return common.WrapError(err, common.ErrorCodeProcessingFailed, "Intent processing failed")
	}

	// Update intent status
	intent.Status = common.IntentStatusProcessed
	intent.ProcessedAt = time.Now().Unix()

	// Store in memory for quick access
	m.mu.Lock()
	m.intents[intent.ID] = intent
	m.mu.Unlock()

	// Start lifecycle tracking
	m.lifecycle.StartTracking(intent)

	// Attempt matching if enabled
	if m.config.EnableMatching {
		go m.attemptMatching(ctx, intent)
	}

	m.logger.Infof("Intent processed successfully: %s", intent.ID)
	return nil
}

// GetIntentStatus retrieves the current status of an intent
func (m *Manager) GetIntentStatus(ctx context.Context, id string) (*common.Intent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	intent, exists := m.intents[id]
	if !exists {
		return nil, common.ErrIntentNotFound
	}

	return intent, nil
}

// CancelIntent cancels an intent
func (m *Manager) CancelIntent(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	intent, exists := m.intents[id]
	if !exists {
		return common.ErrIntentNotFound
	}

	// Update status to failed
	intent.Status = common.IntentStatusFailed
	intent.Error = "Intent cancelled by user"

	// Stop lifecycle tracking
	m.lifecycle.StopTracking(id)

	return nil
}

// updateMetrics updates processing metrics
func (m *Manager) updateMetrics(success bool, latency int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if success {
		m.metrics.IntentsProcessed++
	} else {
		m.metrics.IntentsFailed++
	}

	// Update average latency
	totalProcessed := m.metrics.IntentsProcessed + m.metrics.IntentsFailed
	if totalProcessed > 0 {
		m.metrics.ProcessingTime = time.Duration((int64(m.metrics.ProcessingTime)*(totalProcessed-1) + latency*int64(time.Millisecond)) / totalProcessed)
	}
}

// attemptMatching attempts to find matches for an intent
func (m *Manager) attemptMatching(ctx context.Context, intent *common.Intent) {
	if m.matcher == nil {
		return
	}

	m.logger.Debugf("Attempting to match intent: %s", intent.ID)

	// Get candidate intents for matching
	candidates := m.getCandidateIntents(intent)
	if len(candidates) == 0 {
		m.logger.Debugf("No candidates found for intent: %s", intent.ID)
		return
	}

	// Find matches
	matches, err := m.matcher.FindMatches(ctx, intent, candidates)
	if err != nil {
		m.logger.Errorf("Matching failed for intent %s: %v", intent.ID, err)
		return
	}

	if len(matches) > 0 {
		m.logger.Infof("Found %d matches for intent: %s", len(matches), intent.ID)

		// Update intent with matches
		m.mu.Lock()
		if storedIntent, exists := m.intents[intent.ID]; exists {
			storedIntent.Status = common.IntentStatusMatched
			for _, match := range matches {
				// Add matched intent IDs (this would need to be extracted from match details)
				if matchedID, ok := match.Details["matched_intent_id"].(string); ok {
					storedIntent.MatchedIntents = append(storedIntent.MatchedIntents, matchedID)
				}
			}
		}
		m.mu.Unlock()

		// Update lifecycle status
		m.lifecycle.UpdateStatus(intent.ID, common.IntentStatusMatched)

		// TODO: Add IntentsMatched metric
	}
}

// getCandidateIntents returns candidate intents for matching
func (m *Manager) getCandidateIntents(intent *common.Intent) []*common.Intent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	candidates := make([]*common.Intent, 0)

	// Find intents of compatible types that are not expired
	for _, candidate := range m.intents {
		if candidate.ID == intent.ID {
			continue // Skip self
		}

		// Check if candidate is still active
		if candidate.Status == common.IntentStatusExpired ||
			candidate.Status == common.IntentStatusCompleted ||
			candidate.Status == common.IntentStatusFailed {
			continue
		}

		// Check TTL expiration
		if common.Times.IsExpired(candidate.Timestamp, candidate.TTL) {
			continue
		}

		// Add to candidates
		candidates = append(candidates, candidate)
	}

	return candidates
}

// cleanupExpiredIntents removes expired intents from memory
func (m *Manager) cleanupExpiredIntents() {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiredCount := 0

	for id, intent := range m.intents {
		if common.Times.IsExpired(intent.Timestamp, intent.TTL) {
			intent.Status = common.IntentStatusExpired
			m.lifecycle.UpdateStatus(id, common.IntentStatusExpired)
			delete(m.intents, id)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		m.logger.Infof("Cleaned up %d expired intents", expiredCount)
		// TODO: Add IntentsExpired metric
	}
}

// Start starts the intent manager background processes
func (m *Manager) Start(ctx context.Context) error {
	m.logger.Info("Starting Intent Manager")

	// Start cleanup ticker
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.cleanupExpiredIntents()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// Stop stops the intent manager
func (m *Manager) Stop(ctx context.Context) error {
	m.logger.Info("Stopping Intent Manager")
	return nil
}
