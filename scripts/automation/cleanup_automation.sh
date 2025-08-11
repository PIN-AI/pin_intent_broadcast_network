#!/bin/bash

# PIN 自动化测试清理脚本
# 停止所有自动化测试节点并清理环境

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
    echo -e "${AUTO_COLOR_MAGENTA}    PIN 自动化测试环境清理    ${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}================================================${AUTO_COLOR_NC}"
    echo ""
}

# 停止单个节点
stop_node() {
    local node_id=$1
    local node_name=$(eval echo \$AUTO_NODE${node_id}_NAME)
    local pid_file=$(eval echo \$AUTO_NODE${node_id}_PID_FILE)
    local status_file=$(eval echo \$AUTO_NODE${node_id}_STATUS_FILE)
    
    log_step "停止节点$node_id ($node_name)..."
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        
        if kill -0 "$pid" 2>/dev/null; then
            log_info "发现运行中的进程 PID: $pid"
            
            # 发送TERM信号进行优雅停止
            log_info "发送优雅停止信号..."
            kill -TERM "$pid" 2>/dev/null || true
            
            # 等待最多10秒让进程优雅退出
            local wait_count=0
            while [ $wait_count -lt 10 ] && kill -0 "$pid" 2>/dev/null; do
                echo -n "."
                sleep 1
                wait_count=$((wait_count + 1))
            done
            
            # 检查进程是否还在运行
            if kill -0 "$pid" 2>/dev/null; then
                echo ""
                log_warn "进程未响应优雅停止信号，强制终止..."
                kill -KILL "$pid" 2>/dev/null || true
                sleep 1
            else
                echo ""
                log_info "进程已优雅停止"
            fi
            
            # 最终检查
            if kill -0 "$pid" 2>/dev/null; then
                log_error "无法停止进程 PID: $pid"
                return 1
            else
                log_info "节点$node_id 已成功停止"
            fi
        else
            log_warn "PID文件存在但进程不在运行"
        fi
        
        # 删除PID文件
        rm -f "$pid_file"
    else
        log_info "节点$node_id 未运行 (无PID文件)"
    fi
    
    # 更新状态文件
    if [ -f "$status_file" ]; then
        echo "stopped" > "$status_file"
    fi
    
    return 0
}

# 停止所有节点
stop_all_nodes() {
    log_step "停止所有自动化测试节点..."
    
    local stopped_count=0
    local failed_count=0
    
    # 按逆序停止节点 (先停止依赖的节点)
    for node_id in 4 3 2 1; do
        if stop_node $node_id; then
            stopped_count=$((stopped_count + 1))
        else
            failed_count=$((failed_count + 1))
        fi
        
        # 节点间停止间隔
        if [ $node_id -gt 1 ]; then
            sleep 1
        fi
    done
    
    echo ""
    log_info "停止节点统计: 成功 $stopped_count, 失败 $failed_count"
}

# 清理端口占用
cleanup_ports() {
    log_step "检查并清理端口占用..."
    
    local ports=($AUTO_NODE1_HTTP_PORT $AUTO_NODE1_GRPC_PORT $AUTO_NODE1_P2P_PORT
                 $AUTO_NODE2_HTTP_PORT $AUTO_NODE2_GRPC_PORT $AUTO_NODE2_P2P_PORT
                 $AUTO_NODE3_HTTP_PORT $AUTO_NODE3_GRPC_PORT $AUTO_NODE3_P2P_PORT
                 $AUTO_NODE4_HTTP_PORT $AUTO_NODE4_GRPC_PORT $AUTO_NODE4_P2P_PORT)
    
    local occupied_ports=()
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            local pid=$(lsof -Pi :$port -sTCP:LISTEN -t)
            log_warn "端口 $port 仍被占用，进程 PID: $pid"
            occupied_ports+=("$port:$pid")
        fi
    done
    
    if [ ${#occupied_ports[@]} -gt 0 ]; then
        log_warn "发现 ${#occupied_ports[@]} 个端口仍被占用"
        echo "占用详情:"
        for port_info in "${occupied_ports[@]}"; do
            local port=${port_info%:*}
            local pid=${port_info#*:}
            echo "  端口 $port: PID $pid"
        done
        
        echo ""
        read -p "是否强制终止这些进程? (y/N): " -n 1 -r
        echo ""
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            for port_info in "${occupied_ports[@]}"; do
                local pid=${port_info#*:}
                log_info "终止进程 PID: $pid"
                kill -KILL "$pid" 2>/dev/null || true
            done
            
            # 再次检查
            sleep 1
            local remaining_ports=0
            for port in "${ports[@]}"; do
                if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
                    remaining_ports=$((remaining_ports + 1))
                fi
            done
            
            if [ $remaining_ports -eq 0 ]; then
                log_info "所有端口已释放"
            else
                log_warn "仍有 $remaining_ports 个端口被占用"
            fi
        fi
    else
        log_info "所有自动化测试端口已释放"
    fi
}

# 清理文件和目录
cleanup_files() {
    log_step "清理临时文件和状态..."
    
    # 清理PID文件
    local pid_files=($AUTO_NODE1_PID_FILE $AUTO_NODE2_PID_FILE $AUTO_NODE3_PID_FILE $AUTO_NODE4_PID_FILE)
    for pid_file in "${pid_files[@]}"; do
        if [ -f "$pid_file" ]; then
            rm -f "$pid_file"
            log_info "删除PID文件: $(basename $pid_file)"
        fi
    done
    
    # 清理状态文件
    if [ -d "$AUTO_STATUS_DIR" ]; then
        rm -f "$AUTO_STATUS_DIR"/*.status
        log_info "清理状态文件"
    fi
    
    # 清理日志链接 (保留实际日志文件)
    if [ -d "$AUTO_LOGS_DIR" ]; then
        for i in {1..4}; do
            local log_link="$AUTO_LOGS_DIR/node${i}.log"
            if [ -L "$log_link" ]; then
                rm -f "$log_link"
            fi
        done
        log_info "清理日志链接"
    fi
}

# 显示清理选项
show_cleanup_options() {
    echo ""
    log_step "清理选项"
    echo ""
    echo "选择要执行的清理操作:"
    echo "  1. 仅停止进程 (保留配置和日志)"
    echo "  2. 停止进程 + 清理临时文件"
    echo "  3. 完全清理 (包括配置文件和日志)"
    echo "  4. 取消清理"
    echo ""
    
    while true; do
        read -p "请选择 [1-4]: " -n 1 -r choice
        echo ""
        
        case $choice in
            1)
                return 1
                ;;
            2)
                return 2
                ;;
            3)
                return 3
                ;;
            4)
                log_info "清理已取消"
                exit 0
                ;;
            *)
                echo "无效选择，请输入 1-4"
                ;;
        esac
    done
}

# 完全清理环境
full_cleanup() {
    log_step "执行完全清理..."
    
    # 删除配置文件
    local config_files=($AUTO_NODE1_CONFIG_FILE $AUTO_NODE2_CONFIG_FILE $AUTO_NODE3_CONFIG_FILE $AUTO_NODE4_CONFIG_FILE)
    for config_file in "${config_files[@]}"; do
        if [ -f "$config_file" ]; then
            rm -f "$config_file"
            log_info "删除配置文件: $(basename $config_file)"
        fi
    done
    
    # 删除测试数据目录
    if [ -d "$AUTO_TEST_DATA_DIR" ]; then
        echo ""
        log_warn "即将删除整个自动化测试数据目录: $AUTO_TEST_DATA_DIR"
        read -p "确定要继续吗? 这将删除所有日志和数据 (y/N): " -n 1 -r
        echo ""
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$AUTO_TEST_DATA_DIR"
            log_info "删除测试数据目录: $AUTO_TEST_DATA_DIR"
        else
            log_info "保留测试数据目录"
        fi
    fi
}

# 验证清理结果
verify_cleanup() {
    log_step "验证清理结果..."
    
    local issues=0
    
    # 检查进程
    for node_id in {1..4}; do
        local pid_file=$(eval echo \$AUTO_NODE${node_id}_PID_FILE)
        if [ -f "$pid_file" ]; then
            log_warn "PID文件仍存在: $pid_file"
            issues=$((issues + 1))
        fi
    done
    
    # 检查端口
    local ports=($AUTO_NODE1_HTTP_PORT $AUTO_NODE1_GRPC_PORT $AUTO_NODE1_P2P_PORT
                 $AUTO_NODE2_HTTP_PORT $AUTO_NODE2_GRPC_PORT $AUTO_NODE2_P2P_PORT
                 $AUTO_NODE3_HTTP_PORT $AUTO_NODE3_GRPC_PORT $AUTO_NODE3_P2P_PORT
                 $AUTO_NODE4_HTTP_PORT $AUTO_NODE4_GRPC_PORT $AUTO_NODE4_P2P_PORT)
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_warn "端口仍被占用: $port"
            issues=$((issues + 1))
        fi
    done
    
    if [ $issues -eq 0 ]; then
        log_info "清理验证通过，无遗留问题"
    else
        log_warn "发现 $issues 个清理问题"
    fi
    
    return $issues
}

# 显示清理摘要
show_cleanup_summary() {
    local cleanup_level=$1
    
    echo ""
    log_step "清理完成"
    echo ""
    
    case $cleanup_level in
        1)
            echo -e "${AUTO_COLOR_GREEN}✓ 所有自动化测试进程已停止${AUTO_COLOR_NC}"
            echo -e "  配置文件和日志已保留"
            ;;
        2)
            echo -e "${AUTO_COLOR_GREEN}✓ 进程已停止，临时文件已清理${AUTO_COLOR_NC}"
            echo -e "  配置文件和日志已保留"
            ;;
        3)
            echo -e "${AUTO_COLOR_GREEN}✓ 完全清理已完成${AUTO_COLOR_NC}"
            echo -e "  配置文件、日志和数据目录已删除"
            ;;
    esac
    
    echo ""
    echo -e "${AUTO_COLOR_YELLOW}重新启动自动化测试:${AUTO_COLOR_NC}"
    if [ $cleanup_level -eq 3 ]; then
        echo "  1. ./scripts/automation/setup_automation_env.sh"
        echo "  2. ./scripts/automation/start_automation_test.sh"
    else
        echo "  ./scripts/automation/start_automation_test.sh"
    fi
    echo ""
}

# 检查运行状态
check_running_status() {
    log_step "检查自动化测试运行状态..."
    
    local running_nodes=0
    for node_id in {1..4}; do
        local pid_file=$(eval echo \$AUTO_NODE${node_id}_PID_FILE)
        if [ -f "$pid_file" ]; then
            local pid=$(cat "$pid_file")
            if kill -0 "$pid" 2>/dev/null; then
                running_nodes=$((running_nodes + 1))
                local node_name=$(eval echo \$AUTO_NODE${node_id}_NAME)
                log_info "节点$node_id ($node_name) 正在运行，PID: $pid"
            fi
        fi
    done
    
    if [ $running_nodes -eq 0 ]; then
        log_info "没有发现运行中的自动化测试节点"
        return 1
    fi
    
    log_warn "发现 $running_nodes 个运行中的节点"
    return 0
}

# 主函数
main() {
    show_header
    
    # 检查运行状态
    if ! check_running_status; then
        echo ""
        read -p "没有运行中的节点，是否继续清理文件? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "清理已取消"
            exit 0
        fi
    fi
    
    # 显示清理选项
    cleanup_level=$(show_cleanup_options)
    
    # 执行清理
    stop_all_nodes
    cleanup_ports
    
    if [ $cleanup_level -ge 2 ]; then
        cleanup_files
    fi
    
    if [ $cleanup_level -eq 3 ]; then
        full_cleanup
    fi
    
    # 验证清理结果
    verify_cleanup
    
    # 显示清理摘要
    show_cleanup_summary $cleanup_level
}

# 处理参数
case "${1:-}" in
    "--help"|"-h")
        echo "PIN 自动化测试环境清理工具"
        echo ""
        echo "用法: $0 [选项]"
        echo ""
        echo "选项:"
        echo "  --help, -h      显示帮助信息"
        echo "  --force         强制清理，不询问确认"
        echo "  --full          执行完全清理 (包括配置和日志)"
        echo ""
        echo "功能:"
        echo "  - 停止所有自动化测试节点进程"
        echo "  - 清理端口占用"
        echo "  - 删除临时文件和状态"
        echo "  - 可选择性删除配置文件和日志"
        echo ""
        ;;
    "--force")
        # 强制清理模式
        show_header
        check_running_status || true
        stop_all_nodes
        cleanup_ports
        cleanup_files
        verify_cleanup
        show_cleanup_summary 2
        ;;
    "--full")
        # 完全清理模式
        show_header
        check_running_status || true
        stop_all_nodes
        cleanup_ports
        cleanup_files
        full_cleanup
        verify_cleanup
        show_cleanup_summary 3
        ;;
    *)
        main "$@"
        ;;
esac