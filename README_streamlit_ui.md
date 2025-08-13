# PIN Intent Network - POC Demo Frontend

Real-time monitoring dashboard for the 4-node PIN automation system. Provides comprehensive visualization of intent publishing, bidding, and matching flow with 5-second auto-refresh.

## Recent Updates & Optimizations

### ✅ Latest Fixes (2024)
- **🔄 Auto-refresh mechanism**: Fixed 5-second auto-refresh with visual countdown and manual refresh button
- **💰 Agent Activity data**: Fixed total_bids and success_rate display issues with proper field mapping
- **📡 Recent Intents**: Fixed sender field display and improved data extraction
- **🎯 Recent Matches**: Added intent_id column, removed currency symbols from bid_amount, fixed time_ago display
- **⏰ Timestamp conversion**: Fixed millisecond to second conversion for accurate time display
- **🎨 UI improvements**: Enhanced error handling, better visual feedback, and improved data validation

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

### 1. 📊 System Metrics Overview (Top)
- Active nodes count with delta indicators
- Total intents processed
- Active bids in the system
- Completed matches count

### 2. 📡 Intent Flow Monitoring
- **Recent Intents table** with complete intent_id, type, sender, status, broadcasts, bids, and time_ago
- Intent type distribution pie chart
- Real-time intent publishing activity
- Status tracking (Validated, Broadcasted, Matched, etc.)

### 3. 💰 Bidding Activity Tracking
- **Agent Activity table** showing agent_id, type, status, total_bids, success_rate, earnings
- Service Agent performance statistics with estimated success rates
- Bidding performance comparison charts
- Real-time agent activity monitoring
- Debug information toggle for troubleshooting

### 4. 🎯 Matching Results
- **Recent Matches table** with match_id, **intent_id**, winner, bid_amount (no currency symbol), total_bids, status, time_ago
- Matching statistics and algorithm distribution
- Block Builder performance metrics
- Match status distribution charts

### 5. 🖥️ Node Status Overview
- Health status of all 4 nodes with color-coded indicators
- Response times and connectivity status
- Last check timestamps
- Port information and error details

### 6. 🌐 P2P Network Status
- Connected peer counts
- Message flow statistics (sent/received)
- Network ID and Host ID information
- Subscribed topics list

### 7. 🔄 Auto-Refresh Controls (Bottom)
- Visual countdown timer with progress bar
- Manual refresh button
- Last refresh timestamp
- Auto-refresh status indicators

## API Integration

The dashboard consumes REST APIs from the PIN automation nodes with enhanced field mapping:

### Node 1 (Intent Publisher - 8100)
- `GET /health` - Health check
- `GET /pinai_intent/intent/list?limit=10` - Intent list with proper field mapping:
  - `id` → `intent_id`
  - `type` → `intent_type` 
  - `senderId` → `sender_id`
  - `timestamp` → `created_at` (seconds)
- `GET /pinai_intent/execution/metrics` - System metrics

### Node 2 & 3 (Service Agents - 8101, 8102)
- `GET /health` - Health check  
- `GET /pinai_intent/execution/agents/status` - Agent status with field mapping:
  - `agentId` → `agent_id`
  - `agentType` → `agent_type`
  - `successfulBids` → Used for both `total_bids_submitted` and `successful_bids`
  - `lastActivity` → `last_activity` (seconds)
- `GET /pinai_intent/intent/list` - Received intents

### Node 4 (Block Builder - 8103)
- `GET /health` - Health check
- `GET /pinai_intent/execution/builders/status` - Builder status
- `GET /pinai_intent/execution/matches/history?limit=10` - Match history with field mapping:
  - `intentId` → `intent_id`
  - `winningAgent` → `winning_agent_id`
  - `winningBid` → `winning_bid_amount`
  - `totalBids` → `total_bids`
  - `matchedAt` → `matched_at` (milliseconds → seconds conversion)

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

**Port already in use:**
```bash
# Use different port
STREAMLIT_PORT=8081 ./scripts/start_streamlit_ui.sh
STREAMLIT_PORT=8082 ./scripts/start_streamlit_ui.sh
STREAMLIT_PORT=8083 ./scripts/start_streamlit_ui.sh

# Or kill existing process
lsof -ti:8080 | xargs kill
```

**Agent Activity shows total_bids as 0:**
- This is expected if `processedIntents` is 0 in the API
- The dashboard now uses `successfulBids` to estimate total bids
- Success rate is calculated based on estimated total bids (assumes ~80% success rate)
- Enable debug mode with the "🔍 show debug info" checkbox to see raw data

**Recent Matches time_ago shows negative values:**
- Fixed: Dashboard now properly converts millisecond timestamps to seconds
- If still occurring, check that API returns valid timestamps

**Data not refreshing:**
- Check the auto-refresh countdown at the bottom of the dashboard
- Use the "🔄 refresh" button for manual refresh
- Verify all 4 PIN nodes are responding to API calls

**Dashboard loading slowly:**
- Check network connectivity to PIN nodes
- Verify API timeout settings in config (default: 3 seconds)
- Monitor system resources
- Check for high API response times in node status panel

**Missing dependencies:**
```bash
# Reinstall dependencies
uv pip install -e . --force-reinstall
```

### Known Issues & Solutions

**Agent Activity Success Rate Calculation:**
- The API provides `successfulBids` but not total bids attempted
- Dashboard estimates total bids assuming ~80% success rate
- This provides more realistic success rate display than 100%
- Future API enhancement needed for accurate total bid counts

**Intent Bid Count Always 0:**
- API doesn't provide `bid_count` field for intents
- This is expected behavior, not a bug
- Bid information is available in the Matching Results panel

**Time Display Issues:**
- Fixed: Millisecond timestamps now properly converted to seconds
- If you see negative time values, check API timestamp format
- Dashboard handles both second and millisecond timestamps automatically

**Data Refresh Behavior:**
- Dashboard uses 5-second caching to reduce API load
- Manual refresh bypasses cache
- Some data may appear stale during high activity periods

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
    # Process and return data with proper field mapping
    return safe_extract_new_data(result)

def safe_extract_new_data(data: dict) -> NewDataModel:
    """Safely extract data with field mapping and type conversion."""
    # Handle timestamp conversion if needed
    timestamp = data.get("timestamp", "0")
    try:
        timestamp = int(timestamp)
        # Convert milliseconds to seconds if needed
        if timestamp > 1000000000000:
            timestamp = timestamp // 1000
    except (ValueError, TypeError):
        timestamp = 0
    
    return NewDataModel(
        field1=data.get("apiField1", data.get("altField1", "default")),
        timestamp=timestamp
    )
```

**2. Create new UI component:**
```python  
# In ui_components.py
def render_new_panel(data: Dict[str, Any]) -> None:
    """Render new dashboard panel with error handling."""
    if not data:
        render_error_panel("New Panel", "No data available")
        return
    
    st.subheader("New Feature")
    # Use render_component_with_error_handling for robustness
    render_component_with_error_handling("New Feature", _render_new_content, data)

def _render_new_content(data):
    """Internal render function with actual implementation."""
    # Implement visualization
    df = create_new_dataframe(data)
    st.dataframe(df, use_container_width=True)
```

**3. Update main dashboard:**
```python
# In main.py, add to render_dashboard()
st.markdown("---")
st.subheader("🆕 New Feature Panel")
render_component_with_error_handling("New Feature", render_new_panel, new_data)
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

## Version History

### v1.2.0 (Latest)
- ✅ Fixed auto-refresh mechanism with visual countdown
- ✅ Enhanced Agent Activity data display with proper field mapping
- ✅ Added intent_id column to Recent Matches table
- ✅ Fixed timestamp conversion for accurate time_ago display
- ✅ Removed currency symbols from bid_amount display
- ✅ Improved error handling and debug information
- ✅ Enhanced UI components with better visual feedback

### v1.1.0
- ✅ Implemented 5-second auto-refresh functionality
- ✅ Added comprehensive error handling and fallback UI
- ✅ Enhanced API client with concurrent requests
- ✅ Improved data validation and type safety

### v1.0.0
- ✅ Initial release with basic dashboard functionality
- ✅ 4-node monitoring capability
- ✅ Real-time data visualization
- ✅ REST API integration

## License

This POC demo frontend is part of the PIN Intent Network project. See project root for license information.