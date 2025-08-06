package execution

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"pin_intent_broadcast_network/internal/biz/block_builder"
	"pin_intent_broadcast_network/internal/biz/service_agent"
	"pin_intent_broadcast_network/internal/transport"
)

// AutomationManager manages the lifecycle of service agents and block builders
type AutomationManager struct {
	// Dependencies
	transportMgr transport.TransportManager
	logger       *zap.Logger

	// Configuration (unified)
	config *AgentsConfig

	// Runtime state
	mu        sync.RWMutex
	isRunning bool
	agents    map[string]*service_agent.Agent
	builders  map[string]*block_builder.BlockBuilder
	metrics   *AutomationMetrics
	status    *AutomationStatus

	// Background tasks
	metricsUpdateTicker *time.Ticker
	statusUpdateTicker  *time.Ticker
	cancelFunc          context.CancelFunc
}

// AutomationMetrics tracks overall automation performance
type AutomationMetrics struct {
	// Agent metrics
	TotalAgents          int32   `json:"total_agents"`
	ActiveAgents         int32   `json:"active_agents"`
	TotalIntentsReceived int64   `json:"total_intents_received"`
	TotalBidsSubmitted   int64   `json:"total_bids_submitted"`
	TotalBidsWon         int64   `json:"total_bids_won"`
	AgentSuccessRate     float64 `json:"agent_success_rate"`

	// Builder metrics
	TotalBuilders         int32   `json:"total_builders"`
	ActiveBuilders        int32   `json:"active_builders"`
	TotalSessionsCreated  int64   `json:"total_sessions_created"`
	TotalMatchesCompleted int64   `json:"total_matches_completed"`
	TotalSessionsExpired  int64   `json:"total_sessions_expired"`
	BuilderSuccessRate    float64 `json:"builder_success_rate"`

	// Overall metrics
	AverageResponseTime int64     `json:"average_response_time_ms"`
	SystemUptime        int64     `json:"system_uptime_seconds"`
	LastUpdated         time.Time `json:"last_updated"`
}

// AutomationStatus tracks current system status
type AutomationStatus struct {
	Status             string            `json:"status"` // "starting", "running", "stopping", "stopped", "error"
	StartTime          time.Time         `json:"start_time"`
	LastActivity       time.Time         `json:"last_activity"`
	ConnectedPeers     int32             `json:"connected_peers"`
	P2PStatus          string            `json:"p2p_status"`
	ComponentStatus    map[string]string `json:"component_status"`
	ErrorCount         int32             `json:"error_count"`
	LastError          string            `json:"last_error"`
	HealthCheckPassing bool              `json:"health_check_passing"`
}

// NewAutomationManager creates a new automation manager
func NewAutomationManager(transportMgr transport.TransportManager, logger *zap.Logger) *AutomationManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &AutomationManager{
		transportMgr: transportMgr,
		logger:       logger.Named("automation_manager"),
		agents:       make(map[string]*service_agent.Agent),
		builders:     make(map[string]*block_builder.BlockBuilder),
		metrics: &AutomationMetrics{
			LastUpdated: time.Now(),
		},
		status: &AutomationStatus{
			Status:          "stopped",
			ComponentStatus: make(map[string]string),
			LastActivity:    time.Now(),
		},
	}
}

// Start starts the automation manager and all configured components
func (am *AutomationManager) Start(ctx context.Context) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.isRunning {
		return fmt.Errorf("automation manager already running")
	}

	am.logger.Info("Starting PIN automation manager...")
	am.status.Status = "starting"
	am.status.StartTime = time.Now()
	am.status.LastActivity = time.Now()

	// Load configurations
	if err := am.loadConfigurations(); err != nil {
		am.status.Status = "error"
		am.status.LastError = err.Error()
		am.status.ErrorCount++
		return fmt.Errorf("failed to load configurations: %w", err)
	}

	// Create context for background tasks
	backgroundCtx, cancelFunc := context.WithCancel(ctx)
	am.cancelFunc = cancelFunc

	// Start agents
	if err := am.startAgents(backgroundCtx); err != nil {
		am.status.Status = "error"
		am.status.LastError = err.Error()
		am.status.ErrorCount++
		cancelFunc()
		return fmt.Errorf("failed to start agents: %w", err)
	}

	// Start builders
	if err := am.startBuilders(backgroundCtx); err != nil {
		am.status.Status = "error"
		am.status.LastError = err.Error()
		am.status.ErrorCount++
		cancelFunc()
		return fmt.Errorf("failed to start builders: %w", err)
	}

	// Start background tasks
	am.startBackgroundTasks(backgroundCtx)

	am.isRunning = true
	am.status.Status = "running"
	am.status.HealthCheckPassing = true

	am.logger.Info("PIN automation manager started successfully",
		zap.Int("total_agents", len(am.agents)),
		zap.Int("total_builders", len(am.builders)),
	)

	return nil
}

// Stop stops the automation manager and all components
func (am *AutomationManager) Stop() error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.isRunning {
		return fmt.Errorf("automation manager not running")
	}

	am.logger.Info("Stopping PIN automation manager...")
	am.status.Status = "stopping"
	am.status.LastActivity = time.Now()

	// Cancel background tasks
	if am.cancelFunc != nil {
		am.cancelFunc()
	}

	// Stop background tickers
	if am.metricsUpdateTicker != nil {
		am.metricsUpdateTicker.Stop()
	}
	if am.statusUpdateTicker != nil {
		am.statusUpdateTicker.Stop()
	}

	// Stop all agents
	for agentID, agent := range am.agents {
		if err := agent.Stop(); err != nil {
			am.logger.Error("Failed to stop agent",
				zap.String("agent_id", agentID),
				zap.Error(err),
			)
			am.status.ErrorCount++
		} else {
			am.status.ComponentStatus[agentID] = "stopped"
		}
	}

	// Stop all builders
	for builderID, builder := range am.builders {
		if err := builder.Stop(); err != nil {
			am.logger.Error("Failed to stop builder",
				zap.String("builder_id", builderID),
				zap.Error(err),
			)
			am.status.ErrorCount++
		} else {
			am.status.ComponentStatus[builderID] = "stopped"
		}
	}

	am.isRunning = false
	am.status.Status = "stopped"
	am.status.HealthCheckPassing = false

	am.logger.Info("PIN automation manager stopped")
	return nil
}

// IsRunning returns whether the automation manager is running
func (am *AutomationManager) IsRunning() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.isRunning
}

// GetStatus returns the current automation status
func (am *AutomationManager) GetStatus() *AutomationStatus {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Update P2P status
	if am.transportMgr != nil {
		transportMetrics := am.transportMgr.GetTransportMetrics()
		am.status.ConnectedPeers = int32(transportMetrics.ConnectedPeerCount)

		if transportMetrics.ConnectedPeerCount > 0 {
			am.status.P2PStatus = "connected"
		} else {
			am.status.P2PStatus = "disconnected"
		}
	}

	// Create a copy to avoid concurrent access issues
	status := *am.status
	status.ComponentStatus = make(map[string]string)
	for k, v := range am.status.ComponentStatus {
		status.ComponentStatus[k] = v
	}

	return &status
}

// GetMetrics returns current automation metrics
func (am *AutomationManager) GetMetrics() *AutomationMetrics {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Update metrics from components
	am.updateMetricsFromComponents()

	// Create a copy
	metrics := *am.metrics
	return &metrics
}

// GetAgents returns all registered agents
func (am *AutomationManager) GetAgents() map[string]*service_agent.Agent {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make(map[string]*service_agent.Agent)
	for k, v := range am.agents {
		result[k] = v
	}
	return result
}

// GetBuilders returns all registered builders
func (am *AutomationManager) GetBuilders() map[string]*block_builder.BlockBuilder {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make(map[string]*block_builder.BlockBuilder)
	for k, v := range am.builders {
		result[k] = v
	}
	return result
}

// StartAgent starts a specific agent
func (am *AutomationManager) StartAgent(ctx context.Context, agentID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	agent, exists := am.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	if agent.IsRunning() {
		return fmt.Errorf("agent %s already running", agentID)
	}

	if err := agent.Start(ctx); err != nil {
		am.status.ComponentStatus[agentID] = "error"
		am.status.ErrorCount++
		return fmt.Errorf("failed to start agent %s: %w", agentID, err)
	}

	am.status.ComponentStatus[agentID] = "running"
	am.logger.Info("Agent started", zap.String("agent_id", agentID))
	return nil
}

// StopAgent stops a specific agent
func (am *AutomationManager) StopAgent(agentID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	agent, exists := am.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	if !agent.IsRunning() {
		return fmt.Errorf("agent %s not running", agentID)
	}

	if err := agent.Stop(); err != nil {
		am.status.ComponentStatus[agentID] = "error"
		am.status.ErrorCount++
		return fmt.Errorf("failed to stop agent %s: %w", agentID, err)
	}

	am.status.ComponentStatus[agentID] = "stopped"
	am.logger.Info("Agent stopped", zap.String("agent_id", agentID))
	return nil
}

// StartBuilder starts a specific builder
func (am *AutomationManager) StartBuilder(ctx context.Context, builderID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	builder, exists := am.builders[builderID]
	if !exists {
		return fmt.Errorf("builder %s not found", builderID)
	}

	if builder.IsRunning() {
		return fmt.Errorf("builder %s already running", builderID)
	}

	if err := builder.Start(ctx); err != nil {
		am.status.ComponentStatus[builderID] = "error"
		am.status.ErrorCount++
		return fmt.Errorf("failed to start builder %s: %w", builderID, err)
	}

	am.status.ComponentStatus[builderID] = "running"
	am.logger.Info("Builder started", zap.String("builder_id", builderID))
	return nil
}

// StopBuilder stops a specific builder
func (am *AutomationManager) StopBuilder(builderID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	builder, exists := am.builders[builderID]
	if !exists {
		return fmt.Errorf("builder %s not found", builderID)
	}

	if !builder.IsRunning() {
		return fmt.Errorf("builder %s not running", builderID)
	}

	if err := builder.Stop(); err != nil {
		am.status.ComponentStatus[builderID] = "error"
		am.status.ErrorCount++
		return fmt.Errorf("failed to stop builder %s: %w", builderID, err)
	}

	am.status.ComponentStatus[builderID] = "stopped"
	am.logger.Info("Builder stopped", zap.String("builder_id", builderID))
	return nil
}

// loadConfigurations loads unified agent and builder configuration from YAML file
func (am *AutomationManager) loadConfigurations() error {
	// Load unified configuration (agents + builders)
	configData, err := os.ReadFile("configs/agents_config.yaml")
	if err != nil {
		return fmt.Errorf("failed to read agents config file: %w", err)
	}

	am.config = &AgentsConfig{}
	if err := yaml.Unmarshal(configData, am.config); err != nil {
		return fmt.Errorf("failed to parse agents config: %w", err)
	}

	am.logger.Info("Configurations loaded successfully",
		zap.Int("agents_count", len(am.config.Agents)),
		zap.Int("builders_count", len(am.config.Builders.Configs)),
		zap.Bool("builders_enabled", am.config.Builders.Enabled),
	)

	return nil
}

// startAgents initializes and starts all configured agents
func (am *AutomationManager) startAgents(ctx context.Context) error {
	if !am.config.Automation.Enabled {
		am.logger.Info("Agent automation disabled in configuration")
		return nil
	}

	for _, agentConfig := range am.config.Agents {
		// Convert YAML config to service_agent.AgentConfig
		saConfig := am.convertToServiceAgentConfig(agentConfig)

		// Create agent
		agent := service_agent.NewAgent(saConfig, am.transportMgr, am.logger)
		am.agents[agentConfig.AgentID] = agent

		// Start agent if auto_start is enabled
		if agentConfig.AutoStart {
			if err := agent.Start(ctx); err != nil {
				am.logger.Error("Failed to start agent",
					zap.String("agent_id", agentConfig.AgentID),
					zap.Error(err),
				)
				am.status.ComponentStatus[agentConfig.AgentID] = "error"
				am.status.ErrorCount++
				continue
			}
			am.status.ComponentStatus[agentConfig.AgentID] = "running"
			am.logger.Info("Agent started successfully",
				zap.String("agent_id", agentConfig.AgentID),
				zap.String("agent_type", agentConfig.AgentType),
			)
		} else {
			am.status.ComponentStatus[agentConfig.AgentID] = "stopped"
			am.logger.Info("Agent created but not auto-started",
				zap.String("agent_id", agentConfig.AgentID),
			)
		}
	}

	return nil
}

// startBuilders initializes and starts all configured builders  
func (am *AutomationManager) startBuilders(ctx context.Context) error {
	if !am.config.Builders.Enabled {
		am.logger.Info("Builder automation disabled in configuration")
		return nil
	}

	for _, builderConfig := range am.config.Builders.Configs {
		// Convert YAML config to block_builder.BuilderConfig
		bbConfig := am.convertToBlockBuilderConfig(builderConfig)

		// Create builder
		builder := block_builder.NewBlockBuilder(bbConfig, am.transportMgr, am.logger)
		am.builders[builderConfig.BuilderID] = builder

		// Start builder if auto_start is enabled
		if builderConfig.AutoStart {
			if err := builder.Start(ctx); err != nil {
				am.logger.Error("Failed to start builder",
					zap.String("builder_id", builderConfig.BuilderID),
					zap.Error(err),
				)
				am.status.ComponentStatus[builderConfig.BuilderID] = "error"
				am.status.ErrorCount++
				continue
			}
			am.status.ComponentStatus[builderConfig.BuilderID] = "running"
			am.logger.Info("Builder started successfully",
				zap.String("builder_id", builderConfig.BuilderID),
				zap.String("matching_algorithm", builderConfig.MatchingAlgorithm),
			)
		} else {
			am.status.ComponentStatus[builderConfig.BuilderID] = "stopped"
			am.logger.Info("Builder created but not auto-started",
				zap.String("builder_id", builderConfig.BuilderID),
			)
		}
	}

	return nil
}

// startBackgroundTasks starts monitoring and maintenance tasks
func (am *AutomationManager) startBackgroundTasks(ctx context.Context) {
	// Start metrics update task
	am.metricsUpdateTicker = time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-am.metricsUpdateTicker.C:
				am.updateMetrics()
			}
		}
	}()

	// Start status update task
	am.statusUpdateTicker = time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-am.statusUpdateTicker.C:
				am.updateStatus()
			}
		}
	}()

	am.logger.Info("Background tasks started")
}

// updateMetrics updates automation metrics
func (am *AutomationManager) updateMetrics() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.updateMetricsFromComponents()
}

// updateMetricsFromComponents collects metrics from all components
func (am *AutomationManager) updateMetricsFromComponents() {
	// Reset counters
	am.metrics.TotalAgents = int32(len(am.agents))
	am.metrics.TotalBuilders = int32(len(am.builders))
	am.metrics.ActiveAgents = 0
	am.metrics.ActiveBuilders = 0
	am.metrics.TotalIntentsReceived = 0
	am.metrics.TotalBidsSubmitted = 0
	am.metrics.TotalBidsWon = 0
	am.metrics.TotalSessionsCreated = 0
	am.metrics.TotalMatchesCompleted = 0
	am.metrics.TotalSessionsExpired = 0

	// Collect agent metrics
	for _, agent := range am.agents {
		if agent.IsRunning() {
			am.metrics.ActiveAgents++
		}

		agentMetrics := agent.GetMetrics()
		if agentMetrics != nil {
			am.metrics.TotalIntentsReceived += agentMetrics.IntentsReceived
			am.metrics.TotalBidsSubmitted += agentMetrics.BidsSubmitted
			am.metrics.TotalBidsWon += agentMetrics.BidsWon
		}
	}

	// Collect builder metrics
	for _, builder := range am.builders {
		if builder.IsRunning() {
			am.metrics.ActiveBuilders++
		}

		builderMetrics := builder.GetMetrics()
		if builderMetrics != nil {
			am.metrics.TotalSessionsCreated += builderMetrics.SessionsCreated
			am.metrics.TotalMatchesCompleted += builderMetrics.MatchesCompleted
			am.metrics.TotalSessionsExpired += builderMetrics.SessionsExpired
		}
	}

	// Calculate success rates
	if am.metrics.TotalBidsSubmitted > 0 {
		am.metrics.AgentSuccessRate = float64(am.metrics.TotalBidsWon) / float64(am.metrics.TotalBidsSubmitted)
	}

	if am.metrics.TotalSessionsCreated > 0 {
		am.metrics.BuilderSuccessRate = float64(am.metrics.TotalMatchesCompleted) / float64(am.metrics.TotalSessionsCreated)
	}

	// Update system uptime
	if am.status.Status == "running" {
		am.metrics.SystemUptime = int64(time.Since(am.status.StartTime).Seconds())
	}

	am.metrics.LastUpdated = time.Now()
}

// updateStatus updates automation status
func (am *AutomationManager) updateStatus() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.status.LastActivity = time.Now()

	// Update component status
	for agentID, agent := range am.agents {
		if agent.IsRunning() {
			am.status.ComponentStatus[agentID] = "running"
		} else {
			am.status.ComponentStatus[agentID] = "stopped"
		}
	}

	for builderID, builder := range am.builders {
		if builder.IsRunning() {
			am.status.ComponentStatus[builderID] = "running"
		} else {
			am.status.ComponentStatus[builderID] = "stopped"
		}
	}

	// Update health check
	healthyComponents := 0
	totalComponents := len(am.agents) + len(am.builders)

	for _, status := range am.status.ComponentStatus {
		if status == "running" {
			healthyComponents++
		}
	}

	am.status.HealthCheckPassing = (healthyComponents > 0 && totalComponents > 0)
}
