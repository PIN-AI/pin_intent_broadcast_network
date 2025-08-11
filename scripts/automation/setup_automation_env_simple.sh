#!/bin/bash

# PIN 自动化测试环境初始化脚本（简化版本）
# 创建基本的节点配置文件，暂时不启用自动化功能

set -e

# 加载自动化配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/automation_config.sh"

# 日志函数
log_info() {
    echo -e "${AUTO_COLOR_GREEN}[INFO]${AUTO_COLOR_NC} $1"
}

log_warn() {
    echo -e "${AUTO_COLOR_YELLOW}[WARN]${AUTO_COLOR_NC} $1"
}

log_error() {
    echo -e "${AUTO_COLOR_RED}[ERROR]${AUTO_COLOR_NC} $1"
}

log_step() {
    echo -e "${AUTO_COLOR_BLUE}[STEP]${AUTO_COLOR_NC} $1"
}

# 显示头部信息
show_header() {
    clear
    echo -e "${AUTO_COLOR_MAGENTA}================================================${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}    PIN 自动化测试环境初始化（简化版）    ${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}================================================${AUTO_COLOR_NC}"
    echo ""
    show_auto_config_summary
}

# 创建目录结构
create_directories() {
    log_step "创建目录结构..."
    
    local directories=(
        "$AUTO_TEST_DATA_DIR"
        "$AUTO_STATUS_DIR"
        "$AUTO_LOGS_DIR"
        "$AUTO_NODE1_DATA_DIR"
        "$AUTO_NODE2_DATA_DIR"
        "$AUTO_NODE3_DATA_DIR"
        "$AUTO_NODE4_DATA_DIR"
        "$AUTO_NODE1_DATA_DIR/p2p"
        "$AUTO_NODE2_DATA_DIR/p2p"
        "$AUTO_NODE3_DATA_DIR/p2p"
        "$AUTO_NODE4_DATA_DIR/p2p"
    )
    
    for dir in "${directories[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log_info "创建目录: $dir"
        else
            log_warn "目录已存在: $dir"
        fi
    done
}

# 生成节点配置文件（简化版）
generate_node_configs() {
    log_step "生成节点配置文件（简化版）..."
    
    # 生成四个基本节点配置
    for i in {1..4}; do
        generate_basic_node_config $i
    done
}

# 生成基本节点配置
generate_basic_node_config() {
    local node_id=$1
    local http_port=$(eval echo \$AUTO_NODE${node_id}_HTTP_PORT)
    local grpc_port=$(eval echo \$AUTO_NODE${node_id}_GRPC_PORT)
    local p2p_port=$(eval echo \$AUTO_NODE${node_id}_P2P_PORT)
    local config_file=$(eval echo \$AUTO_NODE${node_id}_CONFIG_FILE)
    local log_file=$(eval echo \$AUTO_NODE${node_id}_LOG_FILE)
    local data_dir=$(eval echo \$AUTO_NODE${node_id}_DATA_DIR)
    local node_name=$(eval echo \$AUTO_NODE${node_id}_NAME)
    
    log_info "生成节点${node_id}配置 ($node_name)..."
    
    # 构建bootstrap_peers数组
    local bootstrap_peers=""
    if [ $node_id -gt 1 ]; then
        bootstrap_peers="  bootstrap_peers:"
        for j in $(seq 1 $((node_id-1))); do
            local prev_p2p_port=$(eval echo \$AUTO_NODE${j}_P2P_PORT)
            bootstrap_peers="$bootstrap_peers
    - /ip4/127.0.0.1/tcp/$prev_p2p_port"
        done
    else
        bootstrap_peers="  bootstrap_peers: []"
    fi
    
    cat > "$config_file" << EOF
server:
  http:
    addr: 0.0.0.0:$http_port
    timeout: 1s
  grpc:
    addr: 0.0.0.0:$grpc_port
    timeout: 1s

data:
  database:
    driver: memory
    source: ""
  redis:
    addr: 127.0.0.1:6379
    password: ""
    db: $((node_id-1))
    dial_timeout: 1s
    read_timeout: 0.2s
    write_timeout: 0.2s

p2p:
  listen_addresses:
    - /ip4/0.0.0.0/tcp/$p2p_port
$bootstrap_peers
  protocol_id: "$AUTO_PROTOCOL_ID"
  enable_mdns: $AUTO_ENABLE_MDNS
  enable_dht: $AUTO_ENABLE_DHT
  data_dir: "./$data_dir/p2p"
  max_connections: $AUTO_MAX_CONNECTIONS
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
    subscription_mode: "all"
    wildcard_patterns:
      - "automation-test.*"
    explicit_topics:
      - "$AUTO_INTENT_TOPIC"
      - "$AUTO_BIDDING_TOPIC"
      - "$AUTO_MATCHING_TOPIC"
    filter:
      allowed_types: []
      blocked_types: []
      allowed_senders: []
      blocked_senders: []
      min_priority: 0
      max_priority: 10
    statistics:
      enabled: true
      retention_period: "3600s"
      aggregation_interval: "60s"
    performance:
      max_subscriptions: 100
      message_buffer_size: 1000
      batch_size: 10
EOF
    
    log_info "节点${node_id}配置已生成: $config_file"
}

# 检查前置条件
check_prerequisites() {
    log_step "检查前置条件..."
    
    # 检查应用程序是否存在
    if [ ! -f "$AUTO_APP_BINARY" ]; then
        log_error "应用程序不存在: $AUTO_APP_BINARY"
        log_error "请先运行 'make build' 构建应用"
        exit 1
    fi
    
    # 检查配置验证
    validate_auto_config
    if [ $? -ne 0 ]; then
        log_error "配置验证失败"
        exit 1
    fi
    
    log_info "前置条件检查通过"
}

# 清理旧环境
cleanup_old_env() {
    log_step "清理旧环境..."
    
    # 停止可能运行的自动化测试进程
    for i in {1..4}; do
        local pid_file=$(eval echo \$AUTO_NODE${i}_PID_FILE)
        if [ -f "$pid_file" ]; then
            local pid=$(cat "$pid_file")
            if kill -0 "$pid" 2>/dev/null; then
                log_warn "停止运行中的节点$i (PID: $pid)"
                kill -TERM "$pid" 2>/dev/null || true
                sleep 1
                if kill -0 "$pid" 2>/dev/null; then
                    kill -KILL "$pid" 2>/dev/null || true
                fi
            fi
            rm -f "$pid_file"
        fi
    done
    
    # 清理状态文件
    rm -f "$AUTO_STATUS_DIR"/*.status 2>/dev/null || true
    
    log_info "环境清理完成"
}

# 验证环境
verify_environment() {
    log_step "验证环境..."
    
    # 检查目录是否存在
    local required_dirs=("$AUTO_TEST_DATA_DIR" "$AUTO_STATUS_DIR")
    for dir in "${required_dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            log_error "必需目录不存在: $dir"
            exit 1
        fi
    done
    
    # 检查配置文件是否存在
    for i in {1..4}; do
        local config_file=$(eval echo \$AUTO_NODE${i}_CONFIG_FILE)
        if [ ! -f "$config_file" ]; then
            log_error "配置文件不存在: $config_file"
            exit 1
        fi
    done
    
    log_info "环境验证通过"
}

# 显示使用说明
show_usage() {
    echo ""
    log_step "自动化测试环境设置完成（简化版）"
    echo ""
    echo -e "${AUTO_COLOR_GREEN}✓ 目录结构已创建${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_GREEN}✓ 基本配置文件已生成${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_GREEN}✓ 环境已验证${AUTO_COLOR_NC}"
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}使用方法:${AUTO_COLOR_NC}"
    echo "1. 手动启动各节点（在不同终端）:"
    echo "   ./scripts/automation/start_node.sh 1  # Intent发布者"
    echo "   ./scripts/automation/start_node.sh 2  # Service Agent 1"
    echo "   ./scripts/automation/start_node.sh 3  # Service Agent 2" 
    echo "   ./scripts/automation/start_node.sh 4  # Block Builder"
    echo ""
    echo "2. 手动测试Intent API:"
    echo "   # 创建Intent"
    echo "   curl -X POST http://localhost:8100/pinai_intent/intent/create \\"
    echo "        -H 'Content-Type: application/json' \\"
    echo "        -d '{\"type\":\"trade\",\"payload\":\"dGVzdA==\",\"sender_id\":\"test-user\"}'"
    echo ""
    echo "   # 查看节点状态"
    echo "   curl http://localhost:8100/health"
    echo "   curl http://localhost:8101/health"
    echo "   curl http://localhost:8102/health"
    echo "   curl http://localhost:8103/health"
    echo ""
    echo "3. 监控P2P连接:"
    echo "   # 查看日志"
    echo "   tail -f test_data/automation/node*/output.log"
    echo ""
    echo "4. 清理环境:"
    echo "   ./scripts/automation/cleanup_automation.sh"
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}注意:${AUTO_COLOR_NC}"
    echo "- 此为简化版配置，自动化功能暂时禁用"
    echo "- 各节点可正常启动并支持手动API测试"
    echo "- P2P网络连接和Intent广播功能可用"
    echo ""
}

# 主函数
main() {
    show_header
    check_prerequisites
    cleanup_old_env
    create_directories
    generate_node_configs
    verify_environment
    show_usage
}

# 执行主函数
main "$@"