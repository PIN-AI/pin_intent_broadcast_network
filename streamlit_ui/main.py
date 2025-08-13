"""
PIN Intent Network POC Demo Frontend - Main Streamlit Application

Real-time monitoring dashboard for the 4-node PIN automation system.
Provides comprehensive visualization of intent publishing, bidding, and matching flow.
"""

import asyncio
import time
import threading
from typing import Dict, Any, Optional
import streamlit as st
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor

# Import local modules
from config import STREAMLIT_CONFIG, REFRESH_INTERVAL_SECONDS, UI_TEXT
from api_client import NodeAPIClient
from data_models import (
    UIState, DataCache, DashboardMetrics, NodeStatus, 
    aggregate_execution_metrics, create_p2p_network_info_from_metrics,
    create_empty_dashboard_metrics
)
from ui_components import (
    render_top_metrics, render_nodes_status_panel, 
    render_intent_monitoring_panel, render_bidding_activity_panel,
    render_matching_results_panel, render_p2p_network_panel,
    render_performance_metrics_panel, render_sidebar_info,
    render_error_panel, render_refresh_indicator, render_component_with_error_handling
)
from utils import get_system_health_score, calculate_delta


class AutoRefreshManager:
    """管理自动刷新机制的类"""
    
    def __init__(self, interval: int = 5):
        self.interval = interval
        self.last_refresh = time.time()
    
    def should_refresh(self) -> bool:
        """检查是否应该刷新"""
        return time.time() - self.last_refresh >= self.interval
    
    def trigger_refresh(self):
        """Mark that refresh should be triggered"""
        self.last_refresh = time.time()
        # Don't call st.rerun() directly here, let main() handle it
    
    def get_countdown(self) -> int:
        """获取倒计时秒数"""
        return max(0, self.interval - int(time.time() - self.last_refresh))
    
    def get_progress(self) -> float:
        """获取刷新进度 (0.0 到 1.0)"""
        elapsed = time.time() - self.last_refresh
        return min(1.0, elapsed / self.interval)
    
    def render_indicator(self):
        """Render improved refresh indicator with better visual feedback"""
        countdown = self.get_countdown()
        progress = self.get_progress()
        
        # Create more prominent refresh status display
        col1, col2 = st.columns([3, 1])
        
        with col1:
            if countdown > 0:
                # Show progress bar with countdown
                st.progress(progress, text=f"🔄 下次自动刷新: {countdown} 秒")
                
                # Add a small status indicator
                status_text = f"⏱️ 自动刷新间隔: 5秒 | 剩余: {countdown}秒"
                st.caption(status_text)
            else:
                st.progress(1.0, text="🔄 正在刷新数据...")
                st.caption("⚡ 正在获取最新数据...")
        
        with col2:
            # Manual refresh button with better styling
            if st.button("🔄 立即刷新", 
                        key="manual_refresh", 
                        help="点击立即刷新所有数据",
                        type="primary"):
                # Reset the refresh timer and trigger refresh
                self.last_refresh = time.time()
                st.session_state.should_refresh = True
                st.rerun()
        
        # Add last refresh time info
        last_refresh_time = datetime.fromtimestamp(self.last_refresh).strftime("%H:%M:%S")
        st.caption(f"📅 上次刷新时间: {last_refresh_time}")


def setup_page_config() -> None:
    """Configure Streamlit page settings."""
    st.set_page_config(**STREAMLIT_CONFIG)
    
    # Custom CSS for better styling
    st.markdown("""
    <style>
    .main {
        padding-top: 2rem;
    }
    
    .metric-container {
        background-color: #f0f2f6;
        padding: 1rem;
        border-radius: 0.5rem;
        margin: 0.5rem 0;
    }
    
    .status-card {
        border: 2px solid;
        border-radius: 10px;
        padding: 15px;
        text-align: center;
        margin: 10px 0;
    }
    
    .error-message {
        color: #dc3545;
        font-weight: bold;
    }
    
    .success-message {
        color: #28a745;
        font-weight: bold;
    }
    
    .warning-message {
        color: #ffc107;
        font-weight: bold;
    }
    
    .refresh-indicator {
        position: fixed;
        top: 10px;
        right: 10px;
        background-color: rgba(0, 0, 0, 0.1);
        padding: 5px 10px;
        border-radius: 5px;
        font-size: 12px;
        color: #666;
    }
    </style>
    """, unsafe_allow_html=True)


def initialize_session_state() -> None:
    """Initialize Streamlit session state variables."""
    if "ui_state" not in st.session_state:
        st.session_state.ui_state = UIState()
        st.session_state.ui_state.last_refresh = time.time()
    
    if "data_cache" not in st.session_state:
        st.session_state.data_cache = DataCache()
    
    if "api_client" not in st.session_state:
        st.session_state.api_client = NodeAPIClient()
    
    if "previous_metrics" not in st.session_state:
        st.session_state.previous_metrics = create_empty_dashboard_metrics()
    
    if "error_count" not in st.session_state:
        st.session_state.error_count = 0


def run_async_in_thread(async_func) -> Dict[str, Any]:
    """
    Run async function in a separate thread with its own event loop.
    This solves the AsyncIO compatibility issue in Streamlit.
    """
    def run_in_thread():
        # Create new event loop for this thread
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            api_client = NodeAPIClient()
            return loop.run_until_complete(api_client.fetch_all_data())
        finally:
            loop.close()
    
    # Execute async function in thread pool
    with ThreadPoolExecutor(max_workers=1) as executor:
        future = executor.submit(run_in_thread)
        try:
            # Wait for result with timeout
            return future.result(timeout=15)  # 15 second timeout
        except Exception as e:
            st.error(f"Failed to fetch data in thread: {str(e)}")
            return create_empty_data_response()


@st.cache_data(ttl=REFRESH_INTERVAL_SECONDS, show_spinner=False)
def fetch_all_data() -> Dict[str, Any]:
    """
    Fetch data from all node APIs with caching.
    Uses thread-based approach to handle AsyncIO in Streamlit.
    """
    try:
        return run_async_in_thread(None)
    except Exception as e:
        st.error(f"Failed to fetch data: {str(e)}")
        return create_empty_data_response()


def create_empty_data_response() -> Dict[str, Any]:
    """Create empty data response for error cases."""
    return {
        "nodes": {},
        "agents": {},
        "builders": {},
        "metrics": {},
        "intents": {},
        "matches": []
    }


def validate_api_data(data: Dict[str, Any]) -> bool:
    """
    Validate API data structure and content.
    
    Args:
        data: API response data dictionary
    
    Returns:
        True if data is valid, False otherwise
    """
    if not isinstance(data, dict):
        return False
    
    required_keys = ["nodes", "agents", "builders", "metrics", "intents", "matches"]
    return all(key in data for key in required_keys)


def safe_extract_data(data: Dict[str, Any], key: str, default_value: Any) -> Any:
    """
    Safely extract data with validation and default fallback.
    
    Args:
        data: Source data dictionary
        key: Key to extract
        default_value: Default value if extraction fails
    
    Returns:
        Extracted value or default
    """
    try:
        value = data.get(key, default_value)
        if value is None:
            return default_value
        return value
    except (AttributeError, KeyError, TypeError):
        return default_value


def process_dashboard_data(data: Dict[str, Any]) -> tuple:
    """
    Process raw API data into dashboard components with comprehensive validation.
    
    Returns:
        Tuple of (dashboard_metrics, nodes_data, agents_data, builders_data, 
                 intents_data, matches_data, network_data)
    """
    # Validate input data
    if not validate_api_data(data):
        st.warning("API数据格式无效，使用默认值")
        return (
            create_empty_dashboard_metrics(),
            {},
            {},
            {},
            {},
            [],
            None
        )
    
    # Safely extract node status data
    nodes_data = safe_extract_data(data, "nodes", {})
    
    # Safely extract metrics data and aggregate
    metrics_data = safe_extract_data(data, "metrics", {})
    dashboard_metrics = aggregate_execution_metrics(metrics_data)
    
    # Count active nodes with validation
    try:
        active_nodes = 0
        for node in nodes_data.values():
            if (hasattr(node, 'is_running') and hasattr(node, 'error') and 
                node.is_running and not node.error):
                active_nodes += 1
        dashboard_metrics.active_nodes = active_nodes
    except (AttributeError, TypeError):
        dashboard_metrics.active_nodes = 0
    
    # Calculate deltas from previous metrics with validation
    try:
        previous = st.session_state.previous_metrics
        dashboard_metrics.delta_nodes = calculate_delta(dashboard_metrics.active_nodes, previous.active_nodes)
        dashboard_metrics.delta_intents = calculate_delta(dashboard_metrics.total_intents, previous.total_intents)
        dashboard_metrics.delta_bids = calculate_delta(dashboard_metrics.active_bids, previous.active_bids)
        dashboard_metrics.delta_matches = calculate_delta(dashboard_metrics.completed_matches, previous.completed_matches)
        
        # Update previous metrics for next comparison
        st.session_state.previous_metrics = DashboardMetrics(
            active_nodes=dashboard_metrics.active_nodes,
            total_intents=dashboard_metrics.total_intents,
            active_bids=dashboard_metrics.active_bids,
            completed_matches=dashboard_metrics.completed_matches
        )
    except (AttributeError, TypeError):
        # If delta calculation fails, set to 0
        dashboard_metrics.delta_nodes = 0
        dashboard_metrics.delta_intents = 0
        dashboard_metrics.delta_bids = 0
        dashboard_metrics.delta_matches = 0
    
    # Safely extract other data
    agents_data = safe_extract_data(data, "agents", {})
    builders_data = safe_extract_data(data, "builders", {})
    intents_data = safe_extract_data(data, "intents", {})
    matches_data = safe_extract_data(data, "matches", [])
    
    # Validate matches_data is a list
    if not isinstance(matches_data, list):
        matches_data = []
    
    # Create P2P network info from metrics with validation
    network_data = None
    try:
        if metrics_data and isinstance(metrics_data, dict):
            # Use metrics from first available node for network info
            for node_metrics in metrics_data.values():
                if (hasattr(node_metrics, 'error') and not node_metrics.error and 
                    hasattr(node_metrics, 'p2p_peers_connected')):
                    network_data = create_p2p_network_info_from_metrics(node_metrics)
                    break
    except (AttributeError, TypeError):
        network_data = None
    
    return (
        dashboard_metrics,
        nodes_data,
        agents_data,
        builders_data,
        intents_data,
        matches_data,
        network_data
    )


def render_dashboard() -> None:
    """Render the main dashboard with all panels and enhanced error handling."""
    # Display header
    st.title("PIN Intent Network - Real-time Monitoring Dashboard")
    
    # Initialize AutoRefreshManager if not exists
    if "refresh_manager" not in st.session_state:
        st.session_state.refresh_manager = AutoRefreshManager(REFRESH_INTERVAL_SECONDS)
    
    refresh_manager = st.session_state.refresh_manager
    
    try:
        # Fetch all data with progress indication
        with st.spinner("Fetching PIN node data..."):
            data = fetch_all_data()
        
        # Validate data before processing
        if not validate_api_data(data):
            st.error("API data validation failed, showing fallback dashboard")
            render_fallback_dashboard()
            return
        
        # Process data with enhanced validation
        (
            dashboard_metrics,
            nodes_data,
            agents_data,
            builders_data,
            intents_data,
            matches_data,
            network_data
        ) = process_dashboard_data(data)
        
        # Check if we have any valid data
        has_valid_data = (
            dashboard_metrics.active_nodes > 0 or
            bool(nodes_data) or
            bool(agents_data) or
            bool(builders_data)
        )
        
        if not has_valid_data:
            st.warning("No valid data available, please check PIN node status")
            render_fallback_dashboard()
            return
        
        # System metrics overview at the top
        st.subheader("📊 System Metrics Overview")
        render_top_metrics(dashboard_metrics)
        
        st.markdown("---")
        
        # Single column layout with components in specified order
        
        # � Iintent Flow Monitoring
        st.subheader("📡 Intent Flow Monitoring")
        render_component_with_error_handling("Intent Flow Monitoring", render_intent_monitoring_panel, intents_data)
        
        st.markdown("---")
        
        # 💰 Bidding Activity Tracking
        st.subheader("💰 Bidding Activity Tracking")
        render_component_with_error_handling("Bidding Activity Tracking", render_bidding_activity_panel, agents_data)
        
        st.markdown("---")
        
        # 🎯 Matching Results
        st.subheader("🎯 Matching Results")
        render_component_with_error_handling("Matching Results", render_matching_results_panel, matches_data)
        
        st.markdown("---")
        
        # 🖥️ Node Status Overview
        st.subheader("🖥️ Node Status Overview")
        render_component_with_error_handling("Node Status Overview", render_nodes_status_panel, nodes_data)
        
        st.markdown("---")
        
        # 🌐 P2P Network Status
        st.subheader("🌐 P2P Network Status")
        render_component_with_error_handling("P2P Network Status", render_p2p_network_panel, network_data)
        
        st.markdown("---")
        
        # Refresh controls at the bottom
        st.subheader("🔄 Auto-Refresh Controls")
        refresh_manager.render_indicator()
        
        # Store refresh status for main() function to handle
        st.session_state.should_refresh = refresh_manager.should_refresh()
        
        # Render sidebar
        render_sidebar_info(st.session_state.ui_state, dashboard_metrics)
        
        # Reset error count on successful refresh
        st.session_state.error_count = 0
        
    except Exception as e:
        st.error(f"Dashboard error: {str(e)}")
        st.session_state.error_count += 1
        
        # Show fallback content after multiple errors
        if st.session_state.error_count > 3:
            st.warning("Multiple connection failures, showing offline mode...")
        else:
            st.info("Showing cached data or default values...")
        
        render_fallback_dashboard()


def render_fallback_dashboard() -> None:
    """Render fallback dashboard when data fetching fails."""
    st.warning("⚠️ Unable to connect to PIN nodes. Showing fallback information.")
    
    # Show basic node information
    st.subheader("Expected Node Configuration")
    
    col1, col2 = st.columns(2)
    
    with col1:
        st.markdown("""
        **Node 1 (8100):** Intent Publisher
        - Publishes intents and provides API services
        - Status: Unknown
        
        **Node 2 (8101):** Service Agent 1 (Trading)
        - Trading agent with automatic bidding
        - Status: Unknown
        """)
    
    with col2:
        st.markdown("""
        **Node 3 (8102):** Service Agent 2 (Data)
        - Data agent with automatic bidding
        - Status: Unknown
        
        **Node 4 (8103):** Block Builder
        - Intent matching coordinator
        - Status: Unknown
        """)
    
    st.markdown("---")
    
    # Connection status with retry button
    col1, col2, col3 = st.columns([2, 1, 2])
    with col2:
        if st.button("🔄 Retry Connection", key="retry_connection"):
            st.session_state.error_count = 0
            st.rerun()
    
    st.info("""
    **Troubleshooting Guide:**
    1. Ensure all 4 PIN nodes are running
    2. Check nodes are accessible on ports 8100-8103
    3. Verify network connection
    4. Run `./scripts/automation/start_automation_test.sh` to start automation system
    5. Check if ports are occupied by other processes
    """)
    
    # Show system requirements
    with st.expander("🔧 System Requirements and Checks"):
        st.markdown("""
        **Start PIN System:**
        ```bash
        # Start complete 4-node automation test
        ./scripts/automation/start_automation_test.sh
        
        # Check node status
        ./scripts/automation/monitor_automation.sh
        
        # Check port usage
        netstat -tulpn | grep :810
        ```
        
        **Manual API Testing:**
        ```bash
        # Test node health status
        curl http://localhost:8100/health
        curl http://localhost:8101/health
        curl http://localhost:8102/health
        curl http://localhost:8103/health
        ```
        """)



def main() -> None:
    """Main Streamlit application entry point."""
    # Setup page configuration
    setup_page_config()
    
    # Initialize session state
    initialize_session_state()
    
    # Render main dashboard
    render_dashboard()
    
    # Handle auto-refresh at the end of main() to ensure it's not interrupted
    if hasattr(st.session_state, 'should_refresh') and st.session_state.should_refresh:
        if "refresh_manager" in st.session_state:
            st.session_state.refresh_manager.last_refresh = time.time()
        time.sleep(0.1)  # Small delay to ensure UI updates
        st.rerun()
    
    # Add footer information
    st.markdown("---")
    with st.expander("ℹ️ About PIN Intent Network Demo"):
        st.markdown("""
        This dashboard provides real-time monitoring of the PIN (P2P Intent Network) automation system:
        
        **Architecture:**
        - **Node 1 (8100):** Intent Publisher - Creates and broadcasts intents
        - **Node 2 (8101):** Service Agent 1 (Trading) - Autonomous bidding entity
        - **Node 3 (8102):** Service Agent 2 (Data) - Autonomous bidding entity  
        - **Node 4 (8103):** Block Builder - Intent matching coordinator
        
        **Features:**
        - 🔄 Auto-refresh every 5 seconds
        - 📡 Real-time intent flow monitoring
        - 💰 Bidding activity tracking
        - 🎯 Matching results visualization
        - 🌐 P2P network status
        - 📊 Performance metrics
        
        **Usage:**
        1. Start the PIN automation system: `./scripts/automation/start_automation_test.sh`
        2. Launch this dashboard: `./scripts/start_streamlit_ui.sh`
        3. Monitor the complete intent → bidding → matching flow
        
        **Data Source:** HTTP APIs from PIN nodes (ports 8100-8103)
        """)


if __name__ == "__main__":
    main()