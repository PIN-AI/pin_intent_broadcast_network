#!/bin/bash

# PIN Intent Broadcast Network - Automation Demo Script
# 演示完整的自动化系统功能

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="http://localhost:8000"
INTENT_API_BASE="${BASE_URL}/pinai_intent/intent"
EXECUTION_API_BASE="${BASE_URL}/pinai_intent/execution"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# 帮助函数
print_header() {
    echo -e "${BLUE}================================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================================================${NC}"
}

print_step() {
    echo -e "${MAGENTA}>>> $1${NC}"
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

# 等待用户输入
wait_for_user() {
    echo -e "${YELLOW}按Enter键继续...${NC}"
    read -r
}

# API调用函数
call_api() {
    local endpoint="$1"
    local method="${2:-GET}"
    local data="${3:-{}}"
    local base_url="${4:-$EXECUTION_API_BASE}"
    
    if [[ "$method" == "POST" ]]; then
        curl -s -X POST \
             -H "Content-Type: application/json" \
             -d "$data" \
             "${base_url}${endpoint}" 2>/dev/null
    else
        curl -s "${base_url}${endpoint}" 2>/dev/null
    fi
}

# 检查服务器状态
check_server() {
    print_step "检查PIN服务器状态"
    
    local health_response
    health_response=$(curl -s "${BASE_URL}/health" 2>/dev/null)
    
    if [[ -n "$health_response" ]]; then
        print_success "PIN服务器正在运行"
        return 0
    else
        print_error "无法连接到PIN服务器"
        print_info "请确保服务器已启动: ./bin/pin_intent_broadcast_network -conf ./configs/config.yaml"
        return 1
    fi
}

# 显示初始系统状态
show_initial_status() {
    print_header "初始系统状态检查"
    
    print_step "获取Service Agents状态"
    local agents_response
    agents_response=$(call_api "/agents/status")
    
    if command -v jq &> /dev/null && [[ -n "$agents_response" ]]; then
        local total_agents
        total_agents=$(echo "$agents_response" | jq -r '.total_agents // 0')
        print_info "发现 $total_agents 个Service Agent"
        
        if [[ "$total_agents" -gt 0 ]]; then
            echo "$agents_response" | jq -r '.agents[] | "  - \(.agent_id): \(.status)"'
        fi
    else
        print_warning "无法解析Agent状态 (可能需要安装jq工具)"
    fi
    
    echo
    print_step "获取Block Builders状态"
    local builders_response
    builders_response=$(call_api "/builders/status")
    
    if command -v jq &> /dev/null && [[ -n "$builders_response" ]]; then
        local total_builders
        total_builders=$(echo "$builders_response" | jq -r '.total_builders // 0')
        print_info "发现 $total_builders 个Block Builder"
        
        if [[ "$total_builders" -gt 0 ]]; then
            echo "$builders_response" | jq -r '.builders[] | "  - \(.builder_id): \(.status)"'
        fi
    else
        print_warning "无法解析Builder状态"
    fi
    
    echo
    wait_for_user
}

# 创建测试意图
create_test_intent() {
    print_header "创建测试意图"
    
    print_step "创建交易类型意图"
    
    # 构建意图数据
    local intent_data
    intent_data=$(cat <<EOF
{
  "type": "trade",
  "payload": "$(echo -n "买入100个ETH，市价订单" | base64)",
  "sender_id": "demo-user-001",
  "priority": 5,
  "ttl": 3600,
  "user_address": "0x1234567890abcdef1234567890abcdef12345678",
  "intent_manifest": {
    "task": "执行加密货币交易",
    "requirements": {
      "asset": "ETH",
      "quantity": "100",
      "order_type": "market"
    },
    "context": "演示自动化出价系统"
  },
  "relevant_tags": [
    {
      "tag_name": "trading_history",
      "tag_fee": "5000",
      "is_tradable": true
    },
    {
      "tag_name": "risk_profile",
      "tag_fee": "3000", 
      "is_tradable": true
    }
  ],
  "max_duration": 1800
}
EOF
    )
    
    print_info "发送意图创建请求..."
    local response
    response=$(call_api "/create" "POST" "$intent_data" "$INTENT_API_BASE")
    
    if command -v jq &> /dev/null && [[ -n "$response" ]]; then
        local success
        success=$(echo "$response" | jq -r '.success // false')
        
        if [[ "$success" == "true" ]]; then
            local intent_id
            intent_id=$(echo "$response" | jq -r '.intent.id // ""')
            print_success "意图创建成功"
            print_info "意图ID: $intent_id"
            
            # 广播意图
            print_step "广播意图到P2P网络"
            local broadcast_data
            broadcast_data=$(cat <<EOF
{
  "intent_id": "$intent_id",
  "topic": "intent-broadcast.trade"
}
EOF
            )
            
            local broadcast_response
            broadcast_response=$(call_api "/broadcast" "POST" "$broadcast_data" "$INTENT_API_BASE")
            
            if [[ -n "$broadcast_response" ]]; then
                local broadcast_success
                broadcast_success=$(echo "$broadcast_response" | jq -r '.success // false')
                
                if [[ "$broadcast_success" == "true" ]]; then
                    print_success "意图广播成功"
                    print_info "广播主题: intent-broadcast.trade"
                else
                    local message
                    message=$(echo "$broadcast_response" | jq -r '.message // "未知错误"')
                    print_error "意图广播失败: $message"
                fi
            fi
            
            echo "$intent_id"  # 返回意图ID给调用者
        else
            local message
            message=$(echo "$response" | jq -r '.message // "未知错误"')
            print_error "意图创建失败: $message"
        fi
    else
        print_error "无法解析意图创建响应"
    fi
    
    echo
    wait_for_user
}

# 监控出价过程
monitor_bidding_process() {
    local intent_id="$1"
    
    if [[ -z "$intent_id" ]]; then
        print_warning "未提供意图ID，跳过出价监控"
        return
    fi
    
    print_header "监控自动出价过程"
    
    print_step "等待Service Agents响应 (15秒监控期)"
    
    for i in {1..15}; do
        echo -ne "\r监控进度: [$i/15] "
        
        # 检查是否有出价
        local bids_response
        bids_response=$(call_api "/intents/$intent_id/bids")
        
        if command -v jq &> /dev/null && [[ -n "$bids_response" ]]; then
            local success
            success=$(echo "$bids_response" | jq -r '.success // false')
            
            if [[ "$success" == "true" ]]; then
                local bids_count
                bids_count=$(echo "$bids_response" | jq -r '.bids | length')
                
                if [[ "$bids_count" -gt 0 ]]; then
                    echo  # 新行
                    print_success "发现 $bids_count 个出价"
                    echo "$bids_response" | jq -r '.bids[] | "  - Agent: \(.agent_id) | 出价: \(.bid_amount) | 类型: \(.agent_type)"'
                    break
                fi
            fi
        fi
        
        sleep 1
    done
    
    echo  # 确保新行
    wait_for_user
}

# 监控匹配过程
monitor_matching_process() {
    print_header "监控自动匹配过程"
    
    print_step "等待Block Builders处理匹配 (20秒监控期)"
    
    for i in {1..20}; do
        echo -ne "\r匹配进度: [$i/20] "
        
        # 检查最新的匹配结果
        local matches_response
        matches_response=$(call_api "/matches/history?limit=1")
        
        if command -v jq &> /dev/null && [[ -n "$matches_response" ]]; then
            local success
            success=$(echo "$matches_response" | jq -r '.success // false')
            
            if [[ "$success" == "true" ]]; then
                local matches_count
                matches_count=$(echo "$matches_response" | jq -r '.matches | length')
                
                if [[ "$matches_count" -gt 0 ]]; then
                    echo  # 新行
                    print_success "发现新的匹配结果"
                    echo "$matches_response" | jq -r '.matches[0] | "  意图ID: \(.intent_id)
  获胜Agent: \(.winning_agent)  
  获胜出价: \(.winning_bid)
  总出价数: \(.total_bids)
  匹配状态: \(.status)
  Block Builder: \(.block_builder_id)"'
                    break
                fi
            fi
        fi
        
        sleep 1
    done
    
    echo  # 确保新行
    wait_for_user
}

# 显示最终系统状态
show_final_status() {
    print_header "最终系统状态"
    
    print_step "系统执行指标"
    local metrics_response
    metrics_response=$(call_api "/metrics")
    
    if command -v jq &> /dev/null && [[ -n "$metrics_response" ]]; then
        local success
        success=$(echo "$metrics_response" | jq -r '.success // false')
        
        if [[ "$success" == "true" ]]; then
            echo "$metrics_response" | jq -r '.metrics | "系统指标总览:
  处理意图总数: \(.total_intents_processed // 0)
  提交出价总数: \(.total_bids_submitted // 0)  
  完成匹配总数: \(.total_matches_completed // 0)
  Agent成功率: \((.success_rate // 0) * 100 | floor)%
  平均响应时间: \(.average_response_time_ms // 0)ms
  活跃Agents: \(.active_agents // 0)
  活跃Builders: \(.active_builders // 0)"'
        fi
    fi
    
    echo
    print_step "最近匹配历史"
    local history_response
    history_response=$(call_api "/matches/history?limit=5")
    
    if command -v jq &> /dev/null && [[ -n "$history_response" ]]; then
        local success
        success=$(echo "$history_response" | jq -r '.success // false')
        
        if [[ "$success" == "true" ]]; then
            local total
            total=$(echo "$history_response" | jq -r '.total // 0')
            print_info "匹配历史总数: $total (显示最近5条)"
            
            if [[ "$total" -gt 0 ]]; then
                echo "$history_response" | jq -r '.matches[] | "  \(.intent_id): \(.winning_agent) (出价: \(.winning_bid))"'
            else
                print_info "暂无匹配历史"
            fi
        fi
    fi
    
    echo
    wait_for_user
}

# 演示Agent控制功能
demo_agent_control() {
    print_header "演示Agent控制功能"
    
    print_step "尝试启动一个停止的Agent"
    
    # 首先获取所有Agent状态
    local agents_response
    agents_response=$(call_api "/agents/status")
    
    if command -v jq &> /dev/null && [[ -n "$agents_response" ]]; then
        local stopped_agent
        stopped_agent=$(echo "$agents_response" | jq -r '.agents[] | select(.status == "stopped") | .agent_id' | head -1)
        
        if [[ -n "$stopped_agent" ]]; then
            print_info "发现停止的Agent: $stopped_agent"
            print_step "启动Agent: $stopped_agent"
            
            local start_response
            start_response=$(call_api "/agents/$stopped_agent/start" "POST")
            
            if [[ -n "$start_response" ]]; then
                local success
                success=$(echo "$start_response" | jq -r '.success // false')
                
                if [[ "$success" == "true" ]]; then
                    print_success "Agent启动成功"
                    
                    # 等待几秒后停止
                    print_step "等待5秒后停止Agent"
                    sleep 5
                    
                    local stop_response
                    stop_response=$(call_api "/agents/$stopped_agent/stop" "POST")
                    
                    if [[ -n "$stop_response" ]]; then
                        local stop_success
                        stop_success=$(echo "$stop_response" | jq -r '.success // false')
                        
                        if [[ "$stop_success" == "true" ]]; then
                            print_success "Agent停止成功"
                        else
                            local message
                            message=$(echo "$stop_response" | jq -r '.message // "未知错误"')
                            print_error "停止Agent失败: $message"
                        fi
                    fi
                else
                    local message
                    message=$(echo "$start_response" | jq -r '.message // "未知错误"')
                    print_error "启动Agent失败: $message"
                fi
            fi
        else
            print_info "所有Agent都在运行中，跳过控制演示"
        fi
    else
        print_warning "无法获取Agent状态进行控制演示"
    fi
    
    echo
    wait_for_user
}

# 显示帮助信息
show_help() {
    echo -e "${BLUE}PIN Intent Broadcast Network - 自动化系统演示${NC}"
    echo
    echo -e "${CYAN}用法:${NC}"
    echo "  $0 [选项]"
    echo
    echo -e "${CYAN}选项:${NC}"
    echo "  --full      - 完整演示 (默认)"
    echo "  --quick     - 快速演示 (跳过等待)"
    echo "  --help      - 显示此帮助信息"
    echo
    echo -e "${CYAN}演示流程:${NC}"
    echo "  1. 检查服务器状态"
    echo "  2. 显示初始系统状态"
    echo "  3. 创建并广播测试意图"
    echo "  4. 监控自动出价过程"
    echo "  5. 监控自动匹配过程"
    echo "  6. 演示Agent控制功能"
    echo "  7. 显示最终系统状态"
    echo
    echo -e "${CYAN}环境要求:${NC}"
    echo "  - PIN服务器运行在 $BASE_URL"
    echo "  - 推荐安装 'jq' 工具进行JSON解析"
    echo
    echo -e "${CYAN}启动服务器:${NC}"
    echo "  ./bin/pin_intent_broadcast_network -conf ./configs/config.yaml"
}

# 主演示函数
main_demo() {
    local quick_mode="$1"
    
    print_header "PIN 自动化系统完整演示"
    print_info "演示内容：Service Agent自动出价 + Block Builder自动匹配"
    echo
    
    if [[ "$quick_mode" != "--quick" ]]; then
        wait_for_user
    fi
    
    # 1. 检查服务器
    if ! check_server; then
        exit 1
    fi
    echo
    
    # 2. 显示初始状态
    show_initial_status
    
    # 3. 创建测试意图
    local intent_id
    intent_id=$(create_test_intent)
    
    # 4. 监控出价过程
    monitor_bidding_process "$intent_id"
    
    # 5. 监控匹配过程
    monitor_matching_process
    
    # 6. 演示Agent控制
    demo_agent_control
    
    # 7. 显示最终状态
    show_final_status
    
    print_header "演示完成"
    print_success "PIN自动化系统功能演示结束"
    print_info "使用 ./scripts/execution_monitor.sh 进行日常监控"
    echo
}

# 脚本入口
case "${1:-}" in
    "--help"|"-h")
        show_help
        ;;
    "--quick")
        main_demo "--quick"
        ;;
    *)
        main_demo
        ;;
esac