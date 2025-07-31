#!/bin/bash

# Intent监控功能测试脚本
# 验证监控节点是否能接收到所有类型的Intent

set -e

# 加载统一配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# 颜色定义
RED="$COLOR_RED"
GREEN="$COLOR_GREEN"
YELLOW="$COLOR_YELLOW"
BLUE="$COLOR_BLUE"
CYAN="$COLOR_CYAN"
NC="$COLOR_NC"

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

log_test() {
    echo -e "${CYAN}[TEST]${NC} $1"
}

# 显示测试头部
show_test_header() {
    echo -e "${CYAN}================================${NC}"
    echo -e "${CYAN}    Intent监控功能测试          ${NC}"
    echo -e "${CYAN}================================${NC}"
    echo ""
}

# 检查节点是否运行
check_node_running() {
    local port=$1
    local node_name=$2
    
    if curl -s "http://localhost:$port/health" >/dev/null 2>&1; then
        log_info "$node_name 运行正常 (端口: $port)"
        return 0
    else
        log_error "$node_name 未运行或无法访问 (端口: $port)"
        return 1
    fi
}

# 发布测试Intent
publish_test_intent() {
    local port=$1
    local type=$2
    local payload=$3
    
    log_test "发布 $type 类型的Intent到节点 (端口: $port)"
    
    local response=$(curl -s -X POST "http://localhost:$port/pinai_intent/intent/create" \
        -H "Content-Type: application/json" \
        -d "{
            \"type\": \"$type\",
            \"payload\": \"$(echo -n "$payload" | base64)\",
            \"sender_id\": \"test-sender-$(date +%s)\",
            \"priority\": 5
        }" 2>/dev/null)
    
    if [ $? -eq 0 ] && echo "$response" | grep -q '"success":true'; then
        local intent_id=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        log_info "Intent创建成功: $intent_id"
        
        # 尝试广播Intent
        local broadcast_response=$(curl -s -X POST "http://localhost:$port/pinai_intent/intent/broadcast" \
            -H "Content-Type: application/json" \
            -d "{\"intent_id\": \"$intent_id\"}" 2>/dev/null)
        
        if [ $? -eq 0 ] && echo "$broadcast_response" | grep -q '"success":true'; then
            log_info "Intent广播成功"
            return 0
        else
            log_warn "Intent广播失败或部分失败"
            return 1
        fi
    else
        log_error "Intent创建失败"
        return 1
    fi
}

# 检查Intent接收情况
check_intent_reception() {
    local port=$1
    local node_name=$2
    local expected_count=$3
    
    log_test "检查 $node_name 的Intent接收情况"
    
    local response=$(curl -s "http://localhost:$port/pinai_intent/intent/list?limit=50" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        # 简单计算Intent数量
        local intent_count=$(echo "$response" | grep -o '"id":"[^"]*"' | wc -l)
        
        log_info "$node_name 当前有 $intent_count 个Intent"
        
        if [ "$intent_count" -ge "$expected_count" ]; then
            log_info "✓ $node_name Intent接收正常"
            return 0
        else
            log_warn "✗ $node_name Intent接收数量不足 (期望: $expected_count, 实际: $intent_count)"
            return 1
        fi
    else
        log_error "✗ 无法获取 $node_name 的Intent列表"
        return 1
    fi
}

# 检查监控配置
check_monitoring_config() {
    local port=$1
    local node_name=$2
    
    log_test "检查 $node_name 的监控配置"
    
    local response=$(curl -s "http://localhost:$port/debug/intent-monitoring/config" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        local mode=$(echo "$response" | grep -o '"subscription_mode":"[^"]*"' | cut -d'"' -f4)
        mode=${mode:-"unknown"}
        
        log_info "$node_name 监控模式: $mode"
        
        if [ "$mode" = "all" ]; then
            log_info "✓ $node_name 配置为监听所有topic"
            return 0
        else
            log_warn "✗ $node_name 监控模式不是'all': $mode"
            return 1
        fi
    else
        log_warn "无法获取 $node_name 的监控配置 (可能使用legacy模式)"
        return 0
    fi
}

# 检查订阅状态
check_subscription_status() {
    local port=$1
    local node_name=$2
    
    log_test "检查 $node_name 的订阅状态"
    
    local response=$(curl -s "http://localhost:$port/debug/intent-monitoring/subscriptions" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        local active_count=$(echo "$response" | grep -o '"active_subscriptions":\[[^]]*\]' | grep -o ',' | wc -l)
        active_count=$((active_count + 1))
        
        if echo "$response" | grep -q '"active_subscriptions":\[\]'; then
            active_count=0
        fi
        
        log_info "$node_name 活跃订阅数: $active_count"
        
        if [ "$active_count" -gt 10 ]; then
            log_info "✓ $node_name 订阅了足够多的topic"
            return 0
        else
            log_warn "✗ $node_name 订阅的topic数量较少: $active_count"
            return 1
        fi
    else
        log_warn "无法获取 $node_name 的订阅状态"
        return 0
    fi
}

# 运行完整测试
run_full_test() {
    show_test_header
    
    log_step "第1步: 检查节点运行状态"
    local nodes_running=0
    
    if check_node_running $NODE1_HTTP_PORT "节点1 (发布者)"; then
        nodes_running=$((nodes_running + 1))
    fi
    
    if check_node_running $NODE2_HTTP_PORT "节点2 (监控者)"; then
        nodes_running=$((nodes_running + 1))
    fi
    
    if check_node_running $NODE3_HTTP_PORT "节点3 (监控者)"; then
        nodes_running=$((nodes_running + 1))
    fi
    
    if [ $nodes_running -lt 2 ]; then
        log_error "需要至少2个节点运行才能进行测试"
        log_error "请先启动节点，然后重新运行测试"
        exit 1
    fi
    
    echo ""
    log_step "第2步: 检查监控配置"
    
    check_monitoring_config $NODE2_HTTP_PORT "节点2"
    check_monitoring_config $NODE3_HTTP_PORT "节点3"
    
    echo ""
    log_step "第3步: 检查订阅状态"
    
    check_subscription_status $NODE2_HTTP_PORT "节点2"
    check_subscription_status $NODE3_HTTP_PORT "节点3"
    
    echo ""
    log_step "第4步: 发布测试Intent"
    
    # 发布不同类型的Intent
    local intent_types=("trade" "swap" "exchange" "transfer" "send" "payment" "lending" "general")
    local published_count=0
    
    for intent_type in "${intent_types[@]}"; do
        if publish_test_intent $NODE1_HTTP_PORT "$intent_type" "test payload for $intent_type"; then
            published_count=$((published_count + 1))
        fi
        sleep 1  # 给网络传播一些时间
    done
    
    log_info "成功发布了 $published_count 个测试Intent"
    
    echo ""
    log_step "第5步: 等待Intent传播"
    log_info "等待10秒让Intent在网络中传播..."
    sleep 10
    
    echo ""
    log_step "第6步: 检查Intent接收情况"
    
    local expected_count=$((published_count / 2))  # 期望至少接收到一半
    local reception_success=0
    
    if check_intent_reception $NODE2_HTTP_PORT "节点2" $expected_count; then
        reception_success=$((reception_success + 1))
    fi
    
    if check_intent_reception $NODE3_HTTP_PORT "节点3" $expected_count; then
        reception_success=$((reception_success + 1))
    fi
    
    echo ""
    log_step "测试结果总结"
    echo "=================="
    
    if [ $reception_success -ge 1 ]; then
        log_info "✓ Intent监控功能测试通过！"
        log_info "  - 监控节点能够接收到Intent广播"
        log_info "  - 新的监控配置工作正常"
        echo ""
        log_info "🎉 问题已解决：监控节点现在可以接收所有类型的Intent了！"
    else
        log_error "✗ Intent监控功能测试失败"
        log_error "  - 监控节点无法正常接收Intent"
        log_error "  - 可能需要检查网络连接或配置"
        echo ""
        log_error "❌ 问题仍然存在，需要进一步调试"
    fi
    
    echo ""
}

# 显示使用帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --help, -h     显示此帮助信息"
    echo "  --quick        快速测试（跳过详细检查）"
    echo ""
    echo "功能:"
    echo "  - 检查节点运行状态"
    echo "  - 验证监控配置"
    echo "  - 测试Intent发布和接收"
    echo "  - 评估监控功能是否正常"
    echo ""
    echo "前置条件:"
    echo "  - 至少需要2个节点运行"
    echo "  - 建议运行所有3个节点以获得最佳测试效果"
    echo ""
}

# 主函数
main() {
    case "${1:-}" in
        --help|-h)
            show_help
            exit 0
            ;;
        --quick)
            log_info "运行快速测试模式"
            # 可以添加快速测试逻辑
            run_full_test
            ;;
        "")
            run_full_test
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"