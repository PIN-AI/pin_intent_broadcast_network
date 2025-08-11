#!/bin/bash

# PIN 自动化测试系统配置文件
# 配置四个节点：1个发布者、2个Service Agent、1个Block Builder

# 基础配置
export AUTOMATION_NODE_COUNT=4
export AUTOMATION_TEST_ENV_NAME="PIN_AUTOMATION_TEST"

# 端口分配 (避免与现有测试冲突)
export AUTO_BASE_HTTP_PORT=8100
export AUTO_BASE_GRPC_PORT=9100
export AUTO_BASE_P2P_PORT=9200

# 节点1配置 (Intent发布者节点)
export AUTO_NODE1_HTTP_PORT=8100
export AUTO_NODE1_GRPC_PORT=9100
export AUTO_NODE1_P2P_PORT=9200
export AUTO_NODE1_CONFIG_FILE="configs/automation_node1.yaml"
export AUTO_NODE1_DATA_DIR="test_data/automation/node1"
export AUTO_NODE1_LOG_FILE="test_data/automation/node1/output.log"
export AUTO_NODE1_PID_FILE="test_data/automation/node1/pid"
export AUTO_NODE1_STATUS_FILE="test_data/automation/status/node1.status"
export AUTO_NODE1_TYPE="PUBLISHER"
export AUTO_NODE1_NAME="Intent发布者"

# 节点2配置 (Service Agent 1 - 交易代理)
export AUTO_NODE2_HTTP_PORT=8101
export AUTO_NODE2_GRPC_PORT=9101
export AUTO_NODE2_P2P_PORT=9201
export AUTO_NODE2_CONFIG_FILE="configs/automation_node2.yaml"
export AUTO_NODE2_DATA_DIR="test_data/automation/node2"
export AUTO_NODE2_LOG_FILE="test_data/automation/node2/output.log"
export AUTO_NODE2_PID_FILE="test_data/automation/node2/pid"
export AUTO_NODE2_STATUS_FILE="test_data/automation/status/node2.status"
export AUTO_NODE2_TYPE="SERVICE_AGENT"
export AUTO_NODE2_NAME="Service Agent 1 (交易)"
export AUTO_NODE2_AGENT_ID="trading-agent-auto-001"

# 节点3配置 (Service Agent 2 - 数据代理)
export AUTO_NODE3_HTTP_PORT=8102
export AUTO_NODE3_GRPC_PORT=9102
export AUTO_NODE3_P2P_PORT=9202
export AUTO_NODE3_CONFIG_FILE="configs/automation_node3.yaml"
export AUTO_NODE3_DATA_DIR="test_data/automation/node3"
export AUTO_NODE3_LOG_FILE="test_data/automation/node3/output.log"
export AUTO_NODE3_PID_FILE="test_data/automation/node3/pid"
export AUTO_NODE3_STATUS_FILE="test_data/automation/status/node3.status"
export AUTO_NODE3_TYPE="SERVICE_AGENT"
export AUTO_NODE3_NAME="Service Agent 2 (数据)"
export AUTO_NODE3_AGENT_ID="data-agent-auto-002"

# 节点4配置 (Block Builder - 匹配节点)
export AUTO_NODE4_HTTP_PORT=8103
export AUTO_NODE4_GRPC_PORT=9103
export AUTO_NODE4_P2P_PORT=9203
export AUTO_NODE4_CONFIG_FILE="configs/automation_node4.yaml"
export AUTO_NODE4_DATA_DIR="test_data/automation/node4"
export AUTO_NODE4_LOG_FILE="test_data/automation/node4/output.log"
export AUTO_NODE4_PID_FILE="test_data/automation/node4/pid"
export AUTO_NODE4_STATUS_FILE="test_data/automation/status/node4.status"
export AUTO_NODE4_TYPE="BLOCK_BUILDER"
export AUTO_NODE4_NAME="Block Builder"
export AUTO_NODE4_BUILDER_ID="auto-builder-001"

# 应用程序配置
export AUTO_APP_BINARY="./bin/pin_intent_broadcast_network"
export AUTO_APP_NAME="pin_intent_broadcast_network"

# 测试参数
export AUTO_TEST_DURATION=300  # 5分钟测试
export AUTO_INTENT_PUBLISH_INTERVAL=5  # 每5秒发布一个Intent
export AUTO_BIDDING_WINDOW=10  # 出价窗口10秒
export AUTO_MATCHING_WINDOW=15  # 匹配窗口15秒

# 目录配置
export AUTO_TEST_DATA_DIR="test_data/automation"
export AUTO_CONFIGS_DIR="configs"
export AUTO_SCRIPTS_DIR="scripts/automation"
export AUTO_STATUS_DIR="test_data/automation/status"
export AUTO_LOGS_DIR="test_data/automation/logs"

# 网络配置
export AUTO_PROTOCOL_ID="/pin-automation/1.0.0"
export AUTO_ENABLE_MDNS=true
export AUTO_ENABLE_DHT=true
export AUTO_MAX_CONNECTIONS=20

# GossipSub 配置
export AUTO_TOPIC_PREFIX="automation-test"
export AUTO_INTENT_TOPIC="${AUTO_TOPIC_PREFIX}.intent-broadcast"
export AUTO_BIDDING_TOPIC="${AUTO_TOPIC_PREFIX}.bidding"
export AUTO_MATCHING_TOPIC="${AUTO_TOPIC_PREFIX}.matching"

# Intent配置
export AUTO_INTENT_TYPES=("trade" "swap" "exchange" "data_access")
export AUTO_INTENT_PRIORITIES=(1 2 3 4 5)

# 颜色配置
export AUTO_COLOR_RED='\033[0;31m'
export AUTO_COLOR_GREEN='\033[0;32m'
export AUTO_COLOR_YELLOW='\033[1;33m'
export AUTO_COLOR_BLUE='\033[0;34m'
export AUTO_COLOR_CYAN='\033[0;36m'
export AUTO_COLOR_MAGENTA='\033[0;35m'
export AUTO_COLOR_WHITE='\033[1;37m'
export AUTO_COLOR_GRAY='\033[0;37m'
export AUTO_COLOR_NC='\033[0m'

# Service Agent配置
export AUTO_AGENT1_STRATEGY="aggressive"  # 激进出价策略
export AUTO_AGENT1_BID_MARGIN=0.20       # 20%利润率
export AUTO_AGENT1_CAPABILITIES="trading,swap,exchange"

export AUTO_AGENT2_STRATEGY="conservative" # 保守出价策略
export AUTO_AGENT2_BID_MARGIN=0.15        # 15%利润率
export AUTO_AGENT2_CAPABILITIES="data_access,computation"

# Block Builder配置
export AUTO_BUILDER_ALGORITHM="highest_bid"  # 最高出价获胜
export AUTO_BUILDER_MIN_BIDS=1              # 最少需要1个出价
export AUTO_BUILDER_BID_COLLECTION_TIME=10  # 收集出价时间10秒

# 工具函数

# 获取自动化节点配置
get_auto_node_config() {
    local node_id=$1
    case $node_id in
        1)
            echo "HTTP_PORT=$AUTO_NODE1_HTTP_PORT"
            echo "GRPC_PORT=$AUTO_NODE1_GRPC_PORT"
            echo "P2P_PORT=$AUTO_NODE1_P2P_PORT"
            echo "CONFIG_FILE=\"$AUTO_NODE1_CONFIG_FILE\""
            echo "DATA_DIR=\"$AUTO_NODE1_DATA_DIR\""
            echo "LOG_FILE=\"$AUTO_NODE1_LOG_FILE\""
            echo "PID_FILE=\"$AUTO_NODE1_PID_FILE\""
            echo "STATUS_FILE=\"$AUTO_NODE1_STATUS_FILE\""
            echo "NODE_TYPE=\"$AUTO_NODE1_TYPE\""
            echo "NODE_NAME=\"$AUTO_NODE1_NAME\""
            ;;
        2)
            echo "HTTP_PORT=$AUTO_NODE2_HTTP_PORT"
            echo "GRPC_PORT=$AUTO_NODE2_GRPC_PORT"
            echo "P2P_PORT=$AUTO_NODE2_P2P_PORT"
            echo "CONFIG_FILE=\"$AUTO_NODE2_CONFIG_FILE\""
            echo "DATA_DIR=\"$AUTO_NODE2_DATA_DIR\""
            echo "LOG_FILE=\"$AUTO_NODE2_LOG_FILE\""
            echo "PID_FILE=\"$AUTO_NODE2_PID_FILE\""
            echo "STATUS_FILE=\"$AUTO_NODE2_STATUS_FILE\""
            echo "NODE_TYPE=\"$AUTO_NODE2_TYPE\""
            echo "NODE_NAME=\"$AUTO_NODE2_NAME\""
            echo "AGENT_ID=\"$AUTO_NODE2_AGENT_ID\""
            ;;
        3)
            echo "HTTP_PORT=$AUTO_NODE3_HTTP_PORT"
            echo "GRPC_PORT=$AUTO_NODE3_GRPC_PORT"
            echo "P2P_PORT=$AUTO_NODE3_P2P_PORT"
            echo "CONFIG_FILE=\"$AUTO_NODE3_CONFIG_FILE\""
            echo "DATA_DIR=\"$AUTO_NODE3_DATA_DIR\""
            echo "LOG_FILE=\"$AUTO_NODE3_LOG_FILE\""
            echo "PID_FILE=\"$AUTO_NODE3_PID_FILE\""
            echo "STATUS_FILE=\"$AUTO_NODE3_STATUS_FILE\""
            echo "NODE_TYPE=\"$AUTO_NODE3_TYPE\""
            echo "NODE_NAME=\"$AUTO_NODE3_NAME\""
            echo "AGENT_ID=\"$AUTO_NODE3_AGENT_ID\""
            ;;
        4)
            echo "HTTP_PORT=$AUTO_NODE4_HTTP_PORT"
            echo "GRPC_PORT=$AUTO_NODE4_GRPC_PORT"
            echo "P2P_PORT=$AUTO_NODE4_P2P_PORT"
            echo "CONFIG_FILE=\"$AUTO_NODE4_CONFIG_FILE\""
            echo "DATA_DIR=\"$AUTO_NODE4_DATA_DIR\""
            echo "LOG_FILE=\"$AUTO_NODE4_LOG_FILE\""
            echo "PID_FILE=\"$AUTO_NODE4_PID_FILE\""
            echo "STATUS_FILE=\"$AUTO_NODE4_STATUS_FILE\""
            echo "NODE_TYPE=\"$AUTO_NODE4_TYPE\""
            echo "NODE_NAME=\"$AUTO_NODE4_NAME\""
            echo "BUILDER_ID=\"$AUTO_NODE4_BUILDER_ID\""
            ;;
        *)
            echo "ERROR: Invalid automation node ID: $node_id"
            return 1
            ;;
    esac
}

# 获取节点HTTP端口
get_auto_node_http_port() {
    local node_id=$1
    case $node_id in
        1) echo $AUTO_NODE1_HTTP_PORT ;;
        2) echo $AUTO_NODE2_HTTP_PORT ;;
        3) echo $AUTO_NODE3_HTTP_PORT ;;
        4) echo $AUTO_NODE4_HTTP_PORT ;;
        *) echo "ERROR: Invalid automation node ID: $node_id"; return 1 ;;
    esac
}

# 获取节点类型
get_auto_node_type() {
    local node_id=$1
    case $node_id in
        1) echo "$AUTO_NODE1_TYPE" ;;
        2) echo "$AUTO_NODE2_TYPE" ;;
        3) echo "$AUTO_NODE3_TYPE" ;;
        4) echo "$AUTO_NODE4_TYPE" ;;
        *) echo "UNKNOWN"; return 1 ;;
    esac
}

# 验证自动化配置
validate_auto_config() {
    local errors=0
    
    # 检查应用程序
    if [ ! -f "$AUTO_APP_BINARY" ]; then
        echo "ERROR: Automation binary not found: $AUTO_APP_BINARY"
        errors=$((errors + 1))
    fi
    
    # 检查端口冲突
    local all_ports=($AUTO_NODE1_HTTP_PORT $AUTO_NODE1_GRPC_PORT $AUTO_NODE1_P2P_PORT
                     $AUTO_NODE2_HTTP_PORT $AUTO_NODE2_GRPC_PORT $AUTO_NODE2_P2P_PORT
                     $AUTO_NODE3_HTTP_PORT $AUTO_NODE3_GRPC_PORT $AUTO_NODE3_P2P_PORT
                     $AUTO_NODE4_HTTP_PORT $AUTO_NODE4_GRPC_PORT $AUTO_NODE4_P2P_PORT)
    
    local unique_ports=($(printf "%s\n" "${all_ports[@]}" | sort -u))
    
    if [ ${#all_ports[@]} -ne ${#unique_ports[@]} ]; then
        echo "ERROR: Port conflicts in automation configuration"
        errors=$((errors + 1))
    fi
    
    return $errors
}

# 显示自动化配置摘要
show_auto_config_summary() {
    echo -e "${AUTO_COLOR_BLUE}PIN 自动化测试系统配置${AUTO_COLOR_NC}"
    echo "============================="
    echo "节点数量: $AUTOMATION_NODE_COUNT"
    echo "应用程序: $AUTO_APP_BINARY"
    echo ""
    echo "节点配置:"
    echo -e "  节点1 (${AUTO_COLOR_GREEN}${AUTO_NODE1_NAME}${AUTO_COLOR_NC}): HTTP=$AUTO_NODE1_HTTP_PORT, gRPC=$AUTO_NODE1_GRPC_PORT, P2P=$AUTO_NODE1_P2P_PORT"
    echo -e "  节点2 (${AUTO_COLOR_YELLOW}${AUTO_NODE2_NAME}${AUTO_COLOR_NC}): HTTP=$AUTO_NODE2_HTTP_PORT, gRPC=$AUTO_NODE2_GRPC_PORT, P2P=$AUTO_NODE2_P2P_PORT"
    echo -e "  节点3 (${AUTO_COLOR_YELLOW}${AUTO_NODE3_NAME}${AUTO_COLOR_NC}): HTTP=$AUTO_NODE3_HTTP_PORT, gRPC=$AUTO_NODE3_GRPC_PORT, P2P=$AUTO_NODE3_P2P_PORT"
    echo -e "  节点4 (${AUTO_COLOR_MAGENTA}${AUTO_NODE4_NAME}${AUTO_COLOR_NC}): HTTP=$AUTO_NODE4_HTTP_PORT, gRPC=$AUTO_NODE4_GRPC_PORT, P2P=$AUTO_NODE4_P2P_PORT"
    echo ""
    echo "测试参数:"
    echo "  测试时长: $AUTO_TEST_DURATION 秒"
    echo "  Intent发布间隔: $AUTO_INTENT_PUBLISH_INTERVAL 秒"
    echo "  出价窗口: $AUTO_BIDDING_WINDOW 秒"
    echo "  匹配窗口: $AUTO_MATCHING_WINDOW 秒"
    echo ""
}

# 如果直接运行此脚本，显示配置摘要
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    show_auto_config_summary
    validate_auto_config
    if [ $? -eq 0 ]; then
        echo -e "${AUTO_COLOR_GREEN}配置验证通过 ✓${AUTO_COLOR_NC}"
    else
        echo -e "${AUTO_COLOR_RED}配置验证失败 ✗${AUTO_COLOR_NC}"
        exit 1
    fi
fi