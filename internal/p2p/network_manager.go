package p2p

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// networkManager unified network manager implementation
type networkManager struct {
	hostManager       HostManager
	discoveryManager  DiscoveryManager
	connectionManager ConnectionManager
	config            *HostConfig
	logger            *zap.Logger
	isRunning         bool
}

// NewNetworkManager creates new network manager
func NewNetworkManager(logger *zap.Logger) NetworkManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &networkManager{
		logger: logger.Named("network_manager"),
	}
}

// Start starts network manager
func (nm *networkManager) Start(ctx context.Context, config *HostConfig) error {
	if nm.isRunning {
		return fmt.Errorf("network manager already running")
	}

	nm.config = config

	// Create and start host manager
	nm.hostManager = NewHostManager(nm.logger)
	if err := nm.hostManager.Start(ctx, config); err != nil {
		return fmt.Errorf("failed to start host manager: %w", err)
	}

	// Get the libp2p host
	host := nm.hostManager.GetHost()
	if host == nil {
		return fmt.Errorf("failed to get libp2p host")
	}

	// Create and start discovery manager
	nm.discoveryManager = NewDiscoveryManager(host, config, nm.logger)
	if err := nm.discoveryManager.Start(ctx); err != nil {
		nm.logger.Error("Failed to start discovery manager", zap.Error(err))
		nm.hostManager.Stop()
		return fmt.Errorf("failed to start discovery manager: %w", err)
	}

	// Create and start connection manager
	nm.connectionManager = NewConnectionManager(host, nm.logger)
	if connMgr, ok := nm.connectionManager.(*connectionManager); ok {
		if err := connMgr.Start(ctx, config.MaxConnections/4, config.MaxConnections); err != nil {
			nm.logger.Error("Failed to start connection manager", zap.Error(err))
			nm.discoveryManager.Stop()
			nm.hostManager.Stop()
			return fmt.Errorf("failed to start connection manager: %w", err)
		}
	}

	nm.isRunning = true
	nm.logger.Info("Network manager started successfully")
	return nil
}

// Stop stops network manager
func (nm *networkManager) Stop() error {
	if !nm.isRunning {
		return fmt.Errorf("network manager not running")
	}

	var errors []error

	// Stop connection manager
	if nm.connectionManager != nil {
		if err := nm.connectionManager.(*connectionManager).Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop connection manager: %w", err))
		}
	}

	// Stop discovery manager
	if nm.discoveryManager != nil {
		if err := nm.discoveryManager.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop discovery manager: %w", err))
		}
	}

	// Stop host manager
	if nm.hostManager != nil {
		if err := nm.hostManager.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop host manager: %w", err))
		}
	}

	nm.isRunning = false

	if len(errors) > 0 {
		nm.logger.Error("Some errors occurred while stopping network manager",
			zap.Int("error_count", len(errors)),
		)
		// Return the first error
		return errors[0]
	}

	nm.logger.Info("Network manager stopped successfully")
	return nil
}

// GetHostManager gets host manager
func (nm *networkManager) GetHostManager() HostManager {
	return nm.hostManager
}

// GetDiscoveryManager gets discovery manager
func (nm *networkManager) GetDiscoveryManager() DiscoveryManager {
	return nm.discoveryManager
}

// GetConnectionManager gets connection manager
func (nm *networkManager) GetConnectionManager() ConnectionManager {
	return nm.connectionManager
}

// GetNetworkStatus gets network status
func (nm *networkManager) GetNetworkStatus() *NetworkStatus {
	if !nm.isRunning || nm.hostManager == nil {
		return &NetworkStatus{
			IsRunning: false,
		}
	}

	host := nm.hostManager.GetHost()
	if host == nil {
		return &NetworkStatus{
			IsRunning: false,
		}
	}

	var connectedPeersCount int
	var discoveredPeersCount int

	if nm.connectionManager != nil {
		connectedPeersCount = nm.connectionManager.GetConnectionCount()
	}

	if nm.discoveryManager != nil {
		discoveredPeersCount = len(nm.discoveryManager.GetConnectedPeers())
	}

	return &NetworkStatus{
		IsRunning:            nm.isRunning,
		LocalPeerID:          host.ID(),
		ListenAddresses:      nm.hostManager.GetListenAddresses(),
		ConnectedPeersCount:  connectedPeersCount,
		DiscoveredPeersCount: discoveredPeersCount,
		StartTime:            time.Now(), // TODO: track actual start time
		Uptime:               time.Since(time.Now()), // TODO: calculate actual uptime
		Config:               nm.config,
	}
}

// GetNetworkMetrics gets network metrics
func (nm *networkManager) GetNetworkMetrics() *NetworkMetrics {
	metrics := &NetworkMetrics{
		LastUpdated: time.Now(),
	}

	if nm.connectionManager != nil {
		metrics.ConnectedPeers = nm.connectionManager.GetConnectionCount()
	}

	if nm.discoveryManager != nil {
		metrics.DiscoveredPeers = int64(len(nm.discoveryManager.GetConnectedPeers()))
	}

	// TODO: Add more detailed metrics collection
	return metrics
}