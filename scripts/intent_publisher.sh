#!/bin/bash

# Intent 自动发布器脚本
# 随机间隔（10-60秒）自动发布不同类型的 Intent

set -e

# 加载统一配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# 配置参数
NODE_PORT=${1:-$NODE1_HTTP_PORT}
MIN_INTERVAL=$DEFAULT_INTENT_PUBLISHER_MIN_INTERVAL
MAX_INTERVAL=$DEFAULT_INTENT_PUBLISHER_MAX_INTERVAL
PUBLISHER_NAME="Intent发布器"

# 颜色定义（使用统一配置）
RED="$COLOR_RED"
GREEN="$COLOR_GREEN"
YELLOW="$COLOR_YELLOW"
BLUE="$COLOR_BLUE"
CYAN="$COLOR_CYAN"
MAGENTA="$COLOR_MAGENTA"
NC="$COLOR_NC"

# 统计变量
TOTAL_PUBLISHED=0
TOTAL_FAILED=0
START_TIME=$(date +%s)

# Intent 类型数组（使用统一配置）
INTENT_TYPES=("${INTENT_TYPES[@]}")

# 测试数据模板
# 使用兼容的关联数组声明方式
# 初始化关联数组
declare -A INTENT_PAYLOADS 2>/dev/null || true
INTENT_PAYLOADS["trade"]="eyJ0b2tlbl9hIjoiRVRIIiwidG9rZW5fYiI6IlVTRFQiLCJhbW91bnRfYSI6IjEuNSIsImFtb3VudF9iIjoiMzAwMCJ9"  # {"token_a":"ETH","token_b":"USDT","amount_a":"1.5","amount_b":"3000"}
INTENT_PAYLOADS["swap"]="eyJ0b2tlbl9pbiI6IkVUSCIsInRva2VuX291dCI6IkJUQyIsImFtb3VudF9pbiI6IjIuMCIsInNsaXBwYWdlIjoiMC4wNSJ9"  # {"token_in":"ETH","token_out":"BTC","amount_in":"2.0","slippage":"0.05"}
INTENT_PAYLOADS["exchange"]="eyJwYWlyIjoiQlRDL1VTRFQiLCJzaWRlIjoiYnV5IiwiYW1vdW50IjoiMC4xIiwicHJpY2UiOiI0NTAwMCJ9"  # {"pair":"BTC/USDT","side":"buy","amount":"0.1","price":"45000"}
INTENT_PAYLOADS["transfer"]="eyJ0byI6IjB4MTIzNDU2Nzg5YWJjZGVmIiwidG9rZW4iOiJFVEgiLCJhbW91bnQiOiIwLjUifQ=="  # {"to":"0x123456789abcdef","token":"ETH","amount":"0.5"}
INTENT_PAYLOADS["send"]="eyJyZWNpcGllbnQiOiIweDk4NzY1NDMyMWZlZGNiYSIsImFzc2V0IjoiVVNEVCIsInZhbHVlIjoiMTAwIn0="  # {"recipient":"0x987654321fedcba","asset":"USDT","value":"100"}
INTENT_PAYLOADS["payment"]="eyJtZXJjaGFudCI6IkNyeXB0b1N0b3JlIiwiYW1vdW50IjoiNTAiLCJjdXJyZW5jeSI6IlVTRFQifQ=="  # {"merchant":"CryptoStore","amount":"50","currency":"USDT"}
INTENT_PAYLOADS["lending"]="eyJhc3NldCI6IkVUSCIsImFtb3VudCI6IjUuMCIsImR1cmF0aW9uIjoiMzAiLCJyYXRlIjoiNS4yIn0="  # {"asset":"ETH","amount":"5.0","duration":"30","rate":"5.2"}
INTENT_PAYLOADS["borrow"]="eyJjb2xsYXRlcmFsIjoiQlRDIiwiYm9ycm93X2Fzc2V0IjoiVVNEVCIsImFtb3VudCI6IjEwMDAwIiwibHR2IjoiNzAifQ=="  # {"collateral":"BTC","borrow_asset":"USDT","amount":"10000","ltv":"70"}
INTENT_PAYLOADS["loan"]="eyJsb2FuX2Ftb3VudCI6IjUwMDAiLCJjb2xsYXRlcmFsIjoiRVRIIiwiaW50ZXJlc3RfcmF0ZSI6IjguNSIsInRlcm0iOiI5MCJ9"  # {"loan_amount":"5000","collateral":"ETH","interest_rate":"8.5","term":"90"}
INTENT_PAYLOADS["investment"]="eyJzdHJhdGVneSI6IkRlRmkiLCJhc3NldHMiOlsiRVRIIiwiQlRDIl0sImFtb3VudCI6IjEwMDAiLCJyaXNrX2xldmVsIjoibWVkaXVtIn0="  # {"strategy":"DeFi","assets":["ETH","BTC"],"amount":"1000","risk_level":"medium"}
INTENT_PAYLOADS["staking"]="eyJ2YWxpZGF0b3IiOiJldGgyLXZhbGlkYXRvci0xIiwiYW1vdW50IjoiMzIuMCIsImR1cmF0aW9uIjoiMzY1In0="  # {"validator":"eth2-validator-1","amount":"32.0","duration":"365"}
INTENT_PAYLOADS["yield"]="eyJwcm90b2NvbCI6IlVuaXN3YXAiLCJwb29sIjoiRVRIL1VTRFQiLCJhbW91bnQiOiIxMDAwIiwiZXhwZWN0ZWRfYXB5IjoiMTIuNSJ9"  # {"protocol":"Uniswap","pool":"ETH/USDT","amount":"1000","expected_apy":"12.5"}

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

log_publish() {
    echo -e "${CYAN}[PUBLISH]${NC} $1"
}

log_stats() {
    echo -e "${MAGENTA}[STATS]${NC} $1"
}

# 显示发布器头部信息
show_publisher_header() {
    clear
    echo -e "${MAGENTA}================================${NC}"
    echo -e "${MAGENTA}      PIN Intent 自动发布器      ${NC}"
    echo -e "${MAGENTA}================================${NC}"
    echo ""
    echo -e "${CYAN}发布器配置:${NC}"
    echo "  目标节点端口: $NODE_PORT"
    echo "  发布间隔: ${MIN_INTERVAL}-${MAX_INTERVAL} 秒"
    echo "  支持Intent类型: ${#INTENT_TYPES[@]} 种"
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

# 生成随机间隔
generate_random_interval() {
    echo $((RANDOM % (MAX_INTERVAL - MIN_INTERVAL + 1) + MIN_INTERVAL))
}

# 获取随机 Intent 类型
get_random_intent_type() {
    local index=$((RANDOM % ${#INTENT_TYPES[@]}))
    echo "${INTENT_TYPES[$index]}"
}

# 生成随机发送者ID
generate_sender_id() {
    local node_id="node1"
    local random_suffix=$(printf "%04d" $((RANDOM % 10000)))
    echo "${node_id}-peer-${random_suffix}"
}

# 生成随机优先级
generate_priority() {
    echo $((RANDOM % 10 + 1))  # 1-10
}

# 创建 Intent
create_intent() {
    local intent_type=$1
    local sender_id=$2
    local priority=$3
    local payload=${INTENT_PAYLOADS[$intent_type]}
    
    if [ -z "$payload" ]; then
        payload="eyJ0eXBlIjoiJGludGVudF90eXBlIiwiZGF0YSI6InRlc3QifQ=="  # {"type":"$intent_type","data":"test"}
    fi
    
    local json_data=$(cat << EOF
{
    "type": "$intent_type",
    "payload": "$payload",
    "sender_id": "$sender_id",
    "priority": $priority,
    "ttl": 300
}
EOF
)
    
    log_publish "创建 Intent: 类型=$intent_type, 发送者=$sender_id, 优先级=$priority"
    
    local response=$(curl -s -X POST "http://localhost:$NODE_PORT/pinai_intent/intent/create" \
        -H "Content-Type: application/json" \
        -d "$json_data" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        # 尝试提取 Intent ID
        local intent_id=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 | head -1)
        if [ -z "$intent_id" ]; then
            # 尝试从 intent 对象中提取
            intent_id=$(echo "$response" | grep -o '"intent":{"id":"[^"]*"' | cut -d'"' -f6)
        fi
        
        if [ -n "$intent_id" ]; then
            log_info "Intent 创建成功: ID=$intent_id"
            
            # 尝试广播 Intent
            broadcast_intent "$intent_id" "$intent_type"
            
            TOTAL_PUBLISHED=$((TOTAL_PUBLISHED + 1))
            return 0
        else
            log_warn "Intent 创建响应异常: $response"
            TOTAL_FAILED=$((TOTAL_FAILED + 1))
            return 1
        fi
    else
        log_error "Intent 创建失败: 网络错误或节点无响应"
        TOTAL_FAILED=$((TOTAL_FAILED + 1))
        return 1
    fi
}

# 广播 Intent
broadcast_intent() {
    local intent_id=$1
    local intent_type=$2
    local topic="intent-broadcast.${intent_type}"
    
    log_publish "广播 Intent: ID=$intent_id, 主题=$topic"
    
    local json_data=$(cat << EOF
{
    "intent_id": "$intent_id",
    "topic": "$topic"
}
EOF
)
    
    local response=$(curl -s -X POST "http://localhost:$NODE_PORT/pinai_intent/intent/broadcast" \
        -H "Content-Type: application/json" \
        -d "$json_data" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        # 检查广播是否成功
        if echo "$response" | grep -q '"success":true'; then
            log_info "Intent 广播成功"
        else
            log_warn "Intent 广播可能失败: $response"
        fi
    else
        log_warn "Intent 广播请求失败"
    fi
}

# 显示统计信息
show_statistics() {
    local current_time=$(date +%s)
    local running_time=$((current_time - START_TIME))
    local hours=$((running_time / 3600))
    local minutes=$(((running_time % 3600) / 60))
    local seconds=$((running_time % 60))
    
    local success_rate=0
    if [ $((TOTAL_PUBLISHED + TOTAL_FAILED)) -gt 0 ]; then
        success_rate=$((TOTAL_PUBLISHED * 100 / (TOTAL_PUBLISHED + TOTAL_FAILED)))
    fi
    
    echo ""
    log_stats "发布统计信息"
    echo "  运行时间: ${hours}h ${minutes}m ${seconds}s"
    echo "  成功发布: $TOTAL_PUBLISHED"
    echo "  发布失败: $TOTAL_FAILED"
    echo "  成功率: ${success_rate}%"
    echo "  平均间隔: $((MIN_INTERVAL + MAX_INTERVAL)) / 2 = $(((MIN_INTERVAL + MAX_INTERVAL) / 2)) 秒"
    echo ""
}

# 显示下次发布倒计时
show_countdown() {
    local interval=$1
    local intent_type=$2
    
    log_step "下次发布: $intent_type (${interval}秒后)"
    
    for ((i=interval; i>0; i--)); do
        printf "\r${YELLOW}[倒计时]${NC} 还有 %02d 秒发布下一个 Intent..." $i
        sleep 1
    done
    printf "\r${YELLOW}[倒计时]${NC} 正在发布 Intent...                    \n"
}

# 处理中断信号
handle_interrupt() {
    echo ""
    log_step "接收到停止信号，正在停止发布器..."
    show_statistics
    log_info "Intent 发布器已停止"
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
    echo "  $0          # 连接到端口8000的节点"
    echo "  $0 8001     # 连接到端口8001的节点"
    echo ""
    echo "支持的Intent类型:"
    for type in "${INTENT_TYPES[@]}"; do
        echo "  - $type"
    done
    echo ""
}

# 主发布循环
main_publish_loop() {
    log_step "开始 Intent 自动发布循环..."
    
    # 设置信号处理
    trap 'handle_interrupt' INT TERM
    
    while true; do
        # 生成随机参数
        local interval=$(generate_random_interval)
        local intent_type=$(get_random_intent_type)
        local sender_id=$(generate_sender_id)
        local priority=$(generate_priority)
        
        # 显示倒计时
        show_countdown $interval $intent_type
        
        # 创建并发布 Intent
        create_intent "$intent_type" "$sender_id" "$priority"
        
        # 显示统计信息（每10次发布显示一次）
        if [ $((TOTAL_PUBLISHED % 10)) -eq 0 ] && [ $TOTAL_PUBLISHED -gt 0 ]; then
            show_statistics
        fi
    done
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
            NODE_PORT=8000
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
    
    show_publisher_header
    check_node_availability
    main_publish_loop
}

# 执行主函数
main "$@"