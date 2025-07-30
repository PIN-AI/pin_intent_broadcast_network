package common

import "fmt"

// IntentError represents an error related to intent operations
type IntentError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error implements the error interface
func (e *IntentError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewIntentError creates a new intent error
func NewIntentError(code, message, details string) *IntentError {
	return &IntentError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Predefined error types for common scenarios
var (
	// Intent not found errors
	ErrIntentNotFound = &IntentError{
		Code:    "INTENT_NOT_FOUND",
		Message: "Intent not found",
	}

	// Validation errors
	ErrIntentInvalidFormat = &IntentError{
		Code:    "INVALID_FORMAT",
		Message: "Intent format is invalid",
	}

	ErrIntentValidationFailed = &IntentError{
		Code:    "VALIDATION_FAILED",
		Message: "Intent validation failed",
	}

	// Security errors
	ErrIntentSignatureFailed = &IntentError{
		Code:    "SIGNATURE_FAILED",
		Message: "Intent signature verification failed",
	}

	ErrIntentPermissionDenied = &IntentError{
		Code:    "PERMISSION_DENIED",
		Message: "Permission denied",
	}

	// Processing errors
	ErrIntentProcessingFailed = &IntentError{
		Code:    "PROCESSING_FAILED",
		Message: "Intent processing failed",
	}

	ErrIntentHandlerNotFound = &IntentError{
		Code:    "HANDLER_NOT_FOUND",
		Message: "No handler found for intent type",
	}

	// Lifecycle errors
	ErrIntentExpired = &IntentError{
		Code:    "INTENT_EXPIRED",
		Message: "Intent has expired",
	}

	ErrIntentAlreadyProcessed = &IntentError{
		Code:    "ALREADY_PROCESSED",
		Message: "Intent has already been processed",
	}

	// Network errors
	ErrNetworkUnavailable = &IntentError{
		Code:    "NETWORK_UNAVAILABLE",
		Message: "Network is unavailable",
	}

	ErrBroadcastFailed = &IntentError{
		Code:    "BROADCAST_FAILED",
		Message: "Failed to broadcast intent",
	}

	// Configuration errors
	ErrInvalidConfiguration = &IntentError{
		Code:    "INVALID_CONFIGURATION",
		Message: "Invalid configuration",
	}

	// Matching errors
	ErrMatchingFailed = &IntentError{
		Code:    "MATCHING_FAILED",
		Message: "Intent matching failed",
	}

	// Storage errors
	ErrStorageUnavailable = &IntentError{
		Code:    "STORAGE_UNAVAILABLE",
		Message: "Storage is unavailable",
	}

	// Timeout errors
	ErrProcessingTimeout = &IntentError{
		Code:    "PROCESSING_TIMEOUT",
		Message: "Processing timeout exceeded",
	}

	// Rate limiting errors
	ErrRateLimitExceeded = &IntentError{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: "Rate limit exceeded",
	}
)

// IsIntentError checks if an error is an IntentError
func IsIntentError(err error) bool {
	_, ok := err.(*IntentError)
	return ok
}

// GetErrorCode extracts the error code from an IntentError
func GetErrorCode(err error) string {
	if intentErr, ok := err.(*IntentError); ok {
		return intentErr.Code
	}
	return "UNKNOWN_ERROR"
}

// WrapError wraps a generic error as an IntentError
func WrapError(err error, code, message string) *IntentError {
	return &IntentError{
		Code:    code,
		Message: message,
		Details: err.Error(),
	}
}

// ValidationError represents a validation-specific error
type ValidationError struct {
	*IntentError
	Field string `json:"field,omitempty"`
	Value string `json:"value,omitempty"`
}

// NewValidationError creates a new validation error
func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		IntentError: &IntentError{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Details: fmt.Sprintf("field: %s, value: %s", field, value),
		},
		Field: field,
		Value: value,
	}
}

// SecurityError represents a security-specific error
type SecurityError struct {
	*IntentError
	Reason string `json:"reason,omitempty"`
}

// NewSecurityError creates a new security error
func NewSecurityError(reason, message string) *SecurityError {
	return &SecurityError{
		IntentError: &IntentError{
			Code:    "SECURITY_ERROR",
			Message: message,
			Details: reason,
		},
		Reason: reason,
	}
}

// NetworkError represents a network-specific error
type NetworkError struct {
	*IntentError
	PeerID string `json:"peer_id,omitempty"`
	Topic  string `json:"topic,omitempty"`
}

// NewNetworkError creates a new network error
func NewNetworkError(peerID, topic, message string) *NetworkError {
	return &NetworkError{
		IntentError: &IntentError{
			Code:    "NETWORK_ERROR",
			Message: message,
			Details: fmt.Sprintf("peer: %s, topic: %s", peerID, topic),
		},
		PeerID: peerID,
		Topic:  topic,
	}
}

// ProcessingError represents a processing-specific error
type ProcessingError struct {
	*IntentError
	Stage    string `json:"stage,omitempty"`
	IntentID string `json:"intent_id,omitempty"`
}

// NewProcessingError creates a new processing error
func NewProcessingError(stage, intentID, message string) *ProcessingError {
	return &ProcessingError{
		IntentError: &IntentError{
			Code:    "PROCESSING_ERROR",
			Message: message,
			Details: fmt.Sprintf("stage: %s, intent: %s", stage, intentID),
		},
		Stage:    stage,
		IntentID: intentID,
	}
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     *IntentError `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
	Timestamp int64        `json:"timestamp"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err *IntentError, requestID string, timestamp int64) *ErrorResponse {
	return &ErrorResponse{
		Error:     err,
		RequestID: requestID,
		Timestamp: timestamp,
	}
}
