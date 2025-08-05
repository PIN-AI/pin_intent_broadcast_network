package service_agent

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"pin_intent_broadcast_network/internal/transport"
	"go.uber.org/zap"
)

// Agent represents a service agent that listens to intents and submits bids
type Agent struct {
	config         *AgentConfig
	transportMgr   transport.TransportManager
	intentListener *IntentListener
	logger         *zap.Logger
	
	// State management
	mu          sync.RWMutex
	status      *AgentStatus
	metrics     *AgentMetrics
	isRunning   bool
	
	// Active intents tracking
	activeIntents map[string]*IntentEvent
}

// NewAgent creates a new service agent
func NewAgent(config *AgentConfig, transportMgr transport.TransportManager, logger *zap.Logger) *Agent {
	if config == nil {
		config = DefaultAgentConfig()
	}
	
	if logger == nil {
		logger = zap.NewNop()
	}
	
	return &Agent{
		config:       config,
		transportMgr: transportMgr,
		logger:       logger.Named("service_agent"),
		status: &AgentStatus{
			AgentID:          config.AgentID,
			Status:           "offline",
			ActiveIntents:    0,
			ProcessedIntents: 0,
			SuccessfulBids:   0,
			TotalEarnings:    "0",
			LastActivity:     time.Now(),
			ConnectedPeers:   0,
		},
		metrics: &AgentMetrics{
			IntentsReceived:     0,
			IntentsFiltered:     0,
			BidsSubmitted:       0,
			BidsWon:             0,
			TotalEarnings:       "0",
			AverageConfidence:   0.0,
			AverageResponseTime: 0,
			LastUpdated:         time.Now(),
		},
		activeIntents: make(map[string]*IntentEvent),
	}
}

// Start starts the service agent
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.isRunning {
		return fmt.Errorf("agent %s already running", a.config.AgentID)
	}
	
	// Initialize intent listener
	a.intentListener = NewIntentListener(a.config, a.transportMgr, a.logger)
	
	// Start intent listener
	if err := a.intentListener.Start(ctx); err != nil {
		return fmt.Errorf("failed to start intent listener: %w", err)
	}
	
	// Subscribe to match results to track bid outcomes
	_, err := a.transportMgr.SubscribeToMatches(a.handleMatchResult)
	if err != nil {
		a.logger.Error("Failed to subscribe to match results", zap.Error(err))
		// Not critical - continue without match result tracking
	}
	
	// Start background tasks
	go a.metricsUpdater(ctx)
	go a.statusUpdater(ctx)
	
	a.isRunning = true
	a.status.Status = "active"
	a.status.LastActivity = time.Now()
	
	a.logger.Info("Service agent started successfully",
		zap.String("agent_id", a.config.AgentID),
		zap.String("agent_type", string(a.config.AgentType)),
		zap.Int("max_concurrent_intents", a.config.MaxConcurrentIntents),
	)
	
	return nil
}

// Stop stops the service agent
func (a *Agent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if !a.isRunning {
		return fmt.Errorf("agent %s not running", a.config.AgentID)
	}
	
	// Stop intent listener
	if a.intentListener != nil {
		if err := a.intentListener.Stop(); err != nil {
			a.logger.Error("Failed to stop intent listener", zap.Error(err))
		}
	}
	
	a.isRunning = false
	a.status.Status = "offline"
	a.status.LastActivity = time.Now()
	
	a.logger.Info("Service agent stopped",
		zap.String("agent_id", a.config.AgentID),
	)
	
	return nil
}

// IsRunning returns whether the agent is running
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isRunning
}

// GetStatus returns the current agent status
func (a *Agent) GetStatus() *AgentStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Update connected peers count
	if a.transportMgr != nil {
		metrics := a.transportMgr.GetTransportMetrics()
		a.status.ConnectedPeers = metrics.ConnectedPeerCount
	}
	
	// Create a copy to avoid concurrent access issues
	status := *a.status
	return &status
}

// GetMetrics returns the current agent metrics
func (a *Agent) GetMetrics() *AgentMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Create a copy to avoid concurrent access issues
	metrics := *a.metrics
	return &metrics
}

// GetConfig returns the agent configuration
func (a *Agent) GetConfig() *AgentConfig {
	return a.config
}

// UpdateConfig updates the agent configuration
func (a *Agent) UpdateConfig(config *AgentConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if config.AgentID != a.config.AgentID {
		return fmt.Errorf("cannot change agent ID")
	}
	
	a.config = config
	a.logger.Info("Agent configuration updated",
		zap.String("agent_id", a.config.AgentID),
	)
	
	return nil
}

// handleMatchResult handles incoming match results to track bid outcomes
func (a *Agent) handleMatchResult(result *transport.MatchResult) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Check if this agent won the bid
	if result.WinningAgent == a.config.AgentID {
		a.metrics.BidsWon++
		a.status.SuccessfulBids++
		
		// Update earnings (simplified calculation)
		// In a real implementation, this would involve more complex accounting
		a.logger.Info("Won bid for intent",
			zap.String("intent_id", result.IntentID),
			zap.String("winning_bid", result.WinningBid),
		)
	}
	
	// Remove from active intents if tracking
	delete(a.activeIntents, result.IntentID)
	a.status.ActiveIntents = len(a.activeIntents)
	
	return nil
}

// metricsUpdater periodically updates agent metrics
func (a *Agent) metricsUpdater(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.updateMetrics()
		}
	}
}

// statusUpdater periodically updates agent status
func (a *Agent) statusUpdater(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.updateStatus()
		}
	}
}

// updateMetrics updates internal metrics
func (a *Agent) updateMetrics() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.metrics.LastUpdated = time.Now()
	
	// Calculate average confidence
	if a.metrics.BidsSubmitted > 0 {
		// This would need to be tracked during bid submission
		// For now, use a placeholder calculation
	}
	
	a.logger.Debug("Metrics updated",
		zap.String("agent_id", a.config.AgentID),
		zap.Int64("bids_submitted", a.metrics.BidsSubmitted),
		zap.Int64("bids_won", a.metrics.BidsWon),
	)
}

// updateStatus updates internal status
func (a *Agent) updateStatus() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.status.LastActivity = time.Now()
	a.status.ActiveIntents = len(a.activeIntents)
	
	// Update status based on activity
	if len(a.activeIntents) >= a.config.MaxConcurrentIntents {
		a.status.Status = "busy"
	} else if a.isRunning {
		a.status.Status = "active"
	} else {
		a.status.Status = "offline"
	}
}