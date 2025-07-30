package network

import (
	"context"
	"sync"
	"time"
)

// Status manages network status information
// This file will contain the implementation for task 7.2
type Status struct {
	connectedPeers   int64
	messagesSent     int64
	messagesReceived int64
	startTime        time.Time
	lastUpdate       time.Time
	healthStatus     string
	mu               sync.RWMutex
	stopCh           chan struct{}
}

// NewStatus creates a new network status manager
func NewStatus() *Status {
	return &Status{
		startTime:    time.Now(),
		lastUpdate:   time.Now(),
		healthStatus: "unknown",
		stopCh:       make(chan struct{}),
	}
}

// GetConnectedPeers returns the number of connected peers
func (s *Status) GetConnectedPeers() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectedPeers
}

// GetMessagesSent returns the number of messages sent
func (s *Status) GetMessagesSent() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.messagesSent
}

// GetMessagesReceived returns the number of messages received
func (s *Status) GetMessagesReceived() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.messagesReceived
}

// GetHealthStatus returns the current health status
func (s *Status) GetHealthStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.healthStatus
}

// GetUptime returns the network uptime
func (s *Status) GetUptime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.startTime)
}

// IncrementConnectedPeers increments the connected peers count
func (s *Status) IncrementConnectedPeers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connectedPeers++
	s.lastUpdate = time.Now()
	s.updateHealthStatus()
}

// DecrementConnectedPeers decrements the connected peers count
func (s *Status) DecrementConnectedPeers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.connectedPeers > 0 {
		s.connectedPeers--
	}
	s.lastUpdate = time.Now()
	s.updateHealthStatus()
}

// IncrementMessagesSent increments the messages sent count
func (s *Status) IncrementMessagesSent() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messagesSent++
	s.lastUpdate = time.Now()
}

// IncrementMessagesReceived increments the messages received count
func (s *Status) IncrementMessagesReceived() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messagesReceived++
	s.lastUpdate = time.Now()
}

// updateHealthStatus updates the health status based on current metrics
func (s *Status) updateHealthStatus() {
	// TODO: Implement in task 7.2
	// This method should be called with mutex already held

	if s.connectedPeers == 0 {
		s.healthStatus = "disconnected"
	} else if s.connectedPeers < 3 {
		s.healthStatus = "poor"
	} else if s.connectedPeers < 10 {
		s.healthStatus = "good"
	} else {
		s.healthStatus = "excellent"
	}

	// Check for recent activity
	if time.Since(s.lastUpdate) > 5*time.Minute {
		s.healthStatus = "stale"
	}
}

// StartMonitoring starts the status monitoring goroutine
func (s *Status) StartMonitoring(ctx context.Context) {
	// TODO: Implement in task 7.2
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			s.updateHealthStatus()
			s.mu.Unlock()
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// StopMonitoring stops the status monitoring
func (s *Status) StopMonitoring() {
	close(s.stopCh)
}

// GetStatusSnapshot returns a snapshot of current status
func (s *Status) GetStatusSnapshot() StatusSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return StatusSnapshot{
		ConnectedPeers:   s.connectedPeers,
		MessagesSent:     s.messagesSent,
		MessagesReceived: s.messagesReceived,
		HealthStatus:     s.healthStatus,
		Uptime:           time.Since(s.startTime),
		LastUpdate:       s.lastUpdate,
	}
}

// StatusSnapshot represents a point-in-time status snapshot
type StatusSnapshot struct {
	ConnectedPeers   int64         `json:"connected_peers"`
	MessagesSent     int64         `json:"messages_sent"`
	MessagesReceived int64         `json:"messages_received"`
	HealthStatus     string        `json:"health_status"`
	Uptime           time.Duration `json:"uptime"`
	LastUpdate       time.Time     `json:"last_update"`
}

// Reset resets all status counters
func (s *Status) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connectedPeers = 0
	s.messagesSent = 0
	s.messagesReceived = 0
	s.startTime = time.Now()
	s.lastUpdate = time.Now()
	s.healthStatus = "unknown"
}

// SetHealthStatus manually sets the health status
func (s *Status) SetHealthStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.healthStatus = status
	s.lastUpdate = time.Now()
}

// IsHealthy returns true if the network is considered healthy
func (s *Status) IsHealthy() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.healthStatus == "good" || s.healthStatus == "excellent"
}

// GetMessageRate returns the message rate (messages per second)
func (s *Status) GetMessageRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uptime := time.Since(s.startTime).Seconds()
	if uptime == 0 {
		return 0
	}

	totalMessages := s.messagesSent + s.messagesReceived
	return float64(totalMessages) / uptime
}
