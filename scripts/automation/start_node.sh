#!/bin/bash

# PIN 自动化测试单节点启动脚本
# 在独立终端中启动指定节点

set -e

# 加载自动化配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/automation_config.sh"

# 获取节点ID参数
NODE_ID="${1:-1}"

if [[ "$NODE_ID" -lt 1 || "$NODE_ID" -gt 4 ]]; then
    echo -e "${AUTO_COLOR_RED}错误: 节点ID必须在1-4之间${AUTO_COLOR_NC}"
    echo "用法: $0 <node_id>"
    echo "  1 - Intent发布者节点"
    echo "  2 - Service Agent 1 (交易代理)"
    echo "  3 - Service Agent 2 (数据代理)"
    echo "  4 - Block Builder (匹配节点)"
    exit 1
fi

# 获取节点配置
eval $(get_auto_node_config $NODE_ID)

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

log_step() {
    echo -e "${AUTO_COLOR_BLUE}[STEP]${AUTO_COLOR_NC} $1"
}

log_node() {
    echo -e "${AUTO_COLOR_CYAN}[NODE$NODE_ID]${AUTO_COLOR_NC} $1"
}

# 显示节点信息头部
show_node_header() {
    clear
    echo -e "${AUTO_COLOR_MAGENTA}================================================${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}    PIN 自动化测试系统 - 节点$NODE_ID    ${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}         $NODE_NAME         ${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}================================================${AUTO_COLOR_NC}"
    echo ""
    echo -e "${AUTO_COLOR_CYAN}节点信息:${AUTO_COLOR_NC}"
    echo "  节点ID: $NODE_ID"
    echo "  节点类型: $NODE_TYPE"
    echo "  节点名称: $NODE_NAME"
    echo "  HTTP端口: $HTTP_PORT"
    echo "  gRPC端口: $GRPC_PORT"
    echo "  P2P端口: $P2P_PORT"
    echo "  配置文件: $CONFIG_FILE"
    echo "  日志文件: $LOG_FILE"
    
    # 显示特殊角色信息
    case $NODE_TYPE in
        "SERVICE_AGENT")
            echo "  Agent ID: $AGENT_ID"
            ;;
        "BLOCK_BUILDER")
            echo "  Builder ID: $BUILDER_ID"
            ;;
    esac
    echo ""
}

# 检查前置条件
check_prerequisites() {
    log_step "检查启动前置条件..."
    
    # 检查应用程序是否存在
    if [ ! -f "$AUTO_APP_BINARY" ]; then
        log_error "应用程序不存在: $AUTO_APP_BINARY"
        log_error "请先运行 'make build' 构建应用"
        exit 1
    fi
    
    # 检查配置文件是否存在
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "配置文件不存在: $CONFIG_FILE"
        log_error "请先运行 './scripts/automation/setup_automation_env.sh' 初始化环境"
        exit 1
    fi
    
    # 检查数据目录
    if [ ! -d "$DATA_DIR" ]; then
        log_error "节点数据目录不存在: $DATA_DIR"
        log_error "请先运行 './scripts/automation/setup_automation_env.sh' 初始化环境"
        exit 1
    fi
    
    # 自动设置节点对应的agents_config配置
    log_step "设置Node$NODE_ID对应的配置..."
    if [ -f "$SCRIPT_DIR/setup_automation_configs.sh" ]; then
        "$SCRIPT_DIR/setup_automation_configs.sh" setup "$NODE_ID" >/dev/null 2>&1 || {
            log_warn "自动配置设置失败，使用现有配置"
        }
        log_info "Node$NODE_ID配置已设置"
    else
        log_warn "配置设置脚本不存在，使用现有配置"
    fi
    
    log_info "前置条件检查通过"
}

# 检查端口可用性
check_port_availability() {
    log_step "检查端口可用性..."
    
    local ports=($HTTP_PORT $GRPC_PORT $P2P_PORT)
    local port_names=("HTTP" "gRPC" "P2P")
    local conflicts=0
    
    for i in "${!ports[@]}"; do
        local port=${ports[$i]}
        local name=${port_names[$i]}
        
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            local pid=$(lsof -Pi :$port -sTCP:LISTEN -t)
            log_error "端口 $port ($name) 已被占用，进程 PID: $pid"
            conflicts=$((conflicts + 1))
        fi
    done
    
    if [ $conflicts -gt 0 ]; then
        log_error "发现 $conflicts 个端口冲突"
        log_error "请使用 './scripts/automation/cleanup_automation.sh' 清理环境或手动停止占用进程"
        exit 1
    fi
    
    log_info "所有端口可用"
}

# 检查节点状态
check_node_status() {
    log_step "检查节点状态..."
    
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_error "节点$NODE_ID 已在运行，PID: $pid"
            log_error "请先停止现有节点或使用 './scripts/automation/cleanup_automation.sh' 清理环境"
            exit 1
        else
            log_warn "发现过期的 PID 文件，正在清理..."
            rm -f "$PID_FILE"
        fi
    fi
    
    log_info "节点状态检查通过"
}

# 启动节点
start_node() {
    log_step "启动节点$NODE_ID..."
    
    # 更新状态文件
    echo "starting" > "$STATUS_FILE"
    
    # 启动节点进程
    log_info "执行命令: $AUTO_APP_BINARY -conf $CONFIG_FILE"
    
    # 启动节点并获取 PID
    $AUTO_APP_BINARY -conf "$CONFIG_FILE" > "$LOG_FILE" 2>&1 &
    local pid=$!
    
    # 保存 PID
    echo $pid > "$PID_FILE"
    
    log_info "节点$NODE_ID 已启动，PID: $pid"
    
    # 等待节点启动
    log_step "等待节点启动..."
    local max_wait=30
    local wait_count=0
    
    while [ $wait_count -lt $max_wait ]; do
        if curl -s "http://localhost:$HTTP_PORT/health" >/dev/null 2>&1; then
            log_info "节点$NODE_ID 启动成功！"
            echo "running" > "$STATUS_FILE"
            break
        fi
        
        # 检查进程是否还在运行
        if ! kill -0 $pid 2>/dev/null; then
            log_error "节点进程意外退出"
            echo "failed" > "$STATUS_FILE"
            show_startup_error
            exit 1
        fi
        
        echo -n "."
        sleep 1
        wait_count=$((wait_count + 1))
    done
    
    if [ $wait_count -ge $max_wait ]; then
        log_error "节点启动超时"
        echo "timeout" > "$STATUS_FILE"
        kill $pid 2>/dev/null || true
        rm -f "$PID_FILE"
        show_startup_error
        exit 1
    fi
}

# 显示启动错误信息
show_startup_error() {
    log_error "节点启动失败，查看错误日志:"
    echo ""
    if [ -f "$LOG_FILE" ]; then
        echo -e "${AUTO_COLOR_RED}=== 错误日志 (最后20行) ===${AUTO_COLOR_NC}"
        tail -20 "$LOG_FILE"
        echo -e "${AUTO_COLOR_RED}=========================${AUTO_COLOR_NC}"
    else
        log_error "日志文件不存在: $LOG_FILE"
    fi
    echo ""
    log_error "可能的解决方案:"
    echo "1. 检查配置文件是否正确: $CONFIG_FILE"
    echo "2. 检查端口是否被占用"
    echo "3. 检查应用程序是否正确构建"
    echo "4. 运行 './scripts/automation/cleanup_automation.sh' 清理环境后重试"
}

# 显示节点运行信息
show_node_info() {
    echo ""
    log_step "节点运行信息"
    echo ""
    echo -e "${AUTO_COLOR_GREEN}✓ 节点$NODE_ID ($NODE_NAME) 运行中${AUTO_COLOR_NC}"
    echo ""
    echo "服务端点:"
    echo "  HTTP API: http://localhost:$HTTP_PORT"
    echo "  gRPC API: localhost:$GRPC_PORT"
    echo "  P2P 地址: /ip4/127.0.0.1/tcp/$P2P_PORT"
    echo ""
    
    # 显示特定节点的功能说明
    case $NODE_TYPE in
        "PUBLISHER")
            echo "节点功能:"
            echo "  - Intent发布者节点"
            echo "  - 只负责发布Intent消息到P2P网络"  
            echo "  - 不启动Service Agent或Block Builder"
            echo "  - Intent类型: trade, swap, exchange, data_access"
            echo "  - 发布间隔: ${AUTO_INTENT_PUBLISH_INTERVAL}秒"
            echo ""
            echo "API测试:"
            echo "  手动发布Intent: curl -X POST http://localhost:$HTTP_PORT/pinai_intent/intent/create \\"
            echo "                   -H 'Content-Type: application/json' \\"
            echo "                   -d '{\"type\":\"trade\",\"payload\":\"dGVzdA==\",\"sender_id\":\"auto-publisher\"}'"
            ;;
        "SERVICE_AGENT")
            echo "节点功能:"
            echo "  - 监听Intent广播消息"
            echo "  - 自动评估并提交出价"
            echo "  - Agent ID: $AGENT_ID"
            local strategy=$(eval echo \$AUTO_AGENT$(($NODE_ID-1))_STRATEGY)
            local margin=$(eval echo \$AUTO_AGENT$(($NODE_ID-1))_BID_MARGIN)
            echo "  - 出价策略: $strategy"
            echo "  - 利润率: $(echo "$margin * 100" | bc)%"
            echo ""
            echo "API测试:"
            echo "  Agent状态: curl http://localhost:$HTTP_PORT/pinai_intent/execution/agents/status"
            echo "  启动Agent: curl -X POST http://localhost:$HTTP_PORT/pinai_intent/execution/agents/$AGENT_ID/start"
            ;;
        "BLOCK_BUILDER")
            echo "节点功能:"
            echo "  - 收集Intent出价信息"
            echo "  - 执行匹配算法选择获胜者"
            echo "  - Builder ID: $BUILDER_ID"
            echo "  - 匹配算法: $AUTO_BUILDER_ALGORITHM"
            echo "  - 出价收集时间: ${AUTO_BUILDER_BID_COLLECTION_TIME}秒"
            echo ""
            echo "API测试:"
            echo "  Builder状态: curl http://localhost:$HTTP_PORT/pinai_intent/execution/builders/status"
            echo "  匹配历史: curl http://localhost:$HTTP_PORT/pinai_intent/execution/matches/history?limit=10"
            ;;
    esac
    
    echo ""
    echo "监控工具:"
    echo "  实时监控: ./scripts/automation/monitor_automation.sh"
    echo "  节点日志: tail -f $LOG_FILE"
    echo ""
    
    case $NODE_TYPE in
        "PUBLISHER")
            echo -e "${AUTO_COLOR_YELLOW}提示: 此节点将自动发布Intent消息${AUTO_COLOR_NC}"
            echo -e "${AUTO_COLOR_YELLOW}      其他节点启动后会自动响应出价${AUTO_COLOR_NC}"
            ;;
        "SERVICE_AGENT")
            echo -e "${AUTO_COLOR_YELLOW}提示: 此Agent会自动监听Intent并出价${AUTO_COLOR_NC}"
            echo -e "${AUTO_COLOR_YELLOW}      确保发布者节点(节点1)已启动${AUTO_COLOR_NC}"
            ;;
        "BLOCK_BUILDER")
            echo -e "${AUTO_COLOR_YELLOW}提示: 此Builder会自动匹配Intent和出价${AUTO_COLOR_NC}"
            echo -e "${AUTO_COLOR_YELLOW}      确保发布者和Agent节点已启动${AUTO_COLOR_NC}"
            ;;
    esac
    echo ""
}

# 实时显示日志
show_live_logs() {
    echo ""
    log_step "实时日志输出 (按 Ctrl+C 停止节点)"
    echo -e "${AUTO_COLOR_CYAN}================================================${AUTO_COLOR_NC}"
    
    # 设置信号处理
    trap 'handle_interrupt' INT TERM
    
    # 实时显示日志
    tail -f "$LOG_FILE" 2>/dev/null || {
        log_warn "无法读取日志文件，显示进程状态..."
        while true; do
            if [ -f "$PID_FILE" ]; then
                local pid=$(cat "$PID_FILE")
                if kill -0 "$pid" 2>/dev/null; then
                    echo "$(date '+%Y-%m-%d %H:%M:%S') [INFO] 节点$NODE_ID 运行中 (PID: $pid)"
                else
                    log_error "节点进程已停止"
                    break
                fi
            else
                log_error "PID 文件不存在"
                break
            fi
            sleep 5
        done
    }
}

# 处理中断信号
handle_interrupt() {
    echo ""
    log_step "接收到停止信号，正在关闭节点$NODE_ID..."
    
    # 停止节点
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "停止节点$NODE_ID (PID: $pid)"
            kill -TERM "$pid" 2>/dev/null || true
            
            # 等待优雅退出
            local wait_count=0
            while [ $wait_count -lt 10 ] && kill -0 "$pid" 2>/dev/null; do
                sleep 1
                wait_count=$((wait_count + 1))
            done
            
            # 强制停止
            if kill -0 "$pid" 2>/dev/null; then
                log_warn "强制停止节点$NODE_ID"
                kill -KILL "$pid" 2>/dev/null || true
            fi
        fi
        rm -f "$PID_FILE"
    fi
    
    # 更新状态
    echo "stopped" > "$STATUS_FILE"
    
    log_info "节点$NODE_ID 已停止"
    echo ""
    exit 0
}

# 主函数
main() {
    show_node_header
    check_prerequisites
    check_port_availability  
    check_node_status
    start_node
    show_node_info
    show_live_logs
}

# 执行主函数
main "$@"