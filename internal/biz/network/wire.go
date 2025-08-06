//go:build wireinject
// +build wireinject

package network

import (
	"github.com/google/wire"

	"pin_intent_broadcast_network/internal/biz/common"
)

// ProviderSet is network providers.
var ProviderSet = wire.NewSet(
	NewManager,
	NewStatus,
	NewTopology,
	NewManagerConfig,
	wire.Bind(new(common.NetworkManager), new(*Manager)),
)

// NewManagerConfig creates a default network manager configuration
func NewManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		EnableTopologyTracking: true,
		StatusUpdateInterval:   30,
		MaxPeers:               100,
		EnableMetrics:          true,
	}
}
