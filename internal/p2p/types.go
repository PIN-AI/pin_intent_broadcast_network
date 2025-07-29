package p2p

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// HostConfig P2P主机配置
type HostConfig struct {
	// ListenAddresses 监听地址列表
	ListenAddresses []string `json:"listen_addresses" yaml:"listen_addresses"`
	// BootstrapPeers 引导节点列表
	BootstrapPeers []string `json:"bootstrap_peers" yaml:"bootstrap_peers"`
	// ProtocolID 协议标识符
	ProtocolID string `json:"protocol_id" yaml:"protocol_id"`
	// EnableMDNS 是否启用mDNS发现
	EnableMDNS bool `json:"enable_mdns" yaml:"enable_mdns"`
	// EnableDHT 是否启用DHT发现
	EnableDHT bool `json:"enable_dht" yaml:"enable_dht"`
	// DataDir 数据目录
	DataDir string `json:"data_dir" yaml:"data_dir"`
	// MaxConnections 最大连接数
	MaxConnections int `json:"max_connections" yaml:"max_connections"`
	// EnableSigning 是否启用消息签名
	EnableSigning bool `json:"enable_signing" yaml:"enable_signing"`
	// PrivateKeyPath 私钥文件路径（可选）
	PrivateKeyPath string `json:"private_key_path,omitempty" yaml:"private_key_path,omitempty"`
}

// ConnectionStatus 连接状态枚举
type ConnectionStatus int

const (
	// ConnectionStatusDisconnected 未连接
	ConnectionStatusDisconnected ConnectionStatus = iota
	// ConnectionStatusConnecting 连接中
	ConnectionStatusConnecting
	// ConnectionStatusConnected 已连接
	ConnectionStatusConnected
	// ConnectionStatusFailed 连接失败
	ConnectionStatusFailed
)

// String 返回连接状态的字符串表示
func (cs ConnectionStatus) String() string {
	switch cs {
	case ConnectionStatusDisconnected:
		return "disconnected"
	case ConnectionStatusConnecting:
		return "connecting"
	case ConnectionStatusConnected:
		return "connected"
	case ConnectionStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	// PeerID 节点ID
	PeerID peer.ID `json:"peer_id"`
	// Status 连接状态
	Status ConnectionStatus `json:"status"`
	// ConnectedAt 连接建立时间
	ConnectedAt time.Time `json:"connected_at"`
	// LastSeen 最后活跃时间
	LastSeen time.Time `json:"last_seen"`
	// RemoteAddr 远程地址
	RemoteAddr multiaddr.Multiaddr `json:"remote_addr"`
	// Direction 连接方向（inbound/outbound）
	Direction string `json:"direction"`
	// Latency 连接延迟
	Latency time.Duration `json:"latency"`
}

// NetworkTopology 网络拓扑信息
type NetworkTopology struct {
	// LocalPeerID 本地节点ID
	LocalPeerID peer.ID `json:"local_peer_id"`
	// ConnectedPeers 已连接的节点列表
	ConnectedPeers []peer.ID `json:"connected_peers"`
	// ConnectionMap 连接信息映射
	ConnectionMap map[peer.ID]ConnectionInfo `json:"connection_map"`
	// CreatedAt 拓扑信息创建时间
	CreatedAt time.Time `json:"created_at"`
	// TotalConnections 总连接数
	TotalConnections int `json:"total_connections"`
}

// NetworkStatus 网络状态信息
type NetworkStatus struct {
	// IsRunning 网络是否运行中
	IsRunning bool `json:"is_running"`
	// LocalPeerID 本地节点ID
	LocalPeerID peer.ID `json:"local_peer_id"`
	// ListenAddresses 监听地址列表
	ListenAddresses []multiaddr.Multiaddr `json:"listen_addresses"`
	// ConnectedPeersCount 已连接节点数量
	ConnectedPeersCount int `json:"connected_peers_count"`
	// DiscoveredPeersCount 已发现节点数量
	DiscoveredPeersCount int `json:"discovered_peers_count"`
	// StartTime 网络启动时间
	StartTime time.Time `json:"start_time"`
	// Uptime 运行时长
	Uptime time.Duration `json:"uptime"`
	// Config 配置信息
	Config *HostConfig `json:"config"`
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	// ConnectedPeers 连接的节点数
	ConnectedPeers int `json:"connected_peers"`
	// ConnectionErrors 连接错误数
	ConnectionErrors int64 `json:"connection_errors"`
	// DiscoveredPeers 发现的节点数
	DiscoveredPeers int64 `json:"discovered_peers"`
	// NetworkLatency 网络延迟
	NetworkLatency time.Duration `json:"network_latency"`
	// BytesSent 发送字节数
	BytesSent int64 `json:"bytes_sent"`
	// BytesReceived 接收字节数
	BytesReceived int64 `json:"bytes_received"`
	// MessagesSent 发送消息数
	MessagesSent int64 `json:"messages_sent"`
	// MessagesReceived 接收消息数
	MessagesReceived int64 `json:"messages_received"`
	// LastUpdated 最后更新时间
	LastUpdated time.Time `json:"last_updated"`
}

// PeerInfo 节点信息
type PeerInfo struct {
	// ID 节点ID
	ID peer.ID `json:"id"`
	// Addresses 节点地址列表
	Addresses []multiaddr.Multiaddr `json:"addresses"`
	// Protocols 支持的协议列表
	Protocols []string `json:"protocols"`
	// AgentVersion 客户端版本
	AgentVersion string `json:"agent_version,omitempty"`
	// ProtocolVersion 协议版本
	ProtocolVersion string `json:"protocol_version,omitempty"`
	// FirstSeen 首次发现时间
	FirstSeen time.Time `json:"first_seen"`
	// LastSeen 最后活跃时间
	LastSeen time.Time `json:"last_seen"`
}

// DiscoveryEvent 发现事件
type DiscoveryEvent struct {
	// Type 事件类型（discovered/lost）
	Type string `json:"type"`
	// PeerInfo 节点信息
	PeerInfo peer.AddrInfo `json:"peer_info"`
	// Timestamp 事件时间戳
	Timestamp time.Time `json:"timestamp"`
	// Source 发现来源（mdns/dht/bootstrap）
	Source string `json:"source"`
}

// ConnectionEvent 连接事件
type ConnectionEvent struct {
	// Type 事件类型（connected/disconnected/failed）
	Type string `json:"type"`
	// PeerID 节点ID
	PeerID peer.ID `json:"peer_id"`
	// RemoteAddr 远程地址
	RemoteAddr multiaddr.Multiaddr `json:"remote_addr"`
	// Direction 连接方向
	Direction string `json:"direction"`
	// Timestamp 事件时间戳
	Timestamp time.Time `json:"timestamp"`
	// Error 错误信息（如果有）
	Error string `json:"error,omitempty"`
}