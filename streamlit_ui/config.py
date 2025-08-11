"""
Configuration settings for PIN POC Demo Frontend.
Defines node configurations, API endpoints, and UI settings.
"""

from typing import Dict, Any


# Node configurations for the 4-node PIN automation system
NODE_CONFIGS: Dict[int, Dict[str, Any]] = {
    1: {
        "name": "Intent Publisher",
        "type": "PUBLISHER",
        "http_port": 8100,
        "base_url": "http://localhost:8100",
        "description": "Publishes intents and provides API services",
        "color": "#FF6B6B",  # Red
        "icon": "📡"
    },
    2: {
        "name": "Service Agent 1 (Trading)",
        "type": "SERVICE_AGENT",
        "http_port": 8101,
        "base_url": "http://localhost:8101",
        "description": "Trading agent with automatic bidding",
        "color": "#4ECDC4",  # Teal
        "icon": "💼"
    },
    3: {
        "name": "Service Agent 2 (Data)",
        "type": "SERVICE_AGENT",
        "http_port": 8102,
        "base_url": "http://localhost:8102",
        "description": "Data agent with automatic bidding",
        "color": "#45B7D1",  # Blue
        "icon": "📊"
    },
    4: {
        "name": "Block Builder",
        "type": "BLOCK_BUILDER",
        "http_port": 8103,
        "base_url": "http://localhost:8103",
        "description": "Intent matching coordinator",
        "color": "#96CEB4",  # Green
        "icon": "🎯"
    }
}

# API configuration
API_TIMEOUT_SECONDS = 3
MAX_RETRIES = 2
RETRY_DELAY_SECONDS = 1

# UI configuration
REFRESH_INTERVAL_SECONDS = 5
MAX_HISTORY_ITEMS = 20
PAGE_TITLE = "PIN Intent Network - POC Demo"
PAGE_ICON = "🌐"

# Streamlit configuration
STREAMLIT_CONFIG = {
    "page_title": PAGE_TITLE,
    "page_icon": PAGE_ICON,
    "layout": "wide",
    "initial_sidebar_state": "expanded",
    "menu_items": {
        "Get Help": None,
        "Report a Bug": None,
        "About": "PIN Intent Network POC Demo Dashboard"
    }
}

# Data visualization configuration
CHART_CONFIG = {
    "metrics_chart_height": 300,
    "intent_flow_height": 250,
    "bidding_activity_height": 200,
    "matching_results_height": 200,
    "network_status_height": 150
}

# Status color mapping
STATUS_COLORS = {
    "running": "#28A745",     # Green
    "active": "#28A745",      # Green
    "idle": "#FFC107",        # Yellow
    "stopped": "#DC3545",     # Red
    "error": "#DC3545",       # Red
    "offline": "#6C757D",     # Gray
    "pending": "#17A2B8",     # Info blue
    "completed": "#28A745",   # Green
    "failed": "#DC3545"       # Red
}

# Intent type configurations
INTENT_TYPES = {
    "exchange": {
        "name": "Exchange",
        "color": "#FF6B6B",
        "icon": "🔄"
    },
    "data": {
        "name": "Data Processing",
        "color": "#4ECDC4",
        "icon": "📊"
    },
    "compute": {
        "name": "Compute Task",
        "color": "#45B7D1",
        "icon": "⚡"
    },
    "storage": {
        "name": "Storage Request",
        "color": "#96CEB4",
        "icon": "💾"
    }
}

# Agent type configurations
AGENT_TYPES = {
    "trading": {
        "name": "Trading Agent",
        "color": "#FF6B6B",
        "icon": "💼"
    },
    "data": {
        "name": "Data Agent",
        "color": "#4ECDC4",
        "icon": "📊"
    },
    "compute": {
        "name": "Compute Agent",
        "color": "#45B7D1",
        "icon": "⚡"
    },
    "storage": {
        "name": "Storage Agent",
        "color": "#96CEB4",
        "icon": "💾"
    }
}

# Matching algorithm configurations
MATCHING_ALGORITHMS = {
    "highest_bid": {
        "name": "Highest Bid",
        "description": "Selects agent with highest bid amount",
        "icon": "📈"
    },
    "reputation_weighted": {
        "name": "Reputation Weighted",
        "description": "Combines bid amount with agent reputation",
        "icon": "⭐"
    },
    "random": {
        "name": "Random Selection",
        "description": "Fair random selection among valid bids",
        "icon": "🎲"
    }
}

# P2P network configuration
P2P_CONFIG = {
    "bootstrap_peers": [
        "/ip4/127.0.0.1/tcp/4001",
        "/ip4/127.0.0.1/tcp/4002",
        "/ip4/127.0.0.1/tcp/4003",
        "/ip4/127.0.0.1/tcp/4004"
    ],
    "topics": [
        "intent-broadcast.*",
        "intent-network/bids/1.0.0",
        "intent-network/matches/1.0.0"
    ]
}

# Error messages
ERROR_MESSAGES = {
    "connection_failed": "Node is offline or unreachable",
    "timeout": "Request timed out",
    "http_error": "HTTP error occurred",
    "invalid_node_id": "Invalid node ID",
    "invalid_agent_node": "Node is not a Service Agent",
    "invalid_builder_node": "Node is not a Block Builder",
    "unknown": "Unknown error occurred",
    "no_data": "No data available"
}

# Default values for empty states
DEFAULT_VALUES = {
    "response_time": 0,
    "success_rate": 0.0,
    "bid_amount": "0.0",
    "earnings": "0.0",
    "count": 0,
    "timestamp": 0
}

# UI text constants
UI_TEXT = {
    "dashboard_title": "PIN Intent Network - Real-time Monitoring Dashboard",
    "node_status_title": "🖥️ Node Status Overview",
    "intent_monitoring_title": "📡 Intent Flow Monitoring",
    "bidding_activity_title": "💰 Bidding Activity Tracking",
    "matching_results_title": "🎯 Matching Results",
    "p2p_network_title": "🌐 P2P Network Status",
    "performance_metrics_title": "📊 Performance Metrics",
    "last_update": "🔄 Last Update",
    "auto_refresh": "Auto-refresh every 5 seconds",
    "error_panel": "❌ {} Unavailable",
    "loading_panel": "Loading {}...",
    "no_data": "No data available",
    "retrying": "Retrying on next refresh..."
}

# Metrics formatting
METRICS_FORMAT = {
    "currency": "$ {:.2f}",
    "percentage": "{:.1f}%",
    "count": "{:,}",
    "response_time": "{} ms",
    "timestamp": "%H:%M:%S"
}


def get_node_config(node_id: int) -> Dict[str, Any]:
    """Get configuration for specific node."""
    return NODE_CONFIGS.get(node_id, {})


def get_all_node_ids() -> list:
    """Get all node IDs."""
    return list(NODE_CONFIGS.keys())


def get_service_agent_nodes() -> list:
    """Get node IDs for Service Agents."""
    return [
        node_id for node_id, config in NODE_CONFIGS.items()
        if config["type"] == "SERVICE_AGENT"
    ]


def get_block_builder_nodes() -> list:
    """Get node IDs for Block Builders."""
    return [
        node_id for node_id, config in NODE_CONFIGS.items()
        if config["type"] == "BLOCK_BUILDER"
    ]


def get_publisher_nodes() -> list:
    """Get node IDs for Intent Publishers."""
    return [
        node_id for node_id, config in NODE_CONFIGS.items()
        if config["type"] == "PUBLISHER"
    ]


def get_status_color(status: str) -> str:
    """Get color for status."""
    return STATUS_COLORS.get(status.lower(), STATUS_COLORS["offline"])


def get_intent_type_config(intent_type: str) -> Dict[str, Any]:
    """Get configuration for intent type."""
    return INTENT_TYPES.get(intent_type.lower(), {
        "name": intent_type.title(),
        "color": "#6C757D",
        "icon": "❓"
    })


def get_agent_type_config(agent_type: str) -> Dict[str, Any]:
    """Get configuration for agent type."""
    return AGENT_TYPES.get(agent_type.lower(), {
        "name": agent_type.title(),
        "color": "#6C757D",
        "icon": "🤖"
    })


def format_metric_value(value: Any, format_type: str) -> str:
    """Format metric value according to type."""
    format_string = METRICS_FORMAT.get(format_type, "{}")
    
    try:
        if format_type == "currency":
            return format_string.format(float(value))
        elif format_type == "percentage":
            return format_string.format(float(value) * 100)
        elif format_type == "count":
            return format_string.format(int(value))
        elif format_type == "response_time":
            return format_string.format(int(value))
        else:
            return str(value)
    except (ValueError, TypeError):
        return str(value)