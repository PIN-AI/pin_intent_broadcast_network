package block_builder

import (
	"github.com/google/wire"
	"go.uber.org/zap"
	
	"pin_intent_broadcast_network/internal/transport"
)

// ProviderSet is block builder providers.
var ProviderSet = wire.NewSet(
	NewBlockBuilder,
	NewMatchingEngine,
	DefaultBuilderConfig,
)

// NewBlockBuilderFromConfig creates a new block builder with the provided config
func NewBlockBuilderFromConfig(config *BuilderConfig, transportMgr transport.TransportManager, logger *zap.Logger) *BlockBuilder {
	return NewBlockBuilder(config, transportMgr, logger)
}