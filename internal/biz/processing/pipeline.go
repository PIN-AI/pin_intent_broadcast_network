package processing

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"pin_intent_broadcast_network/internal/biz/common"
)

// Pipeline implements the ProcessingPipeline interface
// It manages the sequential processing of intents through multiple stages
type Pipeline struct {
	stages []common.ProcessingStage
	config *PipelineConfig
}

// PipelineConfig holds configuration for the processing pipeline
type PipelineConfig struct {
	StageTimeout    time.Duration `yaml:"stage_timeout"`
	PipelineTimeout time.Duration `yaml:"pipeline_timeout"`
	MaxRetries      int           `yaml:"max_retries"`
	EnableAsync     bool          `yaml:"enable_async"`
}

// NewPipeline creates a new processing pipeline
func NewPipeline(config *PipelineConfig) *Pipeline {
	return &Pipeline{
		stages: make([]common.ProcessingStage, 0),
		config: config,
	}
}

// AddStage adds a processing stage to the pipeline
func (p *Pipeline) AddStage(stage common.ProcessingStage) error {
	// TODO: Implement in task 5.2
	p.stages = append(p.stages, stage)

	// Sort stages by priority
	sort.Slice(p.stages, func(i, j int) bool {
		return p.stages[i].GetPriority() > p.stages[j].GetPriority()
	})

	return nil
}

// Process processes an intent through all pipeline stages
func (p *Pipeline) Process(ctx context.Context, intent *common.Intent) error {
	if intent == nil {
		return common.NewIntentError(common.ErrorCodeInvalidFormat, "Intent cannot be nil", "")
	}

	if len(p.stages) == 0 {
		return nil // No stages to process
	}

	// Create pipeline context with timeout
	pipelineCtx, cancel := context.WithTimeout(ctx, p.config.PipelineTimeout)
	defer cancel()

	processedStages := 0

	for i, stage := range p.stages {
		// Check if stage should process this intent
		if !stage.ShouldProcess(intent) {
			continue
		}

		// Create stage context with timeout
		stageCtx, stageCancel := context.WithTimeout(pipelineCtx, p.config.StageTimeout)

		// Process stage with retry logic
		err := p.processStageWithRetry(stageCtx, stage, intent)
		stageCancel()

		if err != nil {
			return common.WrapError(err, common.ErrorCodeProcessingFailed,
				fmt.Sprintf("Stage '%s' failed at position %d", stage.Name(), i))
		}

		processedStages++
	}

	if processedStages == 0 {
		return common.NewIntentError(common.ErrorCodeProcessingFailed,
			"No pipeline stages processed the intent", intent.Type)
	}

	return nil
}

// GetStages returns all pipeline stages
func (p *Pipeline) GetStages() []common.ProcessingStage {
	// Return a copy to prevent external modification
	stages := make([]common.ProcessingStage, len(p.stages))
	copy(stages, p.stages)
	return stages
}

// processStageWithRetry processes a stage with retry logic
func (p *Pipeline) processStageWithRetry(ctx context.Context, stage common.ProcessingStage, intent *common.Intent) error {
	var lastErr error

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		err := stage.Process(ctx, intent)
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

		// Wait before retry (exponential backoff could be added here)
		if attempt < p.config.MaxRetries {
			time.Sleep(time.Millisecond * 100 * time.Duration(attempt+1))
		}
	}

	return lastErr
}

// ValidationStage implements basic validation processing stage
type ValidationStage struct {
	name      string
	validator common.IntentValidator
	priority  int
}

// NewValidationStage creates a new validation stage
func NewValidationStage(validator common.IntentValidator) *ValidationStage {
	return &ValidationStage{
		name:      "validation",
		validator: validator,
		priority:  100, // High priority
	}
}

// Name returns the stage name
func (vs *ValidationStage) Name() string {
	return vs.name
}

// Process processes the intent through validation
func (vs *ValidationStage) Process(ctx context.Context, intent *common.Intent) error {
	return vs.validator.ValidateIntent(ctx, intent)
}

// ShouldProcess determines if this stage should process the intent
func (vs *ValidationStage) ShouldProcess(intent *common.Intent) bool {
	// Always validate intents
	return true
}

// GetPriority returns the stage priority
func (vs *ValidationStage) GetPriority() int {
	return vs.priority
}

// SignatureStage implements signature verification processing stage
type SignatureStage struct {
	name     string
	signer   common.IntentSigner
	priority int
}

// NewSignatureStage creates a new signature stage
func NewSignatureStage(signer common.IntentSigner) *SignatureStage {
	return &SignatureStage{
		name:     "signature",
		signer:   signer,
		priority: 90, // High priority, after validation
	}
}

// Name returns the stage name
func (ss *SignatureStage) Name() string {
	return ss.name
}

// Process processes the intent through signature verification
func (ss *SignatureStage) Process(ctx context.Context, intent *common.Intent) error {
	return ss.signer.VerifySignature(intent)
}

// ShouldProcess determines if this stage should process the intent
func (ss *SignatureStage) ShouldProcess(intent *common.Intent) bool {
	// Only process intents with signatures
	return len(intent.Signature) > 0
}

// GetPriority returns the stage priority
func (ss *SignatureStage) GetPriority() int {
	return ss.priority
}

// EnrichmentStage implements intent enrichment processing stage
type EnrichmentStage struct {
	name     string
	priority int
}

// NewEnrichmentStage creates a new enrichment stage
func NewEnrichmentStage() *EnrichmentStage {
	return &EnrichmentStage{
		name:     "enrichment",
		priority: 80, // Medium priority
	}
}

// Name returns the stage name
func (es *EnrichmentStage) Name() string {
	return es.name
}

// Process processes the intent through enrichment
func (es *EnrichmentStage) Process(ctx context.Context, intent *common.Intent) error {
	// Add processing timestamp if not set
	if intent.ProcessedAt == 0 {
		intent.ProcessedAt = time.Now().Unix()
	}

	// Add default metadata if missing
	if intent.Metadata == nil {
		intent.Metadata = make(map[string]string)
	}

	// Add processing metadata
	intent.Metadata["processed_by"] = "pipeline"
	intent.Metadata["processing_stage"] = "enrichment"
	intent.Metadata["enriched_at"] = time.Now().Format(time.RFC3339)

	return nil
}

// ShouldProcess determines if this stage should process the intent
func (es *EnrichmentStage) ShouldProcess(intent *common.Intent) bool {
	// Always enrich intents
	return true
}

// GetPriority returns the stage priority
func (es *EnrichmentStage) GetPriority() int {
	return es.priority
}

// TransformationStage implements intent transformation processing stage
type TransformationStage struct {
	name         string
	priority     int
	transformers map[string]IntentTransformer
}

// IntentTransformer defines the interface for intent transformers
type IntentTransformer interface {
	Transform(ctx context.Context, intent *common.Intent) error
	GetName() string
}

// NewTransformationStage creates a new transformation stage
func NewTransformationStage() *TransformationStage {
	return &TransformationStage{
		name:         "transformation",
		priority:     70, // Medium priority
		transformers: make(map[string]IntentTransformer),
	}
}

// Name returns the stage name
func (ts *TransformationStage) Name() string {
	return ts.name
}

// Process processes the intent through transformation
func (ts *TransformationStage) Process(ctx context.Context, intent *common.Intent) error {
	// Apply type-specific transformers
	if transformer, exists := ts.transformers[intent.Type]; exists {
		if err := transformer.Transform(ctx, intent); err != nil {
			return common.WrapError(err, common.ErrorCodeProcessingFailed,
				fmt.Sprintf("Transformation failed for type %s", intent.Type))
		}
	}

	// Apply default transformations
	return ts.applyDefaultTransformations(ctx, intent)
}

// ShouldProcess determines if this stage should process the intent
func (ts *TransformationStage) ShouldProcess(intent *common.Intent) bool {
	// Process if we have transformers for this type or default transformations
	_, hasTransformer := ts.transformers[intent.Type]
	return hasTransformer || true // Always apply default transformations
}

// GetPriority returns the stage priority
func (ts *TransformationStage) GetPriority() int {
	return ts.priority
}

// AddTransformer adds a transformer for a specific intent type
func (ts *TransformationStage) AddTransformer(intentType string, transformer IntentTransformer) {
	ts.transformers[intentType] = transformer
}

// applyDefaultTransformations applies default transformations to all intents
func (ts *TransformationStage) applyDefaultTransformations(ctx context.Context, intent *common.Intent) error {
	// Normalize intent type to lowercase
	intent.Type = strings.ToLower(intent.Type)

	// Ensure priority is within valid range
	if intent.Priority < common.PriorityLow {
		intent.Priority = common.PriorityLow
	} else if intent.Priority > common.PriorityUrgent {
		intent.Priority = common.PriorityUrgent
	}

	// Set default TTL if not specified
	if intent.TTL == 0 {
		intent.TTL = int64(common.DefaultTTL.Seconds())
	}

	return nil
}

// FilteringStage implements intent filtering processing stage
type FilteringStage struct {
	name     string
	priority int
	filters  []IntentFilter
}

// IntentFilter defines the interface for intent filters
type IntentFilter interface {
	ShouldAllow(ctx context.Context, intent *common.Intent) (bool, error)
	GetName() string
}

// NewFilteringStage creates a new filtering stage
func NewFilteringStage() *FilteringStage {
	return &FilteringStage{
		name:     "filtering",
		priority: 60, // Lower priority
		filters:  make([]IntentFilter, 0),
	}
}

// Name returns the stage name
func (fs *FilteringStage) Name() string {
	return fs.name
}

// Process processes the intent through filtering
func (fs *FilteringStage) Process(ctx context.Context, intent *common.Intent) error {
	for _, filter := range fs.filters {
		allowed, err := filter.ShouldAllow(ctx, intent)
		if err != nil {
			return common.WrapError(err, common.ErrorCodeProcessingFailed,
				fmt.Sprintf("Filter '%s' failed", filter.GetName()))
		}

		if !allowed {
			return common.NewIntentError(common.ErrorCodeProcessingFailed,
				fmt.Sprintf("Intent filtered out by '%s'", filter.GetName()), intent.ID)
		}
	}

	return nil
}

// ShouldProcess determines if this stage should process the intent
func (fs *FilteringStage) ShouldProcess(intent *common.Intent) bool {
	// Only process if we have filters
	return len(fs.filters) > 0
}

// GetPriority returns the stage priority
func (fs *FilteringStage) GetPriority() int {
	return fs.priority
}

// AddFilter adds a filter to the filtering stage
func (fs *FilteringStage) AddFilter(filter IntentFilter) {
	fs.filters = append(fs.filters, filter)
}

// GetPipelineStats returns pipeline processing statistics
func (p *Pipeline) GetPipelineStats() map[string]interface{} {
	return map[string]interface{}{
		"total_stages":     len(p.stages),
		"pipeline_timeout": p.config.PipelineTimeout.String(),
		"stage_timeout":    p.config.StageTimeout.String(),
		"max_retries":      p.config.MaxRetries,
		"async_enabled":    p.config.EnableAsync,
		"stage_names":      p.getStageNames(),
	}
}

// getStageNames returns the names of all stages in the pipeline
func (p *Pipeline) getStageNames() []string {
	names := make([]string, len(p.stages))
	for i, stage := range p.stages {
		names[i] = stage.Name()
	}
	return names
}
