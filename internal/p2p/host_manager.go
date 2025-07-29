package p2p

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	connmgr "github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// hostManager libp2p host manager implementation
type hostManager struct {
	host      host.Host
	ctx       context.Context
	cancel    context.CancelFunc
	config    *HostConfig
	connMgr   *connmgr.BasicConnMgr
	isRunning atomic.Bool
	mu        sync.RWMutex
	logger    *zap.Logger
}

// NewHostManager creates new host manager
func NewHostManager(logger *zap.Logger) HostManager {
	if logger == nil {
		logger = zap.NewNop()
	}
	
	return &hostManager{
		logger: logger.Named("host_manager"),
	}
}

// Start starts P2P host
func (hm *hostManager) Start(ctx context.Context, config *HostConfig) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Check if already running
	if hm.isRunning.Load() {
		return ErrHostAlreadyRunning
	}

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	hm.config = config
	hm.ctx, hm.cancel = context.WithCancel(ctx)

	// Ensure data directory exists
	if err := EnsureDataDir(config.DataDir); err != nil {
		return fmt.Errorf("failed to ensure data directory: %w", err)
	}

	// Generate or load private key
	privateKey, err := GenerateOrLoadPrivateKey(config.DataDir)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	// Parse listen addresses
	listenAddrs, err := ParseListenAddresses(config.ListenAddresses)
	if err != nil {
		return fmt.Errorf("failed to parse listen addresses: %w", err)
	}

	// Create connection manager
	connMgr, err := hm.createConnectionManager(config)
	if err != nil {
		return fmt.Errorf("failed to create connection manager: %w", err)
	}
	hm.connMgr = connMgr

	// Create libp2p host
	h, err := hm.createLibp2pHost(privateKey, listenAddrs, connMgr)
	if err != nil {
		return fmt.Errorf("failed to create libp2p host: %w", err)
	}
	hm.host = h

	// Mark as running
	hm.isRunning.Store(true)

	hm.logger.Info("P2P host started successfully",
		zap.String("peer_id", h.ID().String()),
		zap.Strings("listen_addresses", config.ListenAddresses),
		zap.String("protocol_id", config.ProtocolID),
		zap.Int("max_connections", config.MaxConnections),
	)

	// Print listen addresses
	for _, addr := range h.Addrs() {
		hm.logger.Info("Listening on address",
			zap.String("address", fmt.Sprintf("%s/p2p/%s", addr, h.ID())),
		)
	}

	return nil
}

// Stop stops P2P host
func (hm *hostManager) Stop() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.isRunning.Load() {
		return ErrHostNotRunning
	}

	// Cancel context
	if hm.cancel != nil {
		hm.cancel()
	}

	// Close host
	if hm.host != nil {
		if err := hm.host.Close(); err != nil {
			hm.logger.Error("Failed to close libp2p host", zap.Error(err))
			return fmt.Errorf("failed to close host: %w", err)
		}
	}

	// Mark as stopped
	hm.isRunning.Store(false)

	hm.logger.Info("P2P host stopped successfully")
	return nil
}

// GetHost gets libp2p host instance
func (hm *hostManager) GetHost() host.Host {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.host
}

// GetPeerID gets local peer ID
func (hm *hostManager) GetPeerID() peer.ID {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	
	if hm.host == nil {
		return ""
	}
	return hm.host.ID()
}

// IsRunning checks if host is running
func (hm *hostManager) IsRunning() bool {
	return hm.isRunning.Load()
}

// GetListenAddresses gets listen address list
func (hm *hostManager) GetListenAddresses() []multiaddr.Multiaddr {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	
	if hm.host == nil {
		return nil
	}
	return hm.host.Addrs()
}

// createConnectionManager creates connection manager
func (hm *hostManager) createConnectionManager(config *HostConfig) (*connmgr.BasicConnMgr, error) {
	// Set connection limits
	lowWater := config.MaxConnections / 4
	highWater := config.MaxConnections

	// Create connection manager
	connMgr, err := connmgr.NewConnManager(lowWater, highWater, connmgr.WithGracePeriod(time.Minute))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	hm.logger.Info("Connection manager created",
		zap.Int("low_water", lowWater),
		zap.Int("high_water", highWater),
	)

	return connMgr, nil
}

// createLibp2pHost creates libp2p host
func (hm *hostManager) createLibp2pHost(privateKey crypto.PrivKey, listenAddrs []multiaddr.Multiaddr, connMgr *connmgr.BasicConnMgr) (host.Host, error) {
	// Build libp2p options
	opts := []libp2p.Option{
		// Identity
		libp2p.Identity(privateKey),
		
		// Listen addresses
		libp2p.ListenAddrs(listenAddrs...),
		
		// Transport protocols
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(libp2pquic.NewTransport),
		
		// Security protocols
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		
		// Stream multiplexing
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
		
		// Connection management
		libp2p.ConnectionManager(connMgr),
		
		// Enable auto relay
		libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{}),
		
		// Enable NAT traversal
		libp2p.EnableNATService(),
		
		// Enable relay service
		libp2p.EnableRelayService(),
	}

	// Create host
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	return h, nil
}