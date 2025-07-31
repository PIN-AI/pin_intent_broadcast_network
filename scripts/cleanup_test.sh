#!/bin/bash

# 多节点测试环境清理脚本
# 用于停止所有节点进程并清理测试数据

set -e

echo "=== PIN 多节点测试环境清理 ==="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 停止节点进程
stop_node_processes() {
    log_step "停止节点进程..."
    
    local stopped_count=0
    
    # 查找并停止所有 pin_intent_broadcast_network 进程
    local pids=$(pgrep -f "pin_intent_broadcast_network" 2>/dev/null || true)
    
    if [ -n "$pids" ]; then
        log_info "发现运行中的节点进程: $pids"
        
        for pid in $pids; do
            log_info "停止进程 PID: $pid"
            kill -TERM $pid 2>/dev/null || true
            stopped_count=$((stopped_count + 1))
        done
        
        # 等待进程优雅退出
        sleep 3
        
        # 强制杀死仍在运行的进程
        local remaining_pids=$(pgrep -f "pin_intent_broadcast_network" 2>/dev/null || true)
        if [ -n "$remaining_pids" ]; then
            log_warn "强制停止剩余进程: $remaining_pids"
            for pid in $remaining_pids; do
                kill -KILL $pid 2>/dev/null || true
            done
        fi
        
        log_info "已停止 $stopped_count 个节点进程"
    else
        log_info "没有发现运行中的节点进程"
    fi
}

# 清理 PID 文件
cleanup_pid_files() {
    log_step "清理 PID 文件..."
    
    local cleaned_count=0
    
    for i in {1..3}; do
        local pid_file="test_data/node${i}/pid"
        if [ -f "$pid_file" ]; then
            rm -f "$pid_file"
            log_info "删除 PID 文件: $pid_file"
            cleaned_count=$((cleaned_count + 1))
        fi
    done
    
    if [ $cleaned_count -gt 0 ]; then
        log_info "已清理 $cleaned_count 个 PID 文件"
    else
        log_info "没有发现 PID 文件"
    fi
}

# 清理日志文件
cleanup_log_files() {
    log_step "清理日志文件..."
    
    local keep_logs=false
    
    # 检查是否保留日志
    if [ "$1" = "--keep-logs" ]; then
        keep_logs=true
        log_info "保留日志文件"
        return
    fi
    
    local cleaned_count=0
    
    # 清理节点日志
    for i in {1..3}; do
        local log_file="test_data/node${i}/output.log"
        if [ -f "$log_file" ]; then
            rm -f "$log_file"
            log_info "删除日志文件: $log_file"
            cleaned_count=$((cleaned_count + 1))
        fi
        
        # 清理日志目录中的其他日志
        if [ -d "test_data/node${i}/logs" ]; then
            rm -rf "test_data/node${i}/logs"/*
            log_info "清理节点${i}日志目录"
        fi
    done
    
    # 清理主日志文件
    if [ -f "server.log" ]; then
        rm -f "server.log"
        log_info "删除主日志文件: server.log"
        cleaned_count=$((cleaned_count + 1))
    fi
    
    if [ $cleaned_count -gt 0 ]; then
        log_info "已清理 $cleaned_count 个日志文件"
    else
        log_info "没有发现日志文件"
    fi
}

# 清理数据库文件
cleanup_database_files() {
    log_step "清理数据库文件..."
    
    local cleaned_count=0
    
    for i in {1..3}; do
        local db_file="test_data/node${i}/data.db"
        if [ -f "$db_file" ]; then
            rm -f "$db_file"
            log_info "删除数据库文件: $db_file"
            cleaned_count=$((cleaned_count + 1))
        fi
    done
    
    if [ $cleaned_count -gt 0 ]; then
        log_info "已清理 $cleaned_count 个数据库文件"
    else
        log_info "没有发现数据库文件"
    fi
}

# 清理 P2P 数据
cleanup_p2p_data() {
    log_step "清理 P2P 数据..."
    
    local cleaned_count=0
    
    for i in {1..3}; do
        local p2p_dir="test_data/node${i}/p2p"
        if [ -d "$p2p_dir" ] && [ "$(ls -A $p2p_dir 2>/dev/null)" ]; then
            rm -rf "$p2p_dir"/*
            log_info "清理节点${i} P2P 数据目录"
            cleaned_count=$((cleaned_count + 1))
        fi
    done
    
    if [ $cleaned_count -gt 0 ]; then
        log_info "已清理 $cleaned_count 个 P2P 数据目录"
    else
        log_info "没有发现 P2P 数据"
    fi
}

# 重置状态文件
reset_status_files() {
    log_step "重置状态文件..."
    
    if [ -d "test_data/status" ]; then
        for i in {1..3}; do
            echo "stopped" > "test_data/status/node${i}.status"
        done
        log_info "重置节点状态文件"
    fi
}

# 检查端口占用
check_port_usage() {
    log_step "检查端口占用情况..."
    
    local ports=(8000 8001 8002 9000 9001 9002 9001 9002 9003)
    local occupied_ports=()
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            occupied_ports+=($port)
        fi
    done
    
    if [ ${#occupied_ports[@]} -gt 0 ]; then
        log_warn "以下端口仍被占用: ${occupied_ports[*]}"
        log_warn "可能需要手动停止相关进程"
    else
        log_info "所有测试端口已释放"
    fi
}

# 显示清理摘要
show_cleanup_summary() {
    log_step "清理摘要"
    
    echo ""
    echo "已执行的清理操作:"
    echo "=================="
    echo "✓ 停止所有节点进程"
    echo "✓ 清理 PID 文件"
    if [ "$1" != "--keep-logs" ]; then
        echo "✓ 清理日志文件"
    else
        echo "- 保留日志文件"
    fi
    echo "✓ 清理数据库文件"
    echo "✓ 清理 P2P 数据"
    echo "✓ 重置状态文件"
    echo ""
    
    echo "保留的文件和目录:"
    echo "=================="
    echo "- 配置文件 (configs/test_node*.yaml)"
    echo "- 目录结构 (test_data/node*/)"
    echo "- 脚本文件 (scripts/)"
    if [ "$1" = "--keep-logs" ]; then
        echo "- 日志文件 (test_data/node*/output.log)"
    fi
    echo ""
}

# 显示使用帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --keep-logs    保留日志文件"
    echo "  --help         显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                # 完全清理"
    echo "  $0 --keep-logs    # 清理但保留日志"
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
        --keep-logs)
            log_info "启动清理程序 (保留日志模式)"
            ;;
        "")
            log_info "启动清理程序 (完全清理模式)"
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
    
    echo ""
    
    stop_node_processes
    cleanup_pid_files
    cleanup_log_files "$1"
    cleanup_database_files
    cleanup_p2p_data
    reset_status_files
    check_port_usage
    show_cleanup_summary "$1"
    
    log_info "测试环境清理完成！"
    log_info "现在可以重新运行 './scripts/setup_test_env.sh' 初始化环境"
    
    echo ""
}

# 执行主函数
main "$@"