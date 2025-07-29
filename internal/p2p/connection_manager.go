package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// connectionManager connection manager implementation
type connectionManager struct {
	host        host.Host
	ctx         context.Context
	cancel      context.CancelFunc
	connections map[peer.ID]*ConnectionInfo
	mu          sync.RWMutex
	logger      *zap.Logger
	isRunning   bool
	lowWater    int
	highWater   int
}

// NewConnectionManager creates new connection manager
func NewConnectionManager(h host.Host, logger *zap.Logger) ConnectionManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &connectionManager{
		host:        h,
		connections: make(map[peer.ID]*ConnectionInfo),
		logger:      logger.Named("connection_manager"),
	}
}

// Start starts connection manager
func (cm *connectionManager) Start(ctx context.Context, lowWater, highWater int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isRunning {
		return fmt.Errorf("connection manager already running")
	}

	cm.ctx, cm.cancel = context.WithCancel(ctx)
	cm.lowWater = lowWater
	cm.highWater = highWater
	cm.isRunning = true

	// Register network event handlers
	cm.host.Network().Notify(&networkNotifee{cm: cm})

	// Start connection monitoring
	go cm.monitorConnections()

	cm.logger.Info("Connection manager started",
		zap.Int("low_water", lowWater),
		zap.Int("high_water", highWater),
	)

	return nil
}

// Stop stops connection manager
func (cm *connectionManager) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.isRunning {
		return fmt.Errorf("connection manager not running")
	}

	if cm.cancel != nil {
		cm.cancel()
	}

	cm.isRunning = false
	cm.logger.Info("Connection manager stopped")
	return nil
}

// ConnectToPeer connects to specified peer
func (cm *connectionManager) ConnectToPeer(ctx context.Context, peerInfo peer.AddrInfo) error {
	// Check if already connected
	if cm.host.Network().Connectedness(peerInfo.ID) == network.Connected {
		cm.logger.Debug("Already connected to peer",
			zap.String("peer_id", FormatPeerID(peerInfo.ID)),
		)
		return nil
	}

	// Check connection limits
	currentConnections := cm.GetConnectionCount()
	if currentConnections >= cm.highWater {
		return ErrConnectionLimitReached
	}

	// Attempt connection
	if err := cm.host.Connect(ctx, peerInfo); err != nil {
		cm.logger.Debug("Failed to connect to peer",
			zap.String("peer_id", FormatPeerID(peerInfo.ID)),
			zap.Error(err),
		)
		return NewConnectionError(peerInfo.ID, err)
	}

	cm.logger.Info("Successfully connected to peer",
		zap.String("peer_id", FormatPeerID(peerInfo.ID)),
	)

	return nil
}

// DisconnectFromPeer disconnects from specified peer
func (cm *connectionManager) DisconnectFromPeer(peerID peer.ID) error {
	// Check if connected
	if cm.host.Network().Connectedness(peerID) != network.Connected {
		return ErrPeerNotFound
	}

	// Close connection
	if err := cm.host.Network().ClosePeer(peerID); err != nil {
		return NewConnectionError(peerID, err)
	}

	cm.logger.Info("Disconnected from peer",
		zap.String("peer_id", FormatPeerID(peerID)),
	)

	return nil
}

// GetConnectionStatus gets connection status with specified peer
func (cm *connectionManager) GetConnectionStatus(peerID peer.ID) ConnectionStatus {
	connectedness := cm.host.Network().Connectedness(peerID)
	switch connectedness {
	case network.Connected:
		return ConnectionStatusConnected
	case network.CanConnect:
		return ConnectionStatusDisconnected
	case network.CannotConnect:
		return ConnectionStatusFailed
	default:
		return ConnectionStatusDisconnected
	}
}

// GetNetworkTopology gets network topology information
func (cm *connectionManager) GetNetworkTopology() *NetworkTopology {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	connectedPeers := make([]peer.ID, 0)
	connectionMap := make(map[peer.ID]ConnectionInfo)

	// Get all connected peers
	for _, conn := range cm.host.Network().Conns() {
		peerID := conn.RemotePeer()
		connectedPeers = append(connectedPeers, peerID)

		// Get connection info
		if connInfo, exists := cm.connections[peerID]; exists {
			connectionMap[peerID] = *connInfo
		} else {
			// Create basic connection info
			connectionMap[peerID] = ConnectionInfo{
				PeerID:      peerID,
				Status:      ConnectionStatusConnected,
				ConnectedAt: time.Now(), // Approximate
				LastSeen:    time.Now(),
				RemoteAddr:  conn.RemoteMultiaddr(),
				Direction:   conn.Stat().Direction.String(),
			}
		}
	}

	return &NetworkTopology{
		LocalPeerID:      cm.host.ID(),
		ConnectedPeers:   connectedPeers,
		ConnectionMap:    connectionMap,
		CreatedAt:        time.Now(),
		TotalConnections: len(connectedPeers),
	}
}

// SetConnectionLimits sets connection limits
func (cm *connectionManager) SetConnectionLimits(low, high int) error {
	if low < 0 || high < 0 || low > high {
		return fmt.Errorf("invalid connection limits: low=%d, high=%d", low, high)
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.lowWater = low
	cm.highWater = high

	cm.logger.Info("Connection limits updated",
		zap.Int("low_water", low),
		zap.Int("high_water", high),
	)

	return nil
}

// GetConnectionCount gets current connection count
func (cm *connectionManager) GetConnectionCount() int {
	return len(cm.host.Network().Conns())
}

// monitorConnections monitors connection health
func (cm *connectionManager) monitorConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.checkConnectionHealth()
		}
	}
}

// checkConnectionHealth checks connection health and cleanup
func (cm *connectionManager) checkConnectionHealth() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	connectedPeers := make(map[peer.ID]bool)

	// Track currently connected peers
	for _, conn := range cm.host.Network().Conns() {
		connectedPeers[conn.RemotePeer()] = true
	}

	// Clean up disconnected peer records
	for peerID, connInfo := range cm.connections {
		if !connectedPeers[peerID] {
			// Peer is no longer connected
			delete(cm.connections, peerID)
			cm.logger.Debug("Cleaned up disconnected peer record",
				zap.String("peer_id", FormatPeerID(peerID)),
				zap.Duration("duration", now.Sub(connInfo.ConnectedAt)),
			)
		} else {
			// Update last seen time
			connInfo.LastSeen = now
			cm.connections[peerID] = connInfo
		}
	}

	currentCount := len(connectedPeers)
	cm.logger.Debug("Connection health check completed",
		zap.Int("connected_peers", currentCount),
		zap.Int("low_water", cm.lowWater),
		zap.Int("high_water", cm.highWater),
	)
}

// updateConnectionInfo updates connection information
func (cm *connectionManager) updateConnectionInfo(peerID peer.ID, conn network.Conn, connected bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if connected {
		// Peer connected
		connInfo := &ConnectionInfo{
			PeerID:      peerID,
			Status:      ConnectionStatusConnected,
			ConnectedAt: time.Now(),
			LastSeen:    time.Now(),
			RemoteAddr:  conn.RemoteMultiaddr(),
			Direction:   conn.Stat().Direction.String(),
		}
		cm.connections[peerID] = connInfo

		cm.logger.Info("Peer connected",
			zap.String("peer_id", FormatPeerID(peerID)),
			zap.String("direction", connInfo.Direction),
			zap.String("remote_addr", FormatMultiaddr(connInfo.RemoteAddr)),
		)
	} else {
		// Peer disconnected
		if connInfo, exists := cm.connections[peerID]; exists {
			duration := time.Since(connInfo.ConnectedAt)
			cm.logger.Info("Peer disconnected",
				zap.String("peer_id", FormatPeerID(peerID)),
				zap.Duration("duration", duration),
			)
		}
		delete(cm.connections, peerID)
	}
}

// networkNotifee handles network events
type networkNotifee struct {
	cm *connectionManager
}

// Listen handles listen event
func (nn *networkNotifee) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose handles listen close event
func (nn *networkNotifee) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Connected handles peer connected event
func (nn *networkNotifee) Connected(net network.Network, conn network.Conn) {
	nn.cm.updateConnectionInfo(conn.RemotePeer(), conn, true)
}

// Disconnected handles peer disconnected event
func (nn *networkNotifee) Disconnected(net network.Network, conn network.Conn) {
	nn.cm.updateConnectionInfo(conn.RemotePeer(), conn, false)
}