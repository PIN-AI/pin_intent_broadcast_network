package processing

import (
	"context"
	"sort"
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
	// TODO: Implement in task 5.2

	// Create pipeline context with timeout
	pipelineCtx, cancel := context.WithTimeout(ctx, p.config.PipelineTimeout)
	defer cancel()

	for _, stage := range p.stages {
		if !stage.ShouldProcess(intent) {
			continue
		}

		// Create stage context with timeout
		stageCtx, stageCancel := context.WithTimeout(pipelineCtx, p.config.StageTimeout)

		err := p.processStageWithRetry(stageCtx, stage, intent)
		stageCancel()

		if err != nil {
			return err
		}
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
