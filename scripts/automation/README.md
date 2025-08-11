# PIN 自动化测试系统

## 概述

PIN 自动化测试系统实现了完整的四节点自动化测试环境，模拟真实的意图广播、自动出价和智能匹配流程。每个节点在独立的终端中运行，提供完整的可观测性和控制能力。

## 架构设计

### 四节点架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   节点1 (8100)  │    │   节点2 (8101)  │    │   节点3 (8102)  │    │   节点4 (8103)  │
│  Intent发布者   │───▶│ Service Agent 1 │    │ Service Agent 2 │◀───│  Block Builder  │
│                 │    │   (交易代理)     │    │   (数据代理)     │    │   (匹配节点)    │
│ 自动发布Intent  │    │   自动出价       │    │   自动出价       │    │   自动匹配      │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 测试流程

1. **Intent发布阶段**
   - 节点1 每15秒自动发布一个Intent
   - Intent类型包括：trade, swap, exchange, data_access
   - 通过P2P网络广播到所有节点

2. **自动出价阶段**
   - 节点2&3 监听Intent广播消息
   - 根据能力匹配和出价策略自动评估
   - 在10秒出价窗口内提交出价

3. **智能匹配阶段**
   - 节点4 收集所有出价信息
   - 使用配置的匹配算法选择获胜者
   - 广播匹配结果到网络

4. **循环测试**
   - 整个流程持续进行5分钟
   - 提供实时监控和状态查看

## 快速开始

### 1. 构建应用程序

```bash
make build
```

### 2. 初始化自动化测试环境

```bash
./scripts/automation/setup_automation_env.sh
```

### 3. 启动自动化测试

```bash
./scripts/automation/start_automation_test.sh
```

这将在4个独立终端中启动所有节点，并开启一个监控终端。

### 4. 手动启动单个节点 (可选)

如果你想要更精细的控制，可以分别启动每个节点：

```bash
# 终端1: Intent发布者
./scripts/automation/start_node.sh 1

# 终端2: Service Agent 1 (交易代理)
./scripts/automation/start_node.sh 2

# 终端3: Service Agent 2 (数据代理)
./scripts/automation/start_node.sh 3

# 终端4: Block Builder
./scripts/automation/start_node.sh 4

# 终端5: 监控 (可选)
./scripts/automation/monitor_automation.sh
```

### 5. 监控测试状态

```bash
./scripts/automation/monitor_automation.sh
```

### 6. 清理环境

```bash
./scripts/automation/cleanup_automation.sh
```

## 节点详情

### 节点1 - Intent发布者 (端口8100)

**功能：**
- 自动发布Intent消息到P2P网络
- 支持多种Intent类型：trade, swap, exchange, data_access
- 每15秒发布一个新的Intent
- 提供手动Intent发布API

**API端点：**
```bash
# 健康检查
curl http://localhost:8100/health

# 手动创建Intent
curl -X POST http://localhost:8100/pinai_intent/intent/create \
  -H "Content-Type: application/json" \
  -d '{"type":"trade","payload":"dGVzdA==","sender_id":"auto-publisher"}'
```

### 节点2 - Service Agent 1 (端口8101)

**功能：**
- 监听Intent广播消息
- 专门处理交易类型：trade, swap, exchange
- 激进出价策略，20%利润率
- 自动评估和提交出价

**配置：**
- Agent ID: `trading-agent-auto-001`
- 出价策略: `aggressive`
- 利润率: 20%
- 能力范围: trading, swap, exchange

**API端点：**
```bash
# Agent状态
curl http://localhost:8101/pinai_intent/execution/agents/status

# 启动Agent
curl -X POST http://localhost:8101/pinai_intent/execution/agents/trading-agent-auto-001/start
```

### 节点3 - Service Agent 2 (端口8102)

**功能：**
- 监听Intent广播消息
- 专门处理数据类型：data_access, computation
- 保守出价策略，15%利润率
- 自动评估和提交出价

**配置：**
- Agent ID: `data-agent-auto-002`
- 出价策略: `conservative`
- 利润率: 15%
- 能力范围: data_access, computation

**API端点：**
```bash
# Agent状态
curl http://localhost:8102/pinai_intent/execution/agents/status

# 启动Agent
curl -X POST http://localhost:8102/pinai_intent/execution/agents/data-agent-auto-002/start
```

### 节点4 - Block Builder (端口8103)

**功能：**
- 收集Intent出价信息
- 使用最高出价算法选择获胜者
- 10秒出价收集窗口
- 广播匹配结果

**配置：**
- Builder ID: `auto-builder-001`
- 匹配算法: `highest_bid`
- 出价收集时间: 10秒
- 最少出价数: 1个

**API端点：**
```bash
# Builder状态
curl http://localhost:8103/pinai_intent/execution/builders/status

# 匹配历史
curl http://localhost:8103/pinai_intent/execution/matches/history?limit=10
```

## 配置说明

### 端口分配

| 节点 | HTTP | gRPC | P2P  |
|------|------|------|------|
| 节点1| 8100 | 9100 | 9200 |
| 节点2| 8101 | 9101 | 9201 |
| 节点3| 8102 | 9102 | 9202 |
| 节点4| 8103 | 9103 | 9203 |

### 测试参数

- **测试时长**: 300秒 (5分钟)
- **Intent发布间隔**: 15秒
- **出价窗口**: 10秒
- **匹配窗口**: 15秒

### 网络配置

- **协议ID**: `/pin-automation/1.0.0`
- **Intent主题**: `automation-test.intent-broadcast`
- **出价主题**: `automation-test.bidding`
- **匹配主题**: `automation-test.matching`

## 监控和调试

### 实时监控

```bash
# 启动实时监控 (3秒刷新)
./scripts/automation/monitor_automation.sh

# 自定义刷新间隔和历史记录数
./scripts/automation/monitor_automation.sh 5 20  # 5秒刷新，20条历史
```

监控界面显示：
- 节点运行状态
- Service Agent出价活动
- Block Builder匹配活动
- 系统性能指标
- 最近匹配历史

### 日志查看

```bash
# 查看所有节点日志
tail -f test_data/automation/logs/node*.log

# 查看特定节点日志
tail -f test_data/automation/node1/output.log  # Intent发布者
tail -f test_data/automation/node2/output.log  # Service Agent 1
tail -f test_data/automation/node3/output.log  # Service Agent 2
tail -f test_data/automation/node4/output.log  # Block Builder
```

### API测试

所有节点都提供完整的HTTP API用于手动测试和调试：

```bash
# 系统指标 (任一节点)
curl http://localhost:8100/pinai_intent/execution/metrics

# 最近匹配历史 (Block Builder)
curl http://localhost:8103/pinai_intent/execution/matches/history?limit=5

# Agent控制
curl -X POST http://localhost:8101/pinai_intent/execution/agents/trading-agent-auto-001/stop
curl -X POST http://localhost:8101/pinai_intent/execution/agents/trading-agent-auto-001/start
```

## 文件结构

```
scripts/automation/
├── automation_config.sh      # 统一配置文件
├── setup_automation_env.sh   # 环境初始化
├── start_automation_test.sh  # 一键启动测试
├── start_node.sh            # 单节点启动器
├── monitor_automation.sh     # 实时监控工具
├── cleanup_automation.sh     # 环境清理工具
└── README.md               # 本文档

test_data/automation/
├── node1/                   # 节点1数据目录
├── node2/                   # 节点2数据目录  
├── node3/                   # 节点3数据目录
├── node4/                   # 节点4数据目录
├── status/                  # 状态文件目录
└── logs/                    # 日志链接目录

configs/
├── automation_node1.yaml   # 节点1配置
├── automation_node2.yaml   # 节点2配置
├── automation_node3.yaml   # 节点3配置
└── automation_node4.yaml   # 节点4配置
```

## 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   ./scripts/automation/cleanup_automation.sh
   ```

2. **节点启动失败**
   - 检查应用程序是否构建: `make build`
   - 检查配置文件是否存在: `ls configs/automation_*.yaml`
   - 查看错误日志: `tail test_data/automation/node*/output.log`

3. **P2P连接问题**
   - 确保所有节点按顺序启动
   - 检查防火墙设置
   - 查看P2P连接日志

4. **自动化功能不工作**
   - 确认节点配置中automation.enabled=true
   - 检查Agent和Builder是否正确配置
   - 查看API响应: `curl http://localhost:810X/pinai_intent/execution/agents/status`

### 调试模式

如果需要更详细的调试信息，可以修改配置文件中的日志级别：

```yaml
log:
  level: "debug"  # 改为debug获取更多信息
```

### 性能调优

如果测试运行缓慢，可以调整配置参数：

```bash
# 编辑 scripts/automation/automation_config.sh
export AUTO_INTENT_PUBLISH_INTERVAL=5    # 降低到5秒
export AUTO_BIDDING_WINDOW=5              # 降低到5秒
```

## 扩展开发

### 添加新的Agent策略

1. 修改 `configs/automation_nodeX.yaml` 中的bidding_strategy
2. 实现对应的策略逻辑
3. 重新生成配置: `./scripts/automation/setup_automation_env.sh`

### 添加新的匹配算法

1. 修改 `configs/automation_node4.yaml` 中的matching_algorithm
2. 实现对应的匹配逻辑
3. 重启Block Builder节点

### 自定义Intent类型

1. 修改 `automation_config.sh` 中的 `AUTO_INTENT_TYPES`
2. 更新Agent的capabilities配置
3. 重新初始化环境

---

**注意**: 自动化测试系统设计为开发和演示用途。在生产环境中使用前，请确保进行适当的安全配置和性能优化。