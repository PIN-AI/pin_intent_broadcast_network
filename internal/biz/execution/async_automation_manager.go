package execution

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"pin_intent_broadcast_network/internal/biz/block_builder"
	"pin_intent_broadcast_network/internal/biz/service_agent"
	"pin_intent_broadcast_network/internal/p2p"
	"pin_intent_broadcast_network/internal/transport"
	"go.uber.org/zap"
)

// AsyncInitializationConfig holds configuration for async initialization
type AsyncInitializationConfig struct {
	Enabled                    bool          `yaml:"enabled"`
	TransportReadinessTimeout  time.Duration `yaml:"transport_readiness_timeout"`
	ComponentStartTimeout      time.Duration `yaml:"component_start_timeout"`
	MaxInitRetries             int           `yaml:"max_init_retries"`
	RetryBackoffInterval       time.Duration `yaml:"retry_backoff_interval"`
}

// DefaultAsyncInitializationConfig returns default async initialization configuration
func DefaultAsyncInitializationConfig() *AsyncInitializationConfig {
	return &AsyncInitializationConfig{
		Enabled:                    true,
		TransportReadinessTimeout:  60 * time.Second,
		ComponentStartTimeout:      30 * time.Second,
		MaxInitRetries:             5,
		RetryBackoffInterval:       2 * time.Second,
	}
}

// AsyncComponentWrapper wraps service agents and builders to implement LifecycleComponent
type AsyncComponentWrapper struct {
	id           string
	componentType string
	state        ComponentState
	mu           sync.RWMutex
	
	// Service Agent or Block Builder
	agent   *service_agent.Agent
	builder *block_builder.BlockBuilder
}

// NewAsyncAgentWrapper creates a wrapper for service agent
func NewAsyncAgentWrapper(agent *service_agent.Agent, agentID string) *AsyncComponentWrapper {
	return &AsyncComponentWrapper{
		id:            agentID,
		componentType: "service_agent",
		state:         ComponentStatePending,
		agent:         agent,
	}
}

// NewAsyncBuilderWrapper creates a wrapper for block builder
func NewAsyncBuilderWrapper(builder *block_builder.BlockBuilder, builderID string) *AsyncComponentWrapper {
	return &AsyncComponentWrapper{
		id:            builderID,
		componentType: "block_builder",
		state:         ComponentStatePending,
		builder:       builder,
	}
}

// LifecycleComponent interface implementation
func (acw *AsyncComponentWrapper) GetID() string {
	return acw.id
}

func (acw *AsyncComponentWrapper) Start(ctx context.Context) error {
	acw.mu.Lock()
	acw.state = ComponentStateStarting
	acw.mu.Unlock()

	var err error
	if acw.agent != nil {
		err = acw.agent.Start(ctx)
	} else if acw.builder != nil {
		err = acw.builder.Start(ctx)
	} else {
		err = fmt.Errorf("no component to start")
	}

	acw.mu.Lock()
	if err != nil {
		acw.state = ComponentStateError
	} else {
		acw.state = ComponentStateRunning
	}
	acw.mu.Unlock()

	return err
}

func (acw *AsyncComponentWrapper) Stop() error {
	acw.mu.Lock()
	acw.state = ComponentStateStopping
	acw.mu.Unlock()

	var err error
	if acw.agent != nil {
		err = acw.agent.Stop()
	} else if acw.builder != nil {
		err = acw.builder.Stop()
	}

	acw.mu.Lock()
	if err != nil {
		acw.state = ComponentStateError
	} else {
		acw.state = ComponentStateStopped
	}
	acw.mu.Unlock()

	return err
}

func (acw *AsyncComponentWrapper) IsRunning() bool {
	if acw.agent != nil {
		return acw.agent.IsRunning()
	} else if acw.builder != nil {
		return acw.builder.IsRunning()
	}
	return false
}

func (acw *AsyncComponentWrapper) GetState() ComponentState {
	acw.mu.RLock()
	defer acw.mu.RUnlock()
	return acw.state
}

// AsyncAutomationManager manages async initialization of automation system
type AsyncAutomationManager struct {
	// Base components
	baseManager      *AutomationManager
	networkManager   p2p.NetworkManager
	transportMgr     transport.TransportManager
	readinessChecker *transport.TransportReadinessChecker
	lifecycleManager *ComponentLifecycleManager
	logger           *zap.Logger
	
	// Configuration
	config *AsyncInitializationConfig
	
	// Async state management
	mu                  sync.RWMutex
	isInitialized       atomic.Bool
	isInitializing      atomic.Bool
	initErrors          chan error
	initWaitGroup       sync.WaitGroup
	initContext         context.Context
	initContextCancel   context.CancelFunc
	
	// Component tracking
	componentWrappers   map[string]*AsyncComponentWrapper
	serviceAgentPhase   atomic.Bool
	blockBuilderPhase   atomic.Bool
	
	// Status tracking
	initStartTime       time.Time
	lastInitError       string
	initRetryCount      int
}

// NewAsyncAutomationManager creates a new async automation manager
func NewAsyncAutomationManager(
	baseManager *AutomationManager,
	networkManager p2p.NetworkManager,
	transportMgr transport.TransportManager,
	logger *zap.Logger,
) *AsyncAutomationManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	readinessChecker := transport.NewTransportReadinessChecker(
		networkManager,
		transportMgr,
		logger,
	)

	lifecycleManager := NewComponentLifecycleManager(logger)

	aam := &AsyncAutomationManager{
		baseManager:       baseManager,
		networkManager:    networkManager,
		transportMgr:      transportMgr,
		readinessChecker:  readinessChecker,
		lifecycleManager:  lifecycleManager,
		logger:            logger.Named("async_automation"),
		config:            DefaultAsyncInitializationConfig(),
		initErrors:        make(chan error, 10),
		componentWrappers: make(map[string]*AsyncComponentWrapper),
	}

	// Setup lifecycle manager callbacks
	aam.setupLifecycleCallbacks()

	return aam
}

// SetConfig updates the async initialization configuration
func (aam *AsyncAutomationManager) SetConfig(config *AsyncInitializationConfig) {
	aam.mu.Lock()
	defer aam.mu.Unlock()
	aam.config = config
}

// StartAsync starts the automation manager asynchronously
func (aam *AsyncAutomationManager) StartAsync(ctx context.Context) error {
	if !aam.config.Enabled {
		aam.logger.Info("Async initialization disabled, falling back to synchronous start")
		return aam.baseManager.Start(ctx)
	}

	if !aam.isInitializing.CompareAndSwap(false, true) {
		return fmt.Errorf("async initialization already in progress")
	}

	aam.logger.Info("Starting async automation initialization")
	aam.mu.Lock()
	aam.initStartTime = time.Now()
	aam.initContext, aam.initContextCancel = context.WithCancel(ctx)
	aam.mu.Unlock()

	// Start async initialization in background
	go aam.performAsyncInitialization()

	return nil
}

// WaitForInitialization waits for the async initialization to complete
func (aam *AsyncAutomationManager) WaitForInitialization(timeout time.Duration) error {
	if aam.isInitialized.Load() {
		return nil
	}

	timeoutChan := make(chan struct{})
	if timeout > 0 {
		go func() {
			time.Sleep(timeout)
			close(timeoutChan)
		}()
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutChan:
			if timeout > 0 {
				return fmt.Errorf("initialization timeout after %v", timeout)
			}
		case <-ticker.C:
			if aam.isInitialized.Load() {
				return nil
			}
			// Check for errors
			select {
			case err := <-aam.initErrors:
				return fmt.Errorf("initialization failed: %w", err)
			default:
			}
		}
	}
}

// IsInitialized returns true if async initialization is complete
func (aam *AsyncAutomationManager) IsInitialized() bool {
	return aam.isInitialized.Load()
}

// IsInitializing returns true if async initialization is in progress  
func (aam *AsyncAutomationManager) IsInitializing() bool {
	return aam.isInitializing.Load()
}

// GetInitializationStatus returns current initialization status
func (aam *AsyncAutomationManager) GetInitializationStatus() map[string]interface{} {
	aam.mu.RLock()
	defer aam.mu.RUnlock()

	status := map[string]interface{}{
		"is_initialized":       aam.isInitialized.Load(),
		"is_initializing":      aam.isInitializing.Load(),
		"init_start_time":      aam.initStartTime,
		"service_agent_phase":  aam.serviceAgentPhase.Load(),
		"block_builder_phase":  aam.blockBuilderPhase.Load(),
		"init_retry_count":     aam.initRetryCount,
		"last_init_error":      aam.lastInitError,
	}

	if aam.readinessChecker != nil {
		status["transport_health"] = aam.readinessChecker.GetLastHealthStatus()
	}

	if aam.lifecycleManager != nil {
		status["component_status"] = aam.lifecycleManager.GetAllComponentsStatus()
	}

	return status
}

// Delegate methods to base manager
func (aam *AsyncAutomationManager) IsRunning() bool {
	return aam.baseManager.IsRunning()
}

func (aam *AsyncAutomationManager) GetStatus() *AutomationStatus {
	status := aam.baseManager.GetStatus()
	
	// Enhance with async status
	if aam.isInitializing.Load() {
		status.Status = "initializing"
	}
	
	return status
}

func (aam *AsyncAutomationManager) GetMetrics() *AutomationMetrics {
	return aam.baseManager.GetMetrics()
}

func (aam *AsyncAutomationManager) GetAgents() map[string]*service_agent.Agent {
	return aam.baseManager.GetAgents()
}

func (aam *AsyncAutomationManager) GetBuilders() map[string]*block_builder.BlockBuilder {
	return aam.baseManager.GetBuilders()
}

func (aam *AsyncAutomationManager) StartAgent(ctx context.Context, agentID string) error {
	return aam.baseManager.StartAgent(ctx, agentID)
}

func (aam *AsyncAutomationManager) StopAgent(agentID string) error {
	return aam.baseManager.StopAgent(agentID)
}

func (aam *AsyncAutomationManager) StartBuilder(ctx context.Context, builderID string) error {
	return aam.baseManager.StartBuilder(ctx, builderID)
}

func (aam *AsyncAutomationManager) StopBuilder(builderID string) error {
	return aam.baseManager.StopBuilder(builderID)
}

func (aam *AsyncAutomationManager) Stop() error {
	aam.logger.Info("Stopping async automation manager")

	// Cancel async initialization if running
	aam.mu.Lock()
	if aam.initContextCancel != nil {
		aam.initContextCancel()
	}
	aam.mu.Unlock()

	// Stop lifecycle manager
	if aam.lifecycleManager != nil {
		if err := aam.lifecycleManager.StopComponents(); err != nil {
			aam.logger.Error("Failed to stop lifecycle components", zap.Error(err))
		}
	}

	// Stop base manager
	return aam.baseManager.Stop()
}

// Internal methods

// performAsyncInitialization performs the actual async initialization
func (aam *AsyncAutomationManager) performAsyncInitialization() {
	defer aam.isInitializing.Store(false)

	retryCount := 0
	for retryCount < aam.config.MaxInitRetries {
		if err := aam.runInitializationPhases(); err != nil {
			retryCount++
			aam.mu.Lock()
			aam.initRetryCount = retryCount
			aam.lastInitError = err.Error()
			aam.mu.Unlock()

			aam.logger.Error("Initialization attempt failed",
				zap.Int("retry", retryCount),
				zap.Int("max_retries", aam.config.MaxInitRetries),
				zap.Error(err),
			)

			if retryCount < aam.config.MaxInitRetries {
				// Clean up partial state before retrying
				aam.cleanupPartialState()
				
				select {
				case <-aam.initContext.Done():
					aam.sendInitError(aam.initContext.Err())
					return
				case <-time.After(aam.config.RetryBackoffInterval):
					continue
				}
			} else {
				aam.sendInitError(fmt.Errorf("initialization failed after %d retries: %w", aam.config.MaxInitRetries, err))
				return
			}
		} else {
			// Initialization successful
			aam.isInitialized.Store(true)
			aam.logger.Info("Async automation initialization completed successfully",
				zap.Int("retry_count", retryCount),
				zap.Duration("total_time", time.Since(aam.initStartTime)),
			)
			return
		}
	}
}

// runInitializationPhases runs the initialization phases in order
func (aam *AsyncAutomationManager) runInitializationPhases() error {
	// Phase 1: Wait for transport readiness
	aam.logger.Info("Phase 1: Waiting for transport readiness")
	readinessCtx, cancel := context.WithTimeout(aam.initContext, aam.config.TransportReadinessTimeout)
	if err := aam.readinessChecker.WaitForTransportReady(readinessCtx); err != nil {
		cancel()
		return fmt.Errorf("transport readiness check failed: %w", err)
	}
	cancel()

	// Phase 2: Load configurations and prepare components
	aam.logger.Info("Phase 2: Loading configurations and preparing components")
	if err := aam.prepareComponents(); err != nil {
		return fmt.Errorf("component preparation failed: %w", err)
	}

	// Phase 3: Start service agents
	aam.logger.Info("Phase 3: Starting service agents")
	if err := aam.startServiceAgents(); err != nil {
		return fmt.Errorf("service agent startup failed: %w", err)
	}
	aam.serviceAgentPhase.Store(true)

	// Phase 4: Start block builders
	aam.logger.Info("Phase 4: Starting block builders")
	if err := aam.startBlockBuilders(); err != nil {
		return fmt.Errorf("block builder startup failed: %w", err)
	}
	aam.blockBuilderPhase.Store(true)

	// Phase 5: Start background tasks
	aam.logger.Info("Phase 5: Starting background tasks")
	if err := aam.startBackgroundTasks(); err != nil {
		return fmt.Errorf("background task startup failed: %w", err)
	}

	return nil
}

// prepareComponents loads configurations and prepares component wrappers
func (aam *AsyncAutomationManager) prepareComponents() error {
	// Load configurations using base manager
	aam.baseManager.mu.Lock()
	defer aam.baseManager.mu.Unlock()

	if err := aam.baseManager.loadConfigurations(); err != nil {
		return fmt.Errorf("failed to load configurations: %w", err)
	}

	// Prepare service agent wrappers
	if aam.baseManager.config != nil && aam.baseManager.config.Automation.Enabled {
		for _, agentConfig := range aam.baseManager.config.Agents {
			saConfig := aam.baseManager.convertToServiceAgentConfig(agentConfig)
			agent := service_agent.NewAgent(saConfig, aam.baseManager.transportMgr, aam.logger)
			
			wrapper := NewAsyncAgentWrapper(agent, agentConfig.AgentID)
			aam.componentWrappers[agentConfig.AgentID] = wrapper
			aam.baseManager.agents[agentConfig.AgentID] = agent

			// Register with lifecycle manager
			priority := 1 // Service agents have priority 1
			if err := aam.lifecycleManager.RegisterComponent(wrapper, priority); err != nil {
				return fmt.Errorf("failed to register service agent %s: %w", agentConfig.AgentID, err)
			}
		}
	}

	// Prepare block builder wrappers
	if aam.baseManager.config != nil && aam.baseManager.config.Builders.Enabled {
		for _, builderConfig := range aam.baseManager.config.Builders.Configs {
			bbConfig := aam.baseManager.convertToBlockBuilderConfig(builderConfig)
			builder := block_builder.NewBlockBuilder(bbConfig, aam.baseManager.transportMgr, aam.logger)
			
			wrapper := NewAsyncBuilderWrapper(builder, builderConfig.BuilderID)
			aam.componentWrappers[builderConfig.BuilderID] = wrapper
			aam.baseManager.builders[builderConfig.BuilderID] = builder

			// Register with lifecycle manager
			priority := 2 // Block builders have priority 2 (after agents)
			if err := aam.lifecycleManager.RegisterComponent(wrapper, priority); err != nil {
				return fmt.Errorf("failed to register block builder %s: %w", builderConfig.BuilderID, err)
			}

			// Set dependency on service agents (builders depend on at least one agent being ready)
			// For now, we don't enforce specific dependencies, just priority ordering
		}
	}

	return nil
}

// startServiceAgents starts service agents using lifecycle manager
func (aam *AsyncAutomationManager) startServiceAgents() error {
	// Filter and start only service agents
	agentComponents := make([]LifecycleComponent, 0)
	for id, wrapper := range aam.componentWrappers {
		if wrapper.componentType == "service_agent" {
			if info, err := aam.lifecycleManager.GetComponentStatus(id); err == nil && info.Priority == 1 {
				agentComponents = append(agentComponents, wrapper)
			}
		}
	}

	if len(agentComponents) == 0 {
		aam.logger.Info("No service agents to start")
		return nil
	}

	// Create background context for long-running agents (not tied to startup timeout)
	backgroundCtx, _ := context.WithCancel(aam.initContext)

	for _, component := range agentComponents {
		// Use backgroundCtx for long-running operations
		if err := component.Start(backgroundCtx); err != nil {
			return fmt.Errorf("failed to start service agent %s: %w", component.GetID(), err)
		}
		aam.logger.Info("Service agent started", zap.String("agent_id", component.GetID()))
	}

	return nil
}

// startBlockBuilders starts block builders using lifecycle manager
func (aam *AsyncAutomationManager) startBlockBuilders() error {
	// Filter and start only block builders
	builderComponents := make([]LifecycleComponent, 0)
	for id, wrapper := range aam.componentWrappers {
		if wrapper.componentType == "block_builder" {
			if info, err := aam.lifecycleManager.GetComponentStatus(id); err == nil && info.Priority == 2 {
				builderComponents = append(builderComponents, wrapper)
			}
		}
	}

	if len(builderComponents) == 0 {
		aam.logger.Info("No block builders to start")
		return nil
	}

	// Start builders with timeout
	startCtx, cancel := context.WithTimeout(aam.initContext, aam.config.ComponentStartTimeout)
	defer cancel()

	for _, component := range builderComponents {
		if err := component.Start(startCtx); err != nil {
			return fmt.Errorf("failed to start block builder %s: %w", component.GetID(), err)
		}
		aam.logger.Info("Block builder started", zap.String("builder_id", component.GetID()))
	}

	return nil
}

// startBackgroundTasks starts background monitoring tasks
func (aam *AsyncAutomationManager) startBackgroundTasks() error {
	// Update base manager state
	aam.baseManager.mu.Lock()
	aam.baseManager.isRunning = true
	aam.baseManager.status.Status = "running"
	aam.baseManager.status.HealthCheckPassing = true
	
	// Create background context for tasks
	backgroundCtx, cancelFunc := context.WithCancel(aam.initContext)
	aam.baseManager.cancelFunc = cancelFunc
	aam.baseManager.mu.Unlock()

	// Start background tasks from base manager
	aam.baseManager.startBackgroundTasks(backgroundCtx)

	return nil
}

// setupLifecycleCallbacks sets up callbacks for lifecycle events
func (aam *AsyncAutomationManager) setupLifecycleCallbacks() {
	// Register transport readiness callback
	aam.readinessChecker.RegisterReadinessCallback(func(tm transport.TransportManager) {
		aam.logger.Info("Transport ready callback triggered")
	})

	// Register component state change callbacks would be added here
	// For now, we rely on the base manager's status tracking
}

// cleanupPartialState cleans up any partial initialization state before retrying
func (aam *AsyncAutomationManager) cleanupPartialState() {
	aam.logger.Info("Cleaning up partial initialization state for retry")
	
	// Stop and clear lifecycle manager components
	if aam.lifecycleManager != nil {
		if err := aam.lifecycleManager.StopComponents(); err != nil {
			aam.logger.Warn("Failed to stop components during cleanup", zap.Error(err))
		}
		aam.lifecycleManager.ClearComponents()
	}
	
	// Reset base automation manager
	if aam.baseManager != nil {
		if err := aam.baseManager.Stop(); err != nil {
			aam.logger.Warn("Failed to stop base manager during cleanup", zap.Error(err))
		}
		// Reset the base manager to clean state
		aam.baseManager = NewAutomationManager(aam.transportMgr, aam.logger)
	}
}

// sendInitError sends an error to the error channel
func (aam *AsyncAutomationManager) sendInitError(err error) {
	select {
	case aam.initErrors <- err:
	default:
		// Channel full, log the error
		aam.logger.Error("Init error channel full, dropping error", zap.Error(err))
	}
}