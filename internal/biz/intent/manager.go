package intent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"pin_intent_broadcast_network/internal/biz/common"
	"pin_intent_broadcast_network/internal/transport"
)

// Manager implements the IntentManager interface
// It coordinates all intent-related operations including creation, validation, processing, and lifecycle management
type Manager struct {
	validator     common.IntentValidator
	signer        common.IntentSigner
	processor     common.IntentProcessor
	matcher       common.IntentMatcher
	lifecycle     common.LifecycleManager
	transportMgr  transport.TransportManager
	config        *Config
	metrics       *Metrics
	logger        *log.Helper
	mu            sync.RWMutex
	intents       map[string]*common.Intent      // In-memory intent storage for quick access
	subscriptions map[string]chan *common.Intent // Intent event subscriptions
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
	transportMgr transport.TransportManager,
	config *Config,
	logger log.Logger,
) *Manager {
	return &Manager{
		validator:     validator,
		signer:        signer,
		processor:     processor,
		matcher:       matcher,
		lifecycle:     lifecycle,
		transportMgr:  transportMgr,
		config:        config,
		metrics:       &Metrics{},
		logger:        log.NewHelper(logger),
		intents:       make(map[string]*common.Intent),
		subscriptions: make(map[string]chan *common.Intent),
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

// CreateIntent implements IntentManager.CreateIntent
func (m *Manager) CreateIntent(ctx context.Context, req *common.CreateIntentRequest) (*common.CreateIntentResponse, error) {
	startTime := time.Now()
	defer func() {
		m.updateMetrics(true, time.Since(startTime).Milliseconds())
	}()

	m.logger.Infof("Creating intent: type=%s, sender=%s", req.Type, req.SenderID)

	// Validate request
	if err := m.validateCreateRequest(req); err != nil {
		m.logger.Errorf("Create intent validation failed: %v", err)
		return &common.CreateIntentResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	// Create intent instance
	intent := &common.Intent{
		ID:        common.GenerateIntentID(),
		Type:      req.Type,
		Payload:   req.Payload,
		Timestamp: time.Now().Unix(),
		SenderID:  req.SenderID,
		Metadata:  req.Metadata,
		Status:    common.IntentStatusCreated,
		Priority:  req.Priority,
		TTL:       req.TTL,
	}

	// Set default values
	if intent.Priority == 0 {
		intent.Priority = common.PriorityNormal
	}
	if intent.TTL == 0 {
		intent.TTL = int64(common.DefaultTTL.Seconds())
	}

	// Validate intent through validator
	if m.validator != nil {
		if err := m.validator.ValidateIntent(ctx, intent); err != nil {
			m.logger.Errorf("Intent validation failed: %v", err)
			return &common.CreateIntentResponse{
				Success: false,
				Message: "Intent validation failed: " + err.Error(),
			}, err
		}
		intent.Status = common.IntentStatusValidated
	}

	// Sign intent if signer is available
	if m.signer != nil && req.PrivateKey != nil {
		if err := m.signer.SignIntent(intent, req.PrivateKey); err != nil {
			m.logger.Errorf("Intent signing failed: %v", err)
			return &common.CreateIntentResponse{
				Success: false,
				Message: "Intent signing failed: " + err.Error(),
			}, err
		}
	}

	// Store intent in memory
	m.mu.Lock()
	m.intents[intent.ID] = intent
	m.metrics.IntentsCreated++
	m.mu.Unlock()

	// Start lifecycle tracking
	if m.lifecycle != nil {
		m.lifecycle.StartTracking(intent)
	}

	m.logger.Infof("Intent created successfully: id=%s", intent.ID)

	return &common.CreateIntentResponse{
		Intent:  intent,
		Success: true,
		Message: "Intent created successfully",
	}, nil
}

// validateCreateRequest validates create intent request
func (m *Manager) validateCreateRequest(req *common.CreateIntentRequest) error {
	if req == nil {
		return common.NewIntentError(common.ErrorCodeInvalidFormat, "Request cannot be nil", "")
	}

	// Validate required fields
	if common.Strings.IsEmpty(req.Type) {
		return common.NewValidationError("type", req.Type, "Intent type cannot be empty")
	}

	if !common.Validation.IsValidIntentType(req.Type) {
		return common.NewValidationError("type", req.Type, "Invalid intent type")
	}

	if len(req.Payload) == 0 {
		return common.NewValidationError("payload", "", "Intent payload cannot be empty")
	}

	if !common.Validation.IsValidPayloadSize(len(req.Payload)) {
		return common.NewValidationError("payload", "", fmt.Sprintf("Payload size %d exceeds maximum allowed", len(req.Payload)))
	}

	if common.Strings.IsEmpty(req.SenderID) {
		return common.NewValidationError("sender_id", req.SenderID, "Sender ID cannot be empty")
	}

	if req.Priority != 0 && !common.Validation.IsValidPriority(req.Priority) {
		return common.NewValidationError("priority", fmt.Sprintf("%d", req.Priority), "Invalid priority value")
	}

	if req.TTL != 0 && !common.Validation.IsValidTTL(req.TTL) {
		return common.NewValidationError("ttl", fmt.Sprintf("%d", req.TTL), "Invalid TTL value")
	}

	return nil
}

// BroadcastIntent implements IntentManager.BroadcastIntent
func (m *Manager) BroadcastIntent(ctx context.Context, req *common.BroadcastIntentRequest) (*common.BroadcastIntentResponse, error) {
	startTime := time.Now()
	defer func() {
		m.updateMetrics(true, time.Since(startTime).Milliseconds())
	}()

	if req == nil || req.Intent == nil {
		return &common.BroadcastIntentResponse{
			Success: false,
			Message: "Invalid broadcast request: intent cannot be nil",
		}, common.NewIntentError(common.ErrorCodeInvalidFormat, "Invalid broadcast request", "")
	}

	intent := req.Intent
	m.logger.Infof("Broadcasting intent: id=%s, type=%s", intent.ID, intent.Type)

	// Check if transport manager is available
	if m.transportMgr == nil {
		m.logger.Error("Transport manager not available")
		return &common.BroadcastIntentResponse{
			Success:  false,
			IntentID: intent.ID,
			Message:  "Transport manager not available",
		}, fmt.Errorf("transport manager not initialized")
	}

	// Determine broadcast topic
	topic := req.Topic
	if topic == "" {
		topic = determineTopicByType(intent.Type)
	}

	// Update intent status to broadcasting
	m.mu.Lock()
	if storedIntent, exists := m.intents[intent.ID]; exists {
		storedIntent.Status = common.IntentStatusBroadcasted
		intent = storedIntent // Use stored intent to ensure consistency
	} else {
		// Intent not found in local storage, store it first
		intent.Status = common.IntentStatusBroadcasted
		m.intents[intent.ID] = intent
	}
	m.mu.Unlock()

	// Create transport message for intent
	transportMsg := &transport.TransportMessage{
		Type:      "intent_broadcast",
		Payload:   serializeIntentForTransport(intent),
		Timestamp: time.Now().UnixMilli(),
		Sender:    intent.SenderID,
		Priority:  intent.Priority,
		Metadata:  make(map[string]string),
	}

	// Generate message ID after setting other fields
	transportMsg.ID = transport.GenerateMessageID(transportMsg)

	// Add intent metadata to transport message
	transportMsg.Metadata["intent_id"] = intent.ID
	transportMsg.Metadata["intent_type"] = intent.Type
	transportMsg.Metadata["broadcast_topic"] = topic
	if intent.Metadata != nil {
		for k, v := range intent.Metadata {
			transportMsg.Metadata["intent_"+k] = v
		}
	}

	// Publish to P2P network
	if err := m.transportMgr.PublishMessage(ctx, topic, transportMsg); err != nil {
		m.logger.Errorf("Failed to broadcast intent %s: %v", intent.ID, err)

		// Revert intent status on failure
		m.mu.Lock()
		if storedIntent, exists := m.intents[intent.ID]; exists {
			storedIntent.Status = common.IntentStatusFailed
			storedIntent.Error = "Broadcast failed: " + err.Error()
		}
		m.mu.Unlock()

		return &common.BroadcastIntentResponse{
			Success:  false,
			IntentID: intent.ID,
			Topic:    topic,
			Message:  "Failed to broadcast intent: " + err.Error(),
		}, err
	}

	// Update lifecycle tracking
	if m.lifecycle != nil {
		m.lifecycle.UpdateStatus(intent.ID, common.IntentStatusBroadcasted)
	}

	// Notify subscribers
	m.notifySubscribers(intent)

	m.logger.Infof("Intent broadcasted successfully: id=%s, topic=%s", intent.ID, topic)

	return &common.BroadcastIntentResponse{
		Success:  true,
		IntentID: intent.ID,
		Topic:    topic,
		Message:  "Intent broadcasted successfully",
	}, nil
}

// notifySubscribers notifies all intent subscribers
func (m *Manager) notifySubscribers(intent *common.Intent) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for subID, ch := range m.subscriptions {
		select {
		case ch <- intent:
			m.logger.Debugf("Notified subscriber %s about intent %s", subID, intent.ID)
		default:
			m.logger.Warnf("Subscriber %s channel full, skipping notification for intent %s", subID, intent.ID)
		}
	}
}

// QueryIntents implements IntentManager.QueryIntents
func (m *Manager) QueryIntents(ctx context.Context, req *common.QueryIntentsRequest) (*common.QueryIntentsResponse, error) {
	m.logger.Debugf("Querying intents: type=%s, limit=%d", req.Type, req.Limit)

	m.mu.RLock()
	defer m.mu.RUnlock()

	var filteredIntents []*common.Intent

	// Filter intents based on request criteria
	for _, intent := range m.intents {
		// Filter by type if specified
		if req.Type != "" && intent.Type != req.Type {
			continue
		}

		// Filter by time range if specified
		if req.StartTime > 0 && intent.Timestamp < req.StartTime {
			continue
		}
		if req.EndTime > 0 && intent.Timestamp > req.EndTime {
			continue
		}

		filteredIntents = append(filteredIntents, intent)
	}

	// Apply offset and limit
	total := int32(len(filteredIntents))
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return &common.QueryIntentsResponse{
			Intents: []*common.Intent{},
			Total:   total,
		}, nil
	}

	// Calculate end index for slicing
	limit := req.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}
	endIndex := offset + limit
	if endIndex > total {
		endIndex = total
	}

	// Slice the results
	resultIntents := filteredIntents[offset:endIndex]

	return &common.QueryIntentsResponse{
		Intents: resultIntents,
		Total:   total,
	}, nil
}

// SubscribeIntents implements IntentManager.SubscribeIntents
func (m *Manager) SubscribeIntents(ctx context.Context, req *common.SubscribeIntentsRequest) (<-chan *common.Intent, error) {
	m.logger.Debugf("Creating intent subscription for types: %v", req.Types)

	// Create subscription channel
	subscriptionID := fmt.Sprintf("sub_%d", time.Now().UnixNano())
	intentChan := make(chan *common.Intent, 100) // Buffered channel

	m.mu.Lock()
	m.subscriptions[subscriptionID] = intentChan
	m.mu.Unlock()

	// Start a goroutine to handle subscription cleanup
	go func() {
		<-ctx.Done()

		m.mu.Lock()
		if ch, exists := m.subscriptions[subscriptionID]; exists {
			close(ch)
			delete(m.subscriptions, subscriptionID)
		}
		m.mu.Unlock()

		m.logger.Debugf("Subscription %s cleaned up", subscriptionID)
	}()

	// Subscribe to transport layer for incoming intents
	if m.transportMgr != nil && len(req.Topics) > 0 {
		for _, topic := range req.Topics {
			_, err := m.transportMgr.SubscribeToTopic(topic, func(msg *transport.TransportMessage) error {
				return m.handleIncomingIntentMessage(ctx, msg, req.Types, intentChan)
			})
			if err != nil {
				m.logger.Errorf("Failed to subscribe to topic %s: %v", topic, err)
			}
		}
	}

	return intentChan, nil
}

// handleIncomingIntentMessage handles incoming intent messages from transport layer
func (m *Manager) handleIncomingIntentMessage(ctx context.Context, msg *transport.TransportMessage, typeFilter []string, intentChan chan *common.Intent) error {
	if msg.Type != "intent_broadcast" {
		return nil // Not an intent message
	}

	// Deserialize intent from message payload
	intent, err := m.deserializeIntentFromTransport(msg.Payload)
	if err != nil {
		m.logger.Errorf("Failed to deserialize intent from transport message: %v", err)
		return err
	}

	// Apply type filter if specified
	if len(typeFilter) > 0 {
		typeMatched := false
		for _, allowedType := range typeFilter {
			if intent.Type == allowedType {
				typeMatched = true
				break
			}
		}
		if !typeMatched {
			return nil // Type not in filter
		}
	}

	// Process the received intent
	if err := m.ProcessIntent(ctx, intent); err != nil {
		m.logger.Errorf("Failed to process received intent %s: %v", intent.ID, err)
		return err
	}

	// Send to subscription channel
	select {
	case intentChan <- intent:
		m.logger.Debugf("Sent intent %s to subscription channel", intent.ID)
	default:
		m.logger.Warnf("Subscription channel full, dropping intent %s", intent.ID)
	}

	return nil
}

// deserializeIntentFromTransport deserializes intent from transport layer payload
func (m *Manager) deserializeIntentFromTransport(payload []byte) (*common.Intent, error) {
	var intent common.Intent
	if err := common.JSON.Unmarshal(payload, &intent); err != nil {
		return nil, fmt.Errorf("failed to deserialize intent: %w", err)
	}
	return &intent, nil
}

// SetTransportManager sets the transport manager for the intent manager
func (m *Manager) SetTransportManager(transportMgr transport.TransportManager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transportMgr = transportMgr
	m.logger.Info("Transport manager updated successfully")
}

// StartIntentSubscription starts subscribing to intent broadcast topics
func (m *Manager) StartIntentSubscription(ctx context.Context) error {
	if m.transportMgr == nil {
		m.logger.Warn("Transport manager not available, skipping intent subscription")
		return nil
	}

	// Subscribe to all intent broadcast topics
	topics := []string{
		"intent-broadcast.trade",
		"intent-broadcast.swap",
		"intent-broadcast.exchange",
		"intent-broadcast.transfer",
		"intent-broadcast.general",
	}

	for _, topic := range topics {
		_, err := m.transportMgr.SubscribeToTopic(topic, func(msg *transport.TransportMessage) error {
			return m.handleIncomingIntentBroadcast(ctx, msg)
		})
		if err != nil {
			m.logger.Errorf("Failed to subscribe to topic %s: %v", topic, err)
			continue
		}
		m.logger.Infof("Subscribed to intent broadcast topic: %s", topic)
	}

	return nil
}

// handleIncomingIntentBroadcast handles incoming intent broadcast messages
func (m *Manager) handleIncomingIntentBroadcast(ctx context.Context, msg *transport.TransportMessage) error {
	if msg.Type != "intent_broadcast" {
		return nil // Not an intent broadcast message
	}

	// Deserialize intent from message payload
	intent, err := deserializeIntentFromTransport(msg.Payload)
	if err != nil {
		m.logger.Errorf("Failed to deserialize intent from transport message: %v", err)
		return err
	}

	// Check if we already have this intent
	m.mu.Lock()
	if _, exists := m.intents[intent.ID]; exists {
		m.mu.Unlock()
		m.logger.Debugf("Intent %s already exists, skipping", intent.ID)
		return nil
	}

	// Store the received intent
	intent.Status = common.IntentStatusReceived
	m.intents[intent.ID] = intent
	m.mu.Unlock()

	m.logger.Infof("Received intent broadcast: id=%s, type=%s, sender=%s", intent.ID, intent.Type, intent.SenderID)

	// Notify local subscribers
	m.notifySubscribers(intent)

	return nil
}

// serializeIntentForTransport serializes an intent for transport
func serializeIntentForTransport(intent *common.Intent) []byte {
	// Use JSON serialization for now
	data, err := common.JSON.Marshal(intent)
	if err != nil {
		return nil
	}
	return data
}

// deserializeIntentFromTransport deserializes an intent from transport data
func deserializeIntentFromTransport(data []byte) (*common.Intent, error) {
	var intent common.Intent
	if err := common.JSON.Unmarshal(data, &intent); err != nil {
		return nil, err
	}
	return &intent, nil
}
