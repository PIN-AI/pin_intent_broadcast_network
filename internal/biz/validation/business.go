package validation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pin_intent_broadcast_network/internal/biz/common"
)

// BusinessValidator implements business rule validation for intents
// This file will contain the implementation for task 3.3
type BusinessValidator struct {
	rules  []common.BusinessRule
	config *BusinessConfig
}

// BusinessConfig holds configuration for business validation
type BusinessConfig struct {
	MaxPayloadSize int           `yaml:"max_payload_size"`
	MaxTTL         time.Duration `yaml:"max_ttl"`
	AllowedTypes   []string      `yaml:"allowed_types"`
}

// NewBusinessValidator creates a new business validator
func NewBusinessValidator(config *BusinessConfig) *BusinessValidator {
	return &BusinessValidator{
		rules:  make([]common.BusinessRule, 0),
		config: config,
	}
}

// ValidateBusinessRules validates business rules for an intent
func (bv *BusinessValidator) ValidateBusinessRules(ctx context.Context, intent *common.Intent) error {
	if intent == nil {
		return common.NewValidationError("intent", "", "Intent cannot be nil")
	}

	// Check if Intent type is allowed
	if !bv.isTypeAllowed(intent.Type) {
		return common.NewValidationError("type", intent.Type,
			fmt.Sprintf("Intent type %s is not allowed", intent.Type))
	}

	// Check payload size
	if bv.config.MaxPayloadSize > 0 && len(intent.Payload) > bv.config.MaxPayloadSize {
		return common.NewValidationError("payload", "",
			fmt.Sprintf("Payload size %d exceeds maximum %d", len(intent.Payload), bv.config.MaxPayloadSize))
	}

	// Check TTL
	if bv.config.MaxTTL > 0 && intent.TTL > int64(bv.config.MaxTTL.Seconds()) {
		return common.NewValidationError("ttl", fmt.Sprintf("%d", intent.TTL),
			fmt.Sprintf("TTL %d exceeds maximum %d", intent.TTL, int64(bv.config.MaxTTL.Seconds())))
	}

	// Check time validity (intent shouldn't be too old)
	intentTime := time.Unix(intent.Timestamp, 0)
	if time.Since(intentTime) > time.Hour {
		return common.NewValidationError("timestamp", fmt.Sprintf("%d", intent.Timestamp),
			"Intent is too old (more than 1 hour)")
	}

	// Check time validity (intent shouldn't be from future)
	if intentTime.After(time.Now().Add(5 * time.Minute)) {
		return common.NewValidationError("timestamp", fmt.Sprintf("%d", intent.Timestamp),
			"Intent timestamp is too far in the future")
	}

	// Check priority bounds
	if intent.Priority < common.PriorityLow || intent.Priority > common.PriorityUrgent {
		return common.NewValidationError("priority", fmt.Sprintf("%d", intent.Priority),
			"Priority must be between 1 and 20")
	}

	// Check TTL minimum value
	if intent.TTL > 0 && intent.TTL < 60 {
		return common.NewValidationError("ttl", fmt.Sprintf("%d", intent.TTL),
			"TTL must be at least 60 seconds")
	}

	// Apply custom business rules
	for _, rule := range bv.rules {
		if err := rule.ValidateBusinessLogic(ctx, intent); err != nil {
			return common.WrapError(err, common.ErrorCodeValidationFailed,
				fmt.Sprintf("Business rule '%s' failed", rule.Name()))
		}
	}

	return nil
}

// isTypeAllowed checks if an intent type is allowed
func (bv *BusinessValidator) isTypeAllowed(intentType string) bool {
	if len(bv.config.AllowedTypes) == 0 {
		return true // No restrictions
	}

	for _, allowedType := range bv.config.AllowedTypes {
		if allowedType == intentType {
			return true
		}
	}

	return false
}

// AddRule adds a business validation rule
func (bv *BusinessValidator) AddRule(rule common.BusinessRule) {
	bv.rules = append(bv.rules, rule)
}

// PayloadSizeRule implements payload size validation
type PayloadSizeRule struct {
	name    string
	maxSize int
}

// NewPayloadSizeRule creates a new payload size rule
func NewPayloadSizeRule(maxSize int) *PayloadSizeRule {
	return &PayloadSizeRule{
		name:    "payload_size",
		maxSize: maxSize,
	}
}

// Name returns the rule name
func (r *PayloadSizeRule) Name() string {
	return r.name
}

// Validate validates the intent
func (r *PayloadSizeRule) Validate(ctx context.Context, intent *common.Intent) error {
	return r.ValidateBusinessLogic(ctx, intent)
}

// GetPriority returns the rule priority
func (r *PayloadSizeRule) GetPriority() int {
	return 80
}

// ValidateBusinessLogic validates the business logic
func (r *PayloadSizeRule) ValidateBusinessLogic(ctx context.Context, intent *common.Intent) error {
	if len(intent.Payload) > r.maxSize {
		return fmt.Errorf("payload size %d exceeds maximum %d", len(intent.Payload), r.maxSize)
	}
	return nil
}

// TTLRule implements TTL validation
type TTLRule struct {
	name   string
	maxTTL int64
}

// NewTTLRule creates a new TTL validation rule
func NewTTLRule(maxTTL int64) *TTLRule {
	return &TTLRule{
		name:   "ttl_validation",
		maxTTL: maxTTL,
	}
}

// Name returns the rule name
func (r *TTLRule) Name() string {
	return r.name
}

// Validate validates the intent
func (r *TTLRule) Validate(ctx context.Context, intent *common.Intent) error {
	return r.ValidateBusinessLogic(ctx, intent)
}

// GetPriority returns the rule priority
func (r *TTLRule) GetPriority() int {
	return 85
}

// ValidateBusinessLogic validates the TTL business logic
func (r *TTLRule) ValidateBusinessLogic(ctx context.Context, intent *common.Intent) error {
	if intent.TTL > r.maxTTL {
		return common.NewValidationError("ttl", fmt.Sprintf("%d", intent.TTL),
			fmt.Sprintf("TTL %d exceeds maximum %d", intent.TTL, r.maxTTL))
	}

	if intent.TTL <= 0 {
		return common.NewValidationError("ttl", fmt.Sprintf("%d", intent.TTL),
			"TTL must be positive")
	}

	return nil
}

// TypeWhitelistRule implements intent type whitelist validation
type TypeWhitelistRule struct {
	name         string
	allowedTypes []string
}

// NewTypeWhitelistRule creates a new type whitelist rule
func NewTypeWhitelistRule(allowedTypes []string) *TypeWhitelistRule {
	return &TypeWhitelistRule{
		name:         "type_whitelist",
		allowedTypes: allowedTypes,
	}
}

// Name returns the rule name
func (r *TypeWhitelistRule) Name() string {
	return r.name
}

// Validate validates the intent
func (r *TypeWhitelistRule) Validate(ctx context.Context, intent *common.Intent) error {
	return r.ValidateBusinessLogic(ctx, intent)
}

// GetPriority returns the rule priority
func (r *TypeWhitelistRule) GetPriority() int {
	return 90
}

// ValidateBusinessLogic validates the type whitelist business logic
func (r *TypeWhitelistRule) ValidateBusinessLogic(ctx context.Context, intent *common.Intent) error {
	if len(r.allowedTypes) == 0 {
		return nil // No restrictions
	}

	for _, allowedType := range r.allowedTypes {
		if intent.Type == allowedType {
			return nil
		}
	}

	return common.NewValidationError("type", intent.Type,
		fmt.Sprintf("Intent type %s is not in allowed types: %v", intent.Type, r.allowedTypes))
}

// RateLimitRule implements rate limiting validation
type RateLimitRule struct {
	name         string
	maxPerMinute int
	senderCounts map[string]*RateCounter
	mu           sync.RWMutex
}

// RateCounter tracks rate limiting for a sender
type RateCounter struct {
	Count     int
	ResetTime time.Time
}

// NewRateLimitRule creates a new rate limit rule
func NewRateLimitRule(maxPerMinute int) *RateLimitRule {
	return &RateLimitRule{
		name:         "rate_limit",
		maxPerMinute: maxPerMinute,
		senderCounts: make(map[string]*RateCounter),
	}
}

// Name returns the rule name
func (r *RateLimitRule) Name() string {
	return r.name
}

// Validate validates the intent
func (r *RateLimitRule) Validate(ctx context.Context, intent *common.Intent) error {
	return r.ValidateBusinessLogic(ctx, intent)
}

// GetPriority returns the rule priority
func (r *RateLimitRule) GetPriority() int {
	return 95
}

// ValidateBusinessLogic validates the rate limit business logic
func (r *RateLimitRule) ValidateBusinessLogic(ctx context.Context, intent *common.Intent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	counter, exists := r.senderCounts[intent.SenderID]

	if !exists || now.After(counter.ResetTime) {
		// Create new counter or reset expired counter
		r.senderCounts[intent.SenderID] = &RateCounter{
			Count:     1,
			ResetTime: now.Add(time.Minute),
		}
		return nil
	}

	if counter.Count >= r.maxPerMinute {
		return common.NewValidationError("rate_limit", intent.SenderID,
			fmt.Sprintf("Rate limit exceeded: %d intents per minute", r.maxPerMinute))
	}

	counter.Count++
	return nil
}

// DuplicateRule implements duplicate intent detection
type DuplicateRule struct {
	name          string
	seenIDs       map[string]time.Time
	mu            sync.RWMutex
	cleanupTicker *time.Ticker
}

// NewDuplicateRule creates a new duplicate detection rule
func NewDuplicateRule() *DuplicateRule {
	rule := &DuplicateRule{
		name:    "duplicate_detection",
		seenIDs: make(map[string]time.Time),
	}

	// Start cleanup ticker
	rule.cleanupTicker = time.NewTicker(5 * time.Minute)
	go rule.cleanup()

	return rule
}

// Name returns the rule name
func (r *DuplicateRule) Name() string {
	return r.name
}

// Validate validates the intent
func (r *DuplicateRule) Validate(ctx context.Context, intent *common.Intent) error {
	return r.ValidateBusinessLogic(ctx, intent)
}

// GetPriority returns the rule priority
func (r *DuplicateRule) GetPriority() int {
	return 100 // Highest priority
}

// ValidateBusinessLogic validates for duplicate intents
func (r *DuplicateRule) ValidateBusinessLogic(ctx context.Context, intent *common.Intent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if seenTime, exists := r.seenIDs[intent.ID]; exists {
		return common.NewValidationError("duplicate", intent.ID,
			fmt.Sprintf("Duplicate intent ID detected, first seen at %v", seenTime))
	}

	r.seenIDs[intent.ID] = time.Now()
	return nil
}

// cleanup removes old entries from the seen IDs map
func (r *DuplicateRule) cleanup() {
	for {
		select {
		case <-r.cleanupTicker.C:
			r.mu.Lock()
			cutoff := time.Now().Add(-time.Hour) // Keep entries for 1 hour
			for id, seenTime := range r.seenIDs {
				if seenTime.Before(cutoff) {
					delete(r.seenIDs, id)
				}
			}
			r.mu.Unlock()
		}
	}
}

// Stop stops the duplicate rule cleanup
func (r *DuplicateRule) Stop() {
	if r.cleanupTicker != nil {
		r.cleanupTicker.Stop()
	}
}
