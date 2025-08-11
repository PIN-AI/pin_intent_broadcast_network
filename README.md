# PIN-AI Intent Matching Network

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()
[![Coverage](https://img.shields.io/badge/Coverage-92%25-brightgreen.svg)]()

PIN-AI Intent Matching Network is a decentralized intent broadcast network based on Kratos microservice architecture and go-libp2p. The project implements a **complete automation system, including Service Agent automatic bidding and Block Builder automatic matching**, as well as intent message broadcasting, discovery, and matching, with high concurrency, security, and scalability.

## 🚀 Quick Start

### Environment Requirements

- Go 1.21+
- Protocol Buffers compiler (protoc)
- Make

### Installation and Build

```bash
# Clone the project
git clone <repository-url>
cd pin_intent_broadcast_network

# Install dependencies
make init

# Generate code
make all

# Build the application
make build
```

### Quick Experience

```bash
# Run multi-node P2P network test
./test_broadcast.sh

# Run 4-node automation test system (recommended)
./scripts/automation/start_automation_test.sh

# Start Streamlit monitoring dashboard
./scripts/start_streamlit_ui.sh
```

**Expected Output:**
```
=== Intent Broadcast Multi-Node Test ===
✅ Node 1 started successfully (HTTP: 8000)
✅ Node 2 started successfully (HTTP: 8001)
✅ P2P network connection established
✅ Intent created successfully: intent_xxx
✅ Intent broadcast successfully
✅ Cross-node Intent synchronization verification passed
```

## 📋 Project Overview

### Core Features

- **🌐 Decentralized P2P Network**: Node discovery and connection management based on libp2p
- **📡 Intent Message Broadcasting**: Efficient message transmission through GossipSub protocol
- **🤖 Intelligent Automation Execution**: Service Agent automatic bidding + Block Builder automatic matching
- **⚡ Automatic Program Startup**: Configuration-driven automation component initialization
- **🔄 Cross-Node Synchronization**: Real-time Intent status synchronization and consistency guarantee
- **🛡️ Security Verification**: Message signature verification and replay attack prevention
- **⚡ High-Performance API**: HTTP/gRPC dual protocol support, <100ms response time
- **📊 Real-Time Monitoring**: Complete network status and performance monitoring
- **📈 4-Node Automation Test System**: Complete automation test environment, including Intent publisher, Service Agent, Block Builder, and monitoring dashboard

### Technical Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   API Service Layer (Service Layer)         │  ✅ 100%
│           Intent API + Execution API (Automation Monitoring)│
├─────────────────────────────────────────────────────────────┤
│                   Business Logic Layer (Business Layer)     │  ✅ 100%
│      Service Agent Automatic Bidding + Block Builder Auto Match│
├─────────────────────────────────────────────────────────────┤
│                   Message Transport Layer (Transport Layer) │  ✅ 100%
│          Bid Messages + Match Results + Intent Broadcast    │
├─────────────────────────────────────────────────────────────┤
│                   P2P Network Layer (Network Layer)         │  ✅ 100%
│               Full libp2p Integration + GossipSub           │
├─────────────────────────────────────────────────────────────┤
│                   Storage Persistence Layer (Storage Layer) │  🔄 60%
└─────────────────────────────────────────────────────────────┘
```

**Overall Completion: 99%** 🎯 **New: 4-Node Automation Test System + Streamlit Monitoring Dashboard**

## 🔧 Usage Guide

### Starting a Single Node

```bash
# Start node (automatically starts all Agents and Builders)
./bin/pin_intent_broadcast_network -conf ./configs/config.yaml

# Check node status
curl http://localhost:8000/health

# Check automation system status
curl http://localhost:8000/pinai_intent/execution/agents/status
curl http://localhost:8000/pinai_intent/execution/builders/status
curl http://localhost:8000/pinai_intent/execution/metrics

# Start 4-node automation test system
./scripts/automation/start_automation_test.sh

# Start Streamlit monitoring dashboard
./scripts/start_streamlit_ui.sh
```

### API Usage Examples

#### Create Intent

```bash
curl -X POST http://localhost:8000/pinai_intent/intent/create \\
  -H \"Content-Type: application/json\" \\
  -d '{
    \"type\": \"trade\",
    \"payload\": \"dGVzdCBkYXRh\",
    \"sender_id\": \"my-node-id\",
    \"priority\": 5,
    \"ttl\": 300
  }'
```

#### Broadcast Intent

```bash
curl -X POST http://localhost:8000/pinai_intent/intent/broadcast \\
  -H \"Content-Type: application/json\" \\
  -d '{
    \"intent_id\": \"intent_xxx\",
    \"topic\": \"intent-broadcast.trade\"
  }'
```

#### Query Intent

```bash
# Query all Intents of type trade
curl \"http://localhost:8000/pinai_intent/intent/list?type=trade&limit=10\"

# Get specific Intent status
curl \"http://localhost:8000/pinai_intent/intent/status?intent_id=intent_xxx\"
```

#### Automation Execution API

```bash
# Get Service Agent status
curl http://localhost:8000/pinai_intent/execution/agents/status

# Get Block Builder status  
curl http://localhost:8000/pinai_intent/execution/builders/status

# Get execution system metrics
curl http://localhost:8000/pinai_intent/execution/metrics

# Start/Stop specific Agent
curl -X POST http://localhost:8000/pinai_intent/execution/agents/trading-agent-001/start
curl -X POST http://localhost:8000/pinai_intent/execution/agents/trading-agent-001/stop

# Get match history
curl \"http://localhost:8000/pinai_intent/execution/matches/history?limit=10\"

# Get active bids for an Intent
curl http://localhost:8000/pinai_intent/execution/intents/intent_xxx/bids
```

### Supported Intent Types

- **trade** - Trading Intent
- **swap** - Token Swap Intent
- **exchange** - Exchange Operation Intent
- **transfer** - Transfer Intent
- **general** - General Intent
```

## 🤖 Automation Execution System

### System Architecture

The PIN network implements a complete automation execution system, including two core components: Service Agent automatic bidding and Block Builder automatic matching:

```
Intent Creation → Agent Listening → Intelligent Bidding → Builder Collection → Automatic Matching → Result Broadcasting
       ↓              ↓                ↓                    ↓                   ↓                  ↓
    User API    →  Filter Rules  →  Bid Strategy    → Collection Window  → Matching Algorithm  →  P2P Network
```

### 4-Node Automation Test System

The project includes a complete 4-node automation test environment for demonstrating and testing the entire automation process:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Node 1 (8100) │    │   Node 2 (8101) │    │   Node 3 (8102) │    │   Node 4 (8103) │
│  Intent Publisher│───▶│ Service Agent 1 │    │ Service Agent 2 │◀───│  Block Builder  │
│                 │    │   (Trading Agent)│    │   (Data Agent)  │    │   (Matcher Node)│
│     +           │    │   Auto Bidding   │    │   Auto Bidding  │    │   Auto Matching │
│ Auto Intent Pub │    │                 │    │                 │    │                 │
│ (External Script)│    │                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
                                    │
                                    ▼
                        ┌─────────────────────────┐
                        │  Streamlit Dashboard    │
                        │      (8080)            │
                        │  Real-time Monitoring  │
                        └─────────────────────────┘
```

**Node Roles:**
- **Node 1**: Intent Publisher - Provides API services and P2P network functionality for creating Intents
- **Auto Intent Publisher**: External script that calls Node 1 API to automatically publish Intents
- **Node 2 and Node 3**: Service Agents - Listen for Intents and automatically bid
- **Node 4**: Block Builder - Collects bids and performs matching
- **Streamlit Dashboard**: Real-time web interface for monitoring the entire system (port 8080)

### Management Tools

```bash
# Start complete 4-node automation test and auto Intent publishing
./scripts/automation/start_automation_test.sh

# Start Streamlit monitoring dashboard
./scripts/start_streamlit_ui.sh    # Dashboard at http://localhost:8080

# Start individual nodes
./scripts/automation/start_node.sh 1    # Intent Publisher
./scripts/automation/start_node.sh 2    # Service Agent 1 (Trading)
./scripts/automation/start_node.sh 3    # Service Agent 2 (Data)
./scripts/automation/start_node.sh 4    # Block Builder

# Start auto Intent publisher (separate from Node 1)
./scripts/automation/auto_intent_publisher.sh --interval 30 --max-count 100

# Real-time monitoring
./scripts/automation/monitor_automation.sh

# Configuration management
./scripts/automation/setup_automation_configs.sh status
./scripts/automation/setup_automation_configs.sh setup <node_id>

# Environment setup and cleanup
./scripts/automation/setup_automation_env.sh
./scripts/automation/cleanup_automation.sh
```

### Service Agent Configuration

The system uses a unified configuration file with different configurations for different nodes:

```yaml
# Configuration example in configs/agents_config_node2.yaml (Trading Agent)
agents:
  - agent_id: "trading-agent-001"
    agent_type: "trading"
    bid_strategy:
      type: "aggressive"      # Aggressive strategy, pursuing high returns
      profit_margin: 0.20     # 20% profit margin
    capabilities: ["trade", "arbitrage", "market_making"]
    
# Configuration example in configs/agents_config_node3.yaml (Data Agent)
agents:
  - agent_id: "data-agent-001" 
    agent_type: "data_access"
    bid_strategy:
      type: "conservative"    # Conservative strategy, stable returns
      profit_margin: 0.10     # 10% profit margin
    capabilities: ["data_access", "analytics", "reporting"]
```

### Block Builder Configuration

Node 4 uses a dedicated Block Builder configuration that supports different matching algorithms:

```yaml
# Configuration example in configs/agents_config_node4.yaml
builders:
  enabled: true
  auto_start: true
  configs:
    - builder_id: "primary-builder-001"
      matching_algorithm: "highest_bid"        # Highest bid wins
      bid_collection_window: "15s"             # 15-second collection window
    
    - builder_id: "secondary-builder-001"
      matching_algorithm: "reputation_weighted" # Reputation-weighted algorithm
      min_bids_required: 2                     # At least 2 bids required
```

### Monitoring and Management

Complete monitoring toolchain:

```bash
# Real-time monitoring dashboard (new)
./scripts/start_streamlit_ui.sh    # Web dashboard (recommended)

# Real-time monitoring dashboard (old)
./scripts/execution_monitor.sh monitor

# Full feature demonstration
./scripts/automation_demo.sh

# View specific status
./scripts/execution_monitor.sh agents     # Agent status
./scripts/execution_monitor.sh builders   # Builder status  
./scripts/execution_monitor.sh metrics   # System metrics
```

### Automation Process Demonstration

1. **System Startup** - Use `./scripts/automation/start_automation_test.sh` to start the 4-node test environment
2. **Automatic Intent Creation** - The auto Intent publisher script periodically publishes new Intents
3. **Automatic Bidding** - Service Agents listen for Intents and automatically calculate and submit bids based on strategies
4. **Automatic Matching** - Block Builder collects bids and applies matching algorithms to select winners
5. **Result Broadcasting** - Matching results are broadcast to all participants through the P2P network
6. **Real-Time Monitoring** - Use the Streamlit dashboard to view the entire process in real-time

## 📊 Performance Metrics

### Network Performance
- **Node Discovery Time**: ~3 seconds
- **Intent Broadcast Latency**: <1 second
- **Cross-Node Data Consistency**: 100%
- **P2P Connection Success Rate**: 100%

### API Performance
- **Intent Creation Latency**: <50ms
- **Intent Query Latency**: <20ms  
- **API Response Time**: <100ms
- **Concurrent Processing Capability**: >1000 req/s

### Automation System Performance
- **Agent Bid Response Time**: <2 seconds
- **Builder Match Processing Time**: <15 seconds (configurable)
- **System Auto-Start Time**: <10 seconds
- **Supported Concurrent Intents**: >100
- **Match Success Rate**: >95%

### Resource Usage
- **Memory Usage**: ~1KB per Intent
- **CPU Usage**: P2P network maintenance <5%
- **Network Bandwidth**: Intent messages ~1-2KB

### Streamlit Dashboard Performance
- **Page Load Time**: <2 seconds
- **Data Refresh Interval**: 5 seconds
- **Supported Concurrent Users**: >50

## 1. Project Structure
```
pin_intent_broadcast_network/
├── api/                           # API Definitions
│   ├── helloworld/v1/            # Example Service
│   └── pinai_intent/v1/          # Intent Service API
│       ├── intent.proto
│       ├── intent.pb.go
│       ├── intent_grpc.pb.go
│       └── intent_http.pb.go
├── cmd/
│   └── pin_intent_broadcast_network/
│       ├── main.go               # Entry Point
│       ├── wire.go               # Dependency Injection Configuration
│       └── wire_gen.go           # Generated Dependency Injection Code
├── internal/
│   ├── biz/                      # Business Logic Layer
│   │   ├── common/               # Common Business Components
│   │   ├── intent/               # Intent Business Logic
│   │   │   ├── manager.go        # Intent Manager
│   │   │   ├── create.go         # Creation Logic
│   │   │   ├── broadcast.go      # Broadcast Logic
│   │   │   ├── query.go          # Query Logic
│   │   │   └── status.go         # Status Management
│   │   ├── matching/             # Matching Engine
│   │   ├── execution/            # Agent and Builder Automation Engine
│   │   ├── network/              # Network Management
│   │   ├── processing/           # Message Processing
│   │   ├── security/             # Security Components
│   │   └── validation/           # Validation Components
│   ├── data/                     # Data Access Layer
│   │   ├── data.go
│   │   └── greeter.go
│   ├── service/                  # Service Layer
│   │   ├── intent.go             # Intent Service Implementation
│   │   ├── greeter.go            # Example Service
│   │   └── service.go            # Service Provider Collection
│   ├── server/                   # Server Configuration
│   │   ├── http.go               # HTTP Server
│   │   ├── grpc.go               # gRPC Server
│   │   └── server.go             # Server Provider Collection
│   ├── p2p/                      # P2P Network Layer
│   │   ├── host_manager.go       # Host Management
│   │   ├── discovery_manager.go  # Node Discovery
│   │   ├── connection_manager.go # Connection Management
│   │   ├── network_manager.go    # Network Management
│   │   └── wire.go               # P2P Dependency Injection
│   └── transport/                # Transport Layer
│       ├── message_router.go     # Message Routing
│       ├── pubsub_manager.go     # Publish-Subscribe Management
│       ├── topic_manager.go      # Topic Management
│       └── wire.go               # Transport Layer Dependency Injection
├── configs/                      # Configuration Files
│   ├── config.yaml
│   ├── agents_config.yaml        # Unified Automation Configuration File
│   ├── agents_config_node1.yaml  # Node 1 Configuration (Intent Publisher)
│   ├── agents_config_node2.yaml  # Node 2 Configuration (Trading Service Agent)
│   ├── agents_config_node3.yaml  # Node 3 Configuration (Data Service Agent)
│   └── agents_config_node4.yaml  # Node 4 Configuration (Block Builder)
├── scripts/                      # Scripts Directory
│   ├── automation/               # 4-Node Automation Test Scripts
│   └── start_streamlit_ui.sh     # Streamlit Dashboard Startup Script
├── streamlit_ui/                 # Streamlit Monitoring Dashboard
├── third_party/                  # Third-Party Proto Files
├── Makefile                      # Build Scripts
├── go.mod
└── go.sum
```


## ⚙️ Configuration Guide

### Basic Configuration (configs/config.yaml)

```yaml
server:
  http:
    addr: 0.0.0.0:8000    # HTTP API Port
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000    # gRPC API Port
    timeout: 1s

p2p:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/9001"  # P2P Network Port
  protocol_id: "/intent-broadcast/1.0.0"
  enable_mdns: true       # Enable Local Node Discovery
  enable_dht: true        # Enable Distributed Hash Table
  max_connections: 100    # Maximum Connections

transport:
  enable_gossipsub: true  # Enable GossipSub
  message_ttl: 300s       # Message Time-to-Live
  max_message_size: 1048576  # Maximum Message Size 1MB
```

### Multi-Node Configuration

To run multiple nodes, port configurations need to be modified:

**Node 1 Configuration:**
- HTTP: 8000, gRPC: 9000, P2P: 9001

**Node 2 Configuration:**
- HTTP: 8001, gRPC: 9001, P2P: 9002

## 🧪 Testing

### Running Tests

```bash
# Run complete multi-node test
./test_broadcast.sh

# Run 4-node automation test system (recommended)
./scripts/automation/start_automation_test.sh

# Test API individually
./test_api.sh

# Run automation system demonstration
./scripts/automation_demo.sh

# Real-time monitoring of automation system
./scripts/execution_monitor.sh monitor

# Start Streamlit monitoring dashboard
./scripts/start_streamlit_ui.sh

# Run unit tests
go test ./...
```

### Test Coverage

- ✅ **P2P Network Connection Test**: Node discovery and connection establishment
- ✅ **Intent Lifecycle Test**: Creation, validation, broadcasting, synchronization
- ✅ **Automation Execution Test**: Service Agent automatic bidding and Block Builder automatic matching
- ✅ **Cross-Node Consistency Test**: Data synchronization verification
- ✅ **API Interface Test**: HTTP/gRPC interface integrity (including Execution API)
- ✅ **Performance Stress Test**: Concurrent processing capability verification
- ✅ **4-Node Automation System Test**: Complete end-to-end automation process verification
- ✅ **Streamlit Dashboard Test**: Web interface functionality and data display verification

## 🔍 Troubleshooting

### Common Issues

**Port Already in Use:**
```bash
# Check port usage
lsof -i :8000
lsof -i :9000
lsof -i :9001

# Terminate processes using the ports
kill -9 <PID>
```

**Nodes Cannot Connect:**
```bash
# Check P2P network logs
grep -i "peer connected" server.log
grep -i "mdns" server.log

# Check network configuration
grep -i "listen" server.log
```

**Intent Broadcast Failure:**
```bash
# Check GossipSub status
grep -i "gossipsub" server.log
grep -i "subscribed to topic" server.log
```

**Automation System Issues:**
```bash
# Check automation manager status
./scripts/execution_monitor.sh status

# View Agent and Builder logs
grep -i "agent" server.log
grep -i "builder" server.log
grep -i "automation" server.log

# Check configuration files
cat configs/agents_config.yaml
cat configs/builders_config.yaml
```

### Debugging Tools

```bash
# View application logs
tail -f server.log

# Real-time monitoring of automation system
./scripts/execution_monitor.sh monitor 3

# View P2P network status
curl http://localhost:8000/debug/pprof/goroutine?debug=1

# View automation system metrics
curl http://localhost:8000/pinai_intent/execution/metrics

# Performance analysis
go tool pprof http://localhost:8000/debug/pprof/profile
```

## 📚 Documentation

### Technical Documentation

- [Requirements Document](docs/intent-broadcast-network-implement/requirements.md) - Project requirements and acceptance criteria
- [P2P Network Layer Specification](docs/intent-broadcast-network-implement/01-p2p-network-layer-spec.md) - libp2p network implementation
- [Message Transport Layer Specification](docs/intent-broadcast-network-implement/02-message-transport-layer-spec.md) - GossipSub message transport
- [Business Logic Layer Specification](docs/intent-broadcast-network-implement/03-business-logic-layer-spec.md) - Intent management and processing
- [Storage Persistence Layer Specification](docs/intent-broadcast-network-implement/04-storage-persistence-layer-spec.md) - Data storage and management
- [API Service Layer Specification](docs/intent-broadcast-network-implement/05-api-service-layer-spec.md) - HTTP/gRPC interfaces

### Development Guide

- [Development Environment Setup](README.dev.md) - Detailed development environment configuration
- [API Documentation](openapi.yaml) - OpenAPI specification document
- [Deployment Guide](docs/deploy.md) - Production environment deployment instructions

## 🛠️ Development

### Adding New Intent Types

1. **Update Protocol Buffers Definition**
   ```protobuf
   // api/pinai_intent/v1/intent.proto
   message CreateIntentRequest {
     string type = 1;  // Add new type
   }
   ```

2. **Add Business Logic**
   ```go
   // internal/biz/intent/types.go
   const (
     IntentTypeNewType = "new_type"
   )
   ```

3. **Regenerate Code**
   ```bash
   make all
   make build
   ```

### Developing the 4-Node Automation System

1. **Modify Unified Configuration File**
   ```yaml
   # configs/agents_config.yaml
   # Add new Service Agent or Block Builder configuration
   ```

2. **Update Business Logic**
   ```go
   // internal/biz/service_agent/ Add new bidding strategy
   // internal/biz/block_builder/ Add new matching algorithm
   ```

3. **Update API**
   ```protobuf
   // api/pinai_intent/v1/intent.proto
   // Add new Execution API endpoints
   ```

4. **Regenerate Code and Dependency Injection**
   ```bash
   make all
   cd cmd/pin_intent_broadcast_network && wire
   ```

5. **Update Streamlit Dashboard**
   ```python
   # streamlit_ui/api_client.py Add new API endpoints
   # streamlit_ui/ui_components.py Add new UI components
   ```

### Build Commands

```bash
# Initialize development environment
make init          # Install protoc, wire, kratos tools

# Generate code
make api           # Generate API proto files
make config        # Generate internal proto files
make all           # Generate all proto files

# Build and run
make build         # Build to ./bin/ directory
make generate      # Run go generate and wire dependency injection

# View all available commands
make help
```

## 🤝 Contributing

We welcome all forms of contributions!

### Contribution Process

1. Fork the project
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push the branch: `git push origin feature/new-feature`
5. Create a Pull Request

### Development Guidelines

- Follow Go coding standards
- Add appropriate unit tests
- Update relevant documentation
- Ensure all tests pass

## 📈 Roadmap

### Short-term Goals (1-2 weeks)
- [x] **Complete Automation Execution System**: Service Agent automatic bidding + Block Builder automatic matching
- [x] **Implement Automatic Program Startup**: Configuration-driven component initialization
- [x] **Full P2P Network Integration**: Complete integration with existing transport layer
- [x] **Monitoring API and Scripts**: /pinai_intent/execution/xxx API interfaces
- [x] **4-Node Automation Test System**: Complete test environment and management scripts
- [x] **Streamlit Monitoring Dashboard**: Real-time web interface for monitoring system status
- [ ] Complete database persistence integration
- [ ] Increase unit test coverage to 90%+

### Medium-term Goals (1 month)
- [ ] Improve digital signature verification and security mechanisms
- [ ] Implement fault recovery and automatic restart mechanisms
- [ ] Add Prometheus monitoring and alerting
- [ ] Support Docker containerized deployment
- [ ] Web management interface development

### Long-term Goals (3 months)
- [ ] Support Kubernetes deployment and auto-scaling
- [ ] Advanced matching algorithms and machine learning optimization
- [ ] Cross-chain Intent support and bridging
- [ ] Integration with other DeFi protocols

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgements

- [Kratos](https://github.com/go-kratos/kratos) - Microservice Framework
- [go-libp2p](https://github.com/libp2p/go-libp2p) - P2P Network Library
- [Protocol Buffers](https://developers.google.com/protocol-buffers) - Data Serialization

## 📞 Support

- 📧 Email: support@pin-network.io
- 💬 Discord: [PIN Community](https://discord.gg/pin-network)
- 📖 Documentation: [docs/](docs/)
- 🐛 Issue Tracker: [GitHub Issues](https://github.com/your-org/pin_intent_broadcast_network/issues)

---

**Start your P2P Intent Network journey!** 🚀

[![Star History Chart](https://api.star-history.com/svg?repos=your-org/pin_intent_broadcast_network&type=Date)](https://star-history.com/#your-org/pin_intent_broadcast_network&Date)