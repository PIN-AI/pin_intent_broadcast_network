package common

import (
	"context"
	"strings"
)

// ManifestValidator defines the interface for intent manifest validation
type ManifestValidator interface {
	ValidateManifest(ctx context.Context, manifest *IntentManifest) error
}

// DefaultManifestValidator provides basic manifest validation
type DefaultManifestValidator struct{}

// NewDefaultManifestValidator creates a new default manifest validator
func NewDefaultManifestValidator() *DefaultManifestValidator {
	return &DefaultManifestValidator{}
}

// ValidateManifest validates an intent manifest
func (v *DefaultManifestValidator) ValidateManifest(ctx context.Context, manifest *IntentManifest) error {
	if manifest == nil {
		return nil // Manifest is optional
	}

	// Validate task description
	if strings.TrimSpace(manifest.Task) == "" {
		return NewValidationError("task", manifest.Task, "Task description cannot be empty when manifest is provided")
	}

	if len(manifest.Task) > 1000 {
		return NewValidationError("task", manifest.Task, "Task description too long (max 1000 characters)")
	}

	// Validate requirements if provided
	if manifest.Requirements != nil {
		for key, value := range manifest.Requirements {
			if strings.TrimSpace(key) == "" {
				return NewValidationError("requirements", key, "Requirement key cannot be empty")
			}
			if len(key) > 100 {
				return NewValidationError("requirements", key, "Requirement key too long (max 100 characters)")
			}
			if len(value) > 500 {
				return NewValidationError("requirements", value, "Requirement value too long (max 500 characters)")
			}
		}
	}

	// Validate context if provided
	if manifest.Context != "" && len(manifest.Context) > 2000 {
		return NewValidationError("context", manifest.Context, "Context too long (max 2000 characters)")
	}

	return nil
}