package execution

import (
	"pin_intent_broadcast_network/internal/p2p"
	"pin_intent_broadcast_network/internal/transport"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ProviderSet is the wire provider set for execution module
var ProviderSet = wire.NewSet(
	NewAutomationManager,
	NewAsyncAutomationManagerForWire,
	NewTransportReadinessCheckerForWire,
	NewComponentLifecycleManagerForWire,
)

// NewTransportReadinessCheckerForWire creates a transport readiness checker for wire injection
func NewTransportReadinessCheckerForWire(networkManager p2p.NetworkManager, transportMgr transport.TransportManager, logger *zap.Logger) *transport.TransportReadinessChecker {
	return transport.NewTransportReadinessChecker(networkManager, transportMgr, logger)
}

// NewComponentLifecycleManagerForWire creates a component lifecycle manager for wire injection
func NewComponentLifecycleManagerForWire(logger *zap.Logger) *ComponentLifecycleManager {
	return NewComponentLifecycleManager(logger)
}

// NewAsyncAutomationManagerForWire creates an async automation manager for wire injection
func NewAsyncAutomationManagerForWire(
	baseManager *AutomationManager,
	networkManager p2p.NetworkManager,
	transportMgr transport.TransportManager,
	logger *zap.Logger,
) *AsyncAutomationManager {
	return NewAsyncAutomationManager(baseManager, networkManager, transportMgr, logger)
}