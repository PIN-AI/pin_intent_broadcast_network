package p2p

import (
	"pin_intent_broadcast_network/internal/conf"

	"github.com/google/wire"
)

// ProviderSet P2P module wire provider set
var ProviderSet = wire.NewSet(
	NewNetworkManager,
	NewP2PConfig,
)

// NewP2PConfig creates P2P configuration from bootstrap config
func NewP2PConfig(c *conf.Bootstrap) (*HostConfig, error) {
	if c.P2P == nil {
		return getDefaultConfig(), nil
	}

	config := &HostConfig{
		ListenAddresses: c.P2P.ListenAddresses,
		BootstrapPeers:  c.P2P.BootstrapPeers,
		ProtocolID:      c.P2P.ProtocolId,
		EnableMDNS:      c.P2P.EnableMdns,
		EnableDHT:       c.P2P.EnableDht,
		DataDir:         c.P2P.DataDir,
		MaxConnections:  int(c.P2P.MaxConnections),
		EnableSigning:   c.P2P.EnableSigning,
	}

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// getDefaultConfig returns default P2P configuration
func getDefaultConfig() *HostConfig {
	return &HostConfig{
		ListenAddresses: []string{
			"/ip4/0.0.0.0/tcp/9001",
			"/ip4/0.0.0.0/udp/9001/quic",
		},
		BootstrapPeers:  []string{},
		ProtocolID:      "/intent-broadcast/1.0.0",
		EnableMDNS:      true,
		EnableDHT:       true,
		DataDir:         "./data/p2p",
		MaxConnections:  100,
		EnableSigning:   true,
	}
}