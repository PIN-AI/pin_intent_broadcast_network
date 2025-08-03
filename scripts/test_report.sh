#!/bin/bash

# 测试结果验证和报告生成脚本
# 验证 Intent 广播功能并生成详细的测试报告

set -e

# 配置参数
NODE_COUNT=3
BASE_HTTP_PORT=8000
REPORT_FILE="test_data/test_report_$(date +%Y%m%d_%H%M%S).md"
TEST_DURATION=${1:-60}  # 默认测试60秒

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# 测试数据收集
declare -A NODE_INTENTS 2>/dev/null || true
declare -A NODE_TYPES 2>/dev/null || true
declare -A INTENT_TIMESTAMPS 2>/dev/null || true
TOTAL_PUBLISHED=0
TOTAL_RECEIVED=0
TEST_START_TIME=$(date +%s)
TEST_END_TIME=0

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

log_test() {
    echo -e "${CYAN}[TEST]${NC} $1"
}

# 显示测试头部
show_test_header() {
    clear
    echo -e "${MAGENTA}================================================================${NC}"
    echo -e "${MAGENTA}                PIN 多节点测试结果验证工具                      ${NC}"
    echo -e "${MAGENTA}================================================================${NC}"
    echo ""
    echo -e "${CYAN}测试配置:${NC}"
    echo "  测试时长: $TEST_DURATION 秒"
    echo "  节点数量: $NODE_COUNT"
    echo "  报告文件: $REPORT_FILE"
    echo "  开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""
}

# 检查所有节点状态
check_all_nodes() {
    log_step "检查所有节点状态..."
    
    local running_nodes=0
    local node_status=""
    
    for i in $(seq 1 $NODE_COUNT); do
        local http_port=$((BASE_HTTP_PORT + i - 1))
        local pid_file="test_data/node${i}/pid"
        
        if [ -f "$pid_file" ]; then
            local pid=$(cat "$pid_file")
            if kill -0 "$pid" 2>/dev/null; then
                if curl -s --max-time 3 "http://localhost:$http_port/health" >/dev/null 2>&1; then
                    log_info "节点${i} (端口$http_port): 运行正常"
                    running_nodes=$((running_nodes + 1))
                    node_status="${node_status}节点${i}:运行 "
                else
                    log_warn "节点${i} (端口$http_port): 服务异常"
                    node_status="${node_status}节点${i}:异常 "
                fi
            else
                log_error "节点${i} (端口$http_port): 进程不存在"
                node_status="${node_status}节点${i}:停止 "
            fi
        else
            log_error "节点${i} (端口$http_port): 未启动"
            node_status="${node_status}节点${i}:未启动 "
        fi
    done
    
    echo ""
    if [ $running_nodes -eq $NODE_COUNT ]; then
        log_info "所有 $NODE_COUNT 个节点运行正常"
        return 0
    else
        log_error "只有 $running_nodes/$NODE_COUNT 个节点运行正常"
        log_error "请确保所有节点都已启动并正常运行"
        return 1
    fi
}

# 收集节点 Intent 数据
collect_node_intents() {
    local node_id=$1
    local http_port=$((BASE_HTTP_PORT + node_id - 1))
    
    log_test "收集节点${node_id}的 Intent 数据..."
    
    local response=$(curl -s --max-time 10 "http://localhost:$http_port/pinai_intent/intent/list?limit=100" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        # 解析 Intent 数据
        local intent_data=$(echo "$response" | grep -o '"intents":\[.*\]' | sed 's/"intents":\[//' | sed 's/\]$//' | tr '},{' '\n')
        local intent_count=0
        
        while IFS= read -r intent_line; do
            if [ -n "$intent_line" ]; then
                intent_line=$(echo "$intent_line" | sed 's/^{//' | sed 's/}$//')
                
                local id=$(echo "$intent_line" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
                local type=$(echo "$intent_line" | grep -o '"type":"[^"]*"' | cut -d'"' -f4)
                local sender_id=$(echo "$intent_line" | grep -o '"sender_id":"[^"]*"' | cut -d'"' -f4)
                local timestamp=$(echo "$intent_line" | grep -o '"timestamp":[0-9]*' | cut -d':' -f2)
                
                if [ -n "$id" ] && [ -n "$type" ]; then
                    NODE_INTENTS["${node_id}_${id}"]="$type|$sender_id|$timestamp"
                    NODE_TYPES["${node_id}_${type}"]=$((${NODE_TYPES["${node_id}_${type}"]:-0} + 1))
                    INTENT_TIMESTAMPS["$id"]="$timestamp"
                    intent_count=$((intent_count + 1))
                fi
            fi
        done <<< "$intent_data"
        
        log_info "节点${node_id}: 收集到 $intent_count 个 Intent"
        return $intent_count
    else
        log_warn "节点${node_id}: 无法获取 Intent 数据"
        return 0
    fi
}

# 执行测试数据收集
perform_test_collection() {
    log_step "开始测试数据收集 (持续 $TEST_DURATION 秒)..."
    echo ""
    
    local collection_start=$(date +%s)
    local collection_end=$((collection_start + TEST_DURATION))
    local last_collection=0
    
    # 显示进度条
    while [ $(date +%s) -lt $collection_end ]; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - collection_start))
        local remaining=$((TEST_DURATION - elapsed))
        local progress=$((elapsed * 100 / TEST_DURATION))
        
        # 每10秒收集一次数据
        if [ $((current_time - last_collection)) -ge 10 ]; then
            echo ""
            log_test "收集进度: ${progress}% (剩余 ${remaining}秒)"
            
            # 收集所有节点数据
            for i in $(seq 1 $NODE_COUNT); do
                collect_node_intents $i >/dev/null
            done
            
            last_collection=$current_time
        fi
        
        # 显示进度条
        local bar_length=50
        local filled_length=$((progress * bar_length / 100))
        local bar=""
        
        for ((j=0; j<filled_length; j++)); do
            bar="${bar}█"
        done
        for ((j=filled_length; j<bar_length; j++)); do
            bar="${bar}░"
        done
        
        printf "\r${CYAN}[进度]${NC} [%s] %d%% (%ds/%ds)" "$bar" "$progress" "$elapsed" "$TEST_DURATION"
        
        sleep 1
    done
    
    echo ""
    echo ""
    TEST_END_TIME=$(date +%s)
    
    # 最终数据收集
    log_step "执行最终数据收集..."
    for i in $(seq 1 $NODE_COUNT); do
        collect_node_intents $i
    done
    
    log_info "测试数据收集完成"
}

# 分析测试结果
analyze_test_results() {
    log_step "分析测试结果..."
    echo ""
    
    # 统计每个节点的 Intent 数量
    local node_intent_counts=()
    local total_intents=0
    
    for i in $(seq 1 $NODE_COUNT); do
        local count=0
        for key in "${!NODE_INTENTS[@]}"; do
            if [[ $key == "${i}_"* ]]; then
                count=$((count + 1))
            fi
        done
        node_intent_counts[$i]=$count
        total_intents=$((total_intents + count))
        log_info "节点${i}: $count 个 Intent"
    done
    
    # 分析 Intent 类型分布
    declare -A global_type_counts
    for key in "${!NODE_TYPES[@]}"; do
        local type=$(echo "$key" | cut -d'_' -f2)
        global_type_counts["$type"]=$((${global_type_counts["$type"]:-0} + ${NODE_TYPES["$key"]}))
    done
    
    # 检查数据一致性
    log_step "检查数据一致性..."
    
    local consistency_issues=0
    local unique_intents=()
    
    # 收集所有唯一的 Intent ID
    for key in "${!NODE_INTENTS[@]}"; do
        local intent_id=$(echo "$key" | cut -d'_' -f2-)
        if [[ ! " ${unique_intents[@]} " =~ " ${intent_id} " ]]; then
            unique_intents+=("$intent_id")
        fi
    done
    
    # 检查每个 Intent 是否在所有节点上都存在
    local broadcast_success=0
    local broadcast_total=${#unique_intents[@]}
    
    for intent_id in "${unique_intents[@]}"; do
        local found_nodes=0
        for i in $(seq 1 $NODE_COUNT); do
            if [[ -n "${NODE_INTENTS["${i}_${intent_id}"]}" ]]; then
                found_nodes=$((found_nodes + 1))
            fi
        done
        
        if [ $found_nodes -eq $NODE_COUNT ]; then
            broadcast_success=$((broadcast_success + 1))
        else
            consistency_issues=$((consistency_issues + 1))
            log_warn "Intent $intent_id 只在 $found_nodes/$NODE_COUNT 个节点上找到"
        fi
    done
    
    # 计算成功率
    local success_rate=0
    if [ $broadcast_total -gt 0 ]; then
        success_rate=$((broadcast_success * 100 / broadcast_total))
    fi
    
    echo ""
    log_info "分析结果:"
    log_info "  唯一 Intent 总数: $broadcast_total"
    log_info "  成功广播数量: $broadcast_success"
    log_info "  广播成功率: ${success_rate}%"
    log_info "  一致性问题: $consistency_issues"
    
    # 返回分析结果
    echo "$broadcast_total|$broadcast_success|$success_rate|$consistency_issues"
}

# 生成测试报告
generate_test_report() {
    local analysis_result="$1"
    
    log_step "生成测试报告: $REPORT_FILE"
    
    # 解析分析结果
    local broadcast_total=$(echo "$analysis_result" | cut -d'|' -f1)
    local broadcast_success=$(echo "$analysis_result" | cut -d'|' -f2)
    local success_rate=$(echo "$analysis_result" | cut -d'|' -f3)
    local consistency_issues=$(echo "$analysis_result" | cut -d'|' -f4)
    
    # 创建报告目录
    mkdir -p "$(dirname "$REPORT_FILE")"
    
    # 生成 Markdown 报告
    cat > "$REPORT_FILE" << EOF
# PIN 多节点测试报告

## 测试概览

- **测试时间**: $(if [[ "$OSTYPE" == "darwin"* ]]; then date -r "$TEST_START_TIME" '+%Y-%m-%d %H:%M:%S'; else date -d "@$TEST_START_TIME" '+%Y-%m-%d %H:%M:%S'; fi) - $(if [[ "$OSTYPE" == "darwin"* ]]; then date -r "$TEST_END_TIME" '+%Y-%m-%d %H:%M:%S'; else date -d "@$TEST_END_TIME" '+%Y-%m-%d %H:%M:%S'; fi)
- **测试时长**: $((TEST_END_TIME - TEST_START_TIME)) 秒
- **节点数量**: $NODE_COUNT
- **报告生成时间**: $(date '+%Y-%m-%d %H:%M:%S')

## 测试结果摘要

| 指标 | 数值 | 状态 |
|------|------|------|
| 唯一 Intent 总数 | $broadcast_total | ✓ |
| 成功广播数量 | $broadcast_success | $([ $broadcast_success -eq $broadcast_total ] && echo "✓" || echo "⚠") |
| 广播成功率 | ${success_rate}% | $([ $success_rate -ge 90 ] && echo "✓" || echo "⚠") |
| 一致性问题 | $consistency_issues | $([ $consistency_issues -eq 0 ] && echo "✓" || echo "⚠") |

## 节点状态

EOF

    # 添加节点状态信息
    for i in $(seq 1 $NODE_COUNT); do
        local http_port=$((BASE_HTTP_PORT + i - 1))
        local intent_count=0
        
        for key in "${!NODE_INTENTS[@]}"; do
            if [[ $key == "${i}_"* ]]; then
                intent_count=$((intent_count + 1))
            fi
        done
        
        local node_type="监控节点"
        if [ $i -eq 1 ]; then
            node_type="发布节点"
        fi
        
        cat >> "$REPORT_FILE" << EOF
### 节点 $i ($node_type)

- **HTTP 端口**: $http_port
- **Intent 数量**: $intent_count
- **状态**: $(curl -s --max-time 3 "http://localhost:$http_port/health" >/dev/null 2>&1 && echo "运行正常" || echo "异常")

EOF
    done
    
    # 添加 Intent 类型分布
    cat >> "$REPORT_FILE" << EOF
## Intent 类型分布

| 类型 | 数量 | 百分比 |
|------|------|--------|
EOF
    
    local total_type_count=0
    for type in "${!global_type_counts[@]}"; do
        total_type_count=$((total_type_count + ${global_type_counts["$type"]}))
    done
    
    for type in "${!global_type_counts[@]}"; do
        local count=${global_type_counts["$type"]}
        local percentage=0
        if [ $total_type_count -gt 0 ]; then
            percentage=$((count * 100 / total_type_count))
        fi
        echo "| $type | $count | ${percentage}% |" >> "$REPORT_FILE"
    done
    
    # 添加详细分析
    cat >> "$REPORT_FILE" << EOF

## 详细分析

### 广播功能验证

EOF
    
    if [ $success_rate -ge 95 ]; then
        echo "✅ **广播功能正常**: 成功率达到 ${success_rate}%，表明 P2P 网络广播功能工作正常。" >> "$REPORT_FILE"
    elif [ $success_rate -ge 80 ]; then
        echo "⚠️ **广播功能基本正常**: 成功率为 ${success_rate}%，存在少量广播失败情况。" >> "$REPORT_FILE"
    else
        echo "❌ **广播功能异常**: 成功率仅为 ${success_rate}%，需要检查网络连接和配置。" >> "$REPORT_FILE"
    fi
    
    cat >> "$REPORT_FILE" << EOF

### 数据一致性验证

EOF
    
    if [ $consistency_issues -eq 0 ]; then
        echo "✅ **数据一致性良好**: 所有 Intent 在各节点间保持一致。" >> "$REPORT_FILE"
    else
        echo "⚠️ **发现数据不一致**: 有 $consistency_issues 个 Intent 存在一致性问题。" >> "$REPORT_FILE"
    fi
    
    # 添加性能指标
    local avg_intents_per_second=0
    if [ $((TEST_END_TIME - TEST_START_TIME)) -gt 0 ]; then
        avg_intents_per_second=$((broadcast_total / (TEST_END_TIME - TEST_START_TIME)))
    fi
    
    cat >> "$REPORT_FILE" << EOF

### 性能指标

- **平均 Intent 生成速率**: $avg_intents_per_second Intent/秒
- **网络延迟**: 待测量
- **内存使用**: 待测量
- **CPU 使用**: 待测量

## 建议和改进

EOF
    
    if [ $success_rate -lt 95 ]; then
        cat >> "$REPORT_FILE" << EOF
### 广播成功率改进建议

1. 检查网络连接稳定性
2. 优化 GossipSub 参数配置
3. 增加重试机制
4. 检查节点间的 P2P 连接状态

EOF
    fi
    
    if [ $consistency_issues -gt 0 ]; then
        cat >> "$REPORT_FILE" << EOF
### 数据一致性改进建议

1. 实现消息确认机制
2. 添加数据同步检查
3. 优化消息传输可靠性
4. 增加错误恢复机制

EOF
    fi
    
    cat >> "$REPORT_FILE" << EOF
### 通用改进建议

1. 添加更详细的性能监控
2. 实现自动化测试流程
3. 增加压力测试场景
4. 优化错误处理和日志记录

## 测试环境信息

- **操作系统**: $(uname -s)
- **Go 版本**: $(go version 2>/dev/null || echo "未检测到")
- **测试工具版本**: v1.0.0
- **配置文件**: configs/test_node*.yaml

---

*报告由 PIN 多节点测试系统自动生成*
EOF
    
    log_info "测试报告已生成: $REPORT_FILE"
}

# 显示测试摘要
show_test_summary() {
    local analysis_result="$1"
    
    local broadcast_total=$(echo "$analysis_result" | cut -d'|' -f1)
    local broadcast_success=$(echo "$analysis_result" | cut -d'|' -f2)
    local success_rate=$(echo "$analysis_result" | cut -d'|' -f3)
    local consistency_issues=$(echo "$analysis_result" | cut -d'|' -f4)
    
    echo ""
    echo -e "${WHITE}┌─────────────────────────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${WHITE}│                                  测试结果摘要                                   │${NC}"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────────────────────────┤${NC}"
    printf "${WHITE}│${NC} 测试时长: %-20s                                              ${WHITE}│${NC}\n" "$((TEST_END_TIME - TEST_START_TIME)) 秒"
    printf "${WHITE}│${NC} 唯一 Intent 总数: %-10s                                          ${WHITE}│${NC}\n" "$broadcast_total"
    printf "${WHITE}│${NC} 成功广播数量: %-10s                                              ${WHITE}│${NC}\n" "$broadcast_success"
    
    local success_color="${GREEN}"
    if [ $success_rate -lt 90 ]; then
        success_color="${YELLOW}"
    fi
    if [ $success_rate -lt 70 ]; then
        success_color="${RED}"
    fi
    
    printf "${WHITE}│${NC} 广播成功率: ${success_color}%-10s${NC}                                            ${WHITE}│${NC}\n" "${success_rate}%"
    
    local consistency_color="${GREEN}"
    if [ $consistency_issues -gt 0 ]; then
        consistency_color="${YELLOW}"
    fi
    
    printf "${WHITE}│${NC} 一致性问题: ${consistency_color}%-10s${NC}                                            ${WHITE}│${NC}\n" "$consistency_issues"
    echo -e "${WHITE}├─────────────────────────────────────────────────────────────────────────────────┤${NC}"
    
    local overall_status="${GREEN}测试通过${NC}"
    if [ $success_rate -lt 90 ] || [ $consistency_issues -gt 0 ]; then
        overall_status="${YELLOW}测试警告${NC}"
    fi
    if [ $success_rate -lt 70 ]; then
        overall_status="${RED}测试失败${NC}"
    fi
    
    printf "${WHITE}│${NC} 总体状态: %-30s                                    ${WHITE}│${NC}\n" "$overall_status"
    printf "${WHITE}│${NC} 报告文件: %-50s                     ${WHITE}│${NC}\n" "$REPORT_FILE"
    echo -e "${WHITE}└─────────────────────────────────────────────────────────────────────────────────┘${NC}"
    echo ""
}

# 显示使用帮助
show_help() {
    echo "PIN 多节点测试结果验证工具"
    echo ""
    echo "用法: $0 [测试时长(秒)]"
    echo ""
    echo "参数:"
    echo "  测试时长    数据收集持续时间，单位秒 (默认: 60)"
    echo ""
    echo "示例:"
    echo "  $0          # 运行60秒测试"
    echo "  $0 120      # 运行120秒测试"
    echo "  $0 300      # 运行5分钟测试"
    echo ""
    echo "功能:"
    echo "  - 验证所有节点运行状态"
    echo "  - 收集 Intent 广播数据"
    echo "  - 分析数据一致性"
    echo "  - 计算广播成功率"
    echo "  - 生成详细测试报告"
    echo ""
    echo "前置条件:"
    echo "  - 所有测试节点必须正在运行"
    echo "  - 节点1应该正在发布 Intent"
    echo "  - 节点2和3应该正在接收 Intent"
    echo ""
}

# 主函数
main() {
    case "${1:-}" in
        --help|-h)
            show_help
            exit 0
            ;;
        [0-9]*)
            TEST_DURATION=$1
            ;;
        "")
            TEST_DURATION=60
            ;;
        *)
            log_error "无效的测试时长: $1"
            show_help
            exit 1
            ;;
    esac
    
    show_test_header
    
    # 检查节点状态
    if ! check_all_nodes; then
        log_error "节点状态检查失败，无法进行测试"
        exit 1
    fi
    
    # 执行测试数据收集
    perform_test_collection
    
    # 分析测试结果
    local analysis_result=$(analyze_test_results)
    
    # 生成测试报告
    generate_test_report "$analysis_result"
    
    # 显示测试摘要
    show_test_summary "$analysis_result"
    
    log_info "测试验证完成！"
    echo ""
    echo -e "${CYAN}查看详细报告:${NC} cat $REPORT_FILE"
    echo -e "${CYAN}或在浏览器中打开:${NC} open $REPORT_FILE"
    echo ""
}

# 执行主函数
main "$@"