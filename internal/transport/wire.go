package transport

import (
	"pin_intent_broadcast_network/internal/conf"
	"pin_intent_broadcast_network/internal/p2p"

	"github.com/google/wire"
	"go.uber.org/zap"
)

// ProviderSet Transport module wire provider set
var ProviderSet = wire.NewSet(
	NewTransportManager,
	NewTransportConfig,
)

// NewTransportConfig creates transport configuration from bootstrap config
func NewTransportConfig(c *conf.Bootstrap) (*TransportConfig, error) {
	if c.Transport == nil {
		return DefaultTransportConfig(), nil
	}
	
	config := &TransportConfig{
		EnableGossipSub:                   c.Transport.EnableGossipsub,
		GossipSubHeartbeatInterval:        c.Transport.GossipsubHeartbeatInterval.AsDuration(),
		GossipSubD:                        int(c.Transport.GossipsubD),
		GossipSubDLo:                      int(c.Transport.GossipsubDLo),
		GossipSubDHi:                      int(c.Transport.GossipsubDHi),
		GossipSubFanoutTTL:                c.Transport.GossipsubFanoutTtl.AsDuration(),
		EnableMessageSigning:              c.Transport.EnableMessageSigning,
		EnableStrictSignatureVerification: c.Transport.EnableStrictSignatureVerification,
		MessageIDCacheSize:                int(c.Transport.MessageIdCacheSize),
		MessageTTL:                        c.Transport.MessageTtl.AsDuration(),
		MaxMessageSize:                    int(c.Transport.MaxMessageSize),
	}
	
	return config, nil
}

// NewTransportManagerWithP2P creates transport manager with P2P network manager
func NewTransportManagerWithP2P(networkManager p2p.NetworkManager, logger *zap.Logger) TransportManager {
	if networkManager == nil {
		logger.Error("Network manager is nil, cannot create transport manager")
		return nil
	}
	
	hostManager := networkManager.GetHostManager()
	if hostManager == nil {
		logger.Error("Host manager is nil, cannot create transport manager")
		return nil
	}
	
	host := hostManager.GetHost()
	if host == nil {
		logger.Error("Libp2p host is nil, cannot create transport manager")
		return nil
	}
	
	return NewTransportManager(host, logger)
}