#!/bin/bash

# PIN Intent Broadcast Network - Execution Monitoring Script
# 监控自动化执行系统的状态和性能指标

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="http://localhost:8000"
EXECUTION_API_BASE="${BASE_URL}/pinai_intent/execution"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 帮助函数
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${CYAN}ℹ${NC} $1"
}

# API调用函数
call_api() {
    local endpoint="$1"
    local method="${2:-GET}"
    
    if [[ "$method" == "POST" ]]; then
        curl -s -X POST \
             -H "Content-Type: application/json" \
             -d '{}' \
             "${EXECUTION_API_BASE}${endpoint}" 2>/dev/null
    else
        curl -s "${EXECUTION_API_BASE}${endpoint}" 2>/dev/null
    fi
}

# 检查服务器是否运行
check_server_status() {
    print_header "检查服务器状态"
    
    local health_response
    health_response=$(curl -s "${BASE_URL}/health" 2>/dev/null)
    
    if [[ -n "$health_response" ]]; then
        print_success "PIN服务器正在运行 (${BASE_URL})"
        return 0
    else
        print_error "无法连接到PIN服务器 (${BASE_URL})"
        return 1
    fi
}

# 获取Service Agents状态
get_agents_status() {
    print_header "Service Agents 状态"
    
    local response
    response=$(call_api "/agents/status")
    
    if [[ -z "$response" ]]; then
        print_error "无法获取Agents状态"
        return 1
    fi
    
    # 解析JSON响应 (需要jq工具，如果没有则使用简单解析)
    if command -v jq &> /dev/null; then
        local success
        success=$(echo "$response" | jq -r '.success // false')
        
        if [[ "$success" == "true" ]]; then
            local total_agents
            local agents_data
            total_agents=$(echo "$response" | jq -r '.total_agents // 0')
            agents_data=$(echo "$response" | jq -r '.agents[]?')
            
            print_success "总Agent数量: $total_agents"
            echo
            
            if [[ -n "$agents_data" ]]; then
                echo -e "${CYAN}Agent详情:${NC}"
                echo "$response" | jq -r '.agents[] | "  ID: \(.agent_id) | 类型: \(.agent_type) | 状态: \(.status) | 活跃意图: \(.active_intents) | 成功出价: \(.successful_bids)"'
            else
                print_info "当前没有运行的Agent"
            fi
        else
            local message
            message=$(echo "$response" | jq -r '.message // "未知错误"')
            print_error "获取Agent状态失败: $message"
        fi
    else
        # 简单解析，不依赖jq
        if echo "$response" | grep -q '"success":true'; then
            print_success "Agent状态获取成功"
            echo -e "${CYAN}原始响应:${NC}"
            echo "$response" | head -10
        else
            print_error "获取Agent状态失败"
        fi
    fi
}

# 获取Block Builders状态
get_builders_status() {
    print_header "Block Builders 状态"
    
    local response
    response=$(call_api "/builders/status")
    
    if [[ -z "$response" ]]; then
        print_error "无法获取Builders状态"
        return 1
    fi
    
    if command -v jq &> /dev/null; then
        local success
        success=$(echo "$response" | jq -r '.success // false')
        
        if [[ "$success" == "true" ]]; then
            local total_builders
            local builders_data
            total_builders=$(echo "$response" | jq -r '.total_builders // 0')
            builders_data=$(echo "$response" | jq -r '.builders[]?')
            
            print_success "总Builder数量: $total_builders"
            echo
            
            if [[ -n "$builders_data" ]]; then
                echo -e "${CYAN}Builder详情:${NC}"
                echo "$response" | jq -r '.builders[] | "  ID: \(.builder_id) | 状态: \(.status) | 活跃会话: \(.active_sessions) | 完成匹配: \(.completed_matches) | 收到出价: \(.total_bids_received)"'
            else
                print_info "当前没有运行的Builder"
            fi
        else
            local message
            message=$(echo "$response" | jq -r '.message // "未知错误"')
            print_error "获取Builder状态失败: $message"
        fi
    else
        if echo "$response" | grep -q '"success":true'; then
            print_success "Builder状态获取成功"
            echo -e "${CYAN}原始响应:${NC}"
            echo "$response" | head -10
        else
            print_error "获取Builder状态失败"
        fi
    fi
}

# 获取执行指标
get_execution_metrics() {
    print_header "执行系统指标"
    
    local response
    response=$(call_api "/metrics")
    
    if [[ -z "$response" ]]; then
        print_error "无法获取执行指标"
        return 1
    fi
    
    if command -v jq &> /dev/null; then
        local success
        success=$(echo "$response" | jq -r '.success // false')
        
        if [[ "$success" == "true" ]]; then
            print_success "执行指标获取成功"
            echo
            
            local metrics
            metrics=$(echo "$response" | jq -r '.metrics')
            
            echo -e "${CYAN}系统指标:${NC}"
            echo "$response" | jq -r '.metrics | "  处理意图总数: \(.total_intents_processed // 0)
  提交出价总数: \(.total_bids_submitted // 0)
  完成匹配总数: \(.total_matches_completed // 0)
  成功率: \((.success_rate // 0) * 100 | floor)%
  平均响应时间: \(.average_response_time_ms // 0)ms
  活跃Agent数: \(.active_agents // 0)
  活跃Builder数: \(.active_builders // 0)
  最后更新: \(.last_updated // 0)"'
        else
            local message
            message=$(echo "$response" | jq -r '.message // "未知错误"')
            print_error "获取执行指标失败: $message"
        fi
    else
        if echo "$response" | grep -q '"success":true'; then
            print_success "执行指标获取成功"
            echo -e "${CYAN}原始响应:${NC}"
            echo "$response" | head -15
        else
            print_error "获取执行指标失败"
        fi
    fi
}

# 获取匹配历史
get_match_history() {
    print_header "匹配历史记录"
    
    local limit="${1:-10}"
    local response
    response=$(call_api "/matches/history?limit=$limit")
    
    if [[ -z "$response" ]]; then
        print_error "无法获取匹配历史"
        return 1
    fi
    
    if command -v jq &> /dev/null; then
        local success
        success=$(echo "$response" | jq -r '.success // false')
        
        if [[ "$success" == "true" ]]; then
            local total
            local matches
            total=$(echo "$response" | jq -r '.total // 0')
            matches=$(echo "$response" | jq -r '.matches[]?')
            
            print_success "匹配历史总数: $total (显示最近 $limit 条)"
            echo
            
            if [[ -n "$matches" ]]; then
                echo -e "${CYAN}最近匹配记录:${NC}"
                echo "$response" | jq -r '.matches[] | "  意图ID: \(.intent_id) | 获胜Agent: \(.winning_agent) | 获胜出价: \(.winning_bid) | 总出价数: \(.total_bids) | 状态: \(.status)"'
            else
                print_info "暂无匹配历史记录"
            fi
        else
            local message
            message=$(echo "$response" | jq -r '.message // "未知错误"')
            print_error "获取匹配历史失败: $message"
        fi
    else
        if echo "$response" | grep -q '"success":true'; then
            print_success "匹配历史获取成功"
            echo -e "${CYAN}原始响应:${NC}"
            echo "$response" | head -20
        else
            print_error "获取匹配历史失败"
        fi
    fi
}

# 启动特定Agent
start_agent() {
    local agent_id="$1"
    
    if [[ -z "$agent_id" ]]; then
        print_error "请提供Agent ID"
        return 1
    fi
    
    print_header "启动Agent: $agent_id"
    
    local response
    response=$(call_api "/agents/$agent_id/start" "POST")
    
    if [[ -z "$response" ]]; then
        print_error "无法启动Agent"
        return 1
    fi
    
    if command -v jq &> /dev/null; then
        local success
        success=$(echo "$response" | jq -r '.success // false')
        
        if [[ "$success" == "true" ]]; then
            print_success "Agent $agent_id 启动成功"
        else
            local message
            message=$(echo "$response" | jq -r '.message // "未知错误"')
            print_error "启动Agent失败: $message"
        fi
    else
        if echo "$response" | grep -q '"success":true'; then
            print_success "Agent $agent_id 启动成功"
        else
            print_error "启动Agent失败"
        fi
    fi
}

# 停止特定Agent
stop_agent() {
    local agent_id="$1"
    
    if [[ -z "$agent_id" ]]; then
        print_error "请提供Agent ID"
        return 1
    fi
    
    print_header "停止Agent: $agent_id"
    
    local response
    response=$(call_api "/agents/$agent_id/stop" "POST")
    
    if [[ -z "$response" ]]; then
        print_error "无法停止Agent"
        return 1
    fi
    
    if command -v jq &> /dev/null; then
        local success
        success=$(echo "$response" | jq -r '.success // false')
        
        if [[ "$success" == "true" ]]; then
            print_success "Agent $agent_id 停止成功"
        else
            local message
            message=$(echo "$response" | jq -r '.message // "未知错误"')
            print_error "停止Agent失败: $message"
        fi
    else
        if echo "$response" | grep -q '"success":true'; then
            print_success "Agent $agent_id 停止成功"
        else
            print_error "停止Agent失败"
        fi
    fi
}

# 实时监控模式
real_time_monitor() {
    local interval="${1:-5}"
    
    print_header "实时监控模式 (刷新间隔: ${interval}秒, 按Ctrl+C退出)"
    
    while true; do
        clear
        echo -e "${YELLOW}PIN执行系统实时监控${NC} - $(date '+%Y-%m-%d %H:%M:%S')"
        echo
        
        # 快速状态检查
        if ! check_server_status >/dev/null 2>&1; then
            print_error "服务器连接失败，等待重连..."
            sleep "$interval"
            continue
        fi
        
        # 获取指标
        get_execution_metrics
        echo
        get_agents_status
        echo
        get_builders_status
        
        echo
        print_info "下次刷新: ${interval}秒后 (Ctrl+C 退出)"
        sleep "$interval"
    done
}

# 显示帮助信息
show_help() {
    echo -e "${BLUE}PIN Intent Broadcast Network - 执行监控脚本${NC}"
    echo
    echo -e "${CYAN}用法:${NC}"
    echo "  $0 [命令] [参数]"
    echo
    echo -e "${CYAN}命令:${NC}"
    echo "  status              - 显示所有状态信息 (默认)"
    echo "  agents              - 显示Service Agents状态"
    echo "  builders            - 显示Block Builders状态"
    echo "  metrics             - 显示执行系统指标"
    echo "  history [数量]      - 显示匹配历史 (默认10条)"
    echo "  start-agent <ID>    - 启动指定Agent"
    echo "  stop-agent <ID>     - 停止指定Agent"
    echo "  monitor [间隔]      - 实时监控模式 (默认5秒刷新)"
    echo "  help                - 显示此帮助信息"
    echo
    echo -e "${CYAN}示例:${NC}"
    echo "  $0                           # 显示所有状态"
    echo "  $0 agents                    # 只显示Agent状态"
    echo "  $0 history 20                # 显示最近20条匹配记录"
    echo "  $0 start-agent trading-agent-001"
    echo "  $0 monitor 3                 # 3秒刷新间隔的实时监控"
    echo
    echo -e "${CYAN}环境配置:${NC}"
    echo "  BASE_URL: $BASE_URL"
    echo "  API接口: $EXECUTION_API_BASE"
    echo
    echo -e "${YELLOW}注意: 需要安装 'jq' 工具以获得更好的JSON解析体验${NC}"
}

# 显示所有状态 (默认行为)
show_all_status() {
    if ! check_server_status; then
        exit 1
    fi
    
    echo
    get_execution_metrics
    echo
    get_agents_status
    echo
    get_builders_status
    echo
    get_match_history 5
}

# 主函数
main() {
    local command="${1:-status}"
    
    case "$command" in
        "status")
            show_all_status
            ;;
        "agents")
            if check_server_status >/dev/null 2>&1; then
                echo
                get_agents_status
            fi
            ;;
        "builders")
            if check_server_status >/dev/null 2>&1; then
                echo
                get_builders_status
            fi
            ;;
        "metrics")
            if check_server_status >/dev/null 2>&1; then
                echo
                get_execution_metrics
            fi
            ;;
        "history")
            if check_server_status >/dev/null 2>&1; then
                echo
                get_match_history "${2:-10}"
            fi
            ;;
        "start-agent")
            if check_server_status >/dev/null 2>&1; then
                echo
                start_agent "$2"
            fi
            ;;
        "stop-agent")
            if check_server_status >/dev/null 2>&1; then
                echo
                stop_agent "$2"
            fi
            ;;
        "monitor")
            real_time_monitor "${2:-5}"
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "未知命令: $command"
            echo
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
main "$@"