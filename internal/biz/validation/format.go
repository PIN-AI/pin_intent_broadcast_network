package validation

import (
	"context"
	"fmt"
	"regexp"

	"pin_intent_broadcast_network/internal/biz/common"
)

// FormatValidator implements format validation for intents
// This file will contain the implementation for task 3.2
type FormatValidator struct {
	rules []common.FormatRule
}

// NewFormatValidator creates a new format validator
func NewFormatValidator() *FormatValidator {
	return &FormatValidator{
		rules: make([]common.FormatRule, 0),
	}
}

// ValidateFormat validates the basic format of an intent
func (fv *FormatValidator) ValidateFormat(intent *common.Intent) error {
	if intent == nil {
		return common.NewValidationError("intent", "", "Intent cannot be nil")
	}

	// Basic field validation
	if common.Strings.IsEmpty(intent.ID) {
		return common.NewValidationError("id", intent.ID, "Intent ID cannot be empty")
	}

	if common.Strings.IsEmpty(intent.Type) {
		return common.NewValidationError("type", intent.Type, "Intent type cannot be empty")
	}

	if len(intent.Payload) == 0 {
		return common.NewValidationError("payload", "", "Intent payload cannot be empty")
	}

	if intent.Timestamp <= 0 {
		return common.NewValidationError("timestamp", fmt.Sprintf("%d", intent.Timestamp),
			"Intent timestamp must be positive")
	}

	if common.Strings.IsEmpty(intent.SenderID) {
		return common.NewValidationError("sender_id", intent.SenderID, "Intent sender ID cannot be empty")
	}

	// Validate ID format (should be hex string)
	if !isValidHexString(intent.ID) {
		return common.NewValidationError("id", intent.ID, "Intent ID must be a valid hex string")
	}

	// Validate intent type format
	if !common.Validation.IsValidIntentType(intent.Type) {
		return common.NewValidationError("type", intent.Type, "Invalid intent type")
	}

	// Validate priority if set
	if intent.Priority != 0 && !common.Validation.IsValidPriority(intent.Priority) {
		return common.NewValidationError("priority", fmt.Sprintf("%d", intent.Priority),
			"Invalid priority value")
	}

	// Validate TTL if set
	if intent.TTL != 0 && !common.Validation.IsValidTTL(intent.TTL) {
		return common.NewValidationError("ttl", fmt.Sprintf("%d", intent.TTL),
			"Invalid TTL value")
	}

	// Validate status
	if !isValidIntentStatus(intent.Status) {
		return common.NewValidationError("status", intent.Status.String(),
			"Invalid intent status")
	}

	// Apply custom format rules
	for _, rule := range fv.rules {
		if err := rule.ValidateFormat(intent); err != nil {
			return common.WrapError(err, common.ErrorCodeInvalidFormat,
				fmt.Sprintf("Format rule '%s' failed", rule.Name()))
		}
	}

	return nil
}

// AddRule adds a format validation rule
func (fv *FormatValidator) AddRule(rule common.FormatRule) {
	fv.rules = append(fv.rules, rule)
}

// BasicFormatRule implements basic format validation
type BasicFormatRule struct {
	name string
}

// NewBasicFormatRule creates a new basic format rule
func NewBasicFormatRule() *BasicFormatRule {
	return &BasicFormatRule{name: "basic_format"}
}

// Name returns the rule name
func (r *BasicFormatRule) Name() string {
	return r.name
}

// Validate validates the intent format
func (r *BasicFormatRule) Validate(ctx context.Context, intent *common.Intent) error {
	// TODO: Implement basic validation logic
	return nil
}

// GetPriority returns the rule priority
func (r *BasicFormatRule) GetPriority() int {
	return 100 // High priority for basic validation
}

// ValidateFormat validates the format of an intent
func (r *BasicFormatRule) ValidateFormat(intent *common.Intent) error {
	// This is the basic format rule implementation
	// More specific validation is done in the main ValidateFormat method

	// Validate metadata format if present
	if intent.Metadata != nil {
		for key, value := range intent.Metadata {
			if common.Strings.IsEmpty(key) {
				return common.NewValidationError("metadata", key, "Metadata key cannot be empty")
			}
			if len(key) > 100 {
				return common.NewValidationError("metadata", key, "Metadata key too long")
			}
			if len(value) > 1000 {
				return common.NewValidationError("metadata", value, "Metadata value too long")
			}
		}
	}

	// Validate signature format if present
	if len(intent.Signature) > 0 {
		if common.Strings.IsEmpty(intent.SignatureAlgorithm) {
			return common.NewValidationError("signature_algorithm", intent.SignatureAlgorithm,
				"Signature algorithm must be specified when signature is present")
		}
	}

	return nil
}

// isValidHexString checks if a string is a valid hexadecimal string
func isValidHexString(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Check if string contains only hex characters
	matched, _ := regexp.MatchString("^[a-fA-F0-9]+$", s)
	return matched && len(s)%2 == 0 // Must be even length
}

// isValidIntentStatus checks if an intent status is valid
func isValidIntentStatus(status common.IntentStatus) bool {
	switch status {
	case common.IntentStatusCreated,
		common.IntentStatusValidated,
		common.IntentStatusBroadcasted,
		common.IntentStatusProcessed,
		common.IntentStatusMatched,
		common.IntentStatusCompleted,
		common.IntentStatusFailed,
		common.IntentStatusExpired:
		return true
	default:
		return false
	}
}

// JSONSchemaRule implements JSON schema validation
type JSONSchemaRule struct {
	name   string
	schema map[string]interface{}
}

// NewJSONSchemaRule creates a new JSON schema validation rule
func NewJSONSchemaRule(schema map[string]interface{}) *JSONSchemaRule {
	return &JSONSchemaRule{
		name:   "json_schema",
		schema: schema,
	}
}

// Name returns the rule name
func (r *JSONSchemaRule) Name() string {
	return r.name
}

// Validate validates the intent against JSON schema
func (r *JSONSchemaRule) Validate(ctx context.Context, intent *common.Intent) error {
	return r.ValidateFormat(intent)
}

// GetPriority returns the rule priority
func (r *JSONSchemaRule) GetPriority() int {
	return 80
}

// ValidateFormat validates the intent format against JSON schema
func (r *JSONSchemaRule) ValidateFormat(intent *common.Intent) error {
	// TODO: Implement JSON schema validation
	// This would require a JSON schema validation library
	// For now, we'll do basic structure validation

	if r.schema == nil {
		return nil
	}

	// Basic validation against schema structure
	// In a real implementation, this would use a proper JSON schema validator
	return nil
}

// PayloadFormatRule validates payload format
type PayloadFormatRule struct {
	name        string
	maxSize     int
	allowedMIME []string
}

// NewPayloadFormatRule creates a new payload format rule
func NewPayloadFormatRule(maxSize int, allowedMIME []string) *PayloadFormatRule {
	return &PayloadFormatRule{
		name:        "payload_format",
		maxSize:     maxSize,
		allowedMIME: allowedMIME,
	}
}

// Name returns the rule name
func (r *PayloadFormatRule) Name() string {
	return r.name
}

// Validate validates the intent
func (r *PayloadFormatRule) Validate(ctx context.Context, intent *common.Intent) error {
	return r.ValidateFormat(intent)
}

// GetPriority returns the rule priority
func (r *PayloadFormatRule) GetPriority() int {
	return 70
}

// ValidateFormat validates the payload format
func (r *PayloadFormatRule) ValidateFormat(intent *common.Intent) error {
	// Check payload size
	if r.maxSize > 0 && len(intent.Payload) > r.maxSize {
		return common.NewValidationError("payload", "",
			fmt.Sprintf("Payload size %d exceeds maximum %d", len(intent.Payload), r.maxSize))
	}

	// Check if payload is valid JSON (basic check)
	if len(intent.Payload) > 0 {
		var temp interface{}
		if err := common.JSON.Unmarshal(intent.Payload, &temp); err != nil {
			return common.NewValidationError("payload", "", "Payload must be valid JSON")
		}
	}

	return nil
}
