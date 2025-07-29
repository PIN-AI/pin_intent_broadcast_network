package p2p

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// ValidateConfig 验证P2P配置
func ValidateConfig(config *HostConfig) error {
	if config == nil {
		return NewConfigError("config", nil, "configuration cannot be nil")
	}

	// 验证监听地址
	if len(config.ListenAddresses) == 0 {
		return NewConfigError("listen_addresses", config.ListenAddresses, "at least one listen address is required")
	}

	for i, addrStr := range config.ListenAddresses {
		if _, err := multiaddr.NewMultiaddr(addrStr); err != nil {
			return NewConfigError(fmt.Sprintf("listen_addresses[%d]", i), addrStr, fmt.Sprintf("invalid multiaddr: %v", err))
		}
	}

	// 验证引导节点地址
	for i, peerStr := range config.BootstrapPeers {
		if _, err := peer.AddrInfoFromString(peerStr); err != nil {
			return NewConfigError(fmt.Sprintf("bootstrap_peers[%d]", i), peerStr, fmt.Sprintf("invalid peer address: %v", err))
		}
	}

	// 验证协议ID
	if config.ProtocolID == "" {
		return NewConfigError("protocol_id", config.ProtocolID, "protocol ID cannot be empty")
	}

	// 验证数据目录
	if config.DataDir == "" {
		return NewConfigError("data_dir", config.DataDir, "data directory cannot be empty")
	}

	// 验证最大连接数
	if config.MaxConnections <= 0 {
		return NewConfigError("max_connections", config.MaxConnections, "max connections must be positive")
	}

	return nil
}

// GenerateOrLoadPrivateKey 生成或加载私钥
func GenerateOrLoadPrivateKey(dataDir string) (crypto.PrivKey, error) {
	keyPath := filepath.Join(dataDir, "private.key")

	// 尝试从文件加载私钥
	if keyData, err := os.ReadFile(keyPath); err == nil {
		return crypto.UnmarshalPrivateKey(keyData)
	}

	// 生成新的私钥
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// 保存私钥到文件
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	keyData, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
		return nil, fmt.Errorf("failed to save private key: %w", err)
	}

	return priv, nil
}

// ParseBootstrapPeers 解析引导节点列表
func ParseBootstrapPeers(bootstrapPeers []string) ([]peer.AddrInfo, error) {
	var peers []peer.AddrInfo

	for i, peerStr := range bootstrapPeers {
		peerInfo, err := peer.AddrInfoFromString(peerStr)
		if err != nil {
			return nil, fmt.Errorf("invalid bootstrap peer %d (%s): %w", i, peerStr, err)
		}
		peers = append(peers, *peerInfo)
	}

	return peers, nil
}

// ParseListenAddresses 解析监听地址列表
func ParseListenAddresses(listenAddrs []string) ([]multiaddr.Multiaddr, error) {
	var addrs []multiaddr.Multiaddr

	for i, addrStr := range listenAddrs {
		addr, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid listen address %d (%s): %w", i, addrStr, err)
		}
		addrs = append(addrs, addr)
	}

	return addrs, nil
}

// FormatPeerID 格式化节点ID为短格式
func FormatPeerID(peerID peer.ID) string {
	str := peerID.String()
	if len(str) > 12 {
		return str[:6] + "..." + str[len(str)-6:]
	}
	return str
}

// FormatMultiaddr 格式化多地址为短格式
func FormatMultiaddr(addr multiaddr.Multiaddr) string {
	str := addr.String()
	if len(str) > 50 {
		return str[:25] + "..." + str[len(str)-25:]
	}
	return str
}

// EnsureDataDir 确保数据目录存在
func EnsureDataDir(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory %s: %w", dataDir, err)
	}
	return nil
}