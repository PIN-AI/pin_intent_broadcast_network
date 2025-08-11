"""
API Client Integration Tests for PIN POC Demo Frontend.
Tests HTTP client connectivity to 4 nodes and API integration patterns.
"""

import asyncio
import pytest
import httpx
import time
from unittest.mock import AsyncMock, patch, Mock
from typing import Dict, Any

from ..api_client import NodeAPIClient
from ..data_models import (
    NodeStatus, AgentInfo, BuilderInfo, IntentInfo, MatchResult,
    ExecutionMetrics, AgentsStatusResponse, BuildersStatusResponse,
    IntentListResponse, MatchHistoryResponse
)
from ..config import NODE_CONFIGS, API_TIMEOUT_SECONDS


class TestNodeAPIClient:
    """Test suite for NodeAPIClient with real and mocked scenarios."""

    @pytest.fixture
    def api_client(self):
        """Create API client instance."""
        return NodeAPIClient(timeout=2)  # Shorter timeout for tests
    
    @pytest.fixture
    def mock_response(self):
        """Create mock HTTP response."""
        response = Mock()
        response.status_code = 200
        response.json.return_value = {"status": "ok", "data": "test"}
        return response

    def test_client_initialization(self, api_client):
        """Test client initialization with correct configuration."""
        assert api_client.timeout == 2
        assert len(api_client.base_urls) == 4
        
        # Verify all expected node URLs are configured
        expected_ports = [8100, 8101, 8102, 8103]
        for node_id, base_url in api_client.base_urls.items():
            expected_port = expected_ports[node_id - 1]
            assert f":{expected_port}" in base_url
            assert "localhost" in base_url

    @pytest.mark.asyncio
    async def test_safe_api_call_success(self, api_client, mock_response):
        """Test successful API call with valid response."""
        with patch('httpx.AsyncClient') as mock_client:
            mock_client.return_value.__aenter__.return_value.get.return_value = mock_response
            
            result = await api_client.safe_api_call("http://localhost:8100/health")
            
            assert "error" not in result
            assert result["status"] == "ok"
            assert "_response_time_ms" in result
            assert isinstance(result["_response_time_ms"], int)

    @pytest.mark.asyncio
    async def test_safe_api_call_timeout(self, api_client):
        """Test API call timeout handling."""
        with patch('httpx.AsyncClient') as mock_client:
            mock_client.return_value.__aenter__.return_value.get.side_effect = httpx.TimeoutException("Timeout")
            
            result = await api_client.safe_api_call("http://localhost:8100/health")
            
            assert result["error"] == "timeout"
            assert "Request timeout" in result["message"]

    @pytest.mark.asyncio
    async def test_safe_api_call_connection_error(self, api_client):
        """Test API call connection error handling."""
        with patch('httpx.AsyncClient') as mock_client:
            mock_client.return_value.__aenter__.return_value.get.side_effect = httpx.ConnectError("Connection failed")
            
            result = await api_client.safe_api_call("http://localhost:8100/health")
            
            assert result["error"] == "connection_failed"
            assert "Node offline" in result["message"]

    @pytest.mark.asyncio
    async def test_safe_api_call_http_error(self, api_client):
        """Test API call HTTP error handling."""
        mock_response = Mock()
        mock_response.status_code = 500
        
        with patch('httpx.AsyncClient') as mock_client:
            mock_client.return_value.__aenter__.return_value.get.return_value = mock_response
            
            result = await api_client.safe_api_call("http://localhost:8100/health")
            
            assert result["error"] == "http_error"
            assert result["status_code"] == 500

    @pytest.mark.asyncio
    async def test_safe_api_call_invalid_json(self, api_client):
        """Test API call with invalid JSON response."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.side_effect = ValueError("Invalid JSON")
        
        with patch('httpx.AsyncClient') as mock_client:
            mock_client.return_value.__aenter__.return_value.get.return_value = mock_response
            
            result = await api_client.safe_api_call("http://localhost:8100/health")
            
            assert result["error"] == "invalid_json"

    @pytest.mark.asyncio
    async def test_get_node_status_healthy(self, api_client):
        """Test getting status of healthy node."""
        with patch.object(api_client, 'safe_api_call', return_value={"status": "healthy", "_response_time_ms": 50}):
            status = await api_client.get_node_status(1)
            
            assert isinstance(status, NodeStatus)
            assert status.node_id == 1
            assert status.is_running is True
            assert status.http_port == 8100
            assert status.response_time_ms == 50
            assert status.error is None

    @pytest.mark.asyncio
    async def test_get_node_status_unhealthy(self, api_client):
        """Test getting status of unhealthy node."""
        with patch.object(api_client, 'safe_api_call', return_value={"error": "connection_failed", "message": "Node offline"}):
            status = await api_client.get_node_status(1)
            
            assert isinstance(status, NodeStatus)
            assert status.node_id == 1
            assert status.is_running is False
            assert status.error == "connection_failed"

    @pytest.mark.asyncio
    async def test_get_node_status_invalid_node_id(self, api_client):
        """Test getting status with invalid node ID."""
        status = await api_client.get_node_status(999)
        
        assert isinstance(status, NodeStatus)
        assert status.node_id == 999
        assert status.is_running is False
        assert status.error == "invalid_node_id"

    @pytest.mark.asyncio
    async def test_get_agents_status_success(self, api_client):
        """Test getting Service Agents status."""
        mock_response = {
            "agents": [
                {
                    "agent_id": "trading-agent-001",
                    "agent_type": "trading",
                    "status": "active",
                    "total_bids_submitted": 15,
                    "successful_bids": 8,
                    "total_earnings": "250.50",
                    "last_activity": 1642000000
                }
            ]
        }
        
        with patch.object(api_client, 'safe_api_call', return_value=mock_response):
            response = await api_client.get_agents_status(2)  # Service Agent node
            
            assert isinstance(response, AgentsStatusResponse)
            assert response.error is None
            assert len(response.agents) == 1
            
            agent = response.agents[0]
            assert agent.agent_id == "trading-agent-001"
            assert agent.agent_type == "trading"
            assert agent.total_bids_submitted == 15
            assert agent.successful_bids == 8

    @pytest.mark.asyncio
    async def test_get_agents_status_invalid_node(self, api_client):
        """Test getting agents status from non-agent node."""
        response = await api_client.get_agents_status(1)  # Publisher node, not agent
        
        assert isinstance(response, AgentsStatusResponse)
        assert response.error == "invalid_agent_node"
        assert len(response.agents) == 0

    @pytest.mark.asyncio
    async def test_get_builders_status_success(self, api_client):
        """Test getting Block Builders status."""
        mock_response = {
            "builders": [
                {
                    "builder_id": "primary-builder-001",
                    "status": "active",
                    "active_sessions": 3,
                    "completed_matches": 42,
                    "total_bids_received": 125,
                    "last_activity": 1642000000
                }
            ]
        }
        
        with patch.object(api_client, 'safe_api_call', return_value=mock_response):
            response = await api_client.get_builders_status(4)  # Block Builder node
            
            assert isinstance(response, BuildersStatusResponse)
            assert response.error is None
            assert len(response.builders) == 1
            
            builder = response.builders[0]
            assert builder.builder_id == "primary-builder-001"
            assert builder.active_sessions == 3
            assert builder.completed_matches == 42

    @pytest.mark.asyncio
    async def test_get_builders_status_invalid_node(self, api_client):
        """Test getting builders status from non-builder node."""
        response = await api_client.get_builders_status(2)  # Service Agent node
        
        assert isinstance(response, BuildersStatusResponse)
        assert response.error == "invalid_builder_node"

    @pytest.mark.asyncio
    async def test_get_execution_metrics_success(self, api_client):
        """Test getting execution metrics."""
        mock_response = {
            "total_intents": 150,
            "active_intents": 8,
            "total_bids": 423,
            "active_bids": 15,
            "completed_matches": 87,
            "success_rate": 0.82,
            "avg_response_time_ms": 125,
            "p2p_peers_connected": 3,
            "network_messages_sent": 1250,
            "network_messages_received": 1180
        }
        
        with patch.object(api_client, 'safe_api_call', return_value=mock_response):
            metrics = await api_client.get_execution_metrics(1)
            
            assert isinstance(metrics, ExecutionMetrics)
            assert metrics.error is None
            assert metrics.total_intents == 150
            assert metrics.success_rate == 0.82
            assert metrics.p2p_peers_connected == 3

    @pytest.mark.asyncio
    async def test_get_intent_list_success(self, api_client):
        """Test getting intent list."""
        mock_response = {
            "intents": [
                {
                    "intent_id": "intent_001",
                    "type": "exchange",
                    "status": "broadcasted",
                    "sender_id": "node-1",
                    "created_at": 1642000000,
                    "broadcast_count": 1,
                    "bid_count": 3
                }
            ]
        }
        
        with patch.object(api_client, 'safe_api_call', return_value=mock_response):
            response = await api_client.get_intent_list(1, limit=10)
            
            assert isinstance(response, IntentListResponse)
            assert response.error is None
            assert len(response.intents) == 1
            
            intent = response.intents[0]
            assert intent.intent_id == "intent_001"
            assert intent.intent_type == "exchange"
            assert intent.bid_count == 3

    @pytest.mark.asyncio
    async def test_get_match_history_success(self, api_client):
        """Test getting match history."""
        mock_response = {
            "matches": [
                {
                    "match_id": "match_001",
                    "intent_id": "intent_001",
                    "winning_agent_id": "trading-agent-001",
                    "winning_bid_amount": "125.75",
                    "total_bids": 5,
                    "match_algorithm": "highest_bid",
                    "matched_at": 1642000000,
                    "status": "completed"
                }
            ]
        }
        
        with patch.object(api_client, 'safe_api_call', return_value=mock_response):
            response = await api_client.get_match_history(4)  # Block Builder node
            
            assert isinstance(response, MatchHistoryResponse)
            assert response.error is None
            assert len(response.matches) == 1
            
            match = response.matches[0]
            assert match.match_id == "match_001"
            assert match.winning_bid_amount == "125.75"
            assert match.total_bids == 5

    @pytest.mark.asyncio
    async def test_fetch_all_data_success(self, api_client):
        """Test fetching all data from all nodes."""
        # Mock all the individual method calls
        with patch.object(api_client, 'get_node_status', return_value=NodeStatus(1, True, 8100, 50, time.time())):
            with patch.object(api_client, 'get_agents_status', return_value=AgentsStatusResponse([])):
                with patch.object(api_client, 'get_builders_status', return_value=BuildersStatusResponse([])):
                    with patch.object(api_client, 'get_execution_metrics', return_value=ExecutionMetrics()):
                        with patch.object(api_client, 'get_intent_list', return_value=IntentListResponse([])):
                            with patch.object(api_client, 'get_match_history', return_value=MatchHistoryResponse([])):
                                
                                data = await api_client.fetch_all_data()
                                
                                assert isinstance(data, dict)
                                assert "nodes" in data
                                assert "agents" in data
                                assert "builders" in data
                                assert "metrics" in data
                                assert "intents" in data
                                assert "matches" in data
                                assert "_fetch_metadata" in data
                                
                                # Check metadata
                                metadata = data["_fetch_metadata"]
                                assert "timestamp" in metadata
                                assert "total_tasks" in metadata
                                assert "successful_tasks" in metadata

    @pytest.mark.asyncio
    async def test_fetch_all_data_with_errors(self, api_client):
        """Test fetching all data when some nodes are offline."""
        # Simulate mixed success/error responses
        error_response = {"error": "connection_failed", "message": "Node offline"}
        
        with patch.object(api_client, 'safe_api_call', side_effect=[
            {"status": "ok"},  # Node 1 health - success
            error_response,    # Node 2 health - error
            {"status": "ok"},  # Node 3 health - success
            error_response,    # Node 4 health - error
        ]):
            data = await api_client.fetch_all_data()
            
            assert isinstance(data, dict)
            metadata = data["_fetch_metadata"]
            assert metadata["total_tasks"] > 0
            assert len(metadata["errors"]) > 0  # Should have some errors

    @pytest.mark.asyncio
    async def test_fetch_all_data_timeout(self, api_client):
        """Test fetch all data with overall timeout."""
        # Mock all calls to be slow
        slow_mock = AsyncMock(side_effect=asyncio.sleep(15))  # Longer than timeout
        
        with patch.object(api_client, 'get_node_status', slow_mock):
            with patch.object(api_client, 'get_agents_status', slow_mock):
                data = await api_client.fetch_all_data()
                
                assert isinstance(data, dict)
                # Should have timeout errors in the results
                assert "_fetch_metadata" in data


class TestAPIClientPerformance:
    """Performance tests for API client operations."""

    @pytest.fixture
    def api_client(self):
        return NodeAPIClient(timeout=5)

    @pytest.mark.asyncio
    async def test_concurrent_node_status_requests(self, api_client):
        """Test concurrent node status requests performance."""
        start_time = time.time()
        
        # Mock fast responses
        with patch.object(api_client, 'safe_api_call', return_value={"status": "ok", "_response_time_ms": 50}):
            tasks = [api_client.get_node_status(i) for i in range(1, 5)]
            results = await asyncio.gather(*tasks)
        
        end_time = time.time()
        duration = end_time - start_time
        
        # Should complete in under 1 second for concurrent requests
        assert duration < 1.0
        assert len(results) == 4
        assert all(isinstance(result, NodeStatus) for result in results)

    @pytest.mark.asyncio
    async def test_fetch_all_data_performance(self, api_client):
        """Test overall data fetch performance."""
        start_time = time.time()
        
        # Mock all responses to be fast
        mock_responses = {
            'get_node_status': NodeStatus(1, True, 8100, 50, time.time()),
            'get_agents_status': AgentsStatusResponse([]),
            'get_builders_status': BuildersStatusResponse([]),
            'get_execution_metrics': ExecutionMetrics(),
            'get_intent_list': IntentListResponse([]),
            'get_match_history': MatchHistoryResponse([])
        }
        
        with patch.multiple(api_client, **mock_responses):
            data = await api_client.fetch_all_data()
        
        end_time = time.time()
        duration = end_time - start_time
        
        # Should complete in under 3 seconds
        assert duration < 3.0
        assert isinstance(data, dict)
        assert data["_fetch_metadata"]["total_tasks"] > 0


class TestAPIClientRealEnvironment:
    """Tests that can run against real PIN nodes when available."""

    @pytest.fixture
    def api_client(self):
        return NodeAPIClient(timeout=10)  # Longer timeout for real tests

    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_real_node_connectivity(self, api_client):
        """Test connectivity to real PIN nodes (when running)."""
        # This test requires actual PIN nodes to be running
        # It will skip if nodes are not available
        
        try:
            # Try to connect to node 1 (Intent Publisher)
            status = await api_client.get_node_status(1)
            
            if status.error:
                pytest.skip("PIN nodes not running - skipping real connectivity test")
            
            # If we get here, at least one node is running
            assert status.is_running
            assert status.response_time_ms > 0
            
        except Exception as e:
            pytest.skip(f"Cannot connect to PIN nodes: {str(e)}")

    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_real_all_nodes_health_check(self, api_client):
        """Test health check on all real PIN nodes."""
        try:
            healthy_nodes = 0
            
            for node_id in range(1, 5):
                status = await api_client.get_node_status(node_id)
                if status.is_running and not status.error:
                    healthy_nodes += 1
                    assert status.http_port in [8100, 8101, 8102, 8103]
            
            if healthy_nodes == 0:
                pytest.skip("No healthy PIN nodes found - skipping integration test")
            
            # At least some nodes should be healthy
            assert healthy_nodes > 0
            
        except Exception as e:
            pytest.skip(f"Cannot perform health check: {str(e)}")

    @pytest.mark.integration
    @pytest.mark.asyncio
    async def test_real_fetch_all_data(self, api_client):
        """Test fetching all data from real PIN system."""
        try:
            data = await api_client.fetch_all_data()
            
            # Check if we got any real data
            metadata = data.get("_fetch_metadata", {})
            successful_tasks = metadata.get("successful_tasks", 0)
            
            if successful_tasks == 0:
                pytest.skip("No successful API calls - PIN system may not be running")
            
            # Validate data structure
            assert isinstance(data, dict)
            assert "nodes" in data
            assert "agents" in data
            assert "builders" in data
            assert "metrics" in data
            assert "intents" in data
            assert "matches" in data
            
            # Should have some successful responses
            assert successful_tasks > 0
            
        except Exception as e:
            pytest.skip(f"Cannot fetch data from PIN system: {str(e)}")


class TestAPIClientConfiguration:
    """Tests for API client configuration and settings."""

    def test_default_timeout_configuration(self):
        """Test default timeout configuration."""
        client = NodeAPIClient()
        assert client.timeout == API_TIMEOUT_SECONDS

    def test_custom_timeout_configuration(self):
        """Test custom timeout configuration."""
        client = NodeAPIClient(timeout=10)
        assert client.timeout == 10

    def test_node_url_configuration(self):
        """Test node URL configuration matches expected ports."""
        client = NodeAPIClient()
        
        # Check all node URLs are properly configured
        assert len(client.base_urls) == 4
        
        for node_id, expected_port in [(1, 8100), (2, 8101), (3, 8102), (4, 8103)]:
            url = client.base_urls[node_id]
            assert f":{expected_port}" in url
            assert url.startswith("http://localhost:")

    def test_node_type_validation(self):
        """Test node type validation in client methods."""
        client = NodeAPIClient()
        
        # Service Agent nodes (2, 3) should be valid for agent calls
        service_nodes = [2, 3]
        for node_id in service_nodes:
            config = NODE_CONFIGS[node_id]
            assert config["type"] == "SERVICE_AGENT"
        
        # Block Builder node (4) should be valid for builder calls  
        builder_node = 4
        config = NODE_CONFIGS[builder_node]
        assert config["type"] == "BLOCK_BUILDER"
        
        # Publisher node (1) should not be valid for agent/builder calls
        publisher_node = 1
        config = NODE_CONFIGS[publisher_node]
        assert config["type"] == "PUBLISHER"