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
    render_error_panel, render_refresh_indicator
)
from utils import get_system_health_score, calculate_delta


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
    st.title(UI_TEXT["dashboard_title"])
    render_refresh_indicator()
    
    # Check if it's time to refresh
    ui_state = st.session_state.ui_state
    current_time = time.time()
    
    # Force refresh every 5 seconds
    if current_time - ui_state.last_refresh >= REFRESH_INTERVAL_SECONDS:
        ui_state.last_refresh = current_time
        st.rerun()
    
    try:
        # Fetch all data with progress indication
        with st.spinner("正在获取PIN节点数据..."):
            data = fetch_all_data()
        
        # Validate data before processing
        if not validate_api_data(data):
            st.error("API数据验证失败，显示备用仪表板")
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
            st.warning("暂无有效数据，请检查PIN节点状态")
            render_fallback_dashboard()
            return
        
        # Render top metrics
        render_top_metrics(dashboard_metrics)
        
        st.markdown("---")
        
        # Main content layout
        left_col, right_col = st.columns([3, 2])
        
        # Left column panels
        with left_col:
            st.subheader(UI_TEXT["node_status_title"])
            render_nodes_status_panel(nodes_data)
            
            st.markdown("---")
            st.subheader(UI_TEXT["intent_monitoring_title"])
            render_intent_monitoring_panel(intents_data)
            
            st.markdown("---")
            st.subheader(UI_TEXT["bidding_activity_title"])
            render_bidding_activity_panel(agents_data)
        
        # Right column panels
        with right_col:
            st.subheader(UI_TEXT["matching_results_title"])
            render_matching_results_panel(matches_data)
            
            st.markdown("---")
            st.subheader(UI_TEXT["p2p_network_title"])
            render_p2p_network_panel(network_data)
            
            st.markdown("---")
            st.subheader(UI_TEXT["performance_metrics_title"])
            render_performance_metrics_panel(data.get("metrics", {}))
        
        # Render sidebar
        render_sidebar_info(ui_state, dashboard_metrics)
        
        # Reset error count on successful refresh
        st.session_state.error_count = 0
        
    except Exception as e:
        st.error(f"仪表板错误：{str(e)}")
        st.session_state.error_count += 1
        
        # Show fallback content after multiple errors
        if st.session_state.error_count > 3:
            st.warning("多次连接失败，显示离线模式...")
        else:
            st.info("显示缓存数据或默认值...")
        
        render_fallback_dashboard()


def render_fallback_dashboard() -> None:
    """Render fallback dashboard when data fetching fails."""
    st.warning("⚠️ 无法连接到PIN节点。显示备用信息。")
    
    # Show basic node information
    st.subheader("预期节点配置")
    
    col1, col2 = st.columns(2)
    
    with col1:
        st.markdown("""
        **节点1 (8100):** Intent发布者
        - 发布意图并提供API服务
        - 状态: 未知
        
        **节点2 (8101):** 服务代理1 (交易)
        - 交易代理，具有自动竞标功能
        - 状态: 未知
        """)
    
    with col2:
        st.markdown("""
        **节点3 (8102):** 服务代理2 (数据)
        - 数据代理，具有自动竞标功能
        - 状态: 未知
        
        **节点4 (8103):** 区块构建者
        - 意图匹配协调器
        - 状态: 未知
        """)
    
    st.markdown("---")
    
    # Connection status with retry button
    col1, col2, col3 = st.columns([2, 1, 2])
    with col2:
        if st.button("🔄 重试连接", key="retry_connection"):
            st.session_state.error_count = 0
            st.rerun()
    
    st.info("""
    **故障排除指南：**
    1. 确保所有4个PIN节点都在运行
    2. 检查节点在端口8100-8103上可访问
    3. 验证网络连接
    4. 运行 `./scripts/automation/start_automation_test.sh` 启动自动化系统
    5. 检查端口是否被其他进程占用
    """)
    
    # Show system requirements
    with st.expander("🔧 系统要求和检查"):
        st.markdown("""
        **启动PIN系统：**
        ```bash
        # 启动完整的4节点自动化测试
        ./scripts/automation/start_automation_test.sh
        
        # 检查节点状态
        ./scripts/automation/monitor_automation.sh
        
        # 检查端口占用
        netstat -tulpn | grep :810
        ```
        
        **手动API测试：**
        ```bash
        # 测试节点健康状态
        curl http://localhost:8100/health
        curl http://localhost:8101/health
        curl http://localhost:8102/health
        curl http://localhost:8103/health
        ```
        """)


def handle_auto_refresh() -> None:
    """Handle automatic refresh mechanism."""
    # Create placeholder for refresh countdown
    refresh_placeholder = st.empty()
    
    ui_state = st.session_state.ui_state
    current_time = time.time()
    time_since_refresh = current_time - ui_state.last_refresh
    time_until_refresh = REFRESH_INTERVAL_SECONDS - time_since_refresh
    
    if time_until_refresh <= 0:
        # Time to refresh
        ui_state.last_refresh = current_time
        st.rerun()
    else:
        # Show countdown
        countdown_text = f"Next refresh in {int(time_until_refresh)}s"
        refresh_placeholder.markdown(
            f'<div class="refresh-indicator">{countdown_text}</div>',
            unsafe_allow_html=True
        )


def main() -> None:
    """Main Streamlit application entry point."""
    # Setup page configuration
    setup_page_config()
    
    # Initialize session state
    initialize_session_state()
    
    # Handle auto-refresh
    handle_auto_refresh()
    
    # Render main dashboard
    render_dashboard()
    
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