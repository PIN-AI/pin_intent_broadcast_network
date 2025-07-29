package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"go.uber.org/zap"
)

// discoveryManager node discovery manager implementation
type discoveryManager struct {
	host             host.Host
	config           *HostConfig
	ctx              context.Context
	cancel           context.CancelFunc
	mdnsService      mdns.Service
	dhtService       *dht.IpfsDHT
	routingDiscovery *routing.RoutingDiscovery
	mu               sync.RWMutex
	peers            map[peer.ID]peer.AddrInfo
	logger           *zap.Logger
	isRunning        bool
}

// NewDiscoveryManager creates new discovery manager
func NewDiscoveryManager(h host.Host, config *HostConfig, logger *zap.Logger) DiscoveryManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &discoveryManager{
		host:   h,
		config: config,
		peers:  make(map[peer.ID]peer.AddrInfo),
		logger: logger.Named("discovery_manager"),
	}
}

// Start starts node discovery service
func (dm *discoveryManager) Start(ctx context.Context) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.isRunning {
		return fmt.Errorf("discovery manager already running")
	}

	dm.ctx, dm.cancel = context.WithCancel(ctx)
	dm.isRunning = true

	// Start mDNS discovery
	if dm.config.EnableMDNS {
		if err := dm.startMDNS(); err != nil {
			dm.logger.Error("Failed to start mDNS discovery", zap.Error(err))
			return fmt.Errorf("failed to start mDNS: %w", err)
		}
		dm.logger.Info("mDNS discovery started")
	}

	// Start DHT discovery
	if dm.config.EnableDHT {
		if err := dm.startDHT(); err != nil {
			dm.logger.Error("Failed to start DHT discovery", zap.Error(err))
			return fmt.Errorf("failed to start DHT: %w", err)
		}
		dm.logger.Info("DHT discovery started")
	}

	// Connect to bootstrap peers
	if err := dm.ConnectToBootstrapPeers(dm.ctx); err != nil {
		dm.logger.Warn("Failed to connect to some bootstrap peers", zap.Error(err))
		// Don't return error for bootstrap connection failures
	}

	dm.logger.Info("Discovery manager started successfully")
	return nil
}

// Stop stops node discovery service
func (dm *discoveryManager) Stop() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if !dm.isRunning {
		return fmt.Errorf("discovery manager not running")
	}

	// Cancel context
	if dm.cancel != nil {
		dm.cancel()
	}

	// Stop mDNS service
	if dm.mdnsService != nil {
		dm.mdnsService.Close()
		dm.mdnsService = nil
	}

	// Stop DHT service
	if dm.dhtService != nil {
		dm.dhtService.Close()
		dm.dhtService = nil
	}

	dm.isRunning = false
	dm.logger.Info("Discovery manager stopped successfully")
	return nil
}

// DiscoverPeers discovers peers in specified namespace
func (dm *discoveryManager) DiscoverPeers(ctx context.Context, namespace string) (<-chan peer.AddrInfo, error) {
	if !dm.isRunning {
		return nil, fmt.Errorf("discovery manager not running")
	}

	if dm.routingDiscovery == nil {
		return nil, ErrDHTNotEnabled
	}

	peerChan, err := dm.routingDiscovery.FindPeers(ctx, namespace)
	if err != nil {
		return nil, NewDiscoveryError(namespace, err)
	}

	dm.logger.Info("Started peer discovery",
		zap.String("namespace", namespace),
	)

	return peerChan, nil
}

// Advertise advertises self in specified namespace
func (dm *discoveryManager) Advertise(ctx context.Context, namespace string) error {
	if !dm.isRunning {
		return fmt.Errorf("discovery manager not running")
	}

	if dm.routingDiscovery == nil {
		return ErrDHTNotEnabled
	}

	_, err := dm.routingDiscovery.Advertise(ctx, namespace)
	if err != nil {
		return NewDiscoveryError(namespace, err)
	}

	dm.logger.Info("Started advertising",
		zap.String("namespace", namespace),
	)

	return nil
}

// GetConnectedPeers gets connected peer list
func (dm *discoveryManager) GetConnectedPeers() []peer.ID {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	peers := make([]peer.ID, 0, len(dm.peers))
	for id := range dm.peers {
		peers = append(peers, id)
	}
	return peers
}

// ConnectToBootstrapPeers connects to bootstrap peers
func (dm *discoveryManager) ConnectToBootstrapPeers(ctx context.Context) error {
	if len(dm.config.BootstrapPeers) == 0 {
		dm.logger.Info("No bootstrap peers configured")
		return nil
	}

	bootstrapPeers, err := ParseBootstrapPeers(dm.config.BootstrapPeers)
	if err != nil {
		return fmt.Errorf("failed to parse bootstrap peers: %w", err)
	}

	var successCount int
	for i, peerInfo := range bootstrapPeers {
		if err := dm.host.Connect(ctx, peerInfo); err != nil {
			dm.logger.Warn("Failed to connect to bootstrap peer",
				zap.Int("index", i),
				zap.String("peer_id", peerInfo.ID.String()),
				zap.Error(err),
			)
			continue
		}

		dm.addPeer(peerInfo)
		successCount++
		dm.logger.Info("Connected to bootstrap peer",
			zap.String("peer_id", peerInfo.ID.String()),
		)
	}

	dm.logger.Info("Bootstrap connection completed",
		zap.Int("total_peers", len(bootstrapPeers)),
		zap.Int("successful_connections", successCount),
	)

	return nil
}

// startMDNS starts mDNS discovery service
func (dm *discoveryManager) startMDNS() error {
	notifee := &discoveryNotifee{dm: dm}
	mdnsService := mdns.NewMdnsService(dm.host, dm.config.ProtocolID, notifee)
	
	if err := mdnsService.Start(); err != nil {
		return fmt.Errorf("failed to start mDNS service: %w", err)
	}

	dm.mdnsService = mdnsService
	return nil
}

// startDHT starts DHT discovery service
func (dm *discoveryManager) startDHT() error {
	// Create DHT
	dhtService, err := dht.New(dm.ctx, dm.host)
	if err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}

	// Bootstrap DHT
	if err := dhtService.Bootstrap(dm.ctx); err != nil {
		dhtService.Close()
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	dm.dhtService = dhtService
	dm.routingDiscovery = routing.NewRoutingDiscovery(dhtService)

	// Start periodic advertising
	go dm.advertiseLoop()

	// Start periodic peer finding
	go dm.findPeersLoop()

	return nil
}

// advertiseLoop periodically advertises self
func (dm *discoveryManager) advertiseLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			if dm.routingDiscovery != nil {
				if _, err := dm.routingDiscovery.Advertise(dm.ctx, dm.config.ProtocolID); err != nil {
					dm.logger.Debug("Failed to advertise", zap.Error(err))
				}
			}
		}
	}
}

// findPeersLoop periodically finds peers
func (dm *discoveryManager) findPeersLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			if dm.routingDiscovery != nil {
				peerChan, err := dm.routingDiscovery.FindPeers(dm.ctx, dm.config.ProtocolID)
				if err != nil {
					dm.logger.Debug("Failed to find peers", zap.Error(err))
					continue
				}

				go dm.handleDiscoveredPeers(peerChan)
			}
		}
	}
}

// handleDiscoveredPeers handles discovered peers
func (dm *discoveryManager) handleDiscoveredPeers(peerChan <-chan peer.AddrInfo) {
	for peerInfo := range peerChan {
		// Skip self
		if peerInfo.ID == dm.host.ID() {
			continue
		}

		// Try to connect
		ctx, cancel := context.WithTimeout(dm.ctx, 10*time.Second)
		if err := dm.host.Connect(ctx, peerInfo); err != nil {
			dm.logger.Debug("Failed to connect to discovered peer",
				zap.String("peer_id", peerInfo.ID.String()),
				zap.Error(err),
			)
			cancel()
			continue
		}
		cancel()

		dm.addPeer(peerInfo)
		dm.logger.Info("Connected to discovered peer",
			zap.String("peer_id", peerInfo.ID.String()),
		)
	}
}

// addPeer adds peer to peer list
func (dm *discoveryManager) addPeer(peerInfo peer.AddrInfo) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.peers[peerInfo.ID] = peerInfo
}

// removePeer removes peer from peer list
func (dm *discoveryManager) removePeer(peerID peer.ID) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.peers, peerID)
}

// discoveryNotifee handles mDNS discovery notifications
type discoveryNotifee struct {
	dm *discoveryManager
}

// HandlePeerFound handles found peer event
func (n *discoveryNotifee) HandlePeerFound(peerInfo peer.AddrInfo) {
	// Skip self
	if peerInfo.ID == n.dm.host.ID() {
		return
	}

	// Try to connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := n.dm.host.Connect(ctx, peerInfo); err != nil {
		n.dm.logger.Debug("Failed to connect to mDNS discovered peer",
			zap.String("peer_id", peerInfo.ID.String()),
			zap.Error(err),
		)
		return
	}

	n.dm.addPeer(peerInfo)
	n.dm.logger.Info("Connected to mDNS discovered peer",
		zap.String("peer_id", peerInfo.ID.String()),
	)
}