# PIN Intent Network - POC Demo Frontend

Real-time monitoring dashboard for the 4-node PIN automation system. Provides comprehensive visualization of intent publishing, bidding, and matching flow with 5-second auto-refresh.

## Overview

The POC Demo Frontend is a Streamlit-based web dashboard that integrates with the existing PIN automation system via HTTP APIs. It provides real-time monitoring of:

- 🖥️ **Node Status**: Health and connectivity of all 4 nodes
- 📡 **Intent Flow**: Intent publishing and broadcasting activity
- 💰 **Bidding Activity**: Service Agent bidding performance and statistics
- 🎯 **Matching Results**: Block Builder matching outcomes and history
- 🌐 **P2P Network**: Network connectivity and message flow
- 📊 **Performance Metrics**: System-wide performance indicators

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Streamlit Dashboard (Port 8080)              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │ Node Status │ │Intent Flow  │ │ Bidding     │ │ Matching    │ │
│  │   Panel     │ │   Panel     │ │   Panel     │ │   Panel     │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
└─────────────────────────┬───────────────────────────────────────┘
                          │ HTTP APIs
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    PIN Automation System                        │
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│ │   Node 1    │ │   Node 2    │ │   Node 3    │ │   Node 4    │ │
│ │ Publisher   │ │Service Agent│ │Service Agent│ │Block Builder│ │
│ │  (8100)     │ │   (8101)    │ │   (8102)    │ │   (8103)    │ │
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

### System Requirements
- Python 3.8+ (managed by uv)
- [uv](https://docs.astral.sh/uv/) package manager
- PIN automation system running (4 nodes on ports 8100-8103)

### Install uv (if not already installed)
```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

## Quick Start

### 1. Start PIN Automation System
First, ensure the 4-node automation system is running:

```bash
# Start complete automation system
./scripts/automation/start_automation_test.sh

# OR start individual nodes
./scripts/automation/start_node.sh 1  # Intent Publisher (8100)
./scripts/automation/start_node.sh 2  # Service Agent 1 (8101)
./scripts/automation/start_node.sh 3  # Service Agent 2 (8102)  
./scripts/automation/start_node.sh 4  # Block Builder (8103)
```

### 2. Launch Streamlit Dashboard
```bash
# Launch dashboard (automatically installs dependencies)
./scripts/start_streamlit_ui.sh

# Dashboard will be available at: http://localhost:8080
```

### 3. Monitor the System
The dashboard automatically refreshes every 5 seconds and displays:
- Real-time node health status
- Intent publishing and broadcasting activity
- Service Agent bidding performance
- Block Builder matching results
- P2P network connectivity
- System performance metrics

## Manual Setup

If you prefer to set up the environment manually:

```bash
# Create virtual environment
uv venv

# Install dependencies
uv pip install -e .

# Run Streamlit directly
uv run streamlit run streamlit_ui/main.py --server.port=8080
```

## Configuration

### Environment Variables
```bash
# Streamlit server configuration
export STREAMLIT_PORT=8080          # Dashboard port (default: 8080)
export STREAMLIT_HOST=localhost     # Dashboard host (default: localhost)

# Node API endpoints (defaults)
export NODE1_URL=http://localhost:8100  # Intent Publisher
export NODE2_URL=http://localhost:8101  # Service Agent 1
export NODE3_URL=http://localhost:8102  # Service Agent 2
export NODE4_URL=http://localhost:8103  # Block Builder
```

### Dashboard Settings
The dashboard configuration can be modified in `streamlit_ui/config.py`:

```python
# Refresh interval (seconds)
REFRESH_INTERVAL_SECONDS = 5

# API timeout (seconds)
API_TIMEOUT_SECONDS = 3

# Node configurations
NODE_CONFIGS = {
    1: {"name": "Intent Publisher", "base_url": "http://localhost:8100"},
    2: {"name": "Service Agent 1", "base_url": "http://localhost:8101"},
    3: {"name": "Service Agent 2", "base_url": "http://localhost:8102"},
    4: {"name": "Block Builder", "base_url": "http://localhost:8103"}
}
```

## Dashboard Panels

### 1. Node Status Overview
- Health status of all 4 nodes
- Response times and connectivity
- Last check timestamps
- Color-coded status indicators

### 2. Intent Flow Monitoring  
- Recent intent publications
- Intent type distribution
- Broadcasting activity
- Status tracking (pending, broadcasted, matched)

### 3. Bidding Activity Tracking
- Service Agent performance statistics
- Bid submission counts and success rates
- Earnings tracking
- Agent activity monitoring

### 4. Matching Results
- Recent matching outcomes
- Winning bid details
- Matching algorithm usage
- Success/failure statistics

### 5. P2P Network Status
- Connected peer counts
- Message flow statistics
- Network topology information
- Topic subscriptions

### 6. Performance Metrics
- System-wide performance indicators
- Response time analytics
- Success rate trends
- Resource utilization

## API Integration

The dashboard consumes REST APIs from the PIN automation nodes:

### Node 1 (Intent Publisher - 8100)
- `GET /health` - Health check
- `GET /pinai_intent/intent/list` - Intent list
- `GET /pinai_intent/execution/metrics` - System metrics

### Node 2 & 3 (Service Agents - 8101, 8102)
- `GET /health` - Health check  
- `GET /pinai_intent/execution/agents/status` - Agent status
- `GET /pinai_intent/intent/list` - Received intents

### Node 4 (Block Builder - 8103)
- `GET /health` - Health check
- `GET /pinai_intent/execution/builders/status` - Builder status
- `GET /pinai_intent/execution/matches/history` - Match history

## Troubleshooting

### Common Issues

**Dashboard shows "Node Offline" errors:**
```bash
# Check if PIN nodes are running
curl http://localhost:8100/health  # Intent Publisher
curl http://localhost:8101/health  # Service Agent 1
curl http://localhost:8102/health  # Service Agent 2
curl http://localhost:8103/health  # Block Builder

# Restart PIN automation if needed
./scripts/automation/start_automation_test.sh
```

**Port 8080 already in use:**
```bash
# Use different port
STREAMLIT_PORT=8081 ./scripts/start_streamlit_ui.sh

# Or kill existing process
lsof -ti:8080 | xargs kill
```

**Dashboard loading slowly:**
- Check network connectivity to PIN nodes
- Verify API timeout settings in config
- Monitor system resources

**Missing dependencies:**
```bash
# Reinstall dependencies
uv pip install -e . --force-reinstall
```

### Debugging

**Enable debug mode:**
```bash
# Set environment variable for verbose logging
export STREAMLIT_LOGGER_LEVEL=debug
./scripts/start_streamlit_ui.sh
```

**Check API responses manually:**
```bash
# Test API endpoints directly
curl -v http://localhost:8100/pinai_intent/execution/metrics
curl -v http://localhost:8101/pinai_intent/execution/agents/status
curl -v http://localhost:8103/pinai_intent/execution/builders/status
```

## Development

### Project Structure
```
streamlit_ui/
├── __init__.py           # Package initialization
├── main.py               # Main Streamlit application
├── api_client.py         # HTTP API client for node communication
├── data_models.py        # Data structures and models
├── ui_components.py      # Streamlit UI component functions
├── config.py             # Configuration settings
└── utils.py              # Utility functions

scripts/
└── start_streamlit_ui.sh # Startup script

pyproject.toml            # UV project configuration
README_streamlit_ui.md    # This documentation
```

### Adding New Features

**1. Add new API endpoint:**
```python
# In api_client.py
async def get_new_endpoint(self, node_id: int) -> ResponseModel:
    url = f"{self.base_urls[node_id]}/new/endpoint"
    result = await self.safe_api_call(url)
    # Process and return data
```

**2. Create new UI component:**
```python  
# In ui_components.py
def render_new_panel(data: Dict[str, Any]) -> None:
    """Render new dashboard panel."""
    st.subheader("New Feature")
    # Implement visualization
```

**3. Update main dashboard:**
```python
# In main.py, add to render_dashboard()
st.subheader("New Feature Panel")
render_new_panel(new_data)
```

### Testing

**Run with test data:**
```bash
# Start dashboard without PIN nodes (shows fallback UI)
./scripts/start_streamlit_ui.sh
```

**API testing:**
```bash
# Test individual API endpoints
python -c "
import asyncio
from streamlit_ui.api_client import NodeAPIClient
client = NodeAPIClient()
data = asyncio.run(client.fetch_all_data())
print(data)
"
```

## Performance Optimization

### Caching Strategy
- API responses cached for 5 seconds using `@st.cache_data`
- Historical data cached in session state
- Efficient data aggregation to minimize API calls

### Resource Management
- Asynchronous API calls for concurrent node communication
- Connection pooling via httpx
- Graceful error handling with fallback UI

## Security Considerations

- Dashboard runs on localhost by default
- No authentication required (local development tool)
- API calls use HTTP (not HTTPS) for local development
- No sensitive data persistence

For production deployment, consider:
- HTTPS configuration
- Authentication mechanism
- Network security policies
- API rate limiting

## Contributing

1. Follow existing code structure and patterns
2. Add type hints to all functions
3. Update documentation for new features
4. Test with running PIN automation system
5. Ensure 5-second refresh cycle works properly

## License

This POC demo frontend is part of the PIN Intent Network project. See project root for license information.