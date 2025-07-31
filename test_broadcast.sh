#!/bin/bash

# Intent 广播多节点测试脚本
# 用于测试 Intent P2P 广播功能

set -e

echo "=== Intent 广播多节点测试 ==="

# 配置
BASE_PORT=8000
P2P_BASE_PORT=9001
NODE_COUNT=2
TEST_INTENT_TYPE="trade"
TEST_PAYLOAD="dGVzdCBwYXlsb2Fk"  # base64 encoded "test payload"

# 创建测试数据目录
mkdir -p test_data/node1 test_data/node2

# 生成节点配置
generate_config() {
    local node_id=$1
    local http_port=$((BASE_PORT + node_id - 1))
    local grpc_port=$((http_port + 1000))
    local p2p_port=$((P2P_BASE_PORT + node_id - 1))
    
    cat > test_data/node${node_id}/config.yaml << EOF
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

p2p:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/${p2p_port}"
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
EOF
}

# 启动节点
start_node() {
    local node_id=$1
    local http_port=$((BASE_PORT + node_id - 1))
    local config_file="test_data/node${node_id}/config.yaml"
    
    echo "启动节点 ${node_id} (HTTP: ${http_port})..."
    
    # 构建并启动应用
    ./bin/pin_intent_broadcast_network -conf ${config_file} > test_data/node${node_id}/output.log 2>&1 &
    local pid=$!
    echo $pid > test_data/node${node_id}/pid
    
    echo "节点 ${node_id} 启动完成 (PID: ${pid})"
    
    # 等待节点启动
    sleep 3
    
    # 检查节点是否正常运行
    if curl -s http://localhost:${http_port}/health > /dev/null; then
        echo "节点 ${node_id} 健康检查通过"
    else
        echo "警告: 节点 ${node_id} 健康检查失败"
    fi
}

# 停止节点
stop_node() {
    local node_id=$1
    local pid_file="test_data/node${node_id}/pid"
    
    if [ -f $pid_file ]; then
        local pid=$(cat $pid_file)
        echo "停止节点 ${node_id} (PID: ${pid})..."
        kill -TERM $pid 2>/dev/null || true
        sleep 2
        kill -KILL $pid 2>/dev/null || true
        rm -f $pid_file
    fi
}

# 清理函数
cleanup() {
    echo "清理测试环境..."
    for i in $(seq 1 $NODE_COUNT); do
        stop_node $i
    done
    # rm -rf test_data
}

# 注册清理函数
trap cleanup EXIT

# 创建 Intent
create_intent() {
    local node_port=$1
    local sender_id=$2
    
    echo "在节点 (port: ${node_port}) 创建 Intent..." >&2
    
    local response=$(curl -s -X POST http://localhost:${node_port}/pinai_intent/intent/create \
        -H "Content-Type: application/json" \
        -d "{
            \"type\": \"${TEST_INTENT_TYPE}\",
            \"payload\": \"${TEST_PAYLOAD}\",
            \"sender_id\": \"${sender_id}\",
            \"priority\": 5,
            \"ttl\": 300
        }")
    
    echo "创建响应: $response" >&2
    
    # 提取 Intent ID (从intent对象中)
    local intent_id=$(echo $response | grep -o '"intent":{"id":"[^"]*"' | cut -d'"' -f6)
    echo "Intent ID: $intent_id" >&2
    echo $intent_id
}

# 广播 Intent
broadcast_intent() {
    local node_port=$1
    local intent_id=$2
    local topic=$3
    
    echo "从节点 (port: ${node_port}) 广播 Intent ${intent_id}..."
    
    local response=$(curl -s -X POST http://localhost:${node_port}/pinai_intent/intent/broadcast \
        -H "Content-Type: application/json" \
        -d "{
            \"intent_id\": \"${intent_id}\",
            \"topic\": \"${topic}\"
        }")
    
    echo "广播响应: $response"
}

# 查询 Intent
query_intents() {
    local node_port=$1
    
    echo "查询节点 (port: ${node_port}) 的 Intent 列表..."
    
    local response=$(curl -s "http://localhost:${node_port}/pinai_intent/intent/list?type=${TEST_INTENT_TYPE}&limit=10")
    echo "查询响应: $response"
}

# 检查网络状态
check_network_status() {
    local node_port=$1
    
    echo "检查节点 (port: ${node_port}) 的网络状态..."
    
    local response=$(curl -s "http://localhost:${node_port}/api/v1/network/status" 2>/dev/null || echo "Network status API not available")
    echo "网络状态: $response"
}

# 主测试流程
main() {
    echo "=== 准备测试环境 ==="
    
    # 确保编译完成
    if [ ! -f "./bin/pin_intent_broadcast_network" ]; then
        echo "构建应用..."
        make build
    fi
    
    # 生成配置文件
    for i in $(seq 1 $NODE_COUNT); do
        generate_config $i
    done
    
    echo "=== 启动节点 ==="
    for i in $(seq 1 $NODE_COUNT); do
        start_node $i
    done
    
    # 等待节点完全启动
    echo "等待节点完全启动..."
    sleep 5
    
    echo "=== 检查网络状态 ==="
    for i in $(seq 1 $NODE_COUNT); do
        local port=$((BASE_PORT + i - 1))
        check_network_status $port
    done
    
    echo "=== 执行 Intent 广播测试 ==="
    
    # 在节点1创建Intent
    local node1_port=$BASE_PORT
    local intent_id=$(create_intent $node1_port "node1-peer-id")
    
    if [ -z "$intent_id" ]; then
        echo "错误: 无法创建 Intent"
        exit 1
    fi
    
    sleep 2
    
    # 广播Intent
    broadcast_intent $node1_port $intent_id "intent-broadcast.trade"
    
    sleep 3
    
    echo "=== 验证广播结果 ==="
    
    # 检查所有节点的Intent列表
    for i in $(seq 1 $NODE_COUNT); do
        local port=$((BASE_PORT + i - 1))
        echo "--- 节点 ${i} Intent 列表 ---"
        query_intents $port
        echo ""
    done
    
    # 检查日志
    echo "=== 节点日志 ==="
    for i in $(seq 1 $NODE_COUNT); do
        echo "--- 节点 ${i} 日志 (最后20行) ---"
        tail -20 test_data/node${i}/output.log || echo "日志文件不存在"
        echo ""
    done
    
    echo "=== 测试完成 ==="
}

# 执行测试
main "$@"