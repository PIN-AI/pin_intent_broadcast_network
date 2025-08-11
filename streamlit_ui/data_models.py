"""
Data models for PIN POC Demo Frontend.
Defines data structures for API responses and UI state management.
"""

from dataclasses import dataclass
from typing import List, Optional, Any
import time


@dataclass
class NodeStatus:
    """Node health status model."""
    node_id: int
    is_running: bool
    http_port: int
    response_time_ms: int
    last_check: float
    error: Optional[str] = None


@dataclass
class AgentInfo:
    """Service Agent information model."""
    agent_id: str
    agent_type: str
    status: str
    total_bids_submitted: int
    successful_bids: int
    total_earnings: str
    last_activity: int


@dataclass
class BuilderInfo:
    """Block Builder information model."""
    builder_id: str
    status: str
    active_sessions: int
    completed_matches: int
    total_bids_received: int
    last_activity: int


@dataclass
class IntentInfo:
    """Intent information model."""
    intent_id: str
    intent_type: str
    status: str
    sender_id: str
    created_at: int
    broadcast_count: int
    bid_count: int


@dataclass
class MatchResult:
    """Match result information model."""
    match_id: str
    intent_id: str
    winning_agent_id: str
    winning_bid_amount: str
    total_bids: int
    match_algorithm: str
    matched_at: int
    status: str


@dataclass
class ExecutionMetrics:
    """System performance metrics model."""
    total_intents: int = 0
    active_intents: int = 0
    total_bids: int = 0
    active_bids: int = 0
    completed_matches: int = 0
    success_rate: float = 0.0
    avg_response_time_ms: int = 0
    p2p_peers_connected: int = 0
    network_messages_sent: int = 0
    network_messages_received: int = 0
    error: Optional[str] = None


@dataclass
class AgentsStatusResponse:
    """Response from agents status API."""
    agents: List[AgentInfo]
    error: Optional[str] = None


@dataclass
class BuildersStatusResponse:
    """Response from builders status API."""
    builders: List[BuilderInfo]
    error: Optional[str] = None


@dataclass
class IntentListResponse:
    """Response from intent list API."""
    intents: List[IntentInfo]
    error: Optional[str] = None


@dataclass
class MatchHistoryResponse:
    """Response from match history API."""
    matches: List[MatchResult]
    error: Optional[str] = None


@dataclass
class P2PNetworkInfo:
    """P2P network information model."""
    total_peers: int = 0
    connected_peers: int = 0
    bootstrap_peers: List[str] = None
    topics_subscribed: List[str] = None
    messages_sent: int = 0
    messages_received: int = 0
    network_id: str = ""
    host_id: str = ""
    
    def __post_init__(self):
        if self.bootstrap_peers is None:
            self.bootstrap_peers = []
        if self.topics_subscribed is None:
            self.topics_subscribed = []


@dataclass
class DashboardMetrics:
    """Aggregated dashboard metrics."""
    active_nodes: int = 0
    total_intents: int = 0
    active_bids: int = 0
    completed_matches: int = 0
    success_rate: float = 0.0
    avg_response_time: int = 0
    p2p_peers: int = 0
    
    # Delta values for metric changes
    delta_nodes: int = 0
    delta_intents: int = 0
    delta_bids: int = 0
    delta_matches: int = 0


@dataclass
class UIState:
    """UI state management."""
    last_refresh: float = 0.0
    auto_refresh_enabled: bool = True
    refresh_interval: int = 5
    selected_node: Optional[int] = None
    error_count: int = 0
    
    def should_refresh(self) -> bool:
        """Check if UI should refresh based on interval."""
        return (
            self.auto_refresh_enabled and 
            time.time() - self.last_refresh >= self.refresh_interval
        )
    
    def mark_refreshed(self) -> None:
        """Mark the UI as refreshed."""
        self.last_refresh = time.time()
    
    def increment_errors(self) -> None:
        """Increment error count."""
        self.error_count += 1


class DataCache:
    """Simple in-memory data cache for historical data."""
    
    def __init__(self, max_size: int = 100):
        self.max_size = max_size
        self._metrics_history: List[DashboardMetrics] = []
        self._intents_history: List[IntentInfo] = []
        self._matches_history: List[MatchResult] = []
        self._agents_history: List[AgentInfo] = []
    
    def add_metrics(self, metrics: DashboardMetrics) -> None:
        """Add metrics to history."""
        self._metrics_history.append(metrics)
        if len(self._metrics_history) > self.max_size:
            self._metrics_history.pop(0)
    
    def add_intents(self, intents: List[IntentInfo]) -> None:
        """Add intents to history."""
        self._intents_history.extend(intents)
        if len(self._intents_history) > self.max_size:
            self._intents_history = self._intents_history[-self.max_size:]
    
    def add_matches(self, matches: List[MatchResult]) -> None:
        """Add matches to history."""
        self._matches_history.extend(matches)
        if len(self._matches_history) > self.max_size:
            self._matches_history = self._matches_history[-self.max_size:]
    
    def add_agents(self, agents: List[AgentInfo]) -> None:
        """Add agents to history."""
        self._agents_history.extend(agents)
        if len(self._agents_history) > self.max_size:
            self._agents_history = self._agents_history[-self.max_size:]
    
    def get_metrics_history(self) -> List[DashboardMetrics]:
        """Get metrics history."""
        return self._metrics_history.copy()
    
    def get_intents_history(self) -> List[IntentInfo]:
        """Get intents history."""
        return self._intents_history.copy()
    
    def get_matches_history(self) -> List[MatchResult]:
        """Get matches history."""
        return self._matches_history.copy()
    
    def get_agents_history(self) -> List[AgentInfo]:
        """Get agents history."""
        return self._agents_history.copy()
    
    def clear(self) -> None:
        """Clear all cached data."""
        self._metrics_history.clear()
        self._intents_history.clear()
        self._matches_history.clear()
        self._agents_history.clear()


def create_empty_dashboard_metrics() -> DashboardMetrics:
    """Create empty dashboard metrics with default values."""
    return DashboardMetrics()


def aggregate_execution_metrics(metrics_by_node: dict) -> DashboardMetrics:
    """
    Aggregate execution metrics from multiple nodes into dashboard metrics.
    
    Args:
        metrics_by_node: Dict of node_id -> ExecutionMetrics
    
    Returns:
        DashboardMetrics: Aggregated metrics
    """
    dashboard = DashboardMetrics()
    
    valid_metrics = [
        m for m in metrics_by_node.values() 
        if isinstance(m, ExecutionMetrics) and m.error is None
    ]
    
    if not valid_metrics:
        return dashboard
    
    # Aggregate totals
    dashboard.total_intents = sum(m.total_intents for m in valid_metrics)
    dashboard.active_bids = sum(m.active_bids for m in valid_metrics)
    dashboard.completed_matches = sum(m.completed_matches for m in valid_metrics)
    
    # Calculate averages
    total_nodes = len(valid_metrics)
    dashboard.success_rate = sum(m.success_rate for m in valid_metrics) / total_nodes
    dashboard.avg_response_time = sum(m.avg_response_time_ms for m in valid_metrics) // total_nodes
    dashboard.p2p_peers = max(m.p2p_peers_connected for m in valid_metrics)
    
    return dashboard


def create_p2p_network_info_from_metrics(metrics: ExecutionMetrics) -> P2PNetworkInfo:
    """Create P2P network info from execution metrics."""
    return P2PNetworkInfo(
        connected_peers=metrics.p2p_peers_connected,
        messages_sent=metrics.network_messages_sent,
        messages_received=metrics.network_messages_received,
        total_peers=metrics.p2p_peers_connected,
        topics_subscribed=["intent-broadcast.*", "intent-network/bids/1.0.0", "intent-network/matches/1.0.0"],
        network_id="PIN-automation-network",
        host_id="auto-generated"
    )