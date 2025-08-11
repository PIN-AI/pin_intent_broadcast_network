#!/bin/bash

# PIN 自动化测试环境初始化脚本
# 创建配置文件、目录结构和必要的初始化

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
    echo -e "${AUTO_COLOR_MAGENTA}    PIN 自动化测试环境初始化    ${AUTO_COLOR_NC}"
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
    )
    
    for dir in "${directories[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log_info "创建目录: $dir"
        else
            log_warn "目录已存在: $dir"
        fi
    done
    
    # 创建日志目录的符号链接便于访问
    for i in {1..4}; do
        local node_data_dir=$(eval echo \$AUTO_NODE${i}_DATA_DIR)
        local log_link="$AUTO_LOGS_DIR/node${i}.log"
        local actual_log=$(eval echo \$AUTO_NODE${i}_LOG_FILE)
        
        if [ ! -L "$log_link" ] && [ ! -f "$log_link" ]; then
            ln -sf "../../$(basename $(dirname $actual_log))/$(basename $actual_log)" "$log_link"
            log_info "创建日志链接: $log_link -> $actual_log"
        fi
    done
}

# 生成节点配置文件
generate_node_configs() {
    log_step "生成节点配置文件..."
    
    # 节点1配置 (Intent发布者)
    generate_node1_config
    
    # 节点2配置 (Service Agent 1)
    generate_node2_config
    
    # 节点3配置 (Service Agent 2)
    generate_node3_config
    
    # 节点4配置 (Block Builder)
    generate_node4_config
}

# 生成节点1配置 (Intent发布者)
generate_node1_config() {
    log_info "生成节点1配置 (Intent发布者)..."
    
    cat > "$AUTO_NODE1_CONFIG_FILE" << EOF
server:
  http:
    addr: 0.0.0.0:$AUTO_NODE1_HTTP_PORT
    timeout: 1s
  grpc:
    addr: 0.0.0.0:$AUTO_NODE1_GRPC_PORT
    timeout: 1s

data:
  database:
    driver: memory
    source: ""
  redis:
    addr: 127.0.0.1:6379
    password: ""
    db: 0
    dial_timeout: 1s
    read_timeout: 0.2s
    write_timeout: 0.2s

p2p:
  listen_addresses:
    - /ip4/0.0.0.0/tcp/$AUTO_NODE1_P2P_PORT
  protocol_id: "$AUTO_PROTOCOL_ID"
  enable_mdns: $AUTO_ENABLE_MDNS
  enable_dht: $AUTO_ENABLE_DHT
  data_dir: "./test_data/automation/node1/p2p"
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
    
    log_info "节点1配置已生成: $AUTO_NODE1_CONFIG_FILE"
}

# 生成节点2配置 (Service Agent 1)
generate_node2_config() {
    log_info "生成节点2配置 (Service Agent 1)..."
    
    cat > "$AUTO_NODE2_CONFIG_FILE" << EOF
server:
  http:
    addr: 0.0.0.0:$AUTO_NODE2_HTTP_PORT
    timeout: 1s
  grpc:
    addr: 0.0.0.0:$AUTO_NODE2_GRPC_PORT
    timeout: 1s

data:
  database:
    driver: memory
    source: ""
  redis:
    addr: 127.0.0.1:6379
    password: ""
    db: 1
    dial_timeout: 1s
    read_timeout: 0.2s
    write_timeout: 0.2s

p2p:
  listen_addresses:
    - /ip4/0.0.0.0/tcp/$AUTO_NODE2_P2P_PORT
  protocol_id: "$AUTO_PROTOCOL_ID"
  enable_mdns: $AUTO_ENABLE_MDNS
  enable_dht: $AUTO_ENABLE_DHT
  data_dir: "./test_data/automation/node2/p2p"
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
      allowed_types: ["trade", "swap", "exchange"]
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
    
    log_info "节点2配置已生成: $AUTO_NODE2_CONFIG_FILE"
}

# 生成节点3配置 (Service Agent 2)
generate_node3_config() {
    log_info "生成节点3配置 (Service Agent 2)..."
    
    cat > "$AUTO_NODE3_CONFIG_FILE" << EOF
server:
  http:
    addr: 0.0.0.0:$AUTO_NODE3_HTTP_PORT
    timeout: 1s
  grpc:
    addr: 0.0.0.0:$AUTO_NODE3_GRPC_PORT
    timeout: 1s

data:
  database:
    driver: memory
    source: ""
  redis:
    addr: 127.0.0.1:6379
    password: ""
    db: 2
    dial_timeout: 1s
    read_timeout: 0.2s
    write_timeout: 0.2s

p2p:
  listen_addresses:
    - /ip4/0.0.0.0/tcp/$AUTO_NODE3_P2P_PORT
  protocol_id: "$AUTO_PROTOCOL_ID"
  enable_mdns: $AUTO_ENABLE_MDNS
  enable_dht: $AUTO_ENABLE_DHT
  data_dir: "./test_data/automation/node3/p2p"
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
      allowed_types: ["data_access", "computation"]
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
    
    log_info "节点3配置已生成: $AUTO_NODE3_CONFIG_FILE"
}

# 生成节点4配置 (Block Builder)
generate_node4_config() {
    log_info "生成节点4配置 (Block Builder)..."
    
    cat > "$AUTO_NODE4_CONFIG_FILE" << EOF
server:
  http:
    addr: 0.0.0.0:$AUTO_NODE4_HTTP_PORT
    timeout: 1s
  grpc:
    addr: 0.0.0.0:$AUTO_NODE4_GRPC_PORT
    timeout: 1s

data:
  database:
    driver: memory
    source: ""
  redis:
    addr: 127.0.0.1:6379
    password: ""
    db: 3
    dial_timeout: 1s
    read_timeout: 0.2s
    write_timeout: 0.2s

p2p:
  listen_addresses:
    - /ip4/0.0.0.0/tcp/$AUTO_NODE4_P2P_PORT
  protocol_id: "$AUTO_PROTOCOL_ID"
  enable_mdns: $AUTO_ENABLE_MDNS
  enable_dht: $AUTO_ENABLE_DHT
  data_dir: "./test_data/automation/node4/p2p"
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
    
    log_info "节点4配置已生成: $AUTO_NODE4_CONFIG_FILE"
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
    log_step "自动化测试环境设置完成"
    echo ""
    echo -e "${AUTO_COLOR_GREEN}✓ 目录结构已创建${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_GREEN}✓ 配置文件已生成${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_GREEN}✓ 环境已验证${AUTO_COLOR_NC}"
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}使用方法:${AUTO_COLOR_NC}"
    echo "1. 启动自动化测试:"
    echo "   ./scripts/automation/start_automation_test.sh"
    echo ""
    echo "2. 在单独终端启动各节点:"
    echo "   ./scripts/automation/start_node.sh 1  # Intent发布者"
    echo "   ./scripts/automation/start_node.sh 2  # Service Agent 1"
    echo "   ./scripts/automation/start_node.sh 3  # Service Agent 2" 
    echo "   ./scripts/automation/start_node.sh 4  # Block Builder"
    echo ""
    echo "3. 监控测试状态:"
    echo "   ./scripts/automation/monitor_automation.sh"
    echo ""
    echo "4. 清理环境:"
    echo "   ./scripts/automation/cleanup_automation.sh"
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