#!/bin/bash

# PIN 自动化测试监控脚本
# 实时监控自动化测试的运行状态

set -e

# 加载自动化配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/automation_config.sh"

# 监控参数
REFRESH_INTERVAL="${1:-3}"  # 刷新间隔，默认3秒
MAX_HISTORY="${2:-10}"      # 最大历史记录数，默认10

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

# API调用函数
call_api() {
    local endpoint="$1"
    local base_url="$2"
    local timeout="${3:-2}"
    
    curl -s --connect-timeout $timeout "$base_url$endpoint" 2>/dev/null || echo '{"error": "connection_failed"}'
}

# 获取节点状态
get_node_status() {
    local node_id=$1
    local http_port=$(get_auto_node_http_port $node_id)
    
    if curl -s --connect-timeout 1 "http://localhost:$http_port/health" >/dev/null 2>&1; then
        echo "running"
    else
        echo "stopped"
    fi
}

# 显示节点状态
show_nodes_status() {
    echo -e "${AUTO_COLOR_CYAN}█ 节点状态${AUTO_COLOR_NC}"
    echo "├─────────────────────────────────────────────────────────────"
    
    for node_id in {1..4}; do
        local node_name=$(eval echo \$AUTO_NODE${node_id}_NAME)
        local node_type=$(eval echo \$AUTO_NODE${node_id}_TYPE)
        local http_port=$(get_auto_node_http_port $node_id)
        local status=$(get_node_status $node_id)
        
        local status_icon status_color
        if [ "$status" = "running" ]; then
            status_icon="●"
            status_color="$AUTO_COLOR_GREEN"
        else
            status_icon="○"
            status_color="$AUTO_COLOR_RED"
        fi
        
        printf "│ 节点%d: %s%-20s%s [%s:%d] %s%s%s\n" \
            $node_id \
            "$status_color" "$node_name" "$AUTO_COLOR_NC" \
            "$node_type" $http_port \
            "$status_color" "$status_icon" "$AUTO_COLOR_NC"
    done
    
    echo "└─────────────────────────────────────────────────────────────"
}

# 显示Service Agent状态
show_agents_status() {
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}█ Service Agents${AUTO_COLOR_NC}"
    echo "├─────────────────────────────────────────────────────────────"
    
    local total_agents=0
    local active_agents=0
    
    for node_id in 2 3; do  # Service Agent节点
        local http_port=$(get_auto_node_http_port $node_id)
        local agent_id=$(eval echo \$AUTO_NODE${node_id}_AGENT_ID)
        
        if [ "$(get_node_status $node_id)" = "running" ]; then
            local agents_response=$(call_api "/pinai_intent/execution/agents/status" "http://localhost:$http_port")
            
            if command -v jq &> /dev/null && [[ "$agents_response" != *"error"* ]]; then
                local agent_count=$(echo "$agents_response" | jq -r '.totalAgents // 0')
                local active_count=$(echo "$agents_response" | jq -r '.agents[] | select(.status == "active") | .agentId' | wc -l)
                
                total_agents=$((total_agents + agent_count))
                active_agents=$((active_agents + active_count))
                
                if [ $agent_count -gt 0 ]; then
                    echo "$agents_response" | jq -r --arg color "$AUTO_COLOR_YELLOW" --arg nc "$AUTO_COLOR_NC" \
                        '.agents[] | "│ \($color)\(.agentId)\($nc): \(.status) | 出价: \(.totalBidsSubmitted // .successfulBids // 0) | 成功: \(.successfulBids // 0)"'
                else
                    printf "│ %s节点%d%s: 无Agent运行\n" "$AUTO_COLOR_RED" $node_id "$AUTO_COLOR_NC"
                fi
            else
                printf "│ %s节点%d%s: API连接失败\n" "$AUTO_COLOR_RED" $node_id "$AUTO_COLOR_NC"
            fi
        else
            printf "│ %s节点%d%s: 节点未运行\n" "$AUTO_COLOR_RED" $node_id "$AUTO_COLOR_NC"
        fi
    done
    
    echo "├─────────────────────────────────────────────────────────────"
    printf "│ 总计: %d个Agent，%d个活跃\n" $total_agents $active_agents
    echo "└─────────────────────────────────────────────────────────────"
}

# 显示Block Builder状态
show_builders_status() {
    echo ""
    echo -e "${AUTO_COLOR_MAGENTA}█ Block Builders${AUTO_COLOR_NC}"
    echo "├─────────────────────────────────────────────────────────────"
    
    local node_id=4  # Block Builder节点
    local http_port=$(get_auto_node_http_port $node_id)
    
    if [ "$(get_node_status $node_id)" = "running" ]; then
        local builders_response=$(call_api "/pinai_intent/execution/builders/status" "http://localhost:$http_port" 5)
        
        if command -v jq &> /dev/null && [[ "$builders_response" != *"error"* ]] && [[ "$builders_response" != *"connection_failed"* ]]; then
            local success=$(echo "$builders_response" | jq -r '.success // false' 2>/dev/null)
            local builder_count=$(echo "$builders_response" | jq -r '.totalBuilders // 0' 2>/dev/null)
            
            if [ "$success" = "true" ] && [ "$builder_count" -gt 0 ]; then
                echo "$builders_response" | jq -r --arg color "$AUTO_COLOR_MAGENTA" --arg nc "$AUTO_COLOR_NC" \
                    '.builders[] | "│ \($color)\(.builderId)\($nc): \(.status) | 匹配: \(.completedMatches // 0) | 会话: \(.activeSessions // 0)"'
            else
                printf "│ %s无Builder运行或Builder服务未启动%s\n" "$AUTO_COLOR_YELLOW" "$AUTO_COLOR_NC"
            fi
        else
            printf "│ %s节点4运行中但Builder服务API不可用%s\n" "$AUTO_COLOR_YELLOW" "$AUTO_COLOR_NC"
        fi
    else
        printf "│ %s节点4未运行%s\n" "$AUTO_COLOR_RED" "$AUTO_COLOR_NC"
    fi
    
    echo "└─────────────────────────────────────────────────────────────"
}

# 查询节点的Intent列表
query_node_intents() {
    local node_id=$1
    local limit=${2:-10}
    local http_port=$(get_auto_node_http_port $node_id)
    
    if [ "$(get_node_status $node_id)" = "running" ]; then
        local response=$(call_api "/pinai_intent/intent/list?limit=$limit" "http://localhost:$http_port" 3)
        if [[ "$response" != *"error"* ]] && [[ "$response" != *"connection_failed"* ]]; then
            echo "$response"
            return 0
        fi
    fi
    return 1
}

# 解析Intent数据 (使用jq简化)
parse_intent_data() {
    local json_data="$1"
    
    if command -v jq &> /dev/null; then
        echo "$json_data" | jq -r '.intents[]? | "\(.id)|\(.type)|\(.senderId)|\(.timestamp)|\(.status)"' 2>/dev/null
    else
        # 回退到原来的文本解析方法
        local intents_array=$(echo "$json_data" | grep -o '"intents":\[.*\]' | sed 's/"intents":\[//' | sed 's/\]$//')
        
        if [ -z "$intents_array" ]; then
            return
        fi
        
        echo "$intents_array" | sed 's/}, *{/}\n{/g' | while IFS= read -r intent_line; do
            if [ -n "$intent_line" ]; then
                intent_line=$(echo "$intent_line" | sed 's/^{//' | sed 's/}$//')
                
                local id=$(echo "$intent_line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
                local type=$(echo "$intent_line" | grep -o '"type":"[^"]*"' | cut -d'"' -f4)
                local sender_id=$(echo "$intent_line" | grep -o '"senderId":"[^"]*"' | cut -d'"' -f4)
                local timestamp=$(echo "$intent_line" | grep -o '"timestamp":"[^"]*"' | cut -d'"' -f4)
                local status=$(echo "$intent_line" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
                
                id=${id:-"unknown"}
                type=${type:-"unknown"}
                sender_id=${sender_id:-"unknown"}
                timestamp=${timestamp:-0}
                status=${status:-"unknown"}
                
                echo "$id|$type|$sender_id|$timestamp|$status"
            fi
        done
    fi
}

# 格式化时间戳
format_timestamp() {
    local timestamp=$1
    if [ -n "$timestamp" ] && [ "$timestamp" != "0" ]; then
        # 处理字符串形式的时间戳
        if [[ "$timestamp" =~ ^[0-9]+$ ]]; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                date -r "$timestamp" '+%H:%M:%S' 2>/dev/null || echo "Unknown"
            else
                date -d "@$timestamp" '+%H:%M:%S' 2>/dev/null || echo "Unknown"
            fi
        else
            echo "Unknown"
        fi
    else
        echo "Unknown"
    fi
}

# 显示Service Agent收到的Intent列表
show_agent_intents() {
    echo ""
    echo -e "${AUTO_COLOR_CYAN}█ Service Agent Intent 列表${AUTO_COLOR_NC}"
    echo "├─────────────────────────────────────────────────────────────"
    
    local has_intents=false
    
    for node_id in 2 3; do  # Service Agent节点
        local node_name=$(eval echo \$AUTO_NODE${node_id}_NAME)
        local http_port=$(get_auto_node_http_port $node_id)
        
        # 直接调用API而不依赖query_node_intents
        if [ "$(get_node_status $node_id)" = "running" ]; then
            local intent_response=$(call_api "/pinai_intent/intent/list?limit=5" "http://localhost:$http_port" 3)
            
            # Debug output - remove in production
            echo "DEBUG: Node $node_id response: ${intent_response:0:100}..." >> /tmp/monitor_debug.log
            
            if [[ "$intent_response" != *"error"* ]] && [[ "$intent_response" != *"connection_failed"* ]] && [ -n "$intent_response" ]; then
                # 检查是否包含实际的intent数据
                if echo "$intent_response" | grep -q '"intents":\['; then
                    local parsed_intents=$(parse_intent_data "$intent_response")
                    
                    if [ -n "$parsed_intents" ]; then
                        printf "│ %s%s%s:\n" "$AUTO_COLOR_YELLOW" "$node_name" "$AUTO_COLOR_NC"
                        
                        local intent_count=0
                        while IFS='|' read -r id type sender_id timestamp status; do
                            if [ -n "$id" ] && [ "$intent_count" -lt 3 ]; then  # 最多显示3个
                                local short_id=$(echo "$id" | cut -c1-12)
                                local formatted_time=$(format_timestamp "$timestamp")
                                
                                printf "│   %s%s%s... | %s%s%s | %s%s%s | %s\n" \
                                    "$AUTO_COLOR_WHITE" "$short_id" "$AUTO_COLOR_NC" \
                                    "$AUTO_COLOR_GREEN" "$type" "$AUTO_COLOR_NC" \
                                    "$AUTO_COLOR_BLUE" "$formatted_time" "$AUTO_COLOR_NC" \
                                    "$status"
                                
                                intent_count=$((intent_count + 1))
                                has_intents=true
                            fi
                        done <<< "$parsed_intents"
                        
                        if [ "$intent_count" -eq 0 ]; then
                            printf "│   %sIntent数据解析失败%s\n" "$AUTO_COLOR_GRAY" "$AUTO_COLOR_NC"
                        fi
                    else
                        printf "│ %s%s%s: %s无法解析Intent数据%s\n" "$AUTO_COLOR_YELLOW" "$node_name" "$AUTO_COLOR_NC" "$AUTO_COLOR_GRAY" "$AUTO_COLOR_NC"
                    fi
                else
                    printf "│ %s%s%s: %s空Intent列表%s\n" "$AUTO_COLOR_YELLOW" "$node_name" "$AUTO_COLOR_NC" "$AUTO_COLOR_GRAY" "$AUTO_COLOR_NC"
                fi
            else
                printf "│ %s%s%s: %sAPI调用失败%s\n" "$AUTO_COLOR_YELLOW" "$node_name" "$AUTO_COLOR_NC" "$AUTO_COLOR_RED" "$AUTO_COLOR_NC"
            fi
        else
            printf "│ %s%s%s: %s节点未运行%s\n" "$AUTO_COLOR_YELLOW" "$node_name" "$AUTO_COLOR_NC" "$AUTO_COLOR_RED" "$AUTO_COLOR_NC"
        fi
    done
    
    if [ "$has_intents" = false ]; then
        printf "│ %s没有Service Agent收到Intent数据%s\n" "$AUTO_COLOR_GRAY" "$AUTO_COLOR_NC"
    fi
    
    echo "└─────────────────────────────────────────────────────────────"
}

# 显示系统指标
show_system_metrics() {
    echo ""
    echo -e "${AUTO_COLOR_BLUE}█ 系统指标${AUTO_COLOR_NC}"
    echo "├─────────────────────────────────────────────────────────────"
    
    # 尝试从任一运行的节点获取指标
    local metrics_found=false
    
    for node_id in {1..4}; do
        if [ "$(get_node_status $node_id)" = "running" ]; then
            local http_port=$(get_auto_node_http_port $node_id)
            local metrics_response=$(call_api "/pinai_intent/execution/metrics" "http://localhost:$http_port")
            
            if command -v jq &> /dev/null && [[ "$metrics_response" != *"error"* ]]; then
                local success=$(echo "$metrics_response" | jq -r '.success // false')
                
                if [ "$success" = "true" ]; then
                    local total_intents=$(echo "$metrics_response" | jq -r '.metrics.total_intents_processed // 0')
                    local total_bids=$(echo "$metrics_response" | jq -r '.metrics.total_bids_submitted // 0')
                    local total_matches=$(echo "$metrics_response" | jq -r '.metrics.total_matches_completed // 0')
                    local success_rate=$(echo "$metrics_response" | jq -r '.metrics.success_rate // 0')
                    local avg_response_time=$(echo "$metrics_response" | jq -r '.metrics.average_response_time_ms // 0')
                    
                    printf "│ Intent处理: %s%-10d%s | 出价提交: %s%-10d%s | 完成匹配: %s%-10d%s\n" \
                        "$AUTO_COLOR_GREEN" $total_intents "$AUTO_COLOR_NC" \
                        "$AUTO_COLOR_YELLOW" $total_bids "$AUTO_COLOR_NC" \
                        "$AUTO_COLOR_MAGENTA" $total_matches "$AUTO_COLOR_NC"
                    
                    local success_percent=$(echo "scale=1; $success_rate * 100" | bc 2>/dev/null || echo "0.0")
                    printf "│ 成功率: %s%.1f%%%s | 平均响应时间: %s%dms%s\n" \
                        "$AUTO_COLOR_GREEN" "$success_percent" "$AUTO_COLOR_NC" \
                        "$AUTO_COLOR_BLUE" $avg_response_time "$AUTO_COLOR_NC"
                    
                    metrics_found=true
                    break
                fi
            fi
        fi
    done
    
    if [ "$metrics_found" = false ]; then
        printf "│ %s无法获取系统指标 (节点未运行或API不可用)%s\n" "$AUTO_COLOR_RED" "$AUTO_COLOR_NC"
    fi
    
    echo "└─────────────────────────────────────────────────────────────"
}

# 显示最近匹配历史
show_recent_matches() {
    echo ""
    echo -e "${AUTO_COLOR_CYAN}█ 最近匹配历史 (最多${MAX_HISTORY}条)${AUTO_COLOR_NC}"
    echo "├─────────────────────────────────────────────────────────────"
    
    # 尝试从Block Builder节点获取匹配历史
    local node_id=4
    if [ "$(get_node_status $node_id)" = "running" ]; then
        local http_port=$(get_auto_node_http_port $node_id)
        local history_response=$(call_api "/pinai_intent/execution/matches/history?limit=$MAX_HISTORY" "http://localhost:$http_port")
        
        if command -v jq &> /dev/null && [[ "$history_response" != *"error"* ]]; then
            local success=$(echo "$history_response" | jq -r '.success // false')
            
            if [ "$success" = "true" ]; then
                local total_matches=$(echo "$history_response" | jq -r '.total // 0')
                
                if [ $total_matches -gt 0 ]; then
                    echo "$history_response" | jq -r --arg color "$AUTO_COLOR_GREEN" --arg nc "$AUTO_COLOR_NC" \
                        '.matches[] | "│ \(.intent_id[0:8])... → \($color)\(.winning_agent)\($nc) (出价: \(.winning_bid)) [\(.status)]"'
                else
                    printf "│ %s暂无匹配历史%s\n" "$AUTO_COLOR_YELLOW" "$AUTO_COLOR_NC"
                fi
            else
                printf "│ %s获取匹配历史失败%s\n" "$AUTO_COLOR_RED" "$AUTO_COLOR_NC"
            fi
        else
            printf "│ %sAPI连接失败%s\n" "$AUTO_COLOR_RED" "$AUTO_COLOR_NC"
        fi
    else
        printf "│ %sBlock Builder节点未运行%s\n" "$AUTO_COLOR_RED" "$AUTO_COLOR_NC"
    fi
    
    echo "└─────────────────────────────────────────────────────────────"
}

# 显示监控头部
show_monitor_header() {
    local current_time=$(date '+%Y-%m-%d %H:%M:%S')
    
    clear
    echo -e "${AUTO_COLOR_WHITE}════════════════════════════════════════════════════════════════${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_WHITE}    PIN 自动化测试系统 - 实时监控    ${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_WHITE}════════════════════════════════════════════════════════════════${AUTO_COLOR_NC}"
    echo ""
    printf "监控时间: %s | 刷新间隔: %ds | 历史记录: %d条\n" "$current_time" "$REFRESH_INTERVAL" "$MAX_HISTORY"
    echo ""
}

# 显示控制说明
show_controls() {
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}控制说明: ${AUTO_COLOR_NC}按 Ctrl+C 退出监控 | 按任意键刷新"
    echo ""
}

# 主监控循环
main_monitor_loop() {
    while true; do
        show_monitor_header
        show_nodes_status
        show_agents_status
        show_agent_intents
        show_builders_status
        show_system_metrics
        show_recent_matches
        show_controls
        
        # 等待指定时间或用户输入
        read -t $REFRESH_INTERVAL -n 1 key 2>/dev/null || true
    done
}

# 处理中断信号
handle_interrupt() {
    echo ""
    echo -e "${AUTO_COLOR_GREEN}监控已停止${AUTO_COLOR_NC}"
    echo ""
    exit 0
}

# 显示帮助信息
show_help() {
    echo "PIN 自动化测试监控工具"
    echo ""
    echo "用法: $0 [刷新间隔] [历史记录数]"
    echo ""
    echo "参数:"
    echo "  刷新间隔    监控刷新间隔(秒)，默认3秒"
    echo "  历史记录数  显示的最大历史匹配数，默认10条"
    echo ""
    echo "示例:"
    echo "  $0           # 默认参数 (3秒刷新，10条历史)"
    echo "  $0 5         # 5秒刷新，10条历史"
    echo "  $0 2 20      # 2秒刷新，20条历史"
    echo ""
    echo "监控内容:"
    echo "  - 节点运行状态"
    echo "  - Service Agent出价活动"
    echo "  - Block Builder匹配活动"
    echo "  - 系统性能指标"
    echo "  - 最近匹配历史"
    echo ""
}

# 主函数
main() {
    # 设置信号处理
    trap 'handle_interrupt' INT TERM
    
    # 验证参数
    if [[ "$REFRESH_INTERVAL" -lt 1 || "$REFRESH_INTERVAL" -gt 60 ]]; then
        echo -e "${AUTO_COLOR_RED}错误: 刷新间隔必须在1-60秒之间${AUTO_COLOR_NC}"
        exit 1
    fi
    
    if [[ "$MAX_HISTORY" -lt 1 || "$MAX_HISTORY" -gt 100 ]]; then
        echo -e "${AUTO_COLOR_RED}错误: 历史记录数必须在1-100之间${AUTO_COLOR_NC}"
        exit 1
    fi
    
    # 检查是否有节点运行 - 注释掉以允许监控即使在节点检测失败时也能运行
    # local running_nodes=0
    # for node_id in {1..4}; do
    #     if [ "$(get_node_status $node_id)" = "running" ]; then
    #         running_nodes=$((running_nodes + 1))
    #     fi
    # done
    # 
    # if [ $running_nodes -eq 0 ]; then
    #     echo -e "${AUTO_COLOR_YELLOW}警告: 没有检测到运行中的自动化测试节点${AUTO_COLOR_NC}"
    #     echo "请先运行 './scripts/automation/start_automation_test.sh' 启动测试"
    #     echo ""
    #     read -p "是否继续监控? (y/N): " -n 1 -r
    #     echo ""
    #     if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    #         exit 0
    #     fi
    # fi
    
    # 开始监控
    main_monitor_loop
}

# 处理参数
case "${1:-}" in
    "--help"|"-h")
        show_help
        ;;
    *)
        main "$@"
        ;;
esac