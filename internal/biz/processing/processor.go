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
	// TODO: Implement in task 5.1
	// 1. Validate intent
	// 2. Apply outgoing processing pipeline
	// 3. Update metrics
	return nil
}

// RegisterHandler registers an intent handler
func (p *Processor) RegisterHandler(intentType string, handler common.IntentHandler) error {
	// TODO: Implement in task 5.1
	return p.registry.RegisterHandler(intentType, handler)
}

// UnregisterHandler unregisters an intent handler
func (p *Processor) UnregisterHandler(intentType string) error {
	// TODO: Implement in task 5.1
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
	// TODO: Implement pipeline processing
	return p.pipeline.Process(ctx, intent)
}

// findHandlers finds appropriate handlers for an intent type
func (p *Processor) findHandlers(intentType string) ([]common.IntentHandler, error) {
	// TODO: Implement handler finding logic
	return p.registry.ListHandlers()[intentType], nil
}

// executeHandlers executes handlers for an intent
func (p *Processor) executeHandlers(ctx context.Context, intent *common.Intent, handlers []common.IntentHandler) error {
	// TODO: Implement handler execution logic
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
