package processing

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"pin_intent_broadcast_network/internal/biz/common"
)

// Processor implements the IntentProcessor interface
// It coordinates intent processing through pipelines and handlers
type Processor struct {
	handlers map[string][]common.IntentHandler
	pipeline common.ProcessingPipeline
	registry common.HandlerRegistry
	metrics  *ProcessingMetrics
	config   *ProcessorConfig
	logger   *log.Helper
	mu       sync.RWMutex
}

// ProcessorConfig holds configuration for the processor
type ProcessorConfig struct {
	MaxConcurrentProcessing int  `yaml:"max_concurrent_processing"`
	EnableAsync             bool `yaml:"enable_async"`
	RetryAttempts           int  `yaml:"retry_attempts"`
	TimeoutSeconds          int  `yaml:"timeout_seconds"`
}

// ProcessingMetrics holds metrics for processing operations
type ProcessingMetrics struct {
	ProcessedCount   int64
	FailedCount      int64
	AverageLatency   int64
	ActiveProcessing int64
}

// NewProcessor creates a new intent processor
func NewProcessor(
	pipeline common.ProcessingPipeline,
	registry common.HandlerRegistry,
	config *ProcessorConfig,
	logger log.Logger,
) *Processor {
	return &Processor{
		handlers: make(map[string][]common.IntentHandler),
		pipeline: pipeline,
		registry: registry,
		metrics:  &ProcessingMetrics{},
		config:   config,
		logger:   log.NewHelper(logger),
	}
}

// ProcessIncomingIntent processes an incoming intent
func (p *Processor) ProcessIncomingIntent(ctx context.Context, intent *common.Intent) error {
	if intent == nil {
		return common.NewIntentError(common.ErrorCodeInvalidFormat, "Intent cannot be nil", "")
	}

	startTime := time.Now()
	p.logger.Infof("Processing incoming intent: %s, type: %s", intent.ID, intent.Type)

	// Update active processing count
	p.mu.Lock()
	p.metrics.ActiveProcessing++
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.metrics.ActiveProcessing--
		p.mu.Unlock()

		latency := time.Since(startTime).Milliseconds()
		p.updateMetrics(true, latency)
	}()

	// 1. Pipeline preprocessing
	if err := p.processWithPipeline(ctx, intent); err != nil {
		p.logger.Errorf("Pipeline processing failed for intent %s: %v", intent.ID, err)
		p.updateMetrics(false, time.Since(startTime).Milliseconds())
		return common.WrapError(err, common.ErrorCodeProcessingFailed, "Pipeline processing failed")
	}

	// 2. Find handlers
	handlers, err := p.findHandlers(intent.Type)
	if err != nil {
		p.logger.Errorf("Failed to find handlers for intent %s: %v", intent.ID, err)
		p.updateMetrics(false, time.Since(startTime).Milliseconds())
		return common.WrapError(err, common.ErrorCodeHandlerNotFound, "No handlers found")
	}

	if len(handlers) == 0 {
		p.logger.Warnf("No handlers available for intent type: %s", intent.Type)
		return common.NewIntentError(common.ErrorCodeHandlerNotFound,
			"No handlers available for intent type", intent.Type)
	}

	// 3. Execute handlers
	if err := p.executeHandlers(ctx, intent, handlers); err != nil {
		p.logger.Errorf("Handler execution failed for intent %s: %v", intent.ID, err)
		p.updateMetrics(false, time.Since(startTime).Milliseconds())
		return common.WrapError(err, common.ErrorCodeProcessingFailed, "Handler execution failed")
	}

	p.logger.Infof("Intent processed successfully: %s", intent.ID)
	return nil
}

// ProcessOutgoingIntent processes an outgoing intent
func (p *Processor) ProcessOutgoingIntent(ctx context.Context, intent *common.Intent) error {
	if intent == nil {
		return common.NewIntentError(common.ErrorCodeInvalidFormat, "Intent cannot be nil", "")
	}

	startTime := time.Now()
	p.logger.Infof("Processing outgoing intent: %s, type: %s", intent.ID, intent.Type)

	// Update active processing count
	p.mu.Lock()
	p.metrics.ActiveProcessing++
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.metrics.ActiveProcessing--
		p.mu.Unlock()

		latency := time.Since(startTime).Milliseconds()
		p.updateMetrics(true, latency)
	}()

	// 1. Validate intent before processing
	if intent.Status != common.IntentStatusCreated && intent.Status != common.IntentStatusValidated {
		return common.NewIntentError(common.ErrorCodeInvalidFormat,
			"Intent must be in created or validated status for outgoing processing", intent.Status.String())
	}

	// 2. Apply outgoing processing pipeline
	if err := p.processWithPipeline(ctx, intent); err != nil {
		p.logger.Errorf("Outgoing pipeline processing failed for intent %s: %v", intent.ID, err)
		p.updateMetrics(false, time.Since(startTime).Milliseconds())
		return common.WrapError(err, common.ErrorCodeProcessingFailed, "Outgoing pipeline processing failed")
	}

	// 3. Update intent status
	intent.Status = common.IntentStatusProcessed
	intent.ProcessedAt = time.Now().Unix()

	p.logger.Infof("Outgoing intent processed successfully: %s", intent.ID)
	return nil
}

// RegisterHandler registers an intent handler
func (p *Processor) RegisterHandler(intentType string, handler common.IntentHandler) error {
	if common.Strings.IsEmpty(intentType) {
		return common.NewIntentError(common.ErrorCodeInvalidConfiguration, "Intent type cannot be empty", "")
	}

	if handler == nil {
		return common.NewIntentError(common.ErrorCodeInvalidConfiguration, "Handler cannot be nil", "")
	}

	p.logger.Infof("Registering handler for intent type: %s", intentType)
	return p.registry.RegisterHandler(intentType, handler)
}

// UnregisterHandler unregisters an intent handler
func (p *Processor) UnregisterHandler(intentType string) error {
	if common.Strings.IsEmpty(intentType) {
		return common.NewIntentError(common.ErrorCodeInvalidConfiguration, "Intent type cannot be empty", "")
	}

	p.logger.Infof("Unregistering handler for intent type: %s", intentType)
	return p.registry.UnregisterHandler(intentType)
}

// GetProcessingStatus returns the current processing status
func (p *Processor) GetProcessingStatus() *common.ProcessingStatus {
	// TODO: Implement in task 5.1
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &common.ProcessingStatus{
		ActiveIntents:  int(p.metrics.ActiveProcessing),
		ProcessedCount: p.metrics.ProcessedCount,
		FailedCount:    p.metrics.FailedCount,
		AverageLatency: p.metrics.AverageLatency,
		HandlerStatus:  make(map[string]interface{}),
		PipelineStatus: make(map[string]interface{}),
	}
}

// processWithPipeline processes an intent through the pipeline
func (p *Processor) processWithPipeline(ctx context.Context, intent *common.Intent) error {
	if p.pipeline == nil {
		p.logger.Warn("No processing pipeline configured")
		return nil
	}

	// Create context with timeout
	pipelineCtx, cancel := context.WithTimeout(ctx, time.Duration(p.config.TimeoutSeconds)*time.Second)
	defer cancel()

	return p.pipeline.Process(pipelineCtx, intent)
}

// findHandlers finds appropriate handlers for an intent type
func (p *Processor) findHandlers(intentType string) ([]common.IntentHandler, error) {
	if p.registry == nil {
		return nil, common.NewIntentError(common.ErrorCodeHandlerNotFound, "Handler registry not initialized", "")
	}

	handlers := p.registry.ListHandlers()[intentType]
	if len(handlers) == 0 {
		return nil, common.NewIntentError(common.ErrorCodeHandlerNotFound,
			"No handlers found for intent type", intentType)
	}

	return handlers, nil
}

// executeHandlers executes handlers for an intent
func (p *Processor) executeHandlers(ctx context.Context, intent *common.Intent, handlers []common.IntentHandler) error {
	var lastErr error
	successCount := 0

	for i, handler := range handlers {
		p.logger.Debugf("Executing handler %d/%d for intent %s", i+1, len(handlers), intent.ID)

		// Create handler context with timeout
		handlerCtx, cancel := context.WithTimeout(ctx, time.Duration(p.config.TimeoutSeconds)*time.Second)

		// Execute handler with retry logic
		err := p.executeHandlerWithRetry(handlerCtx, intent, handler)
		cancel()

		if err != nil {
			p.logger.Errorf("Handler execution failed for intent %s: %v", intent.ID, err)
			lastErr = err

			// Continue to next handler if not in strict mode
			if !p.config.EnableAsync {
				continue
			} else {
				// In async mode, we can try other handlers
				continue
			}
		} else {
			successCount++
			p.logger.Debugf("Handler executed successfully for intent %s", intent.ID)
		}
	}

	// Check if at least one handler succeeded
	if successCount == 0 && lastErr != nil {
		return common.WrapError(lastErr, common.ErrorCodeProcessingFailed, "All handlers failed")
	}

	if successCount == 0 {
		return common.NewIntentError(common.ErrorCodeProcessingFailed, "No handlers executed successfully", "")
	}

	p.logger.Infof("Successfully executed %d/%d handlers for intent %s", successCount, len(handlers), intent.ID)
	return nil
}

// updateMetrics updates processing metrics
func (p *Processor) updateMetrics(success bool, latency int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if success {
		p.metrics.ProcessedCount++
	} else {
		p.metrics.FailedCount++
	}

	// Update average latency
	totalProcessed := p.metrics.ProcessedCount + p.metrics.FailedCount
	if totalProcessed > 0 {
		p.metrics.AverageLatency = (p.metrics.AverageLatency*(totalProcessed-1) + latency) / totalProcessed
	}
}

// executeHandlerWithRetry executes a handler with retry logic
func (p *Processor) executeHandlerWithRetry(ctx context.Context, intent *common.Intent, handler common.IntentHandler) error {
	var lastErr error

	for attempt := 0; attempt <= p.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			p.logger.Debugf("Retrying handler execution for intent %s, attempt %d/%d",
				intent.ID, attempt, p.config.RetryAttempts)

			// Wait before retry with exponential backoff
			backoffDuration := time.Duration(attempt*attempt) * 100 * time.Millisecond
			select {
			case <-time.After(backoffDuration):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := handler.Handle(ctx, intent)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	return lastErr
}

// Start starts the processor background processes
func (p *Processor) Start(ctx context.Context) error {
	p.logger.Info("Starting Intent Processor")

	// Start metrics collection if enabled
	go p.startMetricsCollection(ctx)

	return nil
}

// Stop stops the processor
func (p *Processor) Stop(ctx context.Context) error {
	p.logger.Info("Stopping Intent Processor")
	return nil
}

// startMetricsCollection starts background metrics collection
func (p *Processor) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.collectMetrics()
		case <-ctx.Done():
			return
		}
	}
}

// collectMetrics collects and logs processing metrics
func (p *Processor) collectMetrics() {
	p.mu.RLock()
	metrics := *p.metrics
	p.mu.RUnlock()

	p.logger.Infof("Processing metrics - Processed: %d, Failed: %d, Active: %d, Avg Latency: %dms",
		metrics.ProcessedCount, metrics.FailedCount, metrics.ActiveProcessing, metrics.AverageLatency)
}

// GetHealthStatus returns the health status of the processor
func (p *Processor) GetHealthStatus() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := map[string]interface{}{
		"status":            "healthy",
		"processed_count":   p.metrics.ProcessedCount,
		"failed_count":      p.metrics.FailedCount,
		"active_processing": p.metrics.ActiveProcessing,
		"average_latency":   p.metrics.AverageLatency,
		"success_rate":      p.calculateSuccessRate(),
	}

	// Check if processor is overloaded
	if p.metrics.ActiveProcessing > int64(p.config.MaxConcurrentProcessing) {
		status["status"] = "overloaded"
	}

	return status
}

// calculateSuccessRate calculates the success rate
func (p *Processor) calculateSuccessRate() float64 {
	total := p.metrics.ProcessedCount + p.metrics.FailedCount
	if total == 0 {
		return 0.0
	}
	return float64(p.metrics.ProcessedCount) / float64(total) * 100.0
}
