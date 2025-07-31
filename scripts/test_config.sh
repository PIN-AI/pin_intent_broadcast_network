#!/bin/bash

# 测试统一配置的脚本

# 加载统一配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh"

echo "=== 统一配置测试 ==="
echo ""

echo "基础配置:"
echo "  节点数量: $NODE_COUNT"
echo "  应用程序: $APP_BINARY"
echo ""

echo "端口配置:"
for i in $(seq 1 $NODE_COUNT); do
    local http_port=$(get_node_http_port $i)
    local node_type=$(get_node_type $i)
    
    case $i in
        1) 
            echo "  节点$i ($node_type): HTTP=$NODE1_HTTP_PORT, gRPC=$NODE1_GRPC_PORT, P2P=$NODE1_P2P_PORT"
            ;;
        2) 
            echo "  节点$i ($node_type): HTTP=$NODE2_HTTP_PORT, gRPC=$NODE2_GRPC_PORT, P2P=$NODE2_P2P_PORT"
            ;;
        3) 
            echo "  节点$i ($node_type): HTTP=$NODE3_HTTP_PORT, gRPC=$NODE3_GRPC_PORT, P2P=$NODE3_P2P_PORT"
            ;;
    esac
done

echo ""
echo "测试配置:"
echo "  默认测试时长: $DEFAULT_TEST_DURATION 秒"
echo "  Intent发布间隔: $DEFAULT_INTENT_PUBLISHER_MIN_INTERVAL-$DEFAULT_INTENT_PUBLISHER_MAX_INTERVAL 秒"
echo "  监控刷新间隔: $DEFAULT_MONITOR_REFRESH_INTERVAL 秒"
echo ""

echo "配置验证:"
validate_config
if [ $? -eq 0 ]; then
    echo "✓ 配置验证通过"
else
    echo "✗ 配置验证失败"
fi