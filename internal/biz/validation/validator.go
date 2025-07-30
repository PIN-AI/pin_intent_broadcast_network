package validation

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/libp2p/go-libp2p/core/peer"

	"pin_intent_broadcast_network/internal/biz/common"
)

// Validator implements the IntentValidator interface
// It coordinates all validation operations including format, business rules, and permissions
type Validator struct {
	formatRules   []common.FormatRule
	businessRules []common.BusinessRule
	permissionMgr PermissionManager
	config        *Config
	logger        *log.Helper
	mu            sync.RWMutex
}

// Config holds configuration for the validator
type Config struct {
	EnableStrict    bool     `yaml:"enable_strict"`
	MaxPayloadSize  int      `yaml:"max_payload_size"`
	MaxTTL          int64    `yaml:"max_ttl"`
	AllowedTypes    []string `yaml:"allowed_types"`
	ValidationCache bool     `yaml:"validation_cache"`
}

// NewValidator creates a new validator instance
func NewValidator(config *Config, logger log.Logger) *Validator {
	return &Validator{
		formatRules:   make([]common.FormatRule, 0),
		businessRules: make([]common.BusinessRule, 0),
		config:        config,
		logger:        log.NewHelper(logger),
	}
}

// ValidateIntent validates an intent using all validation rules
func (v *Validator) ValidateIntent(ctx context.Context, intent *common.Intent) error {
	if intent == nil {
		return common.NewValidationError("intent", "", "Intent cannot be nil")
	}

	v.logger.Debugf("Validating intent: %s, type: %s", intent.ID, intent.Type)

	// 1. Format validation
	if err := v.ValidateFormat(intent); err != nil {
		v.logger.Errorf("Format validation failed for intent %s: %v", intent.ID, err)
		return common.WrapError(err, common.ErrorCodeValidationFailed, "Format validation failed")
	}

	// 2. Business rules validation
	if err := v.ValidateBusinessRules(ctx, intent); err != nil {
		v.logger.Errorf("Business rules validation failed for intent %s: %v", intent.ID, err)
		return common.WrapError(err, common.ErrorCodeValidationFailed, "Business rules validation failed")
	}

	// 3. Permission validation (if sender ID is available)
	if intent.SenderID != "" {
		if err := v.ValidatePermissions(intent, peer.ID(intent.SenderID)); err != nil {
			v.logger.Errorf("Permission validation failed for intent %s: %v", intent.ID, err)
			return common.WrapError(err, common.ErrorCodePermissionDenied, "Permission validation failed")
		}
	}

	v.logger.Debugf("Intent validation successful: %s", intent.ID)
	return nil
}

// ValidateFormat validates the format of an intent
func (v *Validator) ValidateFormat(intent *common.Intent) error {
	// Create format validator if not exists
	if len(v.formatRules) == 0 {
		v.addDefaultFormatRules()
	}

	// Apply all format rules
	for _, rule := range v.formatRules {
		if err := rule.ValidateFormat(intent); err != nil {
			return common.WrapError(err, common.ErrorCodeInvalidFormat,
				fmt.Sprintf("Format rule '%s' failed", rule.Name()))
		}
	}

	return nil
}

// ValidateBusinessRules validates business rules for an intent
func (v *Validator) ValidateBusinessRules(ctx context.Context, intent *common.Intent) error {
	// Create business validator if not exists
	if len(v.businessRules) == 0 {
		v.addDefaultBusinessRules()
	}

	// Apply all business rules
	for _, rule := range v.businessRules {
		if err := rule.ValidateBusinessLogic(ctx, intent); err != nil {
			return common.WrapError(err, common.ErrorCodeValidationFailed,
				fmt.Sprintf("Business rule '%s' failed", rule.Name()))
		}
	}

	return nil
}

// ValidatePermissions validates permissions for an intent
func (v *Validator) ValidatePermissions(intent *common.Intent, sender peer.ID) error {
	if v.permissionMgr == nil {
		v.permissionMgr = NewPermissionManager(&PermissionConfig{
			EnablePermissionCheck: v.config.EnableStrict,
		})
	}

	return v.permissionMgr.CheckPermission(intent, sender)
}

// RegisterRule registers a validation rule
func (v *Validator) RegisterRule(rule common.ValidationRule) error {
	if rule == nil {
		return common.NewIntentError(common.ErrorCodeInvalidConfiguration, "Rule cannot be nil", "")
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	// Check if it's a format rule
	if formatRule, ok := rule.(common.FormatRule); ok {
		v.formatRules = append(v.formatRules, formatRule)
		v.sortFormatRulesByPriority()
	}

	// Check if it's a business rule
	if businessRule, ok := rule.(common.BusinessRule); ok {
		v.businessRules = append(v.businessRules, businessRule)
		v.sortBusinessRulesByPriority()
	}

	v.logger.Infof("Registered validation rule: %s", rule.Name())
	return nil
}

// PermissionManager manages permission validation
type PermissionManager struct {
	permissions map[string][]Permission
	config      *PermissionConfig
	mu          sync.RWMutex
}

// PermissionConfig holds configuration for permission validation
type PermissionConfig struct {
	EnablePermissionCheck bool     `yaml:"enable_permission_check"`
	DefaultPermissions    []string `yaml:"default_permissions"`
	AdminPeers            []string `yaml:"admin_peers"`
}

// Permission represents a permission rule
type Permission struct {
	Subject string // peer ID or "*" for all
	Action  string // action type or "*" for all
	Object  string // intent type or "*" for all
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager(config *PermissionConfig) *PermissionManager {
	if config == nil {
		config = &PermissionConfig{
			EnablePermissionCheck: false,
		}
	}

	return &PermissionManager{
		permissions: make(map[string][]Permission),
		config:      config,
	}
}

// CheckPermission checks if a sender has permission for an intent
func (pm *PermissionManager) CheckPermission(intent *common.Intent, sender peer.ID) error {
	if !pm.config.EnablePermissionCheck {
		return nil // Permission check disabled
	}

	pm.mu.RLock()
	permissions, exists := pm.permissions[intent.Type]
	pm.mu.RUnlock()

	if !exists {
		return nil // No specific permissions required
	}

	senderStr := sender.String()

	// Check if sender is admin
	for _, adminPeer := range pm.config.AdminPeers {
		if adminPeer == senderStr {
			return nil // Admin has all permissions
		}
	}

	// Check specific permissions
	for _, perm := range permissions {
		if perm.Subject == "*" || perm.Subject == senderStr {
			if perm.Action == "*" || perm.Action == "create" {
				return nil // Permission granted
			}
		}
	}

	return common.NewSecurityError("permission_denied",
		fmt.Sprintf("Permission denied for sender %s", senderStr))
}

// addDefaultFormatRules adds default format validation rules
func (v *Validator) addDefaultFormatRules() {
	// Add basic format rule
	basicRule := NewBasicFormatRule()
	v.formatRules = append(v.formatRules, basicRule)
}

// addDefaultBusinessRules adds default business validation rules
func (v *Validator) addDefaultBusinessRules() {
	// Add payload size rule
	if v.config.MaxPayloadSize > 0 {
		payloadRule := NewPayloadSizeRule(v.config.MaxPayloadSize)
		v.businessRules = append(v.businessRules, payloadRule)
	}

	// Add TTL rule
	if v.config.MaxTTL > 0 {
		ttlRule := NewTTLRule(v.config.MaxTTL)
		v.businessRules = append(v.businessRules, ttlRule)
	}

	// Add type whitelist rule
	if len(v.config.AllowedTypes) > 0 {
		typeRule := NewTypeWhitelistRule(v.config.AllowedTypes)
		v.businessRules = append(v.businessRules, typeRule)
	}
}

// sortFormatRulesByPriority sorts format rules by priority (highest first)
func (v *Validator) sortFormatRulesByPriority() {
	for i := 0; i < len(v.formatRules)-1; i++ {
		for j := 0; j < len(v.formatRules)-i-1; j++ {
			if v.formatRules[j].GetPriority() < v.formatRules[j+1].GetPriority() {
				v.formatRules[j], v.formatRules[j+1] = v.formatRules[j+1], v.formatRules[j]
			}
		}
	}
}

// sortBusinessRulesByPriority sorts business rules by priority (highest first)
func (v *Validator) sortBusinessRulesByPriority() {
	for i := 0; i < len(v.businessRules)-1; i++ {
		for j := 0; j < len(v.businessRules)-i-1; j++ {
			if v.businessRules[j].GetPriority() < v.businessRules[j+1].GetPriority() {
				v.businessRules[j], v.businessRules[j+1] = v.businessRules[j+1], v.businessRules[j]
			}
		}
	}
}

// GetValidationStats returns validation statistics
func (v *Validator) GetValidationStats() map[string]interface{} {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return map[string]interface{}{
		"format_rules_count":   len(v.formatRules),
		"business_rules_count": len(v.businessRules),
		"strict_mode_enabled":  v.config.EnableStrict,
		"max_payload_size":     v.config.MaxPayloadSize,
		"max_ttl":              v.config.MaxTTL,
		"allowed_types":        v.config.AllowedTypes,
	}
}
