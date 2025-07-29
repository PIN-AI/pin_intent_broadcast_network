package p2p

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// HostManager 管理libp2p主机的生命周期和基本操作
type HostManager interface {
	// Start 启动P2P主机
	Start(ctx context.Context, config *HostConfig) error
	// Stop 停止P2P主机
	Stop() error
	// GetHost 获取libp2p主机实例
	GetHost() host.Host
	// GetPeerID 获取本节点的Peer ID
	GetPeerID() peer.ID
	// IsRunning 检查主机是否正在运行
	IsRunning() bool
	// GetListenAddresses 获取监听地址列表
	GetListenAddresses() []multiaddr.Multiaddr
}

// DiscoveryManager 管理节点发现机制
type DiscoveryManager interface {
	// Start 启动节点发现服务
	Start(ctx context.Context) error
	// Stop 停止节点发现服务
	Stop() error
	// DiscoverPeers 发现指定命名空间的节点
	DiscoverPeers(ctx context.Context, namespace string) (<-chan peer.AddrInfo, error)
	// Advertise 在指定命名空间广告自己
	Advertise(ctx context.Context, namespace string) error
	// GetConnectedPeers 获取已连接的节点列表
	GetConnectedPeers() []peer.ID
	// ConnectToBootstrapPeers 连接到引导节点
	ConnectToBootstrapPeers(ctx context.Context) error
}

// ConnectionManager 管理P2P连接
type ConnectionManager interface {
	// ConnectToPeer 连接到指定节点
	ConnectToPeer(ctx context.Context, peerInfo peer.AddrInfo) error
	// DisconnectFromPeer 断开与指定节点的连接
	DisconnectFromPeer(peerID peer.ID) error
	// GetConnectionStatus 获取与指定节点的连接状态
	GetConnectionStatus(peerID peer.ID) ConnectionStatus
	// GetNetworkTopology 获取网络拓扑信息
	GetNetworkTopology() *NetworkTopology
	// SetConnectionLimits 设置连接数限制
	SetConnectionLimits(low, high int) error
	// GetConnectionCount 获取当前连接数
	GetConnectionCount() int
}

// NetworkManager 网络层统一管理接口
type NetworkManager interface {
	// Start 启动网络管理器
	Start(ctx context.Context, config *HostConfig) error
	// Stop 停止网络管理器
	Stop() error
	// GetHostManager 获取主机管理器
	GetHostManager() HostManager
	// GetDiscoveryManager 获取发现管理器
	GetDiscoveryManager() DiscoveryManager
	// GetConnectionManager 获取连接管理器
	GetConnectionManager() ConnectionManager
	// GetNetworkStatus 获取网络状态
	GetNetworkStatus() *NetworkStatus
}