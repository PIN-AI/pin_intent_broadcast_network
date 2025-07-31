#!/bin/bash

# PIN 多节点测试系统统一配置文件
# 所有脚本都应该 source 这个配置文件来获取统一的参数

# 基础配置
export NODE_COUNT=3
export TEST_ENV_NAME="PIN_MULTI_NODE_TEST"

# 端口配置
export BASE_HTTP_PORT=8000
export BASE_GRPC_PORT=9000
export BASE_P2P_PORT=9001

# 节点1配置（发布者节点）
export NODE1_HTTP_PORT=8000
export NODE1_GRPC_PORT=9000
export NODE1_P2P_PORT=9001
export NODE1_CONFIG_FILE="configs/test_node1.yaml"
export NODE1_DATA_DIR="test_data/node1"
export NODE1_LOG_FILE="test_data/node1/output.log"
export NODE1_PID_FILE="test_data/node1/pid"
export NODE1_STATUS_FILE="test_data/status/node1.status"

# 节点2配置（监控节点）
export NODE2_HTTP_PORT=8011
export NODE2_GRPC_PORT=9011
export NODE2_P2P_PORT=9012
export NODE2_CONFIG_FILE="configs/test_node2.yaml"
export NODE2_DATA_DIR="test_data/node2"
export NODE2_LOG_FILE="test_data/node2/output.log"
export NODE2_PID_FILE="test_data/node2/pid"
export NODE2_STATUS_FILE="test_data/status/node2.status"

# 节点3配置（监控节点）
export NODE3_HTTP_PORT=8022
export NODE3_GRPC_PORT=9022
export NODE3_P2P_PORT=9023
export NODE3_CONFIG_FILE="configs/test_node3.yaml"
export NODE3_DATA_DIR="test_data/node3"
export NODE3_LOG_FILE="test_data/node3/output.log"
export NODE3_PID_FILE="test_data/node3/pid"
export NODE3_STATUS_FILE="test_data/status/node3.status"

# 应用程序配置
export APP_BINARY="./bin/pin_intent_broadcast_network"
export APP_NAME="pin_intent_broadcast_network"

# 测试配置
export DEFAULT_TEST_DURATION=60
export DEFAULT_INTENT_PUBLISHER_MIN_INTERVAL=3
export DEFAULT_INTENT_PUBLISHER_MAX_INTERVAL=10
export DEFAULT_MONITOR_REFRESH_INTERVAL=3
export DEFAULT_NETWORK_CHECK_INTERVAL=5

# 目录配置
export TEST_DATA_DIR="test_data"
export CONFIGS_DIR="configs"
export SCRIPTS_DIR="scripts"
export STATUS_DIR="test_data/status"

# 日志配置
export LOG_LEVEL="info"
export LOG_FORMAT="json"

# 网络配置
export PROTOCOL_ID="/intent-broadcast/1.0.0"
export ENABLE_MDNS=true
export ENABLE_DHT=true
export MAX_CONNECTIONS=50
export ENABLE_SIGNING=true

# GossipSub 配置
export GOSSIPSUB_D=6
export GOSSIPSUB_D_LO=4
export GOSSIPSUB_D_HI=12
export GOSSIPSUB_HEARTBEAT_INTERVAL="1s"
export GOSSIPSUB_FANOUT_TTL="60s"
export ENABLE_MESSAGE_SIGNING=true
export ENABLE_STRICT_SIGNATURE_VERIFICATION=true
export MESSAGE_ID_CACHE_SIZE=1000
export MESSAGE_TTL="300s"
export MAX_MESSAGE_SIZE=1048576

# Intent 类型配置
export INTENT_TYPES=(
    "trade"
    "swap"
    "exchange"
    "transfer"
    "send"
    "payment"
    "lending"
    "borrow"
    "loan"
    "investment"
    "staking"
    "yield"
)

# 主题配置
export TOPIC_PREFIX="intent-broadcast"
export DEFAULT_TOPIC="${TOPIC_PREFIX}.general"

# 颜色配置
export COLOR_RED='\033[0;31m'
export COLOR_GREEN='\033[0;32m'
export COLOR_YELLOW='\033[1;33m'
export COLOR_BLUE='\033[0;34m'
export COLOR_CYAN='\033[0;36m'
export COLOR_MAGENTA='\033[0;35m'
export COLOR_WHITE='\033[1;37m'
export COLOR_NC='\033[0m' # No Color

# 工具函数

# 获取节点配置
get_node_config() {
    local node_id=$1
    case $node_id in
        1)
            echo "HTTP_PORT=$NODE1_HTTP_PORT"
            echo "GRPC_PORT=$NODE1_GRPC_PORT"
            echo "P2P_PORT=$NODE1_P2P_PORT"
            echo "CONFIG_FILE=$NODE1_CONFIG_FILE"
            echo "DATA_DIR=$NODE1_DATA_DIR"
            echo "LOG_FILE=$NODE1_LOG_FILE"
            echo "PID_FILE=$NODE1_PID_FILE"
            echo "STATUS_FILE=$NODE1_STATUS_FILE"
            ;;
        2)
            echo "HTTP_PORT=$NODE2_HTTP_PORT"
            echo "GRPC_PORT=$NODE2_GRPC_PORT"
            echo "P2P_PORT=$NODE2_P2P_PORT"
            echo "CONFIG_FILE=$NODE2_CONFIG_FILE"
            echo "DATA_DIR=$NODE2_DATA_DIR"
            echo "LOG_FILE=$NODE2_LOG_FILE"
            echo "PID_FILE=$NODE2_PID_FILE"
            echo "STATUS_FILE=$NODE2_STATUS_FILE"
            ;;
        3)
            echo "HTTP_PORT=$NODE3_HTTP_PORT"
            echo "GRPC_PORT=$NODE3_GRPC_PORT"
            echo "P2P_PORT=$NODE3_P2P_PORT"
            echo "CONFIG_FILE=$NODE3_CONFIG_FILE"
            echo "DATA_DIR=$NODE3_DATA_DIR"
            echo "LOG_FILE=$NODE3_LOG_FILE"
            echo "PID_FILE=$NODE3_PID_FILE"
            echo "STATUS_FILE=$NODE3_STATUS_FILE"
            ;;
        *)
            echo "ERROR: Invalid node ID: $node_id"
            return 1
            ;;
    esac
}

# 获取节点HTTP端口
get_node_http_port() {
    local node_id=$1
    case $node_id in
        1) echo $NODE1_HTTP_PORT ;;
        2) echo $NODE2_HTTP_PORT ;;
        3) echo $NODE3_HTTP_PORT ;;
        *) echo "ERROR: Invalid node ID: $node_id"; return 1 ;;
    esac
}

# 获取节点类型
get_node_type() {
    local node_id=$1
    case $node_id in
        1) echo "发布者" ;;
        2|3) echo "监控者" ;;
        *) echo "未知"; return 1 ;;
    esac
}

# 验证配置
validate_config() {
    local errors=0
    
    # 检查应用程序是否存在
    if [ ! -f "$APP_BINARY" ]; then
        echo "ERROR: Application binary not found: $APP_BINARY"
        errors=$((errors + 1))
    fi
    
    # 检查端口是否有冲突
    local all_ports=($NODE1_HTTP_PORT $NODE1_GRPC_PORT $NODE1_P2P_PORT 
                     $NODE2_HTTP_PORT $NODE2_GRPC_PORT $NODE2_P2P_PORT 
                     $NODE3_HTTP_PORT $NODE3_GRPC_PORT $NODE3_P2P_PORT)
    
    local unique_ports=($(printf "%s\n" "${all_ports[@]}" | sort -u))
    
    if [ ${#all_ports[@]} -ne ${#unique_ports[@]} ]; then
        echo "ERROR: Port conflicts detected in configuration"
        errors=$((errors + 1))
    fi
    
    return $errors
}

# 显示配置摘要
show_config_summary() {
    echo "PIN 多节点测试系统配置摘要"
    echo "============================"
    echo "节点数量: $NODE_COUNT"
    echo "应用程序: $APP_BINARY"
    echo ""
    echo "端口分配:"
    echo "  节点1: HTTP=$NODE1_HTTP_PORT, gRPC=$NODE1_GRPC_PORT, P2P=$NODE1_P2P_PORT"
    echo "  节点2: HTTP=$NODE2_HTTP_PORT, gRPC=$NODE2_GRPC_PORT, P2P=$NODE2_P2P_PORT"
    echo "  节点3: HTTP=$NODE3_HTTP_PORT, gRPC=$NODE3_GRPC_PORT, P2P=$NODE3_P2P_PORT"
    echo ""
    echo "测试配置:"
    echo "  默认测试时长: $DEFAULT_TEST_DURATION 秒"
    echo "  Intent发布间隔: $DEFAULT_INTENT_PUBLISHER_MIN_INTERVAL-$DEFAULT_INTENT_PUBLISHER_MAX_INTERVAL 秒"
    echo "  监控刷新间隔: $DEFAULT_MONITOR_REFRESH_INTERVAL 秒"
    echo ""
}

# 如果直接运行此脚本，显示配置摘要
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    show_config_summary
    validate_config
    if [ $? -eq 0 ]; then
        echo "配置验证通过 ✓"
    else
        echo "配置验证失败 ✗"
        exit 1
    fi
fi