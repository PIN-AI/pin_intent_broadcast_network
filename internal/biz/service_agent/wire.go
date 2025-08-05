package service_agent

import (
	"github.com/google/wire"
	"go.uber.org/zap"
	
	"pin_intent_broadcast_network/internal/transport"
)

// ProviderSet is service agent providers.
var ProviderSet = wire.NewSet(
	NewAgent,
	NewIntentListener,
	NewBidDecisionManager,
	DefaultAgentConfig,
)

// NewAgentFromConfig creates a new agent with the provided config
func NewAgentFromConfig(config *AgentConfig, transportMgr transport.TransportManager, logger *zap.Logger) *Agent {
	return NewAgent(config, transportMgr, logger)
}