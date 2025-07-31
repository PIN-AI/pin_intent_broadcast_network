#!/bin/bash

# 网络状态监控脚本
# 显示P2P网络连接状态和topic订阅信息

set -e

# 加载统一配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# 配置参数
NODE_PORT=${1:-$NODE1_HTTP_PORT}
MONITOR_NAME="网络状态监控器"
REFRESH_INTERVAL=5

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

# 显示头部信息
show_header() {
    clear
    echo -e "${MAGENTA}================================${NC}"
    echo -e "${MAGENTA}      PIN 网络状态监控器        ${NC}"
    echo -e "${MAGENTA}================================${NC}"
    echo ""
    echo -e "${CYAN}监控配置:${NC}"
    echo "  目标节点端口: $NODE_PORT"
    echo "  刷新间隔: $REFRESH_INTERVAL 秒"
    echo "  监控时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""
}

# 检查节点可用性
check_node_availability() {
    log_step "检查目标节点可用性..."
    
    if curl -s "http://localhost:$NODE_PORT/health" >/dev/null 2>&1; then
        log_info "节点连接成功 (端口: $NODE_PORT)"
        return 0
    else
        log_error "无法连接到节点 (端口: $NODE_PORT)"
        log_error "请确保节点已启动并监听在指定端口"
        exit 1
    fi
}

# 查询P2P连接状态
query_p2p_status() {
    local response=$(curl -s "http://localhost:$NODE_PORT/debug/p2p/status" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        echo "$response"
        return 0
    else
        return 1
    fi
}

# 查询topic订阅状态
query_topic_subscriptions() {
    local response=$(curl -s "http://localhost:$NODE_PORT/debug/topics/subscriptions" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        echo "$response"
        return 0
    else
        return 1
    fi
}

# 查询peer连接信息
query_peer_connections() {
    local response=$(curl -s "http://localhost:$NODE_PORT/debug/p2p/peers" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        echo "$response"
        return 0
    else
        return 1
    fi
}

# 显示P2P状态
show_p2p_status() {
    local p2p_data="$1"
    
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                      P2P网络状态                            │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    if [ -n "$p2p_data" ]; then
        # 解析P2P状态信息
        local peer_id=$(echo "$p2p_data" | grep -o '"peer_id":"[^"]*"' | cut -d'"' -f4)
        local connected_peers=$(echo "$p2p_data" | grep -o '"connected_peers":[0-9]*' | cut -d':' -f2)
        local listen_addresses=$(echo "$p2p_data" | grep -o '"listen_addresses":\[[^]]*\]' | sed 's/"listen_addresses":\[//' | sed 's/\]$//' | tr ',' '\n' | wc -l)
        
        peer_id=${peer_id:-"unknown"}
        connected_peers=${connected_peers:-0}
        listen_addresses=${listen_addresses:-0}
        
        # 截断长的peer ID
        local short_peer_id=$(echo "$peer_id" | cut -c1-40)
        if [ ${#peer_id} -gt 40 ]; then
            short_peer_id="${short_peer_id}..."
        fi
        
        printf "${WHITE}│${NC} 节点ID: %-50s ${WHITE}│${NC}\n" "$short_peer_id"
        printf "${WHITE}│${NC} 连接的节点数: %-3d                                      ${WHITE}│${NC}\n" "$connected_peers"
        printf "${WHITE}│${NC} 监听地址数: %-3d                                        ${WHITE}│${NC}\n" "$listen_addresses"
        
        # 显示连接状态
        if [ "$connected_peers" -gt 0 ]; then
            printf "${WHITE}│${NC} 网络状态: ${GREEN}%-10s${NC}                                   ${WHITE}│${NC}\n" "已连接"
        else
            printf "${WHITE}│${NC} 网络状态: ${YELLOW}%-10s${NC}                                   ${WHITE}│${NC}\n" "未连接"
        fi
    else
        printf "${WHITE}│${NC} %-50s                     ${WHITE}│${NC}\n" "P2P状态信息不可用"
    fi
    
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 显示topic订阅状态
show_topic_subscriptions() {
    local topic_data="$1"
    
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                    Topic订阅状态                            │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    if [ -n "$topic_data" ]; then
        # 解析topic订阅信息
        local total_subscriptions=$(echo "$topic_data" | grep -o '"subscribed_topics":\[[^]]*\]' | grep -o ',' | wc -l)
        total_subscriptions=$((total_subscriptions + 1))
        
        # 如果没有逗号，检查是否有内容
        if echo "$topic_data" | grep -q '"subscribed_topics":\[\]'; then
            total_subscriptions=0
        fi
        
        printf "${WHITE}│${NC} 总订阅数: %-3d                                          ${WHITE}│${NC}\n" "$total_subscriptions"
        
        # 显示Intent相关的订阅
        local intent_subscriptions=0
        if echo "$topic_data" | grep -q "intent-broadcast"; then
            intent_subscriptions=$(echo "$topic_data" | grep -o "intent-broadcast[^\"]*" | wc -l)
        fi
        
        printf "${WHITE}│${NC} Intent订阅数: %-3d                                      ${WHITE}│${NC}\n" "$intent_subscriptions"
        
        # 显示订阅状态
        if [ "$total_subscriptions" -gt 0 ]; then
            printf "${WHITE}│${NC} 订阅状态: ${GREEN}%-10s${NC}                                   ${WHITE}│${NC}\n" "活跃"
        else
            printf "${WHITE}│${NC} 订阅状态: ${YELLOW}%-10s${NC}                                   ${WHITE}│${NC}\n" "无订阅"
        fi
    else
        printf "${WHITE}│${NC} %-50s                     ${WHITE}│${NC}\n" "Topic订阅信息不可用"
    fi
    
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 显示peer连接详情
show_peer_connections() {
    local peer_data="$1"
    
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                    Peer连接详情                             │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    if [ -n "$peer_data" ]; then
        # 解析peer连接信息
        local peer_count=$(echo "$peer_data" | grep -o '"peer_id":"[^"]*"' | wc -l)
        
        printf "${WHITE}│${NC} 连接的Peer数量: %-3d                                   ${WHITE}│${NC}\n" "$peer_count"
        
        if [ "$peer_count" -gt 0 ]; then
            echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
            echo -e "${WHITE}│                      Peer列表                               │${NC}"
            echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
            
            # 显示前几个peer的信息
            local count=0
            echo "$peer_data" | grep -o '"peer_id":"[^"]*"' | cut -d'"' -f4 | head -5 | while read -r peer_id; do
                count=$((count + 1))
                local short_peer_id=$(echo "$peer_id" | cut -c1-45)
                if [ ${#peer_id} -gt 45 ]; then
                    short_peer_id="${short_peer_id}..."
                fi
                printf "${WHITE}│${NC} %d. %-50s ${WHITE}│${NC}\n" "$count" "$short_peer_id"
            done
            
            if [ "$peer_count" -gt 5 ]; then
                printf "${WHITE}│${NC} ... 还有 %d 个peer                                      ${WHITE}│${NC}\n" $((peer_count - 5))
            fi
        else
            printf "${WHITE}│${NC} %-50s                     ${WHITE}│${NC}\n" "暂无连接的peer"
        fi
    else
        printf "${WHITE}│${NC} %-50s                     ${WHITE}│${NC}\n" "Peer连接信息不可用"
    fi
    
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 显示网络健康状态
show_network_health() {
    local p2p_data="$1"
    local topic_data="$2"
    
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                    网络健康状态                              │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    local health_score=0
    local max_score=100
    local status_text="未知"
    local status_color="$WHITE"
    
    # 评估P2P连接健康度
    if [ -n "$p2p_data" ]; then
        local connected_peers=$(echo "$p2p_data" | grep -o '"connected_peers":[0-9]*' | cut -d':' -f2)
        connected_peers=${connected_peers:-0}
        
        if [ "$connected_peers" -ge 3 ]; then
            health_score=$((health_score + 40))
        elif [ "$connected_peers" -ge 1 ]; then
            health_score=$((health_score + 20))
        fi
    fi
    
    # 评估topic订阅健康度
    if [ -n "$topic_data" ]; then
        local total_subscriptions=$(echo "$topic_data" | grep -o '"subscribed_topics":\[[^]]*\]' | grep -o ',' | wc -l)
        total_subscriptions=$((total_subscriptions + 1))
        
        if echo "$topic_data" | grep -q '"subscribed_topics":\[\]'; then
            total_subscriptions=0
        fi
        
        if [ "$total_subscriptions" -ge 10 ]; then
            health_score=$((health_score + 40))
        elif [ "$total_subscriptions" -ge 5 ]; then
            health_score=$((health_score + 30))
        elif [ "$total_subscriptions" -ge 1 ]; then
            health_score=$((health_score + 20))
        fi
    fi
    
    # 基础连通性检查
    health_score=$((health_score + 20))
    
    # 确定健康状态
    if [ "$health_score" -ge 80 ]; then
        status_text="优秀"
        status_color="$GREEN"
    elif [ "$health_score" -ge 60 ]; then
        status_text="良好"
        status_color="$CYAN"
    elif [ "$health_score" -ge 40 ]; then
        status_text="一般"
        status_color="$YELLOW"
    else
        status_text="较差"
        status_color="$RED"
    fi
    
    printf "${WHITE}│${NC} 健康评分: ${status_color}%-3d/100${NC}                                    ${WHITE}│${NC}\n" "$health_score"
    printf "${WHITE}│${NC} 网络状态: ${status_color}%-10s${NC}                                   ${WHITE}│${NC}\n" "$status_text"
    
    # 显示建议
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    echo -e "${WHITE}│                        建议                                 │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    if [ "$health_score" -lt 60 ]; then
        printf "${WHITE}│${NC} • 检查网络连接和防火墙设置                              ${WHITE}│${NC}\n"
        printf "${WHITE}│${NC} • 确保其他节点正在运行                                  ${WHITE}│${NC}\n"
        printf "${WHITE}│${NC} • 检查Intent监控配置是否正确                           ${WHITE}│${NC}\n"
    else
        printf "${WHITE}│${NC} • 网络状态良好，继续监控                                ${WHITE}│${NC}\n"
    fi
    
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 主监控循环
main_monitor_loop() {
    log_step "开始网络状态监控循环..."
    
    # 设置信号处理
    trap 'handle_interrupt' INT TERM
    
    while true; do
        # 显示头部
        show_header
        
        # 查询各种状态信息
        local p2p_data=$(query_p2p_status)
        local topic_data=$(query_topic_subscriptions)
        local peer_data=$(query_peer_connections)
        
        # 显示P2P状态
        show_p2p_status "$p2p_data"
        echo ""
        
        # 显示topic订阅状态
        show_topic_subscriptions "$topic_data"
        echo ""
        
        # 显示peer连接详情
        show_peer_connections "$peer_data"
        echo ""
        
        # 显示网络健康状态
        show_network_health "$p2p_data" "$topic_data"
        
        # 显示刷新信息
        echo ""
        echo -e "${BLUE}[刷新]${NC} 下次刷新: ${REFRESH_INTERVAL}秒后 | 按 Ctrl+C 停止监控"
        
        # 等待刷新间隔
        sleep $REFRESH_INTERVAL
    done
}

# 处理中断信号
handle_interrupt() {
    echo ""
    log_step "接收到停止信号，正在停止网络状态监控器..."
    echo ""
    log_info "网络状态监控器已停止"
    echo ""
    exit 0
}

# 显示使用帮助
show_help() {
    echo "用法: $0 [节点端口]"
    echo ""
    echo "参数:"
    echo "  节点端口    目标节点的HTTP端口 (默认: 8000)"
    echo ""
    echo "示例:"
    echo "  $0          # 监控端口8000的节点"
    echo "  $0 8001     # 监控端口8001的节点"
    echo "  $0 8002     # 监控端口8002的节点"
    echo ""
    echo "功能:"
    echo "  - 显示P2P网络连接状态"
    echo "  - 显示topic订阅信息"
    echo "  - 显示peer连接详情"
    echo "  - 评估网络健康状态"
    echo ""
}

# 主函数
main() {
    # 处理命令行参数
    case "${1:-}" in
        --help|-h)
            show_help
            exit 0
            ;;
        "")
            NODE_PORT=$NODE1_HTTP_PORT
            ;;
        [0-9]*)
            NODE_PORT=$1
            ;;
        *)
            log_error "无效的端口号: $1"
            show_help
            exit 1
            ;;
    esac
    
    check_node_availability
    main_monitor_loop
}

# 执行主函数
main "$@"