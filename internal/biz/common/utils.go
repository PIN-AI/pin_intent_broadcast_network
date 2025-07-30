package common

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// StringUtils provides string utility functions
type StringUtils struct{}

// IsEmpty checks if a string is empty or contains only whitespace
func (StringUtils) IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Contains checks if a slice contains a specific string
func (StringUtils) Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate strings from a slice
func (StringUtils) RemoveDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// TruncateString truncates a string to a maximum length
func (StringUtils) TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// TimeUtils provides time utility functions
type TimeUtils struct{}

// Now returns the current Unix timestamp
func (TimeUtils) Now() int64 {
	return time.Now().Unix()
}

// NowNano returns the current Unix timestamp in nanoseconds
func (TimeUtils) NowNano() int64 {
	return time.Now().UnixNano()
}

// IsExpired checks if a timestamp with TTL has expired
func (TimeUtils) IsExpired(timestamp, ttl int64) bool {
	if ttl <= 0 {
		return false // No expiration
	}

	expirationTime := timestamp + ttl
	return time.Now().Unix() > expirationTime
}

// FormatDuration formats a duration in a human-readable format
func (TimeUtils) FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1e6)
	} else if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.2fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.2fh", d.Hours())
	}
}

// ParseDuration parses a duration string with support for various units
func (TimeUtils) ParseDuration(s string) (time.Duration, error) {
	// Support common duration formats
	return time.ParseDuration(s)
}

// JSONUtils provides JSON utility functions
type JSONUtils struct{}

// Marshal marshals an object to JSON bytes
func (JSONUtils) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal unmarshals JSON bytes to an object
func (JSONUtils) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MarshalIndent marshals an object to indented JSON bytes
func (JSONUtils) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// ToJSONString converts an object to a JSON string
func (JSONUtils) ToJSONString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(data)
}

// FromJSONString parses a JSON string to an object
func (JSONUtils) FromJSONString(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// ValidationUtils provides validation utility functions
type ValidationUtils struct{}

// IsValidIntentType checks if an intent type is valid
func (ValidationUtils) IsValidIntentType(intentType string) bool {
	validTypes := []string{
		IntentTypeTrade,
		IntentTypeSwap,
		IntentTypeExchange,
		IntentTypeTransfer,
		IntentTypeSend,
		IntentTypePayment,
		IntentTypeLending,
		IntentTypeBorrow,
		IntentTypeLoan,
		IntentTypeInvestment,
		IntentTypeStaking,
		IntentTypeYield,
	}

	return StringUtils{}.Contains(validTypes, intentType)
}

// IsValidPriority checks if a priority value is valid
func (ValidationUtils) IsValidPriority(priority int32) bool {
	return priority >= PriorityLow && priority <= PriorityUrgent
}

// IsValidTTL checks if a TTL value is valid
func (ValidationUtils) IsValidTTL(ttl int64) bool {
	return ttl > 0 && ttl <= int64(DefaultMaxTTL.Seconds())
}

// IsValidPayloadSize checks if a payload size is valid
func (ValidationUtils) IsValidPayloadSize(size int) bool {
	return size > 0 && size <= DefaultMaxPayloadSize
}

// ConversionUtils provides type conversion utility functions
type ConversionUtils struct{}

// IntentStatusToString converts IntentStatus to string
func (ConversionUtils) IntentStatusToString(status IntentStatus) string {
	return status.String()
}

// StringToIntentStatus converts string to IntentStatus
func (ConversionUtils) StringToIntentStatus(s string) IntentStatus {
	switch strings.ToLower(s) {
	case "created":
		return IntentStatusCreated
	case "validated":
		return IntentStatusValidated
	case "broadcasted":
		return IntentStatusBroadcasted
	case "processed":
		return IntentStatusProcessed
	case "matched":
		return IntentStatusMatched
	case "completed":
		return IntentStatusCompleted
	case "failed":
		return IntentStatusFailed
	case "expired":
		return IntentStatusExpired
	default:
		return IntentStatusCreated
	}
}

// MatchTypeToString converts MatchType to string
func (ConversionUtils) MatchTypeToString(matchType MatchType) string {
	return matchType.String()
}

// StringToMatchType converts string to MatchType
func (ConversionUtils) StringToMatchType(s string) MatchType {
	switch strings.ToLower(s) {
	case "exact":
		return MatchTypeExact
	case "partial":
		return MatchTypePartial
	case "semantic":
		return MatchTypeSemantic
	case "pattern":
		return MatchTypePattern
	default:
		return MatchTypePartial
	}
}

// ContextUtils provides context utility functions
type ContextUtils struct{}

// WithTimeout creates a context with timeout
func (ContextUtils) WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

// WithCancel creates a context with cancel
func (ContextUtils) WithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}

// IsContextCancelled checks if a context is cancelled
func (ContextUtils) IsContextCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// GetContextError returns the context error if any
func (ContextUtils) GetContextError(ctx context.Context) error {
	return ctx.Err()
}

// ReflectionUtils provides reflection utility functions
type ReflectionUtils struct{}

// GetTypeName returns the type name of an object
func (ReflectionUtils) GetTypeName(obj interface{}) string {
	return reflect.TypeOf(obj).String()
}

// IsNil checks if an interface is nil
func (ReflectionUtils) IsNil(obj interface{}) bool {
	if obj == nil {
		return true
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// DeepEqual performs deep equality comparison
func (ReflectionUtils) DeepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// CopyStruct copies fields from source to destination struct
func (ReflectionUtils) CopyStruct(src, dst interface{}) error {
	srcValue := reflect.ValueOf(src)
	dstValue := reflect.ValueOf(dst)

	if srcValue.Kind() == reflect.Ptr {
		srcValue = srcValue.Elem()
	}

	if dstValue.Kind() != reflect.Ptr || dstValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a pointer to struct")
	}

	dstValue = dstValue.Elem()

	if srcValue.Kind() != reflect.Struct {
		return fmt.Errorf("source must be a struct")
	}

	srcType := srcValue.Type()

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		srcFieldType := srcType.Field(i)

		// Find corresponding field in destination
		dstField := dstValue.FieldByName(srcFieldType.Name)
		if !dstField.IsValid() || !dstField.CanSet() {
			continue
		}

		// Check if types are compatible
		if srcField.Type() == dstField.Type() {
			dstField.Set(srcField)
		}
	}

	return nil
}

// SliceUtils provides slice utility functions
type SliceUtils struct{}

// ContainsString checks if a string slice contains a specific string
func (SliceUtils) ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveString removes a string from a slice
func (SliceUtils) RemoveString(slice []string, item string) []string {
	result := make([]string, 0)
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// UniqueStrings returns unique strings from a slice
func (SliceUtils) UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ChunkStrings splits a string slice into chunks of specified size
func (SliceUtils) ChunkStrings(slice []string, chunkSize int) [][]string {
	if chunkSize <= 0 {
		return [][]string{slice}
	}

	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// MapUtils provides map utility functions
type MapUtils struct{}

// MergeStringMaps merges multiple string maps
func (MapUtils) MergeStringMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// CopyStringMap creates a copy of a string map
func (MapUtils) CopyStringMap(original map[string]string) map[string]string {
	copy := make(map[string]string)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// GetMapKeys returns all keys from a string map
func (MapUtils) GetMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// GetMapValues returns all values from a string map
func (MapUtils) GetMapValues(m map[string]string) []string {
	values := make([]string, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// IDUtils provides ID generation utility functions
type IDUtils struct{}

// GenerateIntentID generates a unique ID for an intent
func (IDUtils) GenerateIntentID() string {
	return fmt.Sprintf("intent_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

// GeneratePeerID generates a unique peer ID
func (IDUtils) GeneratePeerID() string {
	return fmt.Sprintf("peer_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

// GenerateSessionID generates a unique session ID
func (IDUtils) GenerateSessionID() string {
	return fmt.Sprintf("session_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// Global utility instances for easy access
var (
	Strings    = StringUtils{}
	Times      = TimeUtils{}
	JSON       = JSONUtils{}
	Validation = ValidationUtils{}
	Conversion = ConversionUtils{}
	Context    = ContextUtils{}
	Reflection = ReflectionUtils{}
	Slices     = SliceUtils{}
	Maps       = MapUtils{}
	IDs        = IDUtils{}
)

// GenerateIntentID is a global function for generating intent IDs
func GenerateIntentID() string {
	return IDs.GenerateIntentID()
}
