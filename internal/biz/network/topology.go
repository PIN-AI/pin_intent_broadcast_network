package network

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"pin_intent_broadcast_network/internal/biz/common"
)

// Topology manages network topology information
// This file will contain the implementation for task 7.3
type Topology struct {
	nodes       map[peer.ID]*NodeInfo
	edges       map[string][]string
	stats       map[string]interface{}
	lastUpdated time.Time
	mu          sync.RWMutex
	stopCh      chan struct{}
}

// NodeInfo represents information about a network node
type NodeInfo struct {
	PeerID      peer.ID   `json:"peer_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeen    time.Time `json:"last_seen"`
	IsActive    bool      `json:"is_active"`
	Connections []peer.ID `json:"connections"`
}

// NewTopology creates a new network topology manager
func NewTopology() *Topology {
	return &Topology{
		nodes:       make(map[peer.ID]*NodeInfo),
		edges:       make(map[string][]string),
		stats:       make(map[string]interface{}),
		lastUpdated: time.Now(),
		stopCh:      make(chan struct{}),
	}
}

// GetTopology returns the current network topology
func (t *Topology) GetTopology() *common.NetworkTopology {
	// TODO: Implement in task 7.3
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Convert internal representation to common format
	nodes := make([]peer.ID, 0, len(t.nodes))
	for peerID := range t.nodes {
		nodes = append(nodes, peerID)
	}

	// Copy edges map
	edges := make(map[string][]string)
	for key, value := range t.edges {
		edges[key] = make([]string, len(value))
		copy(edges[key], value)
	}

	// Copy stats map
	stats := make(map[string]interface{})
	for key, value := range t.stats {
		stats[key] = value
	}

	return &common.NetworkTopology{
		Nodes: nodes,
		Edges: edges,
		Stats: stats,
	}
}

// AddPeer adds a peer to the topology
func (t *Topology) AddPeer(peerID peer.ID) {
	// TODO: Implement in task 7.3
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.nodes[peerID]; !exists {
		t.nodes[peerID] = &NodeInfo{
			PeerID:      peerID,
			ConnectedAt: time.Now(),
			LastSeen:    time.Now(),
			IsActive:    true,
			Connections: make([]peer.ID, 0),
		}

		t.lastUpdated = time.Now()
		t.updateStats()
	}
}

// RemovePeer removes a peer from the topology
func (t *Topology) RemovePeer(peerID peer.ID) {
	// TODO: Implement in task 7.3
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.nodes[peerID]; exists {
		delete(t.nodes, peerID)

		// Remove edges involving this peer
		peerIDStr := peerID.String()
		delete(t.edges, peerIDStr)

		// Remove this peer from other peers' edge lists
		for key, connections := range t.edges {
			filtered := make([]string, 0)
			for _, conn := range connections {
				if conn != peerIDStr {
					filtered = append(filtered, conn)
				}
			}
			t.edges[key] = filtered
		}

		t.lastUpdated = time.Now()
		t.updateStats()
	}
}

// AddConnection adds a connection between two peers
func (t *Topology) AddConnection(peer1, peer2 peer.ID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Ensure both peers exist
	t.ensurePeerExists(peer1)
	t.ensurePeerExists(peer2)

	// Add bidirectional connection
	peer1Str := peer1.String()
	peer2Str := peer2.String()

	// Add peer2 to peer1's connections
	if connections, exists := t.edges[peer1Str]; exists {
		// Check if connection already exists
		for _, conn := range connections {
			if conn == peer2Str {
				return // Connection already exists
			}
		}
		t.edges[peer1Str] = append(connections, peer2Str)
	} else {
		t.edges[peer1Str] = []string{peer2Str}
	}

	// Add peer1 to peer2's connections
	if connections, exists := t.edges[peer2Str]; exists {
		// Check if connection already exists
		for _, conn := range connections {
			if conn == peer1Str {
				return // Connection already exists
			}
		}
		t.edges[peer2Str] = append(connections, peer1Str)
	} else {
		t.edges[peer2Str] = []string{peer1Str}
	}

	t.lastUpdated = time.Now()
	t.updateStats()
}

// RemoveConnection removes a connection between two peers
func (t *Topology) RemoveConnection(peer1, peer2 peer.ID) {
	t.mu.Lock()
	defer t.mu.Unlock()

	peer1Str := peer1.String()
	peer2Str := peer2.String()

	// Remove peer2 from peer1's connections
	if connections, exists := t.edges[peer1Str]; exists {
		filtered := make([]string, 0)
		for _, conn := range connections {
			if conn != peer2Str {
				filtered = append(filtered, conn)
			}
		}
		t.edges[peer1Str] = filtered
	}

	// Remove peer1 from peer2's connections
	if connections, exists := t.edges[peer2Str]; exists {
		filtered := make([]string, 0)
		for _, conn := range connections {
			if conn != peer1Str {
				filtered = append(filtered, conn)
			}
		}
		t.edges[peer2Str] = filtered
	}

	t.lastUpdated = time.Now()
	t.updateStats()
}

// ensurePeerExists ensures a peer exists in the topology
func (t *Topology) ensurePeerExists(peerID peer.ID) {
	if _, exists := t.nodes[peerID]; !exists {
		t.nodes[peerID] = &NodeInfo{
			PeerID:      peerID,
			ConnectedAt: time.Now(),
			LastSeen:    time.Now(),
			IsActive:    true,
			Connections: make([]peer.ID, 0),
		}
	}
}

// updateStats updates topology statistics
func (t *Topology) updateStats() {
	// This method should be called with mutex already held
	totalNodes := len(t.nodes)
	totalEdges := 0

	for _, connections := range t.edges {
		totalEdges += len(connections)
	}

	// Since edges are bidirectional, divide by 2
	totalEdges = totalEdges / 2

	avgConnections := 0.0
	if totalNodes > 0 {
		avgConnections = float64(totalEdges*2) / float64(totalNodes)
	}

	t.stats = map[string]interface{}{
		"total_nodes":      totalNodes,
		"total_edges":      totalEdges,
		"avg_connections":  avgConnections,
		"last_updated":     t.lastUpdated,
		"density":          t.calculateDensity(),
		"clustering_coeff": t.calculateClusteringCoefficient(),
	}
}

// calculateDensity calculates the network density
func (t *Topology) calculateDensity() float64 {
	totalNodes := len(t.nodes)
	if totalNodes < 2 {
		return 0.0
	}

	totalEdges := 0
	for _, connections := range t.edges {
		totalEdges += len(connections)
	}
	totalEdges = totalEdges / 2 // Bidirectional edges

	maxPossibleEdges := totalNodes * (totalNodes - 1) / 2
	return float64(totalEdges) / float64(maxPossibleEdges)
}

// calculateClusteringCoefficient calculates the average clustering coefficient
func (t *Topology) calculateClusteringCoefficient() float64 {
	// TODO: Implement clustering coefficient calculation
	// This is a simplified placeholder implementation
	return 0.0
}

// StartTracking starts topology tracking
func (t *Topology) StartTracking(ctx context.Context) {
	// TODO: Implement in task 7.3
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.performTopologyUpdate()
		case <-t.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// StopTracking stops topology tracking
func (t *Topology) StopTracking() {
	close(t.stopCh)
}

// performTopologyUpdate performs periodic topology updates
func (t *Topology) performTopologyUpdate() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Update last seen times and check for inactive nodes
	now := time.Now()
	for peerID, nodeInfo := range t.nodes {
		// Mark nodes as inactive if not seen for 5 minutes
		if now.Sub(nodeInfo.LastSeen) > 5*time.Minute {
			nodeInfo.IsActive = false
		}

		// Remove nodes that have been inactive for 30 minutes
		if now.Sub(nodeInfo.LastSeen) > 30*time.Minute {
			delete(t.nodes, peerID)
			delete(t.edges, peerID.String())
		}
	}

	t.updateStats()
}

// GetNodeInfo returns information about a specific node
func (t *Topology) GetNodeInfo(peerID peer.ID) *NodeInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if nodeInfo, exists := t.nodes[peerID]; exists {
		// Return a copy to prevent external modification
		return &NodeInfo{
			PeerID:      nodeInfo.PeerID,
			ConnectedAt: nodeInfo.ConnectedAt,
			LastSeen:    nodeInfo.LastSeen,
			IsActive:    nodeInfo.IsActive,
			Connections: append([]peer.ID{}, nodeInfo.Connections...),
		}
	}

	return nil
}

// GetConnectedPeers returns all connected peers for a given peer
func (t *Topology) GetConnectedPeers(peerID peer.ID) []peer.ID {
	t.mu.RLock()
	defer t.mu.RUnlock()

	peerIDStr := peerID.String()
	if connections, exists := t.edges[peerIDStr]; exists {
		peers := make([]peer.ID, 0, len(connections))
		for _, connStr := range connections {
			if connPeerID, err := peer.Decode(connStr); err == nil {
				peers = append(peers, connPeerID)
			}
		}
		return peers
	}

	return []peer.ID{}
}
