#!/bin/bash

# Intent 监控器脚本
# 实时监控和显示接收到的 Intent 信息

set -e

# 加载统一配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# 配置参数
NODE_PORT=${1:-$NODE2_HTTP_PORT}
MONITOR_NAME="Intent监控器"
REFRESH_INTERVAL=$DEFAULT_MONITOR_REFRESH_INTERVAL
MAX_DISPLAY_INTENTS=20

# 颜色定义（使用统一配置）
RED="$COLOR_RED"
GREEN="$COLOR_GREEN"
YELLOW="$COLOR_YELLOW"
BLUE="$COLOR_BLUE"
CYAN="$COLOR_CYAN"
MAGENTA="$COLOR_MAGENTA"
WHITE="$COLOR_WHITE"
NC="$COLOR_NC"

# 统计变量
TOTAL_RECEIVED=0
LAST_RECEIVED_TIME=0
START_TIME=$(date +%s)
LAST_INTENT_COUNT=0

# Intent 类型统计
declare -A INTENT_TYPE_COUNTS 2>/dev/null || true
declare -A LAST_SEEN_INTENTS 2>/dev/null || true

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

log_monitor() {
    echo -e "${CYAN}[MONITOR]${NC} $1"
}

log_stats() {
    echo -e "${MAGENTA}[STATS]${NC} $1"
}

# 显示监控器头部信息
show_monitor_header() {
    clear
    echo -e "${MAGENTA}================================${NC}"
    echo -e "${MAGENTA}      PIN Intent 监控器         ${NC}"
    echo -e "${MAGENTA}================================${NC}"
    echo ""
    echo -e "${CYAN}监控器配置:${NC}"
    echo "  目标节点端口: $NODE_PORT"
    echo "  刷新间隔: $REFRESH_INTERVAL 秒"
    echo "  最大显示数量: $MAX_DISPLAY_INTENTS"
    echo "  启动时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""
}

# 检查节点可用性
check_node_availability() {
    log_step "检查目标节点可用性..."
    
    local max_retries=5
    local retry_count=0
    
    while [ $retry_count -lt $max_retries ]; do
        if curl -s "http://localhost:$NODE_PORT/health" >/dev/null 2>&1; then
            log_info "节点连接成功 (端口: $NODE_PORT)"
            return 0
        fi
        
        retry_count=$((retry_count + 1))
        log_warn "节点连接失败，重试 $retry_count/$max_retries..."
        sleep 2
    done
    
    log_error "无法连接到节点 (端口: $NODE_PORT)"
    log_error "请确保节点已启动并监听在指定端口"
    exit 1
}

# 查询 Intent 列表
query_intents() {
    local limit=${1:-50}
    
    local response=$(curl -s "http://localhost:$NODE_PORT/pinai_intent/intent/list?limit=$limit" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        echo "$response"
        return 0
    else
        return 1
    fi
}

# 解析 Intent 数据
parse_intent_data() {
    local json_data="$1"
    
    # 使用简单的文本处理来解析JSON（避免依赖jq）
    echo "$json_data" | grep -o '"intents":\[.*\]' | sed 's/"intents":\[//' | sed 's/\]$//' | tr '},{' '\n' | while IFS= read -r intent_line; do
        if [ -n "$intent_line" ]; then
            # 清理JSON格式
            intent_line=$(echo "$intent_line" | sed 's/^{//' | sed 's/}$//')
            
            # 提取字段
            local id=$(echo "$intent_line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
            local type=$(echo "$intent_line" | grep -o '"type":"[^"]*"' | cut -d'"' -f4)
            local sender_id=$(echo "$intent_line" | grep -o '"sender_id":"[^"]*"' | cut -d'"' -f4)
            local timestamp=$(echo "$intent_line" | grep -o '"timestamp":[0-9]*' | cut -d':' -f2)
            local status=$(echo "$intent_line" | grep -o '"status":[0-9]*' | cut -d':' -f2)
            local priority=$(echo "$intent_line" | grep -o '"priority":[0-9]*' | cut -d':' -f2)
            
            # 输出解析结果
            echo "$id|$type|$sender_id|$timestamp|$status|$priority"
        fi
    done
}

# 格式化时间戳
format_timestamp() {
    local timestamp=$1
    if [ -n "$timestamp" ] && [ "$timestamp" -gt 0 ]; then
        date -d "@$timestamp" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || echo "Invalid time"
    else
        echo "Unknown"
    fi
}

# 格式化状态
format_status() {
    local status=$1
    case $status in
        0) echo "待处理" ;;
        1) echo "处理中" ;;
        2) echo "已完成" ;;
        3) echo "已失败" ;;
        4) echo "已过期" ;;
        *) echo "未知($status)" ;;
    esac
}

# 计算时间差
time_ago() {
    local timestamp=$1
    local current_time=$(date +%s)
    
    if [ -n "$timestamp" ] && [ "$timestamp" -gt 0 ]; then
        local diff=$((current_time - timestamp))
        
        if [ $diff -lt 60 ]; then
            echo "${diff}秒前"
        elif [ $diff -lt 3600 ]; then
            echo "$((diff / 60))分钟前"
        elif [ $diff -lt 86400 ]; then
            echo "$((diff / 3600))小时前"
        else
            echo "$((diff / 86400))天前"
        fi
    else
        echo "未知"
    fi
}

# 更新统计信息
update_statistics() {
    local intent_data="$1"
    local current_count=0
    
    # 重置类型统计
    for type in "${!INTENT_TYPE_COUNTS[@]}"; do
        INTENT_TYPE_COUNTS[$type]=0
    done
    
    # 统计当前 Intent
    while IFS='|' read -r id type sender_id timestamp status priority; do
        if [ -n "$id" ]; then
            current_count=$((current_count + 1))
            
            # 更新类型统计
            if [ -n "$type" ]; then
                INTENT_TYPE_COUNTS[$type]=$((${INTENT_TYPE_COUNTS[$type]:-0} + 1))
            fi
            
            # 检查是否是新的 Intent
            if [ -z "${LAST_SEEN_INTENTS[$id]}" ]; then
                LAST_SEEN_INTENTS[$id]="$timestamp"
                if [ "$timestamp" -gt "$LAST_RECEIVED_TIME" ]; then
                    LAST_RECEIVED_TIME=$timestamp
                fi
            fi
        fi
    done <<< "$intent_data"
    
    # 更新总接收数量
    if [ $current_count -gt $LAST_INTENT_COUNT ]; then
        TOTAL_RECEIVED=$((TOTAL_RECEIVED + current_count - LAST_INTENT_COUNT))
    fi
    LAST_INTENT_COUNT=$current_count
}

# 显示统计信息
show_statistics() {
    local current_time=$(date +%s)
    local running_time=$((current_time - START_TIME))
    local hours=$((running_time / 3600))
    local minutes=$(((running_time % 3600) / 60))
    local seconds=$((running_time % 60))
    
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                        监控统计信息                          │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    printf "${WHITE}│${NC} 运行时间: %02dh %02dm %02ds                                   ${WHITE}│${NC}\n" $hours $minutes $seconds
    printf "${WHITE}│${NC} 当前Intent数量: %-3d                                      ${WHITE}│${NC}\n" $LAST_INTENT_COUNT
    printf "${WHITE}│${NC} 总接收数量: %-3d                                          ${WHITE}│${NC}\n" $TOTAL_RECEIVED
    
    if [ $LAST_RECEIVED_TIME -gt 0 ]; then
        local last_received_ago=$(time_ago $LAST_RECEIVED_TIME)
        printf "${WHITE}│${NC} 最后接收: %-20s                              ${WHITE}│${NC}\n" "$last_received_ago"
    else
        printf "${WHITE}│${NC} 最后接收: %-20s                              ${WHITE}│${NC}\n" "无"
    fi
    
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    echo -e "${WHITE}│                      按类型统计                             │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────┤${NC}"
    
    local has_types=false
    for type in "${!INTENT_TYPE_COUNTS[@]}"; do
        if [ "${INTENT_TYPE_COUNTS[$type]}" -gt 0 ]; then
            printf "${WHITE}│${NC} %-15s: %-3d                                      ${WHITE}│${NC}\n" "$type" "${INTENT_TYPE_COUNTS[$type]}"
            has_types=true
        fi
    done
    
    if [ "$has_types" = false ]; then
        printf "${WHITE}│${NC} %-50s                     ${WHITE}│${NC}\n" "暂无数据"
    fi
    
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────┘${NC}"
}

# 显示 Intent 列表
show_intent_list() {
    local intent_data="$1"
    local count=0
    
    echo ""
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                                                    最新 Intent 列表                                                                    │${NC}"
    echo -e "${WHITE}├──────────────────────────┬──────────┬─────────────────────┬─────────────────────┬──────────┬────────┬──────────────────────────┤${NC}"
    echo -e "${WHITE}│ Intent ID                │ 类型     │ 发送者              │ 时间                │ 状态     │ 优先级 │ 接收时间                 │${NC}"
    echo -e "${WHITE}├──────────────────────────┼──────────┼─────────────────────┼─────────────────────┼──────────┼────────┼──────────────────────────┤${NC}"
    
    if [ -z "$intent_data" ]; then
        printf "${WHITE}│${NC} %-120s ${WHITE}│${NC}\n" "暂无 Intent 数据"
    else
        while IFS='|' read -r id type sender_id timestamp status priority; do
            if [ -n "$id" ] && [ $count -lt $MAX_DISPLAY_INTENTS ]; then
                local formatted_time=$(format_timestamp "$timestamp")
                local formatted_status=$(format_status "$status")
                local time_ago_str=$(time_ago "$timestamp")
                
                # 截断长字段
                local short_id=$(echo "$id" | cut -c1-24)
                local short_type=$(echo "$type" | cut -c1-8)
                local short_sender=$(echo "$sender_id" | cut -c1-19)
                local short_time=$(echo "$formatted_time" | cut -c1-19)
                local short_status=$(echo "$formatted_status" | cut -c1-8)
                local short_ago=$(echo "$time_ago_str" | cut -c1-24)
                
                # 高亮新接收的 Intent
                if [ "$timestamp" -gt $(($(date +%s) - 30)) ]; then
                    printf "${WHITE}│${GREEN} %-24s ${WHITE}│${GREEN} %-8s ${WHITE}│${GREEN} %-19s ${WHITE}│${GREEN} %-19s ${WHITE}│${GREEN} %-8s ${WHITE}│${GREEN} %-6s ${WHITE}│${GREEN} %-24s ${WHITE}│${NC}\n" \
                        "$short_id" "$short_type" "$short_sender" "$short_time" "$short_status" "$priority" "$short_ago"
                else
                    printf "${WHITE}│${NC} %-24s ${WHITE}│${NC} %-8s ${WHITE}│${NC} %-19s ${WHITE}│${NC} %-19s ${WHITE}│${NC} %-8s ${WHITE}│${NC} %-6s ${WHITE}│${NC} %-24s ${WHITE}│${NC}\n" \
                        "$short_id" "$short_type" "$short_sender" "$short_time" "$short_status" "$priority" "$short_ago"
                fi
                
                count=$((count + 1))
            fi
        done <<< "$intent_data"
        
        if [ $count -eq 0 ]; then
            printf "${WHITE}│${NC} %-120s ${WHITE}│${NC}\n" "暂无有效的 Intent 数据"
        fi
    fi
    
    echo -e "${WHITE}└──────────────────────────┴──────────┴─────────────────────┴─────────────────────┴──────────┴────────┴──────────────────────────┘${NC}"
}

# 显示等待状态
show_waiting_status() {
    local dots_count=$(($(date +%s) % 4))
    local dots=""
    for ((i=0; i<dots_count; i++)); do
        dots="$dots."
    done
    
    echo ""
    echo -e "${YELLOW}等待 Intent 数据$dots${NC}"
    echo -e "${CYAN}提示: 请确保发布者节点正在运行并发布 Intent${NC}"
    echo ""
}

# 主监控循环
main_monitor_loop() {
    log_step "开始 Intent 监控循环..."
    
    # 设置信号处理
    trap 'handle_interrupt' INT TERM
    
    while true; do
        # 清屏并显示头部
        clear
        show_monitor_header
        
        # 查询 Intent 数据
        local response=$(query_intents $MAX_DISPLAY_INTENTS)
        
        if [ $? -eq 0 ] && [ -n "$response" ]; then
            # 解析 Intent 数据
            local intent_data=$(parse_intent_data "$response")
            
            if [ -n "$intent_data" ]; then
                # 更新统计信息
                update_statistics "$intent_data"
                
                # 显示统计信息
                show_statistics
                
                # 显示 Intent 列表
                show_intent_list "$intent_data"
            else
                show_statistics
                show_waiting_status
            fi
        else
            log_error "无法获取 Intent 数据"
            show_waiting_status
        fi
        
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
    log_step "接收到停止信号，正在停止监控器..."
    
    # 显示最终统计
    echo ""
    log_stats "最终统计信息"
    local current_time=$(date +%s)
    local running_time=$((current_time - START_TIME))
    local hours=$((running_time / 3600))
    local minutes=$(((running_time % 3600) / 60))
    local seconds=$((running_time % 60))
    
    echo "  总运行时间: ${hours}h ${minutes}m ${seconds}s"
    echo "  监控的Intent数量: $LAST_INTENT_COUNT"
    echo "  总接收数量: $TOTAL_RECEIVED"
    
    if [ $LAST_RECEIVED_TIME -gt 0 ]; then
        echo "  最后接收时间: $(format_timestamp $LAST_RECEIVED_TIME)"
    fi
    
    echo ""
    log_info "Intent 监控器已停止"
    echo ""
    exit 0
}

# 显示使用帮助
show_help() {
    echo "用法: $0 [节点端口]"
    echo ""
    echo "参数:"
    echo "  节点端口    目标节点的HTTP端口 (默认: 8001)"
    echo ""
    echo "示例:"
    echo "  $0          # 监控端口8001的节点"
    echo "  $0 8000     # 监控端口8000的节点"
    echo "  $0 8002     # 监控端口8002的节点"
    echo ""
    echo "功能:"
    echo "  - 实时显示接收到的 Intent 信息"
    echo "  - 统计不同类型的 Intent 数量"
    echo "  - 显示接收时间和状态信息"
    echo "  - 高亮显示新接收的 Intent"
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
            NODE_PORT=8001
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