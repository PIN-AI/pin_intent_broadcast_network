#!/bin/bash

# 多节点测试环境初始化脚本
# 用于创建必要的目录结构和基础配置

set -e

# 加载统一配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh"

echo "=== PIN 多节点测试环境初始化 ==="

# 配置参数（使用统一配置）
BASE_HTTP_PORT=$BASE_HTTP_PORT
BASE_GRPC_PORT=$BASE_GRPC_PORT
BASE_P2P_PORT=$BASE_P2P_PORT
NODE_COUNT=$NODE_COUNT

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查必要的工具
check_prerequisites() {
    log_step "检查必要工具..."
    
    # 检查 Go 环境
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    # 检查项目是否已构建
    if [ ! -f "./bin/pin_intent_broadcast_network" ]; then
        log_warn "应用未构建，正在构建..."
        make build
        if [ $? -ne 0 ]; then
            log_error "构建失败"
            exit 1
        fi
    fi
    
    log_info "必要工具检查完成"
}

# 创建目录结构
create_directories() {
    log_step "创建目录结构..."
    
    # 创建测试数据目录
    for i in $(seq 1 $NODE_COUNT); do
        mkdir -p "test_data/node${i}/p2p"
        mkdir -p "test_data/node${i}/logs"
        log_info "创建节点${i}目录: test_data/node${i}/"
    done
    
    # 创建脚本目录
    mkdir -p scripts
    log_info "创建脚本目录: scripts/"
    
    # 创建配置目录（如果不存在）
    mkdir -p configs
    log_info "确保配置目录存在: configs/"
    
    log_info "目录结构创建完成"
}

# 检查端口可用性
check_port_availability() {
    local port=$1
    local service_name=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warn "端口 $port ($service_name) 已被占用"
        local pid=$(lsof -Pi :$port -sTCP:LISTEN -t)
        log_warn "占用进程 PID: $pid"
        return 1
    fi
    return 0
}

# 检查所有端口
check_all_ports() {
    log_step "检查端口可用性..."
    
    local port_conflicts=0
    
    for i in $(seq 1 $NODE_COUNT); do
        local http_port=$(get_node_http_port $i)
        local grpc_port=""
        local p2p_port=""
        
        # 获取节点的实际端口配置
        case $i in
            1) grpc_port=$NODE1_GRPC_PORT; p2p_port=$NODE1_P2P_PORT ;;
            2) grpc_port=$NODE2_GRPC_PORT; p2p_port=$NODE2_P2P_PORT ;;
            3) grpc_port=$NODE3_GRPC_PORT; p2p_port=$NODE3_P2P_PORT ;;
        esac
        
        if ! check_port_availability $http_port "Node${i} HTTP"; then
            port_conflicts=$((port_conflicts + 1))
        fi
        
        if ! check_port_availability $grpc_port "Node${i} gRPC"; then
            port_conflicts=$((port_conflicts + 1))
        fi
        
        if ! check_port_availability $p2p_port "Node${i} P2P"; then
            port_conflicts=$((port_conflicts + 1))
        fi
    done
    
    if [ $port_conflicts -gt 0 ]; then
        log_error "发现 $port_conflicts 个端口冲突"
        log_error "请停止占用端口的进程或使用 './scripts/cleanup_test.sh' 清理环境"
        exit 1
    fi
    
    log_info "所有端口检查通过"
}

# 生成节点配置文件
generate_node_config() {
    local node_id=$1
    local http_port=$(get_node_http_port $node_id)
    local grpc_port=""
    local p2p_port=""
    
    # 获取节点的实际端口配置
    case $node_id in
        1) grpc_port=$NODE1_GRPC_PORT; p2p_port=$NODE1_P2P_PORT ;;
        2) grpc_port=$NODE2_GRPC_PORT; p2p_port=$NODE2_P2P_PORT ;;
        3) grpc_port=$NODE3_GRPC_PORT; p2p_port=$NODE3_P2P_PORT ;;
    esac
    
    local config_file="configs/test_node${node_id}.yaml"
    
    log_info "生成节点${node_id}配置: $config_file"
    
    cat > $config_file << EOF
server:
  http:
    addr: 0.0.0.0:${http_port}
    timeout: 1s
  grpc:
    addr: 0.0.0.0:${grpc_port}
    timeout: 1s

data:
  database:
    driver: sqlite
    source: "test_data/node${node_id}/data.db"
  redis:
    addr: 127.0.0.1:6379
    read_timeout: 0.2s
    write_timeout: 0.2s

p2p:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/${p2p_port}"
  bootstrap_peers: []
  protocol_id: "/intent-broadcast/1.0.0"
  enable_mdns: true
  enable_dht: true
  data_dir: "test_data/node${node_id}/p2p"
  max_connections: 50
  enable_signing: true

transport:
  enable_gossipsub: true
  gossipsub_heartbeat_interval: 1s
  gossipsub_d: 6
  gossipsub_d_lo: 4
  gossipsub_d_hi: 12
  gossipsub_fanout_ttl: 60s
  enable_message_signing: true
  enable_strict_signature_verification: true
  message_id_cache_size: 1000
  message_ttl: 300s
  max_message_size: 1048576

  # Intent monitoring configuration
  intent_monitoring:
    # Subscription mode: "wildcard" | "explicit" | "all" | "disabled"
    # Default is "all" - listen to all intent broadcast topics
    subscription_mode: "all"

    # Wildcard patterns for topic matching (used in wildcard mode)
    wildcard_patterns:
      - "intent-broadcast.*"
      - "intent-matching.*"

    # Explicit topic list (used in explicit mode)
    explicit_topics:
      - "intent-broadcast.trade"
      - "intent-broadcast.swap"
      - "intent-broadcast.exchange"
      - "intent-broadcast.transfer"
      - "intent-broadcast.send"
      - "intent-broadcast.payment"
      - "intent-broadcast.lending"
      - "intent-broadcast.borrow"
      - "intent-broadcast.loan"
      - "intent-broadcast.investment"
      - "intent-broadcast.staking"
      - "intent-broadcast.yield"
      - "intent-broadcast.general"

    # Intent filter configuration
    filter:
      allowed_types: [] # Empty means allow all types
      blocked_types: []
      allowed_senders: []
      blocked_senders: []
      min_priority: 0
      max_priority: 10

    # Statistics configuration
    statistics:
      enabled: true
      retention_period: "24h"
      aggregation_interval: "1m"

    # Performance configuration
    performance:
      max_subscriptions: 100
      message_buffer_size: 1000
      batch_size: 10
  max_message_size: 1048576
EOF
}

# 生成所有配置文件
generate_all_configs() {
    log_step "生成节点配置文件..."
    
    for i in $(seq 1 $NODE_COUNT); do
        generate_node_config $i
    done
    
    log_info "所有配置文件生成完成"
}

# 创建测试状态文件
create_status_files() {
    log_step "创建测试状态文件..."
    
    # 创建测试状态目录
    mkdir -p test_data/status
    
    # 创建节点状态文件
    for i in $(seq 1 $NODE_COUNT); do
        echo "stopped" > "test_data/status/node${i}.status"
    done
    
    # 创建测试会话信息文件
    cat > test_data/status/session.info << EOF
# 测试会话信息
session_id=$(date +%Y%m%d_%H%M%S)
created_at=$(date '+%Y-%m-%d %H:%M:%S')
node_count=${NODE_COUNT}
base_http_port=${BASE_HTTP_PORT}
base_grpc_port=${BASE_GRPC_PORT}
base_p2p_port=${BASE_P2P_PORT}
EOF
    
    log_info "测试状态文件创建完成"
}

# 显示环境信息
show_environment_info() {
    log_step "测试环境信息"
    
    echo ""
    echo "节点配置信息:"
    echo "=============="
    for i in $(seq 1 $NODE_COUNT); do
        local http_port=$(get_node_http_port $i)
        local grpc_port=""
        local p2p_port=""
        
        # 获取节点的实际端口配置
        case $i in
            1) grpc_port=$NODE1_GRPC_PORT; p2p_port=$NODE1_P2P_PORT ;;
            2) grpc_port=$NODE2_GRPC_PORT; p2p_port=$NODE2_P2P_PORT ;;
            3) grpc_port=$NODE3_GRPC_PORT; p2p_port=$NODE3_P2P_PORT ;;
        esac
        
        echo "节点${i}:"
        echo "  HTTP端口: $http_port"
        echo "  gRPC端口: $grpc_port"
        echo "  P2P端口:  $p2p_port"
        echo "  配置文件: configs/test_node${i}.yaml"
        echo "  数据目录: test_data/node${i}/"
        echo ""
    done
    
    echo "启动脚本:"
    echo "=========="
    echo "  节点1 (发布者): ./scripts/start_node1.sh"
    echo "  节点2 (监控者): ./scripts/start_node2.sh"
    echo "  节点3 (监控者): ./scripts/start_node3.sh"
    echo ""
    
    echo "工具脚本:"
    echo "=========="
    echo "  Intent发布器: ./scripts/intent_publisher.sh <node_port>"
    echo "  Intent监控器: ./scripts/intent_monitor.sh <node_port>"
    echo "  网络状态检查: ./scripts/network_status.sh <node_port>"
    echo "  环境清理: ./scripts/cleanup_test.sh"
    echo ""
}

# 主函数
main() {
    echo "开始初始化多节点测试环境..."
    echo ""
    
    check_prerequisites
    create_directories
    check_all_ports
    generate_all_configs
    create_status_files
    show_environment_info
    
    log_info "测试环境初始化完成！"
    log_info "现在可以使用启动脚本在不同终端中启动节点"
    
    echo ""
    echo -e "${GREEN}下一步操作:${NC}"
    echo "1. 在终端1中运行: ./scripts/start_node1.sh"
    echo "2. 在终端2中运行: ./scripts/start_node2.sh"
    echo "3. 在终端3中运行: ./scripts/start_node3.sh"
    echo ""
}

# 执行主函数
main "$@"