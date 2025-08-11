#!/bin/bash

# PIN Node1 自动化Intent发布器
# 为自动化测试环境设计的Intent自动发布脚本

set -e

# 加载自动化配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/automation_config.sh"

# 发布器配置
PUBLISHER_NAME="Node1自动Intent发布器"
PUBLISH_INTERVAL=${AUTO_INTENT_PUBLISH_INTERVAL:-5}  # 默认5秒间隔
MAX_INTENTS=${AUTO_INTENT_MAX_COUNT:-0}  # 0表示无限制
PUBLISHER_PID_FILE="$AUTO_STATUS_DIR/intent_publisher.pid"
PUBLISHER_LOG_FILE="$AUTO_LOGS_DIR/intent_publisher.log"

# 统计变量
TOTAL_PUBLISHED=0
TOTAL_FAILED=0
START_TIME=$(date +%s)

# Intent类型和数据模板
INTENT_TYPES=("trade" "swap" "exchange")

# Intent数据模板 - 使用数组方式避免关联数组兼容性问题
get_intent_payload() {
    local intent_type=$1
    case $intent_type in
        "trade")
            echo "eyJ0b2tlbl9hIjoiRVRIIiwidG9rZW5fYiI6IlVTRFQiLCJhbW91bnRfYSI6IjEuNSIsImFtb3VudF9iIjoiMzAwMCJ9"
            ;;
        "swap")
            echo "eyJ0b2tlbl9pbiI6IkVUSCIsInRva2VuX291dCI6IkJUQyIsImFtb3VudF9pbiI6IjIuMCIsInNsaXBwYWdlIjoiMC4wNSJ9"
            ;;
        "exchange")
            echo "eyJwYWlyIjoiQlRDL1VTRFQiLCJzaWRlIjoiYnV5IiwiYW1vdW50IjoiMC4xIiwicHJpY2UiOiI0NTAwMCJ9"
            ;;
        "data_access")
            echo "eyJkYXRhc2V0IjoibWFya2V0LWRhdGEiLCJxdWVyeSI6InByaWNlX2hpc3RvcnkiLCJwYXJhbXMiOnsia2Vs6eoiOiJFVEgvVVNEVCJ9fQ=="
            ;;
        "computation")
            echo "eyJ0YXNrIjoibWFyaWMrdHJhZGluZyIsImFsZ29yaXRobSI6Im1hY2QiLCJwYXJhbXMiOnsicGVyaW9kIjoiMWgiLCJkYXRhX3BvaW50cyI6MTAwfX0="
            ;;
        *)
            echo "eyJ0eXBlIjoiJGludGVudF90eXBlIiwiZGF0YSI6ImF1dG9tYXRpb24tdGVzdCJ9"
            ;;
    esac
}

# 日志函数
log_info() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${AUTO_COLOR_GREEN}[$timestamp][INFO]${AUTO_COLOR_NC} $1" | tee -a "$PUBLISHER_LOG_FILE"
}

log_warn() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${AUTO_COLOR_YELLOW}[$timestamp][WARN]${AUTO_COLOR_NC} $1" | tee -a "$PUBLISHER_LOG_FILE"
}

log_error() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${AUTO_COLOR_RED}[$timestamp][ERROR]${AUTO_COLOR_NC} $1" | tee -a "$PUBLISHER_LOG_FILE"
}

log_publish() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${AUTO_COLOR_CYAN}[$timestamp][PUBLISH]${AUTO_COLOR_NC} $1" | tee -a "$PUBLISHER_LOG_FILE"
}

log_stats() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${AUTO_COLOR_MAGENTA}[$timestamp][STATS]${AUTO_COLOR_NC} $1" | tee -a "$PUBLISHER_LOG_FILE"
}

# 显示发布器头部信息
show_header() {
    echo -e "${AUTO_COLOR_MAGENTA}========================================${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}    PIN Node1 自动Intent发布器    ${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}========================================${AUTO_COLOR_NC}"
    echo ""
    echo -e "${AUTO_COLOR_CYAN}发布器配置:${AUTO_COLOR_NC}"
    echo "  目标节点: Node1 (端口: $AUTO_NODE1_HTTP_PORT)"
    echo "  发布间隔: $PUBLISH_INTERVAL 秒"
    echo "  最大Intent数: $([ $MAX_INTENTS -eq 0 ] && echo "无限制" || echo $MAX_INTENTS)"
    echo "  支持Intent类型: ${#INTENT_TYPES[@]} 种 (${INTENT_TYPES[*]})"
    echo "  启动时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "  PID文件: $PUBLISHER_PID_FILE"
    echo "  日志文件: $PUBLISHER_LOG_FILE"
    echo ""
}

# 检查Node1可用性
check_node1_availability() {
    log_info "检查Node1可用性..."
    
    local max_retries=10
    local retry_count=0
    
    while [ $retry_count -lt $max_retries ]; do
        if curl -s "http://localhost:$AUTO_NODE1_HTTP_PORT/health" >/dev/null 2>&1; then
            log_info "Node1连接成功 (端口: $AUTO_NODE1_HTTP_PORT)"
            return 0
        fi
        
        retry_count=$((retry_count + 1))
        log_warn "Node1连接失败，重试 $retry_count/$max_retries..."
        sleep 3
    done
    
    log_error "无法连接到Node1 (端口: $AUTO_NODE1_HTTP_PORT)"
    log_error "请确保Node1已启动并正常运行"
    exit 1
}

# 获取随机Intent类型
get_random_intent_type() {
    local index=$((RANDOM % ${#INTENT_TYPES[@]}))
    echo "${INTENT_TYPES[$index]}"
}

# 生成发送者ID
generate_sender_id() {
    local random_suffix=$(printf "%04d" $((RANDOM % 10000)))
    echo "node1-auto-publisher-${random_suffix}"
}

# 生成随机优先级
generate_priority() {
    echo $((RANDOM % 5 + 6))  # 6-10，给自动发布的Intent较高优先级
}

# 创建并发布Intent
create_and_publish_intent() {
    local intent_type=$(get_random_intent_type)
    local sender_id=$(generate_sender_id)
    local priority=$(generate_priority)
    local payload=$(get_intent_payload "$intent_type")
    
    log_publish "创建Intent: 类型=$intent_type, 发送者=$sender_id, 优先级=$priority"
    
    local json_data=$(cat << EOF
{
    "type": "$intent_type",
    "payload": "$payload",
    "sender_id": "$sender_id",
    "priority": $priority,
    "ttl": 300,
    "tags": ["automation-test", "node1-publisher"]
}
EOF
)
    
    local response=$(curl -s -X POST "http://localhost:$AUTO_NODE1_HTTP_PORT/pinai_intent/intent/create" \
        -H "Content-Type: application/json" \
        -d "$json_data" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        # 提取Intent ID
        local intent_id=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4 | head -1)
        if [ -z "$intent_id" ]; then
            intent_id=$(echo "$response" | grep -o '"intent":{"id":"[^"]*"' | cut -d'"' -f6)
        fi
        
        if [ -n "$intent_id" ]; then
            log_info "Intent创建成功: ID=$intent_id"
            
            # 广播Intent
            broadcast_intent "$intent_id" "$intent_type"
            
            TOTAL_PUBLISHED=$((TOTAL_PUBLISHED + 1))
            return 0
        else
            log_warn "Intent创建响应异常: $response"
            TOTAL_FAILED=$((TOTAL_FAILED + 1))
            return 1
        fi
    else
        log_error "Intent创建失败: 网络错误或节点无响应"
        TOTAL_FAILED=$((TOTAL_FAILED + 1))
        return 1
    fi
}

# 广播Intent
broadcast_intent() {
    local intent_id=$1
    local intent_type=$2
    local topic="intent.broadcast.${intent_type}"
    
    log_publish "广播Intent: ID=$intent_id, 主题=$topic"
    
    local json_data=$(cat << EOF
{
    "intent_id": "$intent_id",
    "topic": "$topic"
}
EOF
)
    
    local response=$(curl -s -X POST "http://localhost:$AUTO_NODE1_HTTP_PORT/pinai_intent/intent/broadcast" \
        -H "Content-Type: application/json" \
        -d "$json_data" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        if echo "$response" | grep -q '"success":true'; then
            log_info "Intent广播成功"
        else
            log_warn "Intent广播可能失败: $response"
        fi
    else
        log_warn "Intent广播请求失败"
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
    echo "  发布间隔: $PUBLISH_INTERVAL 秒"
    echo ""
}

# 创建PID文件
create_pid_file() {
    echo $$ > "$PUBLISHER_PID_FILE"
    log_info "PID文件已创建: $PUBLISHER_PID_FILE (PID: $$)"
}

# 清理函数
cleanup() {
    log_info "正在停止自动Intent发布器..."
    
    # 移除PID文件
    if [ -f "$PUBLISHER_PID_FILE" ]; then
        rm -f "$PUBLISHER_PID_FILE"
        log_info "PID文件已删除"
    fi
    
    show_statistics
    log_info "自动Intent发布器已停止"
    exit 0
}

# 主发布循环
main_publish_loop() {
    log_info "开始Intent自动发布循环..."
    
    # 设置信号处理
    trap cleanup INT TERM
    
    local count=0
    
    while true; do
        # 检查是否达到最大Intent数量限制
        if [ $MAX_INTENTS -gt 0 ] && [ $count -ge $MAX_INTENTS ]; then
            log_info "已达到最大Intent数量限制: $MAX_INTENTS"
            break
        fi
        
        # 创建并发布Intent
        create_and_publish_intent
        count=$((count + 1))
        
        # 每10次发布显示统计信息
        if [ $((TOTAL_PUBLISHED % 10)) -eq 0 ] && [ $TOTAL_PUBLISHED -gt 0 ]; then
            show_statistics
        fi
        
        # 等待下次发布
        log_info "等待 $PUBLISH_INTERVAL 秒后发布下一个Intent..."
        sleep $PUBLISH_INTERVAL
    done
    
    cleanup
}

# 检查是否已有实例在运行
check_running_instance() {
    if [ -f "$PUBLISHER_PID_FILE" ]; then
        local pid=$(cat "$PUBLISHER_PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_error "自动Intent发布器已在运行 (PID: $pid)"
            log_error "如需重启，请先停止现有实例: kill $pid"
            exit 1
        else
            log_warn "发现孤立的PID文件，正在清理..."
            rm -f "$PUBLISHER_PID_FILE"
        fi
    fi
}

# 显示使用说明
show_usage() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --interval SECONDS    发布间隔（默认: $PUBLISH_INTERVAL 秒）"
    echo "  --max-count COUNT     最大Intent数量（0=无限制，默认: $MAX_INTENTS）"
    echo "  --help               显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                       # 使用默认配置"
    echo "  $0 --interval 60         # 每60秒发布一次"
    echo "  $0 --max-count 100       # 最多发布100个Intent"
    echo ""
}

# 主函数
main() {
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --interval)
                PUBLISH_INTERVAL="$2"
                shift 2
                ;;
            --max-count)
                MAX_INTENTS="$2"
                shift 2
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # 验证参数
    if ! [[ "$PUBLISH_INTERVAL" =~ ^[0-9]+$ ]] || [ "$PUBLISH_INTERVAL" -lt 1 ]; then
        log_error "无效的发布间隔: $PUBLISH_INTERVAL（必须是大于0的整数）"
        exit 1
    fi
    
    if ! [[ "$MAX_INTENTS" =~ ^[0-9]+$ ]]; then
        log_error "无效的最大Intent数量: $MAX_INTENTS（必须是非负整数）"
        exit 1
    fi
    
    # 确保日志目录存在
    mkdir -p "$(dirname "$PUBLISHER_LOG_FILE")"
    
    show_header
    check_running_instance
    create_pid_file
    check_node1_availability
    main_publish_loop
}

# 执行主函数
main "$@"