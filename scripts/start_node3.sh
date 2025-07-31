#!/bin/bash

# 节点3启动脚本 - 监控节点
# 在独立终端中启动节点3，用于监控 Intent

set -e

# 加载统一配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh"

# 节点特定配置
NODE_ID=3
NODE_NAME="监控节点"
CONFIG_FILE="$NODE3_CONFIG_FILE"
HTTP_PORT="$NODE3_HTTP_PORT"
GRPC_PORT="$NODE3_GRPC_PORT"
P2P_PORT="$NODE3_P2P_PORT"
LOG_FILE="$NODE3_LOG_FILE"
PID_FILE="$NODE3_PID_FILE"
STATUS_FILE="$NODE3_STATUS_FILE"

# 颜色定义（使用统一配置）
RED="$COLOR_RED"
GREEN="$COLOR_GREEN"
YELLOW="$COLOR_YELLOW"
BLUE="$COLOR_BLUE"
CYAN="$COLOR_CYAN"
MAGENTA="$COLOR_MAGENTA"
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

log_node() {
    echo -e "${CYAN}[NODE3]${NC} $1"
}

# 显示节点信息头部
show_node_header() {
    clear
    echo -e "${MAGENTA}================================${NC}"
    echo -e "${MAGENTA}    PIN 意图广播网络 - 节点3    ${NC}"
    echo -e "${MAGENTA}         ${NODE_NAME}         ${NC}"
    echo -e "${MAGENTA}================================${NC}"
    echo ""
    echo -e "${CYAN}节点信息:${NC}"
    echo "  节点ID: $NODE_ID"
    echo "  节点类型: $NODE_NAME"
    echo "  HTTP端口: $HTTP_PORT"
    echo "  gRPC端口: $GRPC_PORT"
    echo "  P2P端口: $P2P_PORT"
    echo "  配置文件: $CONFIG_FILE"
    echo "  日志文件: $LOG_FILE"
    echo ""
}

# 检查前置条件
check_prerequisites() {
    log_step "检查启动前置条件..."
    
    # 检查应用是否存在
    if [ ! -f "$APP_BINARY" ]; then
        log_error "应用程序不存在: $APP_BINARY"
        log_error "请先运行 'make build' 构建应用"
        exit 1
    fi
    
    # 检查配置文件是否存在
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "配置文件不存在: $CONFIG_FILE"
        log_error "请先运行 './scripts/setup_test_env.sh' 初始化环境"
        exit 1
    fi
    
    # 检查测试数据目录
    if [ ! -d "$NODE3_DATA_DIR" ]; then
        log_error "节点数据目录不存在: $NODE3_DATA_DIR"
        log_error "请先运行 './scripts/setup_test_env.sh' 初始化环境"
        exit 1
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
        log_error "请使用 './scripts/cleanup_test.sh' 清理环境或手动停止占用进程"
        exit 1
    fi
    
    log_info "所有端口可用"
}

# 检查节点是否已在运行
check_node_status() {
    log_step "检查节点状态..."
    
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_error "节点3已在运行，PID: $pid"
            log_error "请先停止现有节点或使用 './scripts/cleanup_test.sh' 清理环境"
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
    log_step "启动节点3..."
    
    # 更新状态文件
    echo "starting" > "$STATUS_FILE"
    
    # 启动节点进程
    log_info "执行命令: $APP_BINARY -conf $CONFIG_FILE"
    
    # 启动节点并获取 PID
    $APP_BINARY -conf "$CONFIG_FILE" > "$LOG_FILE" 2>&1 &
    local pid=$!
    
    # 保存 PID
    echo $pid > "$PID_FILE"
    
    log_info "节点3已启动，PID: $pid"
    
    # 等待节点启动
    log_step "等待节点启动..."
    local max_wait=30
    local wait_count=0
    
    while [ $wait_count -lt $max_wait ]; do
        if curl -s "http://localhost:$HTTP_PORT/health" >/dev/null 2>&1; then
            log_info "节点3启动成功！"
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
        echo -e "${RED}=== 错误日志 (最后20行) ===${NC}"
        tail -20 "$LOG_FILE"
        echo -e "${RED}=========================${NC}"
    else
        log_error "日志文件不存在: $LOG_FILE"
    fi
    echo ""
    log_error "可能的解决方案:"
    echo "1. 检查配置文件是否正确: $CONFIG_FILE"
    echo "2. 检查端口是否被占用"
    echo "3. 检查应用程序是否正确构建"
    echo "4. 运行 './scripts/cleanup_test.sh' 清理环境后重试"
}

# 显示节点运行信息
show_node_info() {
    echo ""
    log_step "节点运行信息"
    echo ""
    echo -e "${GREEN}✓ 节点3 ($NODE_NAME) 运行中${NC}"
    echo ""
    echo "服务端点:"
    echo "  HTTP API: http://localhost:$HTTP_PORT"
    echo "  gRPC API: localhost:$GRPC_PORT"
    echo "  P2P 地址: /ip4/127.0.0.1/tcp/$P2P_PORT"
    echo ""
    echo "API 测试:"
    echo "  健康检查: curl http://localhost:$HTTP_PORT/health"
    echo "  查询Intent: curl http://localhost:$HTTP_PORT/pinai_intent/intent/list"
    echo ""
    echo "工具脚本:"
    echo "  Intent监控器: ./scripts/intent_monitor.sh $HTTP_PORT"
    echo "  网络状态: ./scripts/network_status.sh $HTTP_PORT"
    echo ""
    echo -e "${YELLOW}提示: 此节点将监控来自其他节点的 Intent 广播${NC}"
    echo -e "${YELLOW}      启动节点1后可以观察到 Intent 的接收情况${NC}"
    echo ""
}

# 启动 Intent 监控器
start_intent_monitor() {
    log_step "启动 Intent 监控器..."
    
    # 等待一段时间确保节点完全启动
    sleep 3
    
    # 检查 intent_monitor.sh 是否存在
    if [ -f "./scripts/intent_monitor.sh" ]; then
        log_info "在后台启动 Intent 监控器..."
        ./scripts/intent_monitor.sh $HTTP_PORT &
        local monitor_pid=$!
        echo $monitor_pid > "$NODE3_DATA_DIR/monitor.pid"
        log_info "Intent 监控器已启动，PID: $monitor_pid"
    else
        log_warn "Intent 监控器脚本不存在，将显示基本日志"
        log_warn "请运行完整的任务实现后再使用监控功能"
    fi
}

# 实时显示日志
show_live_logs() {
    echo ""
    log_step "实时日志输出 (按 Ctrl+C 停止节点)"
    echo -e "${CYAN}================================${NC}"
    
    # 设置信号处理
    trap 'handle_interrupt' INT TERM
    
    # 实时显示日志
    tail -f "$LOG_FILE" 2>/dev/null || {
        log_warn "无法读取日志文件，显示进程状态..."
        while true; do
            if [ -f "$PID_FILE" ]; then
                local pid=$(cat "$PID_FILE")
                if kill -0 "$pid" 2>/dev/null; then
                    echo "$(date '+%Y-%m-%d %H:%M:%S') [INFO] 节点3运行中 (PID: $pid)"
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
    log_step "接收到停止信号，正在关闭节点3..."
    
    # 停止 Intent 监控器
    if [ -f "$NODE3_DATA_DIR/monitor.pid" ]; then
        local monitor_pid=$(cat "$NODE3_DATA_DIR/monitor.pid")
        if kill -0 "$monitor_pid" 2>/dev/null; then
            log_info "停止 Intent 监控器 (PID: $monitor_pid)"
            kill -TERM "$monitor_pid" 2>/dev/null || true
        fi
        rm -f "$NODE3_DATA_DIR/monitor.pid"
    fi
    
    # 停止节点
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "停止节点3 (PID: $pid)"
            kill -TERM "$pid" 2>/dev/null || true
            
            # 等待优雅退出
            local wait_count=0
            while [ $wait_count -lt 10 ] && kill -0 "$pid" 2>/dev/null; do
                sleep 1
                wait_count=$((wait_count + 1))
            done
            
            # 强制停止
            if kill -0 "$pid" 2>/dev/null; then
                log_warn "强制停止节点3"
                kill -KILL "$pid" 2>/dev/null || true
            fi
        fi
        rm -f "$PID_FILE"
    fi
    
    # 更新状态
    echo "stopped" > "$STATUS_FILE"
    
    log_info "节点3已停止"
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
    start_intent_monitor
    show_live_logs
}

# 执行主函数
main "$@"