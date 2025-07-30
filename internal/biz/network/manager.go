package network

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/libp2p/go-libp2p/core/peer"

	"pin_intent_broadcast_network/internal/biz/common"
)

// Manager implements the NetworkManager interface
// This file will contain the implementation for task 7.1
type Manager struct {
	// hostManager p2p.HostManager  // Will be integrated in task 8.2
	// transport   transport.MessageTransport  // Will be integrated in task 8.3
	status   *Status
	topology *Topology
	config   *ManagerConfig
	logger   *log.Helper
	mu       sync.RWMutex
}

// ManagerConfig holds configuration for the network manager
type ManagerConfig struct {
	EnableTopologyTracking bool `yaml:"enable_topology_tracking"`
	StatusUpdateInterval   int  `yaml:"status_update_interval"`
	MaxPeers               int  `yaml:"max_peers"`
	EnableMetrics          bool `yaml:"enable_metrics"`
}

// NewManager creates a new network manager
func NewManager(config *ManagerConfig, logger log.Logger) *Manager {
	return &Manager{
		status:   NewStatus(),
		topology: NewTopology(),
		config:   config,
		logger:   log.NewHelper(logger),
	}
}

// GetNetworkStatus returns the current network status
func (nm *Manager) GetNetworkStatus(ctx context.Context) (*common.NetworkStatusResponse, error) {
	// TODO: Implement in task 7.1
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// This will be implemented once P2P integration is complete
	return &common.NetworkStatusResponse{
		PeerCount:        0,          // nm.hostManager.GetPeerCount(),
		ConnectedPeers:   []string{}, // nm.hostManager.GetConnectedPeerIDs(),
		NetworkHealth:    "unknown",  // nm.status.GetHealthStatus(),
		TopicCount:       0,          // nm.transport.GetTopicCount(),
		MessagesSent:     0,          // nm.status.GetMessagesSent(),
		MessagesReceived: 0,          // nm.status.GetMessagesReceived(),
		Metrics:          make(map[string]interface{}),
	}, nil
}

// GetConnectedPeers returns a list of connected peers
func (nm *Manager) GetConnectedPeers() []peer.ID {
	// TODO: Implement in task 7.1
	// This will be implemented once P2P integration is complete
	return []peer.ID{}
}

// GetNetworkTopology returns the current network topology
func (nm *Manager) GetNetworkTopology() *common.NetworkTopology {
	// TODO: Implement in task 7.1
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	return nm.topology.GetTopology()
}

// HandleNetworkEvent handles network events
func (nm *Manager) HandleNetworkEvent(event common.NetworkEvent) error {
	// TODO: Implement in task 7.1
	nm.logger.Infof("Handling network event: %s for peer %s", event.Type, event.PeerID)

	switch event.Type {
	case "peer_connected":
		return nm.handlePeerConnected(event)
	case "peer_disconnected":
		return nm.handlePeerDisconnected(event)
	case "message_received":
		return nm.handleMessageReceived(event)
	case "message_sent":
		return nm.handleMessageSent(event)
	default:
		nm.logger.Warnf("Unknown network event type: %s", event.Type)
	}

	return nil
}

// handlePeerConnected handles peer connection events
func (nm *Manager) handlePeerConnected(event common.NetworkEvent) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Update topology
	nm.topology.AddPeer(event.PeerID)

	// Update status
	nm.status.IncrementConnectedPeers()

	nm.logger.Infof("Peer connected: %s", event.PeerID)
	return nil
}

// handlePeerDisconnected handles peer disconnection events
func (nm *Manager) handlePeerDisconnected(event common.NetworkEvent) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Update topology
	nm.topology.RemovePeer(event.PeerID)

	// Update status
	nm.status.DecrementConnectedPeers()

	nm.logger.Infof("Peer disconnected: %s", event.PeerID)
	return nil
}

// handleMessageReceived handles message received events
func (nm *Manager) handleMessageReceived(event common.NetworkEvent) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Update status
	nm.status.IncrementMessagesReceived()

	return nil
}

// handleMessageSent handles message sent events
func (nm *Manager) handleMessageSent(event common.NetworkEvent) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Update status
	nm.status.IncrementMessagesSent()

	return nil
}

// Start starts the network manager
func (nm *Manager) Start(ctx context.Context) error {
	nm.logger.Info("Starting network manager")

	// Start status monitoring
	if nm.config.EnableMetrics {
		go nm.status.StartMonitoring(ctx)
	}

	// Start topology tracking
	if nm.config.EnableTopologyTracking {
		go nm.topology.StartTracking(ctx)
	}

	return nil
}

// Stop stops the network manager
func (nm *Manager) Stop(ctx context.Context) error {
	nm.logger.Info("Stopping network manager")

	// Stop status monitoring
	nm.status.StopMonitoring()

	// Stop topology tracking
	nm.topology.StopTracking()

	return nil
}

// GetMetrics returns network metrics
func (nm *Manager) GetMetrics() map[string]interface{} {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	return map[string]interface{}{
		"connected_peers":   nm.status.GetConnectedPeers(),
		"messages_sent":     nm.status.GetMessagesSent(),
		"messages_received": nm.status.GetMessagesReceived(),
		"network_health":    nm.status.GetHealthStatus(),
		"topology_nodes":    len(nm.topology.GetTopology().Nodes),
		"topology_edges":    len(nm.topology.GetTopology().Edges),
	}
}
