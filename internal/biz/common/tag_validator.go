package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// TagValidator defines the interface for tag validation
type TagValidator interface {
	ValidateTag(ctx context.Context, tag *Tag) error
	ValidateTags(ctx context.Context, tags []Tag) error
	CalculateTotalTagFee(tags []Tag) (string, error)
}

// DataVaultPolicyProvider defines the interface for querying Data Vault policies
type DataVaultPolicyProvider interface {
	GetTagPolicy(ctx context.Context, userAddress string, tagName string) (*TagPolicy, error)
	GetUserTagPolicies(ctx context.Context, userAddress string) (map[string]*TagPolicy, error)
}

// TagPolicy represents a tag policy from Data Vault
type TagPolicy struct {
	TagName    string `json:"tag_name"`
	TagFee     string `json:"tag_fee"`
	IsTradable bool   `json:"is_tradable"`
}

// DefaultTagValidator provides basic tag validation for MVP
type DefaultTagValidator struct {
	policyProvider DataVaultPolicyProvider
}

// NewDefaultTagValidator creates a new default tag validator
func NewDefaultTagValidator(policyProvider DataVaultPolicyProvider) *DefaultTagValidator {
	return &DefaultTagValidator{
		policyProvider: policyProvider,
	}
}

// ValidateTag validates a single tag
func (v *DefaultTagValidator) ValidateTag(ctx context.Context, tag *Tag) error {
	if tag == nil {
		return NewValidationError("tag", "", "Tag cannot be nil")
	}

	// Validate tag name
	if strings.TrimSpace(tag.TagName) == "" {
		return NewValidationError("tag_name", tag.TagName, "Tag name cannot be empty")
	}

	// Validate tag name format
	if !isValidTagName(tag.TagName) {
		return NewValidationError("tag_name", tag.TagName, "Invalid tag name format")
	}

	// Validate tag fee format
	if err := v.validateTagFee(tag.TagFee); err != nil {
		return err
	}

	return nil
}

// ValidateTags validates a list of tags
func (v *DefaultTagValidator) ValidateTags(ctx context.Context, tags []Tag) error {
	if len(tags) == 0 {
		return nil // Empty tags list is valid
	}

	// Check for duplicate tag names
	seen := make(map[string]bool)
	for _, tag := range tags {
		if err := v.ValidateTag(ctx, &tag); err != nil {
			return err
		}

		if seen[tag.TagName] {
			return NewValidationError("relevant_tags", tag.TagName, "Duplicate tag name found")
		}
		seen[tag.TagName] = true
	}

	return nil
}

// CalculateTotalTagFee calculates the total fee for all tags
func (v *DefaultTagValidator) CalculateTotalTagFee(tags []Tag) (string, error) {
	if len(tags) == 0 {
		return "0", nil
	}

	totalFee := int64(0)
	for _, tag := range tags {
		fee, err := strconv.ParseInt(tag.TagFee, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid tag fee format for tag %s: %w", tag.TagName, err)
		}
		totalFee += fee
	}

	return strconv.FormatInt(totalFee, 10), nil
}

// validateTagFee validates tag fee format and value
func (v *DefaultTagValidator) validateTagFee(tagFee string) error {
	if strings.TrimSpace(tagFee) == "" {
		return NewValidationError("tag_fee", tagFee, "Tag fee cannot be empty")
	}

	fee, err := strconv.ParseInt(tagFee, 10, 64)
	if err != nil {
		return NewValidationError("tag_fee", tagFee, "Tag fee must be a valid integer")
	}

	if fee < 0 {
		return NewValidationError("tag_fee", tagFee, "Tag fee cannot be negative")
	}

	return nil
}

// isValidTagName checks if tag name follows the expected format
func isValidTagName(tagName string) bool {
	// Basic validation: alphanumeric, underscores, and hyphens
	if len(tagName) < 2 || len(tagName) > 64 {
		return false
	}

	for _, char := range tagName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// MVPDataVaultPolicyProvider provides mock implementation for MVP phase
type MVPDataVaultPolicyProvider struct{}

// NewMVPDataVaultPolicyProvider creates a new MVP policy provider
func NewMVPDataVaultPolicyProvider() *MVPDataVaultPolicyProvider {
	return &MVPDataVaultPolicyProvider{}
}

// GetTagPolicy returns mock tag policy for MVP
func (p *MVPDataVaultPolicyProvider) GetTagPolicy(ctx context.Context, userAddress string, tagName string) (*TagPolicy, error) {
	// MVP: Return mock policy assuming all tags are tradable with default fees
	return &TagPolicy{
		TagName:    tagName,
		TagFee:     "10000", // Default fee in Octa
		IsTradable: true,    // MVP: assume all tags are tradable
	}, nil
}

// GetUserTagPolicies returns all tag policies for a user (MVP mock implementation)
func (p *MVPDataVaultPolicyProvider) GetUserTagPolicies(ctx context.Context, userAddress string) (map[string]*TagPolicy, error) {
	// MVP: Return empty policies map, real implementation would query blockchain
	return make(map[string]*TagPolicy), nil
}