package validation

import (
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"

	"pin_intent_broadcast_network/internal/biz/common"
)

// PermissionValidator implements permission validation for intents
// This file will contain the implementation for task 3.4
type PermissionValidator struct {
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

// NewPermissionValidator creates a new permission validator
func NewPermissionValidator(config *PermissionConfig) *PermissionValidator {
	return &PermissionValidator{
		permissions: make(map[string][]Permission),
		config:      config,
	}
}

// ValidatePermissions validates permissions for an intent
func (pv *PermissionValidator) ValidatePermissions(intent *common.Intent, sender peer.ID) error {
	if intent == nil {
		return common.NewValidationError("intent", "", "Intent cannot be nil")
	}

	if !pv.config.EnablePermissionCheck {
		return nil // Permission check disabled
	}

	return pv.checkPermission(intent, sender)
}

// checkPermission checks if a sender has permission for an intent
func (pv *PermissionValidator) checkPermission(intent *common.Intent, sender peer.ID) error {
	pv.mu.RLock()
	permissions, exists := pv.permissions[intent.Type]
	pv.mu.RUnlock()

	if !exists {
		return nil // No specific permissions required
	}

	senderStr := sender.String()

	// Check if sender is admin
	for _, adminPeer := range pv.config.AdminPeers {
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

// AddPermission adds a permission rule
func (pv *PermissionValidator) AddPermission(intentType string, permission Permission) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	if pv.permissions[intentType] == nil {
		pv.permissions[intentType] = make([]Permission, 0)
	}

	pv.permissions[intentType] = append(pv.permissions[intentType], permission)
}

// RemovePermission removes a permission rule
func (pv *PermissionValidator) RemovePermission(intentType string, subject string) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	permissions := pv.permissions[intentType]
	if permissions == nil {
		return
	}

	filtered := make([]Permission, 0)
	for _, perm := range permissions {
		if perm.Subject != subject {
			filtered = append(filtered, perm)
		}
	}

	pv.permissions[intentType] = filtered
}

// GetPermissions returns all permissions for an intent type
func (pv *PermissionValidator) GetPermissions(intentType string) []Permission {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	permissions := pv.permissions[intentType]
	if permissions == nil {
		return make([]Permission, 0)
	}

	// Return a copy to avoid race conditions
	result := make([]Permission, len(permissions))
	copy(result, permissions)
	return result
}

// HasPermission checks if a peer has a specific permission
func (pv *PermissionValidator) HasPermission(peerID peer.ID, intentType, action string) bool {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	permissions := pv.permissions[intentType]
	if permissions == nil {
		return false
	}

	peerIDStr := peerID.String()

	// Check if peer is admin
	for _, adminPeer := range pv.config.AdminPeers {
		if adminPeer == peerIDStr {
			return true
		}
	}

	// Check specific permissions
	for _, perm := range permissions {
		if (perm.Subject == "*" || perm.Subject == peerIDStr) &&
			(perm.Action == "*" || perm.Action == action) {
			return true
		}
	}

	return false
}

// GrantPermission grants a permission to a peer
func (pv *PermissionValidator) GrantPermission(peerID peer.ID, intentType, action string) {
	permission := Permission{
		Subject: peerID.String(),
		Action:  action,
		Object:  intentType,
	}
	pv.AddPermission(intentType, permission)
}

// RevokePermission revokes a permission from a peer
func (pv *PermissionValidator) RevokePermission(peerID peer.ID, intentType string) {
	pv.RemovePermission(intentType, peerID.String())
}

// ListAllPermissions returns all permissions
func (pv *PermissionValidator) ListAllPermissions() map[string][]Permission {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	result := make(map[string][]Permission)
	for intentType, permissions := range pv.permissions {
		result[intentType] = make([]Permission, len(permissions))
		copy(result[intentType], permissions)
	}

	return result
}

// IsAdmin checks if a peer is an admin
func (pv *PermissionValidator) IsAdmin(peerID peer.ID) bool {
	peerIDStr := peerID.String()
	for _, adminPeer := range pv.config.AdminPeers {
		if adminPeer == peerIDStr {
			return true
		}
	}
	return false
}

// AddAdmin adds a peer as admin
func (pv *PermissionValidator) AddAdmin(peerID peer.ID) {
	peerIDStr := peerID.String()

	// Check if already admin
	for _, adminPeer := range pv.config.AdminPeers {
		if adminPeer == peerIDStr {
			return
		}
	}

	pv.config.AdminPeers = append(pv.config.AdminPeers, peerIDStr)
}

// RemoveAdmin removes a peer from admin list
func (pv *PermissionValidator) RemoveAdmin(peerID peer.ID) {
	peerIDStr := peerID.String()
	filtered := make([]string, 0)

	for _, adminPeer := range pv.config.AdminPeers {
		if adminPeer != peerIDStr {
			filtered = append(filtered, adminPeer)
		}
	}

	pv.config.AdminPeers = filtered
}

// GetPermissionStats returns permission statistics
func (pv *PermissionValidator) GetPermissionStats() map[string]interface{} {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	totalPermissions := 0
	for _, permissions := range pv.permissions {
		totalPermissions += len(permissions)
	}

	return map[string]interface{}{
		"permission_check_enabled": pv.config.EnablePermissionCheck,
		"total_intent_types":       len(pv.permissions),
		"total_permissions":        totalPermissions,
		"admin_peers_count":        len(pv.config.AdminPeers),
		"admin_peers":              pv.config.AdminPeers,
	}
}

// RoleBasedPermission represents a role-based permission
type RoleBasedPermission struct {
	Role        string   `json:"role"`
	IntentTypes []string `json:"intent_types"`
	Actions     []string `json:"actions"`
}

// RoleManager manages role-based permissions
type RoleManager struct {
	roles     map[string]*RoleBasedPermission
	userRoles map[string][]string // peerID -> roles
	mu        sync.RWMutex
}

// NewRoleManager creates a new role manager
func NewRoleManager() *RoleManager {
	return &RoleManager{
		roles:     make(map[string]*RoleBasedPermission),
		userRoles: make(map[string][]string),
	}
}

// AddRole adds a new role
func (rm *RoleManager) AddRole(role *RoleBasedPermission) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.roles[role.Role] = role
}

// AssignRole assigns a role to a user
func (rm *RoleManager) AssignRole(peerID peer.ID, role string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	peerIDStr := peerID.String()
	if rm.userRoles[peerIDStr] == nil {
		rm.userRoles[peerIDStr] = make([]string, 0)
	}

	// Check if role already assigned
	for _, existingRole := range rm.userRoles[peerIDStr] {
		if existingRole == role {
			return
		}
	}

	rm.userRoles[peerIDStr] = append(rm.userRoles[peerIDStr], role)
}

// HasRolePermission checks if a user has permission through roles
func (rm *RoleManager) HasRolePermission(peerID peer.ID, intentType, action string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	peerIDStr := peerID.String()
	userRoles := rm.userRoles[peerIDStr]

	for _, roleName := range userRoles {
		role := rm.roles[roleName]
		if role == nil {
			continue
		}

		// Check if role has permission for this intent type and action
		hasIntentType := len(role.IntentTypes) == 0 // Empty means all types
		for _, allowedType := range role.IntentTypes {
			if allowedType == "*" || allowedType == intentType {
				hasIntentType = true
				break
			}
		}

		hasAction := len(role.Actions) == 0 // Empty means all actions
		for _, allowedAction := range role.Actions {
			if allowedAction == "*" || allowedAction == action {
				hasAction = true
				break
			}
		}

		if hasIntentType && hasAction {
			return true
		}
	}

	return false
}
