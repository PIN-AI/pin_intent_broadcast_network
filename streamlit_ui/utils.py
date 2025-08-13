"""
Utility functions for PIN POC Demo Frontend.
Provides data formatting, time handling, and UI helper functions.
"""

import time
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional, Union
import pandas as pd

from config import METRICS_FORMAT, STATUS_COLORS, DEFAULT_VALUES
from data_models import (
    NodeStatus, AgentInfo, BuilderInfo, IntentInfo, 
    MatchResult, ExecutionMetrics, DashboardMetrics
)


def format_timestamp(timestamp: Union[int, float, str], format_string: str = "%H:%M:%S") -> str:
    """
    Format Unix timestamp to readable string.
    
    Args:
        timestamp: Unix timestamp (seconds) or string
        format_string: DateTime format string
    
    Returns:
        Formatted time string
    """
    if not timestamp or timestamp == 0:
        return "N/A"
    
    try:
        if isinstance(timestamp, str):
            # Try to convert string to float first
            timestamp = float(timestamp)
        
        if isinstance(timestamp, (int, float)):
            dt = datetime.fromtimestamp(timestamp)
            return dt.strftime(format_string)
        else:
            # If it's already a datetime object
            return timestamp.strftime(format_string)
    except (ValueError, OSError, TypeError):
        return "Invalid"


def format_duration(seconds: Union[int, float]) -> str:
    """
    Format duration in seconds to human-readable string.
    
    Args:
        seconds: Duration in seconds
    
    Returns:
        Formatted duration string
    """
    if not seconds or seconds < 0:
        return "0s"
    
    try:
        seconds = int(seconds)
        if seconds < 60:
            return f"{seconds}s"
        elif seconds < 3600:
            minutes = seconds // 60
            remaining_seconds = seconds % 60
            return f"{minutes}m {remaining_seconds}s"
        else:
            hours = seconds // 3600
            remaining_minutes = (seconds % 3600) // 60
            return f"{hours}h {remaining_minutes}m"
    except (ValueError, TypeError):
        return "0s"


def format_currency(amount: Union[str, float], currency: str = "$") -> str:
    """
    Format currency amount.
    
    Args:
        amount: Currency amount
        currency: Currency symbol
    
    Returns:
        Formatted currency string
    """
    try:
        if isinstance(amount, str):
            amount = float(amount)
        return f"{currency}{amount:.2f}"
    except (ValueError, TypeError):
        return f"{currency}0.00"


def format_percentage(value: Union[float, int], decimal_places: int = 1) -> str:
    """
    Format percentage value.
    
    Args:
        value: Decimal value (0.75 for 75%)
        decimal_places: Number of decimal places
    
    Returns:
        Formatted percentage string
    """
    try:
        percentage = float(value) * 100
        return f"{percentage:.{decimal_places}f}%"
    except (ValueError, TypeError):
        return "0.0%"


def format_number(value: Union[int, float], thousands_separator: bool = True) -> str:
    """
    Format number with optional thousands separator.
    
    Args:
        value: Number to format
        thousands_separator: Whether to use thousands separator
    
    Returns:
        Formatted number string
    """
    try:
        if isinstance(value, float):
            if value.is_integer():
                value = int(value)
        
        if thousands_separator:
            return f"{value:,}"
        else:
            return str(value)
    except (ValueError, TypeError):
        return "0"


def get_time_ago(timestamp: Union[int, float, str]) -> str:
    """
    Get human-readable time ago string.
    
    Args:
        timestamp: Unix timestamp or string
    
    Returns:
        Time ago string (e.g., "2m ago", "1h ago")
    """
    if not timestamp or timestamp == 0:
        return "Never"
    
    try:
        if isinstance(timestamp, str):
            timestamp = float(timestamp)
        
        now = time.time()
        diff = now - timestamp
        
        if diff < 60:
            return f"{int(diff)}s ago"
        elif diff < 3600:
            return f"{int(diff // 60)}m ago"
        elif diff < 86400:
            return f"{int(diff // 3600)}h ago"
        else:
            return f"{int(diff // 86400)}d ago"
    except (ValueError, TypeError):
        return "Unknown"


def calculate_success_rate(successful: int, total: int) -> float:
    """
    Calculate success rate as percentage.
    
    Args:
        successful: Number of successful operations
        total: Total number of operations
    
    Returns:
        Success rate as decimal (0.0 to 1.0)
    """
    if total == 0:
        return 0.0
    
    try:
        return float(successful) / float(total)
    except (ValueError, TypeError, ZeroDivisionError):
        return 0.0


def format_intent_status(status: str) -> str:
    """
    Format intent status for display.
    
    Args:
        status: Raw intent status
        
    Returns:
        Formatted status string
    """
    status_mapping = {
        "INTENT_STATUS_UNSPECIFIED": "Unspecified",
        "INTENT_STATUS_CREATED": "Created",
        "INTENT_STATUS_VALIDATED": "Validated", 
        "INTENT_STATUS_BROADCASTED": "Broadcasted",
        "INTENT_STATUS_PROCESSED": "Processing",
        "INTENT_STATUS_MATCHED": "Matched",
        "INTENT_STATUS_COMPLETED": "Completed",
        "INTENT_STATUS_FAILED": "Failed",
        "INTENT_STATUS_EXPIRED": "Expired",
        "INTENT_STATUS_UNKNOWN": "Unknown"
    }
    
    return status_mapping.get(status, status.replace("INTENT_STATUS_", "").title())


def get_status_emoji(status: str) -> str:
    """
    Get emoji for status.
    
    Args:
        status: Status string
    
    Returns:
        Appropriate emoji
    """
    status_emojis = {
        "running": "🟢",
        "active": "🟢", 
        "idle": "🟡",
        "stopped": "🔴",
        "error": "🔴",
        "offline": "⚫",
        "pending": "🔵",
        "completed": "✅",
        "failed": "❌",
        "unknown": "❓"
    }
    
    return status_emojis.get(status.lower(), "❓")


def truncate_string(text: str, max_length: int = 30, suffix: str = "...") -> str:
    """
    Truncate string to maximum length.
    
    Args:
        text: Text to truncate
        max_length: Maximum length
        suffix: Suffix to add when truncating
    
    Returns:
        Truncated string
    """
    if not text:
        return ""
    
    if len(text) <= max_length:
        return text
    
    return text[:max_length - len(suffix)] + suffix


def safe_get(data: Dict[str, Any], key: str, default: Any = None) -> Any:
    """
    Safely get value from dictionary with default.
    
    Args:
        data: Dictionary to get value from
        key: Key to look for
        default: Default value if key not found
    
    Returns:
        Value from dictionary or default
    """
    return data.get(key, default) if data else default


def create_metrics_dataframe(metrics_history: List[DashboardMetrics]) -> pd.DataFrame:
    """
    Create pandas DataFrame from metrics history.
    
    Args:
        metrics_history: List of DashboardMetrics
    
    Returns:
        DataFrame with metrics data
    """
    if not metrics_history:
        return pd.DataFrame()
    
    data = []
    for i, metrics in enumerate(metrics_history):
        data.append({
            "timestamp": i,
            "active_nodes": metrics.active_nodes,
            "total_intents": metrics.total_intents,
            "active_bids": metrics.active_bids,
            "completed_matches": metrics.completed_matches,
            "success_rate": metrics.success_rate,
            "avg_response_time": metrics.avg_response_time,
            "p2p_peers": metrics.p2p_peers
        })
    
    return pd.DataFrame(data)


def create_intents_dataframe(intents: List[IntentInfo]) -> pd.DataFrame:
    """
    Create pandas DataFrame from intent list.
    
    Args:
        intents: List of IntentInfo
    
    Returns:
        DataFrame with intent data
    """
    if not intents:
        return pd.DataFrame()
    
    data = []
    for intent in intents:
        data.append({
            "intent_id": intent.intent_id,
            "type": intent.intent_type,
            "status": format_intent_status(intent.status),
            "sender": intent.sender_id,
            "created_at": format_timestamp(intent.created_at),
            "broadcasts": intent.broadcast_count,
            "bids": intent.bid_count,
            "time_ago": get_time_ago(intent.created_at)
        })
    
    return pd.DataFrame(data)


def create_agents_dataframe(agents: List[AgentInfo]) -> pd.DataFrame:
    """
    Create pandas DataFrame from agent list.
    
    Args:
        agents: List of AgentInfo
    
    Returns:
        DataFrame with agent data
    """
    if not agents:
        return pd.DataFrame()
    
    data = []
    for agent in agents:
        success_rate = calculate_success_rate(
            agent.successful_bids, 
            agent.total_bids_submitted
        )
        
        data.append({
            "agent_id": agent.agent_id,
            "type": agent.agent_type,
            "status": agent.status,
            "total_bids": agent.total_bids_submitted,
            "successful_bids": agent.successful_bids,
            "success_rate": format_percentage(success_rate),
            "earnings": format_currency(agent.total_earnings),
            "last_activity": get_time_ago(agent.last_activity)
        })
    
    return pd.DataFrame(data)


def create_matches_dataframe(matches: List[MatchResult]) -> pd.DataFrame:
    """
    Create pandas DataFrame from match results.
    
    Args:
        matches: List of MatchResult
    
    Returns:
        DataFrame with match data
    """
    if not matches:
        return pd.DataFrame()
    
    data = []
    for match in matches:
        data.append({
            "match_id": match.match_id,
            "intent_id": match.intent_id,  # 完整显示，不截断
            "winner": match.winning_agent_id,  # 完整显示，不截断
            "bid_amount": str(match.winning_bid_amount),  # 不添加货币符号，直接显示数值
            "total_bids": match.total_bids,
            "algorithm": match.match_algorithm,
            "status": match.status,
            "matched_at": format_timestamp(match.matched_at),
            "time_ago": get_time_ago(match.matched_at)
        })
    
    return pd.DataFrame(data)


def calculate_delta(current: int, previous: int) -> int:
    """
    Calculate delta between current and previous values.
    
    Args:
        current: Current value
        previous: Previous value
    
    Returns:
        Delta (positive/negative change)
    """
    try:
        return current - previous
    except (ValueError, TypeError):
        return 0


def is_node_healthy(node: NodeStatus) -> bool:
    """
    Check if node is healthy.
    
    Args:
        node: NodeStatus object
    
    Returns:
        True if node is healthy
    """
    return (
        node.is_running and 
        node.error is None and 
        node.response_time_ms < 5000  # Less than 5 seconds
    )


def get_system_health_score(nodes: List[NodeStatus]) -> float:
    """
    Calculate overall system health score.
    
    Args:
        nodes: List of NodeStatus
    
    Returns:
        Health score from 0.0 to 1.0
    """
    if not nodes:
        return 0.0
    
    healthy_count = sum(1 for node in nodes if is_node_healthy(node))
    return healthy_count / len(nodes)


def format_bytes(bytes_count: Union[int, float]) -> str:
    """
    Format byte count to human readable string.
    
    Args:
        bytes_count: Number of bytes
    
    Returns:
        Formatted byte string
    """
    try:
        bytes_count = float(bytes_count)
        
        if bytes_count < 1024:
            return f"{bytes_count:.0f} B"
        elif bytes_count < 1024 ** 2:
            return f"{bytes_count / 1024:.1f} KB"
        elif bytes_count < 1024 ** 3:
            return f"{bytes_count / (1024 ** 2):.1f} MB"
        else:
            return f"{bytes_count / (1024 ** 3):.1f} GB"
    except (ValueError, TypeError):
        return "0 B"


def validate_node_data(data: Dict[str, Any]) -> bool:
    """
    Validate node data structure.
    
    Args:
        data: Node data dictionary
    
    Returns:
        True if data is valid
    """
    required_fields = ["nodes", "agents", "builders", "metrics", "intents"]
    return all(field in data for field in required_fields)


def extract_error_message(error_data: Dict[str, Any]) -> str:
    """
    Extract error message from error data.
    
    Args:
        error_data: Error data dictionary
    
    Returns:
        Human-readable error message
    """
    if not error_data or "error" not in error_data:
        return "Unknown error"
    
    error_type = error_data.get("error", "unknown")
    error_message = error_data.get("message", "")
    
    error_descriptions = {
        "connection_failed": "Node is offline or unreachable",
        "timeout": "Request timed out - node may be overloaded",
        "http_error": f"HTTP error: {error_message}",
        "invalid_node_id": "Invalid node configuration",
        "unknown": f"Unexpected error: {error_message}"
    }
    
    return error_descriptions.get(error_type, f"Error: {error_message}")


def create_empty_response(response_type: str) -> Dict[str, Any]:
    """
    Create empty response for failed API calls.
    
    Args:
        response_type: Type of response to create
    
    Returns:
        Empty response dictionary
    """
    empty_responses = {
        "agents": {"agents": [], "error": None},
        "builders": {"builders": [], "error": None},
        "intents": {"intents": [], "error": None},
        "matches": {"matches": [], "error": None},
        "metrics": {
            "total_intents": 0,
            "active_intents": 0,
            "total_bids": 0,
            "active_bids": 0,
            "completed_matches": 0,
            "success_rate": 0.0,
            "avg_response_time_ms": 0,
            "p2p_peers_connected": 0,
            "network_messages_sent": 0,
            "network_messages_received": 0,
            "error": None
        }
    }
    
    return empty_responses.get(response_type, {})