package common

import (
	"sync"
	"time"
)

// BusinessMetrics holds all business logic layer metrics
type BusinessMetrics struct {
	// Intent metrics
	IntentsCreated    int64         `json:"intents_created"`
	IntentsProcessed  int64         `json:"intents_processed"`
	IntentsMatched    int64         `json:"intents_matched"`
	IntentsFailed     int64         `json:"intents_failed"`
	IntentsExpired    int64         `json:"intents_expired"`
	ProcessingLatency time.Duration `json:"processing_latency"`

	// Validation metrics
	ValidationErrors  int64 `json:"validation_errors"`
	ValidationSuccess int64 `json:"validation_success"`

	// Security metrics
	SignatureFailures int64 `json:"signature_failures"`
	SignatureSuccess  int64 `json:"signature_success"`

	// Matching metrics
	MatchingAccuracy float64 `json:"matching_accuracy"`
	MatchingAttempts int64   `json:"matching_attempts"`
	MatchingSuccess  int64   `json:"matching_success"`

	// Network metrics
	NetworkPeers     int64         `json:"network_peers"`
	MessagesSent     int64         `json:"messages_sent"`
	MessagesReceived int64         `json:"messages_received"`
	NetworkLatency   time.Duration `json:"network_latency"`

	// Processing metrics
	ProcessingStages   int64 `json:"processing_stages"`
	PipelineExecutions int64 `json:"pipeline_executions"`
	HandlerExecutions  int64 `json:"handler_executions"`

	mu sync.RWMutex
}

// NewBusinessMetrics creates a new business metrics instance
func NewBusinessMetrics() *BusinessMetrics {
	return &BusinessMetrics{}
}

// Intent Metrics Methods

// IncrementIntentsCreated increments the intents created counter
func (m *BusinessMetrics) IncrementIntentsCreated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.IntentsCreated++
}

// IncrementIntentsProcessed increments the intents processed counter
func (m *BusinessMetrics) IncrementIntentsProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.IntentsProcessed++
}

// IncrementIntentsMatched increments the intents matched counter
func (m *BusinessMetrics) IncrementIntentsMatched() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.IntentsMatched++
}

// IncrementIntentsFailed increments the intents failed counter
func (m *BusinessMetrics) IncrementIntentsFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.IntentsFailed++
}

// IncrementIntentsExpired increments the intents expired counter
func (m *BusinessMetrics) IncrementIntentsExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.IntentsExpired++
}

// UpdateProcessingLatency updates the processing latency metric
func (m *BusinessMetrics) UpdateProcessingLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate moving average
	if m.ProcessingLatency == 0 {
		m.ProcessingLatency = latency
	} else {
		// Simple exponential moving average with alpha = 0.1
		m.ProcessingLatency = time.Duration(float64(m.ProcessingLatency)*0.9 + float64(latency)*0.1)
	}
}

// Validation Metrics Methods

// IncrementValidationErrors increments the validation errors counter
func (m *BusinessMetrics) IncrementValidationErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ValidationErrors++
}

// IncrementValidationSuccess increments the validation success counter
func (m *BusinessMetrics) IncrementValidationSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ValidationSuccess++
}

// GetValidationSuccessRate returns the validation success rate
func (m *BusinessMetrics) GetValidationSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.ValidationSuccess + m.ValidationErrors
	if total == 0 {
		return 0.0
	}

	return float64(m.ValidationSuccess) / float64(total)
}

// Security Metrics Methods

// IncrementSignatureFailures increments the signature failures counter
func (m *BusinessMetrics) IncrementSignatureFailures() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SignatureFailures++
}

// IncrementSignatureSuccess increments the signature success counter
func (m *BusinessMetrics) IncrementSignatureSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SignatureSuccess++
}

// GetSignatureSuccessRate returns the signature success rate
func (m *BusinessMetrics) GetSignatureSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.SignatureSuccess + m.SignatureFailures
	if total == 0 {
		return 0.0
	}

	return float64(m.SignatureSuccess) / float64(total)
}

// Matching Metrics Methods

// UpdateMatchingAccuracy updates the matching accuracy metric
func (m *BusinessMetrics) UpdateMatchingAccuracy(accuracy float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate moving average
	if m.MatchingAccuracy == 0 {
		m.MatchingAccuracy = accuracy
	} else {
		// Simple exponential moving average with alpha = 0.1
		m.MatchingAccuracy = m.MatchingAccuracy*0.9 + accuracy*0.1
	}
}

// IncrementMatchingAttempts increments the matching attempts counter
func (m *BusinessMetrics) IncrementMatchingAttempts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MatchingAttempts++
}

// IncrementMatchingSuccess increments the matching success counter
func (m *BusinessMetrics) IncrementMatchingSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MatchingSuccess++
}

// GetMatchingSuccessRate returns the matching success rate
func (m *BusinessMetrics) GetMatchingSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.MatchingAttempts == 0 {
		return 0.0
	}

	return float64(m.MatchingSuccess) / float64(m.MatchingAttempts)
}

// Network Metrics Methods

// SetNetworkPeers sets the network peers count
func (m *BusinessMetrics) SetNetworkPeers(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.NetworkPeers = count
}

// IncrementMessagesSent increments the messages sent counter
func (m *BusinessMetrics) IncrementMessagesSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesSent++
}

// IncrementMessagesReceived increments the messages received counter
func (m *BusinessMetrics) IncrementMessagesReceived() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesReceived++
}

// UpdateNetworkLatency updates the network latency metric
func (m *BusinessMetrics) UpdateNetworkLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate moving average
	if m.NetworkLatency == 0 {
		m.NetworkLatency = latency
	} else {
		// Simple exponential moving average with alpha = 0.1
		m.NetworkLatency = time.Duration(float64(m.NetworkLatency)*0.9 + float64(latency)*0.1)
	}
}

// Processing Metrics Methods

// IncrementProcessingStages increments the processing stages counter
func (m *BusinessMetrics) IncrementProcessingStages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ProcessingStages++
}

// IncrementPipelineExecutions increments the pipeline executions counter
func (m *BusinessMetrics) IncrementPipelineExecutions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PipelineExecutions++
}

// IncrementHandlerExecutions increments the handler executions counter
func (m *BusinessMetrics) IncrementHandlerExecutions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HandlerExecutions++
}

// Snapshot Methods

// GetSnapshot returns a snapshot of current metrics
func (m *BusinessMetrics) GetSnapshot() BusinessMetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return BusinessMetricsSnapshot{
		IntentsCreated:     m.IntentsCreated,
		IntentsProcessed:   m.IntentsProcessed,
		IntentsMatched:     m.IntentsMatched,
		IntentsFailed:      m.IntentsFailed,
		IntentsExpired:     m.IntentsExpired,
		ProcessingLatency:  m.ProcessingLatency,
		ValidationErrors:   m.ValidationErrors,
		ValidationSuccess:  m.ValidationSuccess,
		SignatureFailures:  m.SignatureFailures,
		SignatureSuccess:   m.SignatureSuccess,
		MatchingAccuracy:   m.MatchingAccuracy,
		MatchingAttempts:   m.MatchingAttempts,
		MatchingSuccess:    m.MatchingSuccess,
		NetworkPeers:       m.NetworkPeers,
		MessagesSent:       m.MessagesSent,
		MessagesReceived:   m.MessagesReceived,
		NetworkLatency:     m.NetworkLatency,
		ProcessingStages:   m.ProcessingStages,
		PipelineExecutions: m.PipelineExecutions,
		HandlerExecutions:  m.HandlerExecutions,
		Timestamp:          time.Now(),
	}
}

// BusinessMetricsSnapshot represents a point-in-time metrics snapshot
type BusinessMetricsSnapshot struct {
	IntentsCreated     int64         `json:"intents_created"`
	IntentsProcessed   int64         `json:"intents_processed"`
	IntentsMatched     int64         `json:"intents_matched"`
	IntentsFailed      int64         `json:"intents_failed"`
	IntentsExpired     int64         `json:"intents_expired"`
	ProcessingLatency  time.Duration `json:"processing_latency"`
	ValidationErrors   int64         `json:"validation_errors"`
	ValidationSuccess  int64         `json:"validation_success"`
	SignatureFailures  int64         `json:"signature_failures"`
	SignatureSuccess   int64         `json:"signature_success"`
	MatchingAccuracy   float64       `json:"matching_accuracy"`
	MatchingAttempts   int64         `json:"matching_attempts"`
	MatchingSuccess    int64         `json:"matching_success"`
	NetworkPeers       int64         `json:"network_peers"`
	MessagesSent       int64         `json:"messages_sent"`
	MessagesReceived   int64         `json:"messages_received"`
	NetworkLatency     time.Duration `json:"network_latency"`
	ProcessingStages   int64         `json:"processing_stages"`
	PipelineExecutions int64         `json:"pipeline_executions"`
	HandlerExecutions  int64         `json:"handler_executions"`
	Timestamp          time.Time     `json:"timestamp"`
}

// Reset resets all metrics to zero
func (m *BusinessMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.IntentsCreated = 0
	m.IntentsProcessed = 0
	m.IntentsMatched = 0
	m.IntentsFailed = 0
	m.IntentsExpired = 0
	m.ProcessingLatency = 0
	m.ValidationErrors = 0
	m.ValidationSuccess = 0
	m.SignatureFailures = 0
	m.SignatureSuccess = 0
	m.MatchingAccuracy = 0
	m.MatchingAttempts = 0
	m.MatchingSuccess = 0
	m.NetworkPeers = 0
	m.MessagesSent = 0
	m.MessagesReceived = 0
	m.NetworkLatency = 0
	m.ProcessingStages = 0
	m.PipelineExecutions = 0
	m.HandlerExecutions = 0
}

// GetTotalIntents returns the total number of intents processed
func (m *BusinessMetrics) GetTotalIntents() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.IntentsCreated
}

// GetSuccessRate returns the overall success rate
func (m *BusinessMetrics) GetSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.IntentsProcessed + m.IntentsFailed
	if total == 0 {
		return 0.0
	}

	return float64(m.IntentsProcessed) / float64(total)
}

// GetThroughput returns the processing throughput (intents per second)
func (m *BusinessMetrics) GetThroughput(duration time.Duration) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if duration.Seconds() == 0 {
		return 0.0
	}

	return float64(m.IntentsProcessed) / duration.Seconds()
}

// MetricsCollector interface for collecting metrics from different components
type MetricsCollector interface {
	CollectMetrics() map[string]interface{}
	GetMetricNames() []string
	ResetMetrics()
}

// PrometheusMetrics holds Prometheus-compatible metrics
type PrometheusMetrics struct {
	// Counter metrics
	Counters map[string]int64 `json:"counters"`

	// Gauge metrics
	Gauges map[string]float64 `json:"gauges"`

	// Histogram metrics
	Histograms map[string][]float64 `json:"histograms"`

	mu sync.RWMutex
}

// NewPrometheusMetrics creates a new Prometheus metrics instance
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		Counters:   make(map[string]int64),
		Gauges:     make(map[string]float64),
		Histograms: make(map[string][]float64),
	}
}

// IncrementCounter increments a counter metric
func (pm *PrometheusMetrics) IncrementCounter(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.Counters[name]++
}

// SetGauge sets a gauge metric value
func (pm *PrometheusMetrics) SetGauge(name string, value float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.Gauges[name] = value
}

// ObserveHistogram adds an observation to a histogram metric
func (pm *PrometheusMetrics) ObserveHistogram(name string, value float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.Histograms[name] == nil {
		pm.Histograms[name] = make([]float64, 0)
	}

	pm.Histograms[name] = append(pm.Histograms[name], value)

	// Keep only the last 1000 observations to prevent memory growth
	if len(pm.Histograms[name]) > 1000 {
		pm.Histograms[name] = pm.Histograms[name][len(pm.Histograms[name])-1000:]
	}
}

// GetMetrics returns all metrics in a map format
func (pm *PrometheusMetrics) GetMetrics() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	metrics := make(map[string]interface{})

	// Add counters
	for name, value := range pm.Counters {
		metrics[name] = value
	}

	// Add gauges
	for name, value := range pm.Gauges {
		metrics[name] = value
	}

	// Add histogram summaries
	for name, values := range pm.Histograms {
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			metrics[name+"_sum"] = sum
			metrics[name+"_count"] = len(values)
			metrics[name+"_avg"] = sum / float64(len(values))
		}
	}

	return metrics
}
