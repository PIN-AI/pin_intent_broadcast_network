package p2p

import (
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	// ErrHostNotRunning 主机未运行错误
	ErrHostNotRunning = errors.New("p2p host is not running")

	// ErrHostAlreadyRunning 主机已运行错误
	ErrHostAlreadyRunning = errors.New("p2p host is already running")

	// ErrInvalidConfig 配置无效错误
	ErrInvalidConfig = errors.New("invalid p2p configuration")

	// ErrPeerNotFound 节点未找到错误
	ErrPeerNotFound = errors.New("peer not found")

	// ErrConnectionFailed 连接失败错误
	ErrConnectionFailed = errors.New("connection failed")

	// ErrDiscoveryFailed 发现失败错误
	ErrDiscoveryFailed = errors.New("discovery failed")

	// ErrInvalidPeerID 无效节点ID错误
	ErrInvalidPeerID = errors.New("invalid peer ID")

	// ErrInvalidMultiaddr 无效多地址错误
	ErrInvalidMultiaddr = errors.New("invalid multiaddr")

	// ErrConnectionLimitReached 连接数限制已达到错误
	ErrConnectionLimitReached = errors.New("connection limit reached")

	// ErrDHTNotEnabled DHT未启用错误
	ErrDHTNotEnabled = errors.New("DHT is not enabled")

	// ErrMDNSNotEnabled mDNS未启用错误
	ErrMDNSNotEnabled = errors.New("mDNS is not enabled")
	
	// ErrContextCancelled 上下文取消错误
	ErrContextCancelled = errors.New("context cancelled")
)

// ConfigError 配置错误类型
type ConfigError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error in field '%s' with value '%v': %s", e.Field, e.Value, e.Message)
}

// NewConfigError 创建配置错误
func NewConfigError(field string, value interface{}, message string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// ConnectionError 连接错误类型
type ConnectionError struct {
	PeerID peer.ID
	Cause  error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error with peer %s: %v", e.PeerID.String(), e.Cause)
}

func (e *ConnectionError) Unwrap() error {
	return e.Cause
}

// NewConnectionError 创建连接错误
func NewConnectionError(peerID peer.ID, cause error) *ConnectionError {
	return &ConnectionError{
		PeerID: peerID,
		Cause:  cause,
	}
}

// DiscoveryError 发现错误类型
type DiscoveryError struct {
	Namespace string
	Cause     error
}

func (e *DiscoveryError) Error() string {
	return fmt.Sprintf("discovery error in namespace '%s': %v", e.Namespace, e.Cause)
}

func (e *DiscoveryError) Unwrap() error {
	return e.Cause
}

// NewDiscoveryError 创建发现错误
func NewDiscoveryError(namespace string, cause error) *DiscoveryError {
	return &DiscoveryError{
		Namespace: namespace,
		Cause:     cause,
	}
}