#!/bin/bash

# 网络状态检查工具
# 检查和显示 P2P 网络连接状态、主题订阅等信息

set -e

# 加载统一配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# 配置参数
NODE_PORT=${1:-$NODE1_HTTP_PORT}
CHECK_INTERVAL=$DEFAULT_NETWORK_CHECK_INTERVAL
CONTINUOUS_MODE=false

# 颜色定义（使用统一配置）
RED="$COLOR_RED"
GREEN="$COLOR_GREEN"
YELLOW="$COLOR_YELLOW"
BLUE="$COLOR_BLUE"
CYAN="$COLOR_CYAN"
MAGENTA="$COLOR_MAGENTA"
WHITE="$COLOR_WHITE"
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

log_network() {
    echo -e "${CYAN}[NETWORK]${NC} $1"
}

# 显示工具头部信息
show_header() {
    if [ "$CONTINUOUS_MODE" = true ]; then
        clear
    fi
    
    echo -e "${MAGENTA}================================${NC}"
    echo -e "${MAGENTA}      PIN 网络状态检查工具      ${NC}"
    echo -e "${MAGENTA}================================${NC}"
    echo ""
    echo -e "${CYAN}检查配置:${NC}"
    echo "  目标节点端口: $NODE_PORT"
    echo "  检查时间: $(date '+%Y-%m-%d %H:%M:%S')"
    if [ "$CONTINUOUS_MODE" = true ]; then
        echo "  刷新间隔: $CHECK_INTERVAL 秒"
    fi
    echo ""
}

# 检查节点可用性
check_node_availability() {
    log_step "检查节点可用性..."
    
    if curl -s "http://localhost:$NODE_PORT/health" >/dev/null 2>&1; then
        log_info "节点连接成功 (端口: $NODE_PORT)"
        return 0
    else
        log_error "无法连接到节点 (端口: $NODE_PORT)"
        log_error "请确保节点已启动并监听在指定端口"
        return 1
    fi
}

# 获取节点基本信息
get_node_info() {
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                        节点基本信息                          │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    # 尝试获取节点ID（如果API支持）
    local node_id="未知"
    local node_version="未知"
    local uptime="未知"
    
    # 检查健康状态
    local health_response=$(curl -s "http://localhost:$NODE_PORT/health" 2>/dev/null)
    if [ $? -eq 0 ] && [ -n "$health_response" ]; then
        printf "${WHITE}│${NC} 健康状态: ${GREEN}正常${NC}                                        ${WHITE}│${NC}\n"
    else
        printf "${WHITE}│${NC} 健康状态: ${RED}异常${NC}                                        ${WHITE}│${NC}\n"
    fi
    
    printf "${WHITE}│${NC} HTTP端口: %-10s                                    ${WHITE}│${NC}\n" "$NODE_PORT"
    printf "${WHITE}│${NC} 节点ID: %-20s                              ${WHITE}│${NC}\n" "$node_id"
    printf "${WHITE}│${NC} 版本: %-20s                                ${WHITE}│${NC}\n" "$node_version"
    printf "${WHITE}│${NC} 运行时间: %-20s                            ${WHITE}│${NC}\n" "$uptime"
    
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 检查 P2P 网络连接
check_p2p_connections() {
    echo ""
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                      P2P 网络连接状态                       │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    # 尝试通过不同的API端点获取P2P信息
    local p2p_info=""
    local connected_peers=0
    local listening_addresses=""
    
    # 尝试获取网络状态（这些API可能不存在，需要根据实际实现调整）
    local network_response=$(curl -s "http://localhost:$NODE_PORT/api/v1/network/status" 2>/dev/null)
    if [ $? -ne 0 ] || [ -z "$network_response" ]; then
        network_response=$(curl -s "http://localhost:$NODE_PORT/debug/network" 2>/dev/null)
    fi
    
    if [ $? -eq 0 ] && [ -n "$network_response" ]; then
        # 尝试解析连接的对等节点数量
        connected_peers=$(echo "$network_response" | grep -o '"peers":[0-9]*' | cut -d':' -f2 | head -1)
        if [ -z "$connected_peers" ]; then
            connected_peers=0
        fi
    fi
    
    # 检查其他节点的连接状态
    local node1_status="未知"
    local node2_status="未知"
    local node3_status="未知"
    
    # 检查节点1 (8000)
    if [ "$NODE_PORT" != "8000" ]; then
        if curl -s "http://localhost:8000/health" >/dev/null 2>&1; then
            node1_status="${GREEN}在线${NC}"
        else
            node1_status="${RED}离线${NC}"
        fi
    else
        node1_status="${CYAN}当前节点${NC}"
    fi
    
    # 检查节点2 (8001)
    if [ "$NODE_PORT" != "8001" ]; then
        if curl -s "http://localhost:8001/health" >/dev/null 2>&1; then
            node2_status="${GREEN}在线${NC}"
        else
            node2_status="${RED}离线${NC}"
        fi
    else
        node2_status="${CYAN}当前节点${NC}"
    fi
    
    # 检查节点3 (8002)
    if [ "$NODE_PORT" != "8002" ]; then
        if curl -s "http://localhost:8002/health" >/dev/null 2>&1; then
            node3_status="${GREEN}在线${NC}"
        else
            node3_status="${RED}离线${NC}"
        fi
    else
        node3_status="${CYAN}当前节点${NC}"
    fi
    
    printf "${WHITE}│${NC} 连接的对等节点: %-3d                                    ${WHITE}│${NC}\n" "$connected_peers"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    echo -e "${WHITE}│                      测试节点状态                           │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    printf "${WHITE}│${NC} 节点1 (端口8000): %-30s            ${WHITE}│${NC}\n" "$node1_status"
    printf "${WHITE}│${NC} 节点2 (端口8001): %-30s            ${WHITE}│${NC}\n" "$node2_status"
    printf "${WHITE}│${NC} 节点3 (端口8002): %-30s            ${WHITE}│${NC}\n" "$node3_status"
    
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 检查主题订阅状态
check_topic_subscriptions() {
    echo ""
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                      主题订阅状态                           │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    # 预定义的Intent广播主题
    local topics=(
        "intent-broadcast.trade"
        "intent-broadcast.swap"
        "intent-broadcast.exchange"
        "intent-broadcast.transfer"
        "intent-broadcast.send"
        "intent-broadcast.payment"
        "intent-broadcast.lending"
        "intent-broadcast.borrow"
        "intent-broadcast.loan"
        "intent-broadcast.investment"
        "intent-broadcast.staking"
        "intent-broadcast.yield"
        "intent-broadcast.general"
    )
    
    # 尝试获取主题订阅信息
    local topics_response=$(curl -s "http://localhost:$NODE_PORT/api/v1/pubsub/topics" 2>/dev/null)
    if [ $? -ne 0 ] || [ -z "$topics_response" ]; then
        topics_response=$(curl -s "http://localhost:$NODE_PORT/debug/pubsub" 2>/dev/null)
    fi
    
    local active_topics=0
    
    for topic in "${topics[@]}"; do
        local status="${YELLOW}未知${NC}"
        local peer_count="?"
        
        # 如果有API响应，尝试解析主题状态
        if [ -n "$topics_response" ]; then
            if echo "$topics_response" | grep -q "$topic"; then
                status="${GREEN}已订阅${NC}"
                active_topics=$((active_topics + 1))
                # 尝试获取对等节点数量
                peer_count=$(echo "$topics_response" | grep -A 2 "$topic" | grep -o '"peers":[0-9]*' | cut -d':' -f2 | head -1)
                if [ -z "$peer_count" ]; then
                    peer_count="0"
                fi
            else
                status="${RED}未订阅${NC}"
                peer_count="0"
            fi
        fi
        
        # 截断长主题名
        local short_topic=$(echo "$topic" | cut -c1-25)
        printf "${WHITE}│${NC} %-25s │ %-20s │ %3s peers ${WHITE}│${NC}\n" "$short_topic" "$status" "$peer_count"
    done
    
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    printf "${WHITE}│${NC} 活跃主题总数: %-3d                                      ${WHITE}│${NC}\n" "$active_topics"
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 检查网络延迟和质量
check_network_quality() {
    echo ""
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                      网络质量检查                           │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    # 测试到其他节点的HTTP延迟
    local node_ports=(8000 8001 8002)
    local total_latency=0
    local successful_tests=0
    
    for port in "${node_ports[@]}"; do
        if [ "$port" != "$NODE_PORT" ]; then
            local start_time=$(date +%s%3N)
            if curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1; then
                local end_time=$(date +%s%3N)
                local latency=$((end_time - start_time))
                total_latency=$((total_latency + latency))
                successful_tests=$((successful_tests + 1))
                
                local status_color="${GREEN}"
                if [ $latency -gt 100 ]; then
                    status_color="${YELLOW}"
                fi
                if [ $latency -gt 500 ]; then
                    status_color="${RED}"
                fi
                
                printf "${WHITE}│${NC} 节点 (端口%d): %s%3dms${NC}                                ${WHITE}│${NC}\n" "$port" "$status_color" "$latency"
            else
                printf "${WHITE}│${NC} 节点 (端口%d): ${RED}无响应${NC}                              ${WHITE}│${NC}\n" "$port"
            fi
        fi
    done
    
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    if [ $successful_tests -gt 0 ]; then
        local avg_latency=$((total_latency / successful_tests))
        local quality="${GREEN}良好${NC}"
        
        if [ $avg_latency -gt 100 ]; then
            quality="${YELLOW}一般${NC}"
        fi
        if [ $avg_latency -gt 500 ]; then
            quality="${RED}较差${NC}"
        fi
        
        printf "${WHITE}│${NC} 平均延迟: %3dms                                        ${WHITE}│${NC}\n" "$avg_latency"
        printf "${WHITE}│${NC} 连接质量: %-20s                            ${WHITE}│${NC}\n" "$quality"
        printf "${WHITE}│${NC} 成功连接: %d/%d                                         ${WHITE}│${NC}\n" "$successful_tests" "$((${#node_ports[@]} - 1))"
    else
        printf "${WHITE}│${NC} 平均延迟: ${RED}无法测量${NC}                                  ${WHITE}│${NC}\n"
        printf "${WHITE}│${NC} 连接质量: ${RED}无连接${NC}                                    ${WHITE}│${NC}\n"
        printf "${WHITE}│${NC} 成功连接: 0/%d                                         ${WHITE}│${NC}\n" "$((${#node_ports[@]} - 1))"
    fi
    
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 提供诊断建议
show_diagnostic_suggestions() {
    echo ""
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                        诊断建议                             │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    # 检查是否有节点离线
    local offline_nodes=0
    for port in 8000 8001 8002; do
        if [ "$port" != "$NODE_PORT" ]; then
            if ! curl -s --max-time 2 "http://localhost:$port/health" >/dev/null 2>&1; then
                offline_nodes=$((offline_nodes + 1))
            fi
        fi
    done
    
    if [ $offline_nodes -gt 0 ]; then
        printf "${WHITE}│${NC} ${YELLOW}⚠${NC}  发现 %d 个节点离线                                  ${WHITE}│${NC}\n" "$offline_nodes"
        printf "${WHITE}│${NC}    建议: 检查节点启动状态和端口占用情况                ${WHITE}│${NC}\n"
        printf "${WHITE}│${NC}    命令: ./scripts/cleanup_test.sh && ./scripts/setup_test_env.sh ${WHITE}│${NC}\n"
    else
        printf "${WHITE}│${NC} ${GREEN}✓${NC}  所有测试节点都在线                                  ${WHITE}│${NC}\n"
    fi
    
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    printf "${WHITE}│${NC} 常用诊断命令:                                          ${WHITE}│${NC}\n"
    printf "${WHITE}│${NC}   检查端口占用: lsof -i :8000-8002                     ${WHITE}│${NC}\n"
    printf "${WHITE}│${NC}   查看节点日志: tail -f test_data/node*/output.log     ${WHITE}│${NC}\n"
    printf "${WHITE}│${NC}   重启环境: ./scripts/cleanup_test.sh                  ${WHITE}│${NC}\n"
    printf "${WHITE}│${NC}   网络测试: ping localhost                             ${WHITE}│${NC}\n"
    
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 执行完整的网络状态检查
perform_full_check() {
    show_header
    
    if ! check_node_availability; then
        return 1
    fi
    
    get_node_info
    check_p2p_connections
    check_topic_subscriptions
    check_network_quality
    show_diagnostic_suggestions
    
    return 0
}

# 连续监控模式
continuous_monitor() {
    log_step "启动连续监控模式..."
    
    # 设置信号处理
    trap 'handle_interrupt' INT TERM
    
    while true; do
        if ! perform_full_check; then
            log_error "网络检查失败，等待 $CHECK_INTERVAL 秒后重试..."
        fi
        
        echo ""
        echo -e "${BLUE}[监控]${NC} 下次检查: ${CHECK_INTERVAL}秒后 | 按 Ctrl+C 停止监控"
        sleep $CHECK_INTERVAL
    done
}

# 处理中断信号
handle_interrupt() {
    echo ""
    log_step "接收到停止信号，正在停止网络监控..."
    log_info "网络状态检查工具已停止"
    echo ""
    exit 0
}

# 显示使用帮助
show_help() {
    echo "用法: $0 [选项] [节点端口]"
    echo ""
    echo "参数:"
    echo "  节点端口    目标节点的HTTP端口 (默认: 8000)"
    echo ""
    echo "选项:"
    echo "  -c, --continuous    连续监控模式"
    echo "  -i, --interval N    设置检查间隔 (秒，默认: 5)"
    echo "  -h, --help          显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                  # 检查端口8000的节点"
    echo "  $0 8001             # 检查端口8001的节点"
    echo "  $0 -c 8000          # 连续监控端口8000的节点"
    echo "  $0 -c -i 10 8001    # 每10秒检查一次端口8001的节点"
    echo ""
    echo "功能:"
    echo "  - 检查节点基本信息和健康状态"
    echo "  - 显示 P2P 网络连接状态"
    echo "  - 检查主题订阅情况"
    echo "  - 测试网络延迟和质量"
    echo "  - 提供诊断建议"
    echo ""
}

# 主函数
main() {
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--continuous)
                CONTINUOUS_MODE=true
                shift
                ;;
            -i|--interval)
                CHECK_INTERVAL="$2"
                if ! [[ "$CHECK_INTERVAL" =~ ^[0-9]+$ ]] || [ "$CHECK_INTERVAL" -lt 1 ]; then
                    log_error "无效的检查间隔: $CHECK_INTERVAL"
                    exit 1
                fi
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            [0-9]*)
                NODE_PORT=$1
                shift
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 执行检查
    if [ "$CONTINUOUS_MODE" = true ]; then
        continuous_monitor
    else
        perform_full_check
        echo ""
        echo -e "${CYAN}提示: 使用 -c 选项启用连续监控模式${NC}"
        echo ""
    fi
}

# 执行主函数
main "$@"