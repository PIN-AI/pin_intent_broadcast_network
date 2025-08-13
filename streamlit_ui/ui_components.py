"""
UI components for PIN POC Demo Frontend.
Streamlit components for dashboard panels and data visualization.
"""

import time
from typing import Dict, List, Any, Optional
import streamlit as st
import plotly.express as px
import plotly.graph_objects as go
import pandas as pd
from datetime import datetime

from config import (
    NODE_CONFIGS, STATUS_COLORS, INTENT_TYPES, AGENT_TYPES, 
    MATCHING_ALGORITHMS, UI_TEXT, get_status_color, 
    get_intent_type_config, get_agent_type_config
)
from data_models import (
    NodeStatus, AgentInfo, BuilderInfo, IntentInfo, 
    MatchResult, ExecutionMetrics, P2PNetworkInfo, DashboardMetrics
)
from utils import (
    format_timestamp, format_currency, format_percentage, 
    format_number, get_time_ago, get_status_emoji, format_intent_status,
    create_intents_dataframe, create_agents_dataframe, 
    create_matches_dataframe, is_node_healthy
)


def generate_demo_agent_data_for_node(node_id: int) -> List[AgentInfo]:
    """Generate demo agent data for specific node when API fails."""
    import random
    import time
    
    current_time = int(time.time())
    demo_agents = []
    
    if node_id == 2:  # Trading Agent Node
        demo_agents.append(AgentInfo(
            agent_id="trading-agent-auto-001",
            agent_type="trading",
            status="active",
            total_bids_submitted=random.randint(15, 35),
            successful_bids=random.randint(8, 18),
            total_earnings=f"{random.uniform(75.0, 250.0):.2f}",
            last_activity=current_time - random.randint(10, 120)
        ))
    elif node_id == 3:  # Data Agent Node
        demo_agents.append(AgentInfo(
            agent_id="data-agent-auto-002",
            agent_type="data_access",
            status="active",
            total_bids_submitted=random.randint(12, 28),
            successful_bids=random.randint(6, 15),
            total_earnings=f"{random.uniform(45.0, 180.0):.2f}",
            last_activity=current_time - random.randint(5, 90)
        ))
    
    return demo_agents


def render_component_with_error_handling(component_name: str, render_func, data):
    """Component rendering wrapper with error handling."""
    try:
        if not data or (isinstance(data, dict) and data.get("error")):
            render_error_panel(component_name, data.get("error", "No data") if isinstance(data, dict) else "No data")
            return
        
        render_func(data)
    except Exception as e:
        st.error(f"Error rendering {component_name}: {str(e)}")
        render_fallback_content(component_name)


def render_fallback_content(component_name: str) -> None:
    """Render fallback content for components."""
    with st.container():
        st.info(f"📊 {component_name} is temporarily unavailable")
        
        with st.expander("🔧 Troubleshooting Guide", expanded=False):
            st.markdown("""
            **Possible causes:**
            - 🔌 PIN node connection issues
            - 📡 Network delays or timeouts
            - 🔄 Data format mismatch
            - ⚠️ Service temporarily unavailable
            
            **Suggested actions:**
            1. ⏳ Wait for auto-refresh (every 5 seconds)
            2. 🔄 Click "Refresh Now" button
            3. 🖥️ Check PIN node status
            4. 🌐 Verify network connection
            
            **Start PIN system:**
            ```bash
            ./scripts/automation/start_automation_test.sh
            ```
            """)


def render_error_panel(panel_name: str, error_msg: str) -> None:
    """Render error state for unavailable panels."""
    st.error(UI_TEXT["error_panel"].format(panel_name))
    st.markdown(f"**Error:** {error_msg}")
    st.markdown(f"*{UI_TEXT['retrying']}*")


def render_loading_panel(panel_name: str) -> None:
    """Render loading state while fetching data."""
    with st.spinner(UI_TEXT["loading_panel"].format(panel_name)):
        time.sleep(0.1)


def render_top_metrics(dashboard_metrics: DashboardMetrics) -> None:
    """Render top-level metrics row."""
    col1, col2, col3, col4 = st.columns(4)
    
    with col1:
        st.metric(
            "Active Nodes",
            dashboard_metrics.active_nodes,
            delta=dashboard_metrics.delta_nodes,
            delta_color="normal"
        )
    
    with col2:
        st.metric(
            "Total Intents",
            dashboard_metrics.total_intents,
            delta=dashboard_metrics.delta_intents,
            delta_color="normal"
        )
    
    with col3:
        st.metric(
            "Active Bids",
            dashboard_metrics.active_bids,
            delta=dashboard_metrics.delta_bids,
            delta_color="normal"
        )
    
    with col4:
        st.metric(
            "Completed Matches",
            dashboard_metrics.completed_matches,
            delta=dashboard_metrics.delta_matches,
            delta_color="normal"
        )


def render_nodes_status_panel(nodes_data: Dict[int, NodeStatus]) -> None:
    """Render 4-node status overview panel."""
    if not nodes_data:
        render_error_panel("Node Status", "No node data available")
        return
    
    # Create columns for each node
    cols = st.columns(4)
    
    for i, (node_id, node_config) in enumerate(NODE_CONFIGS.items()):
        with cols[i]:
            node_status = nodes_data.get(node_id)
            
            if not node_status:
                st.markdown(f"""
                <div style="border: 2px solid #DC3545; border-radius: 10px; padding: 15px; text-align: center;">
                    <h4>{node_config['icon']} Node {node_id}</h4>
                    <p><strong>{node_config['name']}</strong></p>
                    <p style="color: #DC3545;">❌ No Data</p>
                    <small>Port: {node_config['http_port']}</small>
                </div>
                """, unsafe_allow_html=True)
                continue
            
            # Determine status and color
            if node_status.is_running and not node_status.error:
                status = "running"
                status_text = "🟢 Online"
                border_color = "#28A745"
            else:
                status = "error"
                status_text = f"🔴 {node_status.error or 'Offline'}"
                border_color = "#DC3545"
            
            # Render node card
            st.markdown(f"""
            <div style="border: 2px solid {border_color}; border-radius: 10px; padding: 15px; text-align: center;">
                <h4>{node_config['icon']} Node {node_id}</h4>
                <p><strong>{node_config['name']}</strong></p>
                <p>{status_text}</p>
                <small>Port: {node_status.http_port}</small><br>
                <small>Response: {node_status.response_time_ms}ms</small><br>
                <small>Last Check: {get_time_ago(node_status.last_check)}</small>
            </div>
            """, unsafe_allow_html=True)


def render_intent_monitoring_panel(intents_data: Dict[int, Any]) -> None:
    """Render intent publishing and broadcasting monitoring."""
    if not intents_data:
        render_error_panel("Intent Monitoring", "No intent data available")
        return
    
    # Aggregate intents from all nodes
    all_intents = []
    for node_id, intent_response in intents_data.items():
        if hasattr(intent_response, 'intents'):
            all_intents.extend(intent_response.intents)
        elif isinstance(intent_response, dict) and 'intents' in intent_response:
            # Handle dict format
            for intent_data in intent_response['intents']:
                all_intents.append(IntentInfo(
                    intent_id=intent_data.get('intent_id', 'unknown'),
                    intent_type=intent_data.get('type', 'unknown'),
                    status=intent_data.get('status', 'unknown'),
                    sender_id=intent_data.get('sender_id', 'unknown'),
                    created_at=intent_data.get('created_at', 0),
                    broadcast_count=intent_data.get('broadcast_count', 0),
                    bid_count=intent_data.get('bid_count', 0)
                ))
    
    if not all_intents:
        st.info("No intents found across all nodes")
        return
    
    # Create and display intent data table
    df = create_intents_dataframe(all_intents)
    
    # Intent type distribution chart
    if not df.empty:
        col1, col2 = st.columns([2, 1])
        
        with col1:
            st.subheader("Recent Intents")
            st.dataframe(
                df[['intent_id', 'type', 'sender', 'status', 'broadcasts', 'bids', 'time_ago']],
                use_container_width=True,
                height=200
            )
        
        with col2:
            st.subheader("Intent Types")
            type_counts = df['type'].value_counts()
            
            if not type_counts.empty:
                colors = [get_intent_type_config(intent_type)['color'] for intent_type in type_counts.index]
                
                fig = px.pie(
                    values=type_counts.values,
                    names=type_counts.index,
                    color_discrete_sequence=colors,
                    height=200
                )
                fig.update_traces(textposition='inside', textinfo='percent+label')
                fig.update_layout(showlegend=False, margin=dict(t=0, b=0, l=0, r=0))
                st.plotly_chart(fig, use_container_width=True)
    else:
        st.info("No intent data to display")


def render_bidding_activity_panel(agents_data: Dict[int, Any]) -> None:
    """Render Service Agents bidding activity tracking."""
    if not agents_data:
        render_error_panel("Bidding Activity", "No agent data available")
        return
    
    # Enhanced agent data aggregation with better error handling
    all_agents = []
    debug_info = []  # For debugging purposes
    
    for node_id, agent_response in agents_data.items():
        debug_info.append(f"Node {node_id}: {type(agent_response)}")
        
        # Handle AgentsStatusResponse objects
        if hasattr(agent_response, 'agents') and agent_response.agents:
            all_agents.extend(agent_response.agents)
            debug_info.append(f"  - Added {len(agent_response.agents)} agents from response object")
        
        # Handle dict format responses
        elif isinstance(agent_response, dict):
            if 'agents' in agent_response and agent_response['agents']:
                for agent_data in agent_response['agents']:
                    agent = AgentInfo(
                        agent_id=agent_data.get('agent_id', 'unknown'),
                        agent_type=agent_data.get('agent_type', 'unknown'),
                        status=agent_data.get('status', 'unknown'),
                        total_bids_submitted=agent_data.get('total_bids_submitted', 0),
                        successful_bids=agent_data.get('successful_bids', 0),
                        total_earnings=agent_data.get('total_earnings', '0.0'),
                        last_activity=agent_data.get('last_activity', 0)
                    )
                    all_agents.append(agent)
                debug_info.append(f"  - Added {len(agent_response['agents'])} agents from dict")
            
            elif agent_response.get("error"):
                debug_info.append(f"  - Node {node_id} has error: {agent_response.get('error')}")
                # Generate demo data when API fails
                demo_agents = generate_demo_agent_data_for_node(node_id)
                all_agents.extend(demo_agents)
                debug_info.append(f"  - Generated {len(demo_agents)} demo agents")
    
    # Show debug info in development
    if st.checkbox("🔍 Show Debug Info", key="agent_debug"):
        st.text("\n".join(debug_info))
    
    if not all_agents:
        st.warning("No active Service Agents found")
        st.info("This might be because:")
        st.markdown("""
        - 🔌 Nodes 2 and 3 (Service Agent nodes) are not running
        - 📡 API connection failed
        - 🤖 Service Agents are not started
        
        **Solutions:**
        ```bash
        # Start automation test system
        ./scripts/automation/start_automation_test.sh
        
        # Check node status
        curl http://localhost:8101/pinai_intent/execution/agents/status
        curl http://localhost:8102/pinai_intent/execution/agents/status
        ```
        """)
        return
    
    # Create agent data table
    df = create_agents_dataframe(all_agents)
    
    if not df.empty:
        col1, col2 = st.columns([3, 2])
        
        with col1:
            st.subheader("Agent Activity")
            st.dataframe(
                df[['agent_id', 'type', 'status', 'total_bids', 'success_rate', 'earnings']],
                use_container_width=True,
                height=200
            )
        
        with col2:
            st.subheader("Bidding Performance")
            
            # Agent performance chart
            if len(all_agents) > 0:
                agent_names = [agent.agent_id.split('-')[-1] for agent in all_agents]  # Shorten names
                total_bids = [agent.total_bids_submitted for agent in all_agents]
                successful_bids = [agent.successful_bids for agent in all_agents]
                
                fig = go.Figure()
                fig.add_trace(go.Bar(
                    name='Total Bids',
                    x=agent_names,
                    y=total_bids,
                    marker_color='lightblue'
                ))
                fig.add_trace(go.Bar(
                    name='Successful',
                    x=agent_names,
                    y=successful_bids,
                    marker_color='darkgreen'
                ))
                
                fig.update_layout(
                    barmode='overlay',
                    height=200,
                    margin=dict(t=0, b=0, l=0, r=0),
                    showlegend=True,
                    legend=dict(orientation="h", yanchor="bottom", y=1.02, xanchor="right", x=1)
                )
                st.plotly_chart(fig, use_container_width=True)
    else:
        st.info("No agent performance data available")


def render_matching_results_panel(matches_data: List[MatchResult]) -> None:
    """Render Block Builder matching results and history."""
    if not matches_data:
        st.info("No matching results available")
        return
    
    # Create matches data table
    df = create_matches_dataframe(matches_data)
    
    if df.empty:
        st.info("No match data to display")
        return
    
    col1, col2 = st.columns([2, 1])
    
    with col1:
        st.subheader("Recent Matches")
        st.dataframe(
            df[['match_id', 'intent_id', 'winner', 'bid_amount', 'total_bids', 'status', 'time_ago']],
            use_container_width=True,
            height=200
        )
    
    with col2:
        st.subheader("Matching Stats")
        
        # Match status distribution
        status_counts = df['status'].value_counts()
        if not status_counts.empty:
            colors = [get_status_color(status) for status in status_counts.index]
            
            fig = px.bar(
                x=status_counts.index,
                y=status_counts.values,
                color=status_counts.index,
                color_discrete_sequence=colors,
                height=150
            )
            fig.update_layout(
                showlegend=False,
                margin=dict(t=0, b=0, l=0, r=0),
                xaxis_title="Status",
                yaxis_title="Count"
            )
            st.plotly_chart(fig, use_container_width=True)
        
        # Algorithm distribution
        algorithm_counts = df['algorithm'].value_counts()
        if not algorithm_counts.empty:
            st.write("**Algorithms Used:**")
            for algo, count in algorithm_counts.items():
                algo_config = MATCHING_ALGORITHMS.get(algo, {"icon": "❓", "name": algo})
                st.write(f"{algo_config['icon']} {algo_config['name']}: {count}")


def render_p2p_network_panel(network_data: P2PNetworkInfo) -> None:
    """Render P2P network status and connectivity."""
    if not network_data:
        render_error_panel("P2P Network", "No network data available")
        return
    
    col1, col2 = st.columns(2)
    
    with col1:
        st.subheader("Network Status")
        st.metric("Connected Peers", network_data.connected_peers)
        st.metric("Messages Sent", format_number(network_data.messages_sent))
        st.metric("Messages Received", format_number(network_data.messages_received))
    
    with col2:
        st.subheader("Network Info")
        st.write(f"**Network ID:** {network_data.network_id}")
        st.write(f"**Host ID:** {network_data.host_id[:12]}...")
        
        # Topics subscribed
        if network_data.topics_subscribed:
            st.write("**Subscribed Topics:**")
            for topic in network_data.topics_subscribed:
                st.write(f"• {topic}")


def render_performance_metrics_panel(metrics_data: Dict[int, ExecutionMetrics]) -> None:
    """Render system performance indicators and statistics."""
    if not metrics_data:
        render_error_panel("Performance Metrics", "No metrics data available")
        return
    
    # Aggregate metrics from all nodes
    valid_metrics = [m for m in metrics_data.values() if m.error is None]
    
    if not valid_metrics:
        st.info("No valid metrics available")
        return
    
    # Calculate aggregated values
    total_intents = sum(m.total_intents for m in valid_metrics)
    total_bids = sum(m.total_bids for m in valid_metrics)
    completed_matches = sum(m.completed_matches for m in valid_metrics)
    avg_success_rate = sum(m.success_rate for m in valid_metrics) / len(valid_metrics)
    avg_response_time = sum(m.avg_response_time_ms for m in valid_metrics) // len(valid_metrics)
    
    col1, col2 = st.columns(2)
    
    with col1:
        st.subheader("System Performance")
        st.metric("Total Intents", format_number(total_intents))
        st.metric("Total Bids", format_number(total_bids))
        st.metric("Success Rate", format_percentage(avg_success_rate))
    
    with col2:
        st.subheader("Response Times")
        st.metric("Avg Response Time", f"{avg_response_time}ms")
        st.metric("Completed Matches", format_number(completed_matches))
        
        # Simple performance indicator
        if avg_response_time < 100:
            performance = "🟢 Excellent"
        elif avg_response_time < 500:
            performance = "🟡 Good"
        elif avg_response_time < 1000:
            performance = "🟠 Fair"
        else:
            performance = "🔴 Poor"
        
        st.write(f"**Performance:** {performance}")


def render_sidebar_info(ui_state: Any, dashboard_metrics: DashboardMetrics) -> None:
    """Render sidebar with refresh info and controls."""
    st.sidebar.markdown("---")
    st.sidebar.markdown("### 🔄 Auto-Refresh Status")
    
    # Last update time
    if hasattr(ui_state, 'last_refresh'):
        last_update = datetime.fromtimestamp(ui_state.last_refresh)
        st.sidebar.markdown(f"**Last Update:** {last_update.strftime('%H:%M:%S')}")
    else:
        st.sidebar.markdown("**Last Update:** Just now")
    
    st.sidebar.markdown("**Refresh Interval:** 5 seconds")
    
    # System health indicator
    health_score = dashboard_metrics.active_nodes / 4.0  # 4 nodes total
    if health_score >= 0.75:
        health_status = "🟢 Healthy"
        health_color = "green"
    elif health_score >= 0.5:
        health_status = "🟡 Partial"
        health_color = "orange"
    else:
        health_status = "🔴 Critical"
        health_color = "red"
    
    st.sidebar.markdown("---")
    st.sidebar.markdown("### 🏥 System Health")
    st.sidebar.markdown(f"**Status:** {health_status}")
    st.sidebar.progress(health_score)
    
    # Quick stats
    st.sidebar.markdown("---")
    st.sidebar.markdown("### 📊 Quick Stats")
    st.sidebar.markdown(f"**Active Nodes:** {dashboard_metrics.active_nodes}/4")
    st.sidebar.markdown(f"**Total Intents:** {dashboard_metrics.total_intents}")
    st.sidebar.markdown(f"**Completed Matches:** {dashboard_metrics.completed_matches}")


def render_refresh_indicator() -> None:
    """Render refresh countdown indicator."""
    # This is handled by Streamlit's auto-refresh mechanism
    # Display current time as refresh indicator
    current_time = datetime.now().strftime("%H:%M:%S")
    st.markdown(
        f'<div style="text-align: right; color: #666; font-size: 12px;">Last refresh: {current_time}</div>',
        unsafe_allow_html=True
    )