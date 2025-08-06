//go:build wireinject
// +build wireinject

package processing

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"pin_intent_broadcast_network/internal/biz/common"
)

// ProviderSet is processing providers.
var ProviderSet = wire.NewSet(
	NewProcessor,
	NewPipeline,
	NewHandlerRegistry,
	NewRoutingEngine,
	NewProcessorConfig,
	NewPipelineConfig,
	NewRoutingConfig,
	wire.Bind(new(common.IntentProcessor), new(*Processor)),
	wire.Bind(new(common.ProcessingPipeline), new(*Pipeline)),
	wire.Bind(new(common.HandlerRegistry), new(*HandlerRegistry)),
)

// NewProcessor creates a new processor with Wire injection
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

// NewProcessorConfig creates a default processor configuration
func NewProcessorConfig() *ProcessorConfig {
	return &ProcessorConfig{
		MaxConcurrentProcessing: 100,
		EnableAsync:             true,
		RetryAttempts:           3,
		TimeoutSeconds:          30,
	}
}

// NewPipelineConfig creates a default pipeline configuration
func NewPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		StageTimeout:    10 * time.Second,
		PipelineTimeout: 60 * time.Second,
		MaxRetries:      3,
		EnableAsync:     true,
	}
}

// NewRoutingConfig creates a default routing configuration
func NewRoutingConfig() *RoutingConfig {
	return &RoutingConfig{
		DefaultStrategy:     "type_based",
		EnableLoadBalancing: true,
		MaxRetries:          3,
		FallbackEnabled:     true,
	}
}
