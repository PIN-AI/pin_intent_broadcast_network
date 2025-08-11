#!/bin/bash

# PIN 自动化测试启动脚本
# 在单独的终端窗口中启动四个节点

set -e

# 加载自动化配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/automation_config.sh"

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

# 显示头部信息
show_header() {
    clear
    echo -e "${AUTO_COLOR_MAGENTA}================================================${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}    PIN 自动化测试系统启动器    ${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}================================================${AUTO_COLOR_NC}"
    echo ""
    echo -e "${AUTO_COLOR_CYAN}自动化测试架构:${AUTO_COLOR_NC}"
    echo -e "  节点1: ${AUTO_COLOR_GREEN}$AUTO_NODE1_NAME${AUTO_COLOR_NC} (端口: $AUTO_NODE1_HTTP_PORT)"
    echo -e "  节点2: ${AUTO_COLOR_YELLOW}$AUTO_NODE2_NAME${AUTO_COLOR_NC} (端口: $AUTO_NODE2_HTTP_PORT)"
    echo -e "  节点3: ${AUTO_COLOR_YELLOW}$AUTO_NODE3_NAME${AUTO_COLOR_NC} (端口: $AUTO_NODE3_HTTP_PORT)"
    echo -e "  节点4: ${AUTO_COLOR_MAGENTA}$AUTO_NODE4_NAME${AUTO_COLOR_NC} (端口: $AUTO_NODE4_HTTP_PORT)"
    echo ""
    echo -e "${AUTO_COLOR_CYAN}测试流程:${AUTO_COLOR_NC}"
    echo "  1. 节点1 发布Intent消息"
    echo "  2. 节点2&3 监听并自动出价"
    echo "  3. 节点4 收集出价并执行匹配"
    echo "  4. 循环进行自动化测试"
    echo ""
}

# 检查环境
check_environment() {
    log_step "检查自动化测试环境..."
    
    # 检查应用程序
    if [ ! -f "$AUTO_APP_BINARY" ]; then
        log_error "应用程序不存在: $AUTO_APP_BINARY"
        log_error "请先运行 'make build' 构建应用"
        exit 1
    fi
    
    # 检查配置文件
    for i in {1..4}; do
        local config_file=$(eval echo \$AUTO_NODE${i}_CONFIG_FILE)
        if [ ! -f "$config_file" ]; then
            log_error "配置文件不存在: $config_file"
            log_error "请先运行 './scripts/automation/setup_automation_env.sh' 初始化环境"
            exit 1
        fi
    done
    
    # 检查目录
    if [ ! -d "$AUTO_TEST_DATA_DIR" ]; then
        log_error "测试数据目录不存在: $AUTO_TEST_DATA_DIR"
        log_error "请先运行 './scripts/automation/setup_automation_env.sh' 初始化环境"
        exit 1
    fi
    
    log_info "环境检查通过"
}

# 检查端口可用性
check_ports() {
    log_step "检查端口可用性..."
    
    local all_ports=($AUTO_NODE1_HTTP_PORT $AUTO_NODE1_GRPC_PORT $AUTO_NODE1_P2P_PORT
                     $AUTO_NODE2_HTTP_PORT $AUTO_NODE2_GRPC_PORT $AUTO_NODE2_P2P_PORT
                     $AUTO_NODE3_HTTP_PORT $AUTO_NODE3_GRPC_PORT $AUTO_NODE3_P2P_PORT
                     $AUTO_NODE4_HTTP_PORT $AUTO_NODE4_GRPC_PORT $AUTO_NODE4_P2P_PORT)
    
    local conflicts=0
    for port in "${all_ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            local pid=$(lsof -Pi :$port -sTCP:LISTEN -t)
            log_error "端口 $port 已被占用，进程 PID: $pid"
            conflicts=$((conflicts + 1))
        fi
    done
    
    if [ $conflicts -gt 0 ]; then
        log_error "发现 $conflicts 个端口冲突"
        log_error "请使用 './scripts/automation/cleanup_automation.sh' 清理环境"
        exit 1
    fi
    
    log_info "所有端口可用"
}

# 检查终端模拟器
detect_terminal() {
    if command -v osascript >/dev/null 2>&1; then
        # macOS
        echo "macos"
    elif command -v gnome-terminal >/dev/null 2>&1; then
        # Linux GNOME
        echo "gnome"
    elif command -v konsole >/dev/null 2>&1; then
        # Linux KDE
        echo "kde"
    elif command -v xterm >/dev/null 2>&1; then
        # Generic X11
        echo "xterm"
    else
        echo "unknown"
    fi
}

# 在新终端窗口中启动节点
launch_node_in_terminal() {
    local node_id=$1
    local node_name=$(eval echo \$AUTO_NODE${node_id}_NAME)
    local terminal_type=$(detect_terminal)
    local start_script="$SCRIPT_DIR/start_node.sh"
    
    log_info "在新终端启动节点$node_id ($node_name)..."
    
    case $terminal_type in
        "macos")
            # macOS Terminal
            osascript -e "tell application \"Terminal\" to do script \"cd '$PWD' && '$start_script' $node_id\""
            ;;
        "gnome")
            # Linux GNOME Terminal
            gnome-terminal --title="PIN节点$node_id - $node_name" -- bash -c "cd '$PWD' && '$start_script' $node_id; exec bash"
            ;;
        "kde")
            # Linux KDE Konsole
            konsole --title="PIN节点$node_id - $node_name" -e bash -c "cd '$PWD' && '$start_script' $node_id; exec bash"
            ;;
        "xterm")
            # Generic xterm
            xterm -title "PIN节点$node_id - $node_name" -e bash -c "cd '$PWD' && '$start_script' $node_id; exec bash" &
            ;;
        *)
            log_warn "无法检测终端类型，请手动启动节点$node_id:"
            echo "  cd $PWD"
            echo "  $start_script $node_id"
            return 1
            ;;
    esac
    
    return 0
}

# 启动所有节点
launch_all_nodes() {
    log_step "启动所有自动化测试节点..."
    
    # 按顺序启动节点，每个节点间隔2秒
    for node_id in {1..4}; do
        local node_name=$(eval echo \$AUTO_NODE${node_id}_NAME)
        
        if launch_node_in_terminal $node_id; then
            log_info "已启动节点$node_id ($node_name)"
        else
            log_error "启动节点$node_id 失败"
        fi
        
        # 等待2秒让节点有时间启动
        if [ $node_id -lt 4 ]; then
            sleep 2
        fi
    done
}

# 等待节点启动完成
wait_for_nodes() {
    log_step "等待所有节点启动完成..."
    
    local max_wait=60
    local wait_count=0
    local ready_nodes=0
    
    while [ $wait_count -lt $max_wait ]; do
        ready_nodes=0
        
        for node_id in {1..4}; do
            local http_port=$(get_auto_node_http_port $node_id)
            if curl -s "http://localhost:$http_port/health" >/dev/null 2>&1; then
                ready_nodes=$((ready_nodes + 1))
            fi
        done
        
        echo -ne "\r等待节点启动: [$ready_nodes/4] "
        
        if [ $ready_nodes -eq 4 ]; then
            echo  # 新行
            log_info "所有节点已成功启动"
            break
        fi
        
        sleep 1
        wait_count=$((wait_count + 1))
    done
    
    if [ $ready_nodes -lt 4 ]; then
        echo  # 新行
        log_warn "只有 $ready_nodes/4 个节点启动成功"
        log_warn "请检查终端窗口中的错误信息"
    fi
}

# 启动自动Intent发布器
launch_intent_publisher() {
    log_step "启动自动Intent发布器..."
    
    local terminal_type=$(detect_terminal)
    local publisher_script="$SCRIPT_DIR/auto_intent_publisher.sh"
    
    # 检查脚本是否存在
    if [ ! -f "$publisher_script" ]; then
        log_error "自动Intent发布器脚本不存在: $publisher_script"
        return 1
    fi
    
    case $terminal_type in
        "macos")
            # macOS Terminal
            osascript -e "tell application \"Terminal\" to do script \"cd '$PWD' && '$publisher_script' --interval ${AUTO_INTENT_PUBLISH_INTERVAL:-30}\""
            ;;
        "gnome")
            # Linux GNOME Terminal
            gnome-terminal --title="PIN自动Intent发布器" -- bash -c "cd '$PWD' && '$publisher_script' --interval ${AUTO_INTENT_PUBLISH_INTERVAL:-30}; exec bash"
            ;;
        "kde")
            # Linux KDE Konsole
            konsole --title="PIN自动Intent发布器" -e bash -c "cd '$PWD' && '$publisher_script' --interval ${AUTO_INTENT_PUBLISH_INTERVAL:-30}; exec bash"
            ;;
        "xterm")
            # Generic xterm
            xterm -title "PIN自动Intent发布器" -e bash -c "cd '$PWD' && '$publisher_script' --interval ${AUTO_INTENT_PUBLISH_INTERVAL:-30}; exec bash" &
            ;;
        *)
            log_warn "无法检测终端类型，请手动启动Intent发布器:"
            echo "  cd $PWD"
            echo "  $publisher_script --interval ${AUTO_INTENT_PUBLISH_INTERVAL:-30}"
            return 1
            ;;
    esac
    
    log_info "自动Intent发布器已启动"
    return 0
}

# 显示测试状态
show_test_status() {
    echo ""
    log_step "自动化测试状态"
    echo ""
    
    for node_id in {1..4}; do
        local http_port=$(get_auto_node_http_port $node_id)
        local node_name=$(eval echo \$AUTO_NODE${node_id}_NAME)
        local node_type=$(eval echo \$AUTO_NODE${node_id}_TYPE)
        
        if curl -s "http://localhost:$http_port/health" >/dev/null 2>&1; then
            case $node_type in
                "PUBLISHER")
                    echo -e "${AUTO_COLOR_GREEN}✓${AUTO_COLOR_NC} 节点$node_id ($node_name) - 运行中 [Intent发布者]"
                    ;;
                "SERVICE_AGENT")
                    echo -e "${AUTO_COLOR_YELLOW}✓${AUTO_COLOR_NC} 节点$node_id ($node_name) - 运行中 [自动出价]"
                    ;;
                "BLOCK_BUILDER")
                    echo -e "${AUTO_COLOR_MAGENTA}✓${AUTO_COLOR_NC} 节点$node_id ($node_name) - 运行中 [匹配处理]"
                    ;;
            esac
        else
            echo -e "${AUTO_COLOR_RED}✗${AUTO_COLOR_NC} 节点$node_id ($node_name) - 未运行"
        fi
    done
}

# 启动监控终端
launch_monitor() {
    log_step "启动监控终端..."
    
    local terminal_type=$(detect_terminal)
    local monitor_script="$SCRIPT_DIR/monitor_automation.sh"
    
    case $terminal_type in
        "macos")
            osascript -e "tell application \"Terminal\" to do script \"cd '$PWD' && '$monitor_script'\""
            ;;
        "gnome")
            gnome-terminal --title="PIN自动化测试监控" -- bash -c "cd '$PWD' && '$monitor_script'; exec bash"
            ;;
        "kde")
            konsole --title="PIN自动化测试监控" -e bash -c "cd '$PWD' && '$monitor_script'; exec bash"
            ;;
        "xterm")
            xterm -title "PIN自动化测试监控" -e bash -c "cd '$PWD' && '$monitor_script'; exec bash" &
            ;;
        *)
            log_warn "请手动启动监控："
            echo "  $monitor_script"
            ;;
    esac
}

# 显示使用说明
show_usage() {
    echo ""
    log_step "自动化测试已启动"
    echo ""
    echo -e "${AUTO_COLOR_GREEN}✓ 所有节点已在独立终端中启动${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_GREEN}✓ 自动Intent发布器已启动${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_GREEN}✓ 监控终端已启动${AUTO_COLOR_NC}"
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}测试说明:${AUTO_COLOR_NC}"
    echo "1. 节点1 + Intent发布器会每${AUTO_INTENT_PUBLISH_INTERVAL:-30}秒自动发布Intent"
    echo "2. 节点2&3 会自动监听并出价"
    echo "3. 节点4 会自动收集出价并匹配"
    echo "4. 测试将持续运行，可手动停止"
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}终端窗口说明:${AUTO_COLOR_NC}"
    echo "  - PIN节点1: Intent发布者（只负责网络和API服务）"
    echo "  - PIN自动Intent发布器: 自动调用API发布Intent"
    echo "  - PIN节点2&3: Service Agent（自动出价）"
    echo "  - PIN节点4: Block Builder（自动匹配）"
    echo "  - PIN自动化测试监控: 实时监控系统状态"
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}监控命令:${AUTO_COLOR_NC}"
    echo "  实时监控: ./scripts/automation/monitor_automation.sh"
    echo "  检查状态: ./scripts/automation/check_automation_status.sh"
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}清理环境:${AUTO_COLOR_NC}"
    echo "  停止测试: ./scripts/automation/cleanup_automation.sh"
    echo ""
    echo -e "${AUTO_COLOR_CYAN}API端点 (用于手动测试):${AUTO_COLOR_NC}"
    for node_id in {1..4}; do
        local http_port=$(get_auto_node_http_port $node_id)
        local node_name=$(eval echo \$AUTO_NODE${node_id}_NAME)
        echo "  节点$node_id ($node_name): http://localhost:$http_port"
    done
    echo ""
}

# 主函数
main() {
    show_header
    check_environment
    check_ports
    launch_all_nodes
    
    echo ""
    log_step "等待节点启动..."
    sleep 5  # 给终端一些时间打开
    
    wait_for_nodes
    show_test_status
    
    # 启动自动Intent发布器（在Node1准备好后）
    if curl -s "http://localhost:$AUTO_NODE1_HTTP_PORT/health" >/dev/null 2>&1; then
        sleep 3  # 额外等待确保Node1完全准备好
        launch_intent_publisher
    else
        log_warn "Node1未就绪，跳过自动Intent发布器启动"
        log_warn "请手动启动: $SCRIPT_DIR/auto_intent_publisher.sh"
    fi
    
    # 启动监控终端
    sleep 2
    launch_monitor
    
    show_usage
}

# 处理参数
case "${1:-}" in
    "--help"|"-h")
        echo "PIN 自动化测试系统启动器"
        echo ""
        echo "用法: $0 [选项]"
        echo ""
        echo "选项:"
        echo "  --help, -h    显示帮助信息"
        echo ""
        echo "功能:"
        echo "  - 在独立终端中启动4个测试节点"
        echo "  - 1个Intent发布者节点"
        echo "  - 2个Service Agent出价节点"
        echo "  - 1个Block Builder匹配节点"
        echo ""
        echo "前置条件:"
        echo "  1. 运行 'make build' 构建应用程序"
        echo "  2. 运行 './scripts/automation/setup_automation_env.sh' 初始化环境"
        echo ""
        ;;
    *)
        main "$@"
        ;;
esac