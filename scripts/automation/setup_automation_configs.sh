#!/bin/bash

# PIN 自动化测试配置分发脚本
# 为不同节点分发正确的agents_config配置

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
    echo -e "${AUTO_COLOR_MAGENTA}================================================${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}    PIN 自动化测试配置分发    ${AUTO_COLOR_NC}"
    echo -e "${AUTO_COLOR_MAGENTA}================================================${AUTO_COLOR_NC}"
    echo ""
    echo "节点角色分配:"
    echo -e "  ${AUTO_COLOR_GREEN}Node1${AUTO_COLOR_NC}: Intent发布者 (不启动automation)"
    echo -e "  ${AUTO_COLOR_YELLOW}Node2${AUTO_COLOR_NC}: Service Agent 1 - 交易代理"
    echo -e "  ${AUTO_COLOR_YELLOW}Node3${AUTO_COLOR_NC}: Service Agent 2 - 数据代理"  
    echo -e "  ${AUTO_COLOR_MAGENTA}Node4${AUTO_COLOR_NC}: Block Builder - 匹配节点"
    echo ""
}

# 检查配置文件是否存在
check_config_files() {
    log_step "检查配置文件..."
    
    local missing=0
    local configs=(
        "configs/agents_config_node1.yaml"
        "configs/agents_config.yaml" 
        "configs/agents_config_node4.yaml"
    )
    
    for config in "${configs[@]}"; do
        if [ ! -f "$config" ]; then
            log_error "缺少配置文件: $config"
            missing=$((missing + 1))
        fi
    done
    
    if [ $missing -gt 0 ]; then
        log_error "请确保所有必需的配置文件都已创建"
        exit 1
    fi
    
    log_info "所有配置文件检查通过"
}

# 为特定节点设置agents_config
setup_node_config() {
    local node_id=$1
    local config_source=""
    local node_name=""
    
    case $node_id in
        1)
            config_source="configs/agents_config_node1.yaml"
            node_name="Intent发布者"
            ;;
        2|3)
            config_source="configs/agents_config.yaml.backup"
            node_name="Service Agent"
            ;;
        4) 
            config_source="configs/agents_config_node4.yaml"
            node_name="Block Builder"
            ;;
        *)
            log_error "无效的节点ID: $node_id"
            return 1
            ;;
    esac
    
    log_info "为Node$node_id ($node_name) 设置配置..."
    
    # 备份现有配置（如果存在）
    if [ -f "configs/agents_config.yaml.backup" ]; then
        log_warn "检测到现有备份，跳过备份步骤"
    else
        if [ -f "configs/agents_config.yaml" ]; then
            cp "configs/agents_config.yaml" "configs/agents_config.yaml.backup"
            log_info "已备份现有配置到 configs/agents_config.yaml.backup"
        fi
    fi
    
    # 复制对应的配置文件
    cp "$config_source" "configs/agents_config.yaml"
    log_info "已应用Node$node_id配置: $config_source -> configs/agents_config.yaml"
}

# 显示当前配置状态
show_current_config() {
    log_step "当前配置状态:"
    
    if [ -f "configs/agents_config.yaml" ]; then
        # 检查automation是否启用
        local automation_enabled=$(grep -A 2 "automation:" configs/agents_config.yaml | grep "enabled:" | awk '{print $2}')
        local agents_count=$(grep -c "agent_id:" configs/agents_config.yaml 2>/dev/null || echo "0")
        agents_count=$(echo "$agents_count" | tr -d '\n')
        local builders_enabled=$(grep -A 5 "builders:" configs/agents_config.yaml | grep "enabled:" | awk '{print $2}')
        
        echo "  配置文件: configs/agents_config.yaml"
        echo "  Automation启用: $automation_enabled"
        echo "  Service Agent数量: $agents_count"
        echo "  Block Builder启用: $builders_enabled"
        
        # 判断节点类型
        if [ "$automation_enabled" = "false" ]; then
            echo -e "  ${AUTO_COLOR_GREEN}-> 当前配置适用于: Node1 (Intent发布者)${AUTO_COLOR_NC}"
        elif [ "$automation_enabled" = "true" ] && [ "$agents_count" -gt "0" ] && [ "$builders_enabled" = "false" ]; then
            echo -e "  ${AUTO_COLOR_YELLOW}-> 当前配置适用于: Node2/Node3 (Service Agent)${AUTO_COLOR_NC}"
        elif [ "$automation_enabled" = "true" ] && [ "$builders_enabled" = "true" ]; then
            echo -e "  ${AUTO_COLOR_MAGENTA}-> 当前配置适用于: Node4 (Block Builder)${AUTO_COLOR_NC}"
        else
            echo -e "  ${AUTO_COLOR_RED}-> 配置状态未知或不完整${AUTO_COLOR_NC}"
        fi
    else
        log_warn "configs/agents_config.yaml 不存在"
    fi
    echo ""
}

# 恢复备份配置
restore_config() {
    log_step "恢复备份配置..."
    
    if [ -f "configs/agents_config.yaml.backup" ]; then
        mv "configs/agents_config.yaml.backup" "configs/agents_config.yaml"
        log_info "已恢复备份配置"
    else
        log_warn "未找到备份文件"
    fi
}

# 验证配置
validate_config() {
    local node_id=$1
    
    log_step "验证Node$node_id配置..."
    
    if [ ! -f "configs/agents_config.yaml" ]; then
        log_error "配置文件不存在: configs/agents_config.yaml"
        return 1
    fi
    
    # 基本语法检查 - 检查关键字段是否存在
    if ! grep -q "automation:" "configs/agents_config.yaml"; then
        log_error "配置文件缺少automation字段"
        return 1
    fi
    
    if ! grep -q "agents:" "configs/agents_config.yaml"; then
        log_error "配置文件缺少agents字段"  
        return 1
    fi
    
    if ! grep -q "builders:" "configs/agents_config.yaml"; then
        log_error "配置文件缺少builders字段"
        return 1
    fi
    
    log_info "配置验证通过"
}

# 显示使用帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  setup <node_id>    为指定节点设置配置 (1-4)"
    echo "  status            显示当前配置状态"
    echo "  restore           恢复备份配置"
    echo "  validate <node_id> 验证节点配置"
    echo "  help              显示此帮助信息"
    echo ""
    echo "节点说明:"
    echo "  1 - Intent发布者节点 (不启动automation)"
    echo "  2 - Service Agent 1 (交易代理)"
    echo "  3 - Service Agent 2 (数据代理)"
    echo "  4 - Block Builder (匹配节点)"
    echo ""
    echo "示例:"
    echo "  $0 setup 1      # 配置Node1为Intent发布者"
    echo "  $0 setup 2      # 配置Node2为Service Agent"
    echo "  $0 status       # 查看当前配置"
    echo "  $0 restore      # 恢复原始配置"
}

# 主函数
main() {
    local action="${1:-help}"
    
    case $action in
        setup)
            if [ -z "$2" ]; then
                log_error "请指定节点ID (1-4)"
                show_help
                exit 1
            fi
            show_header
            check_config_files
            setup_node_config "$2"
            validate_config "$2"
            show_current_config
            echo -e "${AUTO_COLOR_GREEN}Node$2 配置设置完成！${AUTO_COLOR_NC}"
            ;;
        status)
            show_header
            show_current_config
            ;;
        restore)
            show_header
            restore_config
            show_current_config
            ;;
        validate)
            if [ -z "$2" ]; then
                log_error "请指定节点ID (1-4)"
                exit 1
            fi
            validate_config "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知操作: $action"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"