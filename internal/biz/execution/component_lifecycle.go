package execution

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ComponentState represents the current state of a component
type ComponentState string

const (
	ComponentStatePending  ComponentState = "pending"
	ComponentStateStarting ComponentState = "starting" 
	ComponentStateRunning  ComponentState = "running"
	ComponentStateStopping ComponentState = "stopping"
	ComponentStateStopped  ComponentState = "stopped"
	ComponentStateError    ComponentState = "error"
)

// LifecycleComponent represents a component that can be managed by the lifecycle manager
type LifecycleComponent interface {
	// GetID returns the unique identifier for this component
	GetID() string
	
	// Start starts the component
	Start(ctx context.Context) error
	
	// Stop stops the component
	Stop() error
	
	// IsRunning returns true if the component is currently running
	IsRunning() bool
	
	// GetState returns the current component state
	GetState() ComponentState
}

// ComponentInfo holds metadata about a component
type ComponentInfo struct {
	Component    LifecycleComponent `json:"-"`
	ID           string             `json:"id"`
	State        ComponentState     `json:"state"`
	Priority     int                `json:"priority"`
	Dependencies []string           `json:"dependencies"`
	StartedAt    *time.Time         `json:"started_at,omitempty"`
	LastError    string             `json:"last_error,omitempty"`
	ErrorCount   int                `json:"error_count"`
}

// StateCallback is called when a component changes state
type StateCallback func(componentID string, oldState, newState ComponentState)

// ComponentLifecycleConfig holds configuration for the lifecycle manager
type ComponentLifecycleConfig struct {
	StartTimeout     time.Duration `yaml:"start_timeout"`
	StopTimeout      time.Duration `yaml:"stop_timeout"`
	MaxRetries       int           `yaml:"max_retries"`
	RetryInterval    time.Duration `yaml:"retry_interval"`
	DependencyWait   time.Duration `yaml:"dependency_wait"`
	ParallelStart    bool          `yaml:"parallel_start"`
}

// DefaultComponentLifecycleConfig returns default configuration
func DefaultComponentLifecycleConfig() *ComponentLifecycleConfig {
	return &ComponentLifecycleConfig{
		StartTimeout:   30 * time.Second,
		StopTimeout:    15 * time.Second,
		MaxRetries:     3,
		RetryInterval:  2 * time.Second,
		DependencyWait: 60 * time.Second,
		ParallelStart:  true,
	}
}

// ComponentLifecycleManager manages the lifecycle of multiple components
type ComponentLifecycleManager struct {
	logger *zap.Logger
	config *ComponentLifecycleConfig
	
	// Component management
	mu             sync.RWMutex
	components     map[string]*ComponentInfo
	dependencies   map[string][]string  // componentID -> list of dependency IDs
	startOrder     []string             // ordered list of component IDs by priority
	
	// State management
	stateCallbacks map[string][]StateCallback
	startWaitGroup sync.WaitGroup
	stopWaitGroup  sync.WaitGroup
}

// NewComponentLifecycleManager creates a new component lifecycle manager
func NewComponentLifecycleManager(logger *zap.Logger) *ComponentLifecycleManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ComponentLifecycleManager{
		logger:         logger.Named("lifecycle_manager"),
		config:         DefaultComponentLifecycleConfig(),
		components:     make(map[string]*ComponentInfo),
		dependencies:   make(map[string][]string),
		startOrder:     make([]string, 0),
		stateCallbacks: make(map[string][]StateCallback),
	}
}

// SetConfig updates the lifecycle manager configuration
func (clm *ComponentLifecycleManager) SetConfig(config *ComponentLifecycleConfig) {
	clm.mu.Lock()
	defer clm.mu.Unlock()
	clm.config = config
}

// RegisterComponent registers a component with the lifecycle manager
func (clm *ComponentLifecycleManager) RegisterComponent(component LifecycleComponent, priority int) error {
	if component == nil {
		return fmt.Errorf("component cannot be nil")
	}

	componentID := component.GetID()
	if componentID == "" {
		return fmt.Errorf("component ID cannot be empty")
	}

	clm.mu.Lock()
	defer clm.mu.Unlock()

	if _, exists := clm.components[componentID]; exists {
		return fmt.Errorf("component %s already registered", componentID)
	}

	clm.components[componentID] = &ComponentInfo{
		Component:    component,
		ID:           componentID,
		State:        ComponentStatePending,
		Priority:     priority,
		Dependencies: make([]string, 0),
	}

	clm.dependencies[componentID] = make([]string, 0)
	
	// Update start order
	clm.updateStartOrder()

	clm.logger.Info("Component registered",
		zap.String("component_id", componentID),
		zap.Int("priority", priority),
	)

	return nil
}

// SetDependency sets a dependency relationship between components
func (clm *ComponentLifecycleManager) SetDependency(componentID, dependencyID string) error {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	if _, exists := clm.components[componentID]; !exists {
		return fmt.Errorf("component %s not found", componentID)
	}

	if _, exists := clm.components[dependencyID]; !exists {
		return fmt.Errorf("dependency component %s not found", dependencyID)
	}

	// Check for circular dependencies
	if clm.hasCircularDependency(componentID, dependencyID) {
		return fmt.Errorf("circular dependency detected: %s -> %s", componentID, dependencyID)
	}

	// Add dependency
	clm.dependencies[componentID] = append(clm.dependencies[componentID], dependencyID)
	clm.components[componentID].Dependencies = append(clm.components[componentID].Dependencies, dependencyID)

	// Update start order
	clm.updateStartOrder()

	clm.logger.Info("Dependency set",
		zap.String("component_id", componentID),
		zap.String("dependency_id", dependencyID),
	)

	return nil
}

// StartComponents starts all registered components in dependency order
func (clm *ComponentLifecycleManager) StartComponents(ctx context.Context) error {
	clm.logger.Info("Starting all components")

	// Get ordered component list
	clm.mu.RLock()
	orderedComponents := make([]string, len(clm.startOrder))
	copy(orderedComponents, clm.startOrder)
	clm.mu.RUnlock()

	if clm.config.ParallelStart {
		return clm.startComponentsParallel(ctx, orderedComponents)
	}
	return clm.startComponentsSequential(ctx, orderedComponents)
}

// ClearComponents removes all registered components (used for cleanup)
func (clm *ComponentLifecycleManager) ClearComponents() {
	clm.mu.Lock()
	defer clm.mu.Unlock()
	
	clm.logger.Info("Clearing all registered components")
	clm.components = make(map[string]*ComponentInfo)
}

// StopComponents stops all components in reverse dependency order
func (clm *ComponentLifecycleManager) StopComponents() error {
	clm.logger.Info("Stopping all components")

	clm.mu.RLock()
	orderedComponents := make([]string, len(clm.startOrder))
	copy(orderedComponents, clm.startOrder)
	clm.mu.RUnlock()

	// Reverse the order for stopping
	for i := len(orderedComponents)/2 - 1; i >= 0; i-- {
		opp := len(orderedComponents) - 1 - i
		orderedComponents[i], orderedComponents[opp] = orderedComponents[opp], orderedComponents[i]
	}

	var lastErr error
	for _, componentID := range orderedComponents {
		if err := clm.stopComponent(componentID); err != nil {
			clm.logger.Error("Failed to stop component",
				zap.String("component_id", componentID),
				zap.Error(err),
			)
			lastErr = err
		}
	}

	return lastErr
}

// GetComponentStatus returns the status of a specific component
func (clm *ComponentLifecycleManager) GetComponentStatus(componentID string) (*ComponentInfo, error) {
	clm.mu.RLock()
	defer clm.mu.RUnlock()

	info, exists := clm.components[componentID]
	if !exists {
		return nil, fmt.Errorf("component %s not found", componentID)
	}

	// Return a copy to avoid concurrent access issues
	infoCopy := *info
	infoCopy.Dependencies = make([]string, len(info.Dependencies))
	copy(infoCopy.Dependencies, info.Dependencies)

	return &infoCopy, nil
}

// GetAllComponentsStatus returns the status of all components
func (clm *ComponentLifecycleManager) GetAllComponentsStatus() map[string]*ComponentInfo {
	clm.mu.RLock()
	defer clm.mu.RUnlock()

	result := make(map[string]*ComponentInfo)
	for id, info := range clm.components {
		infoCopy := *info
		infoCopy.Dependencies = make([]string, len(info.Dependencies))
		copy(infoCopy.Dependencies, info.Dependencies)
		result[id] = &infoCopy
	}

	return result
}

// RegisterStateCallback registers a callback for component state changes
func (clm *ComponentLifecycleManager) RegisterStateCallback(componentID string, callback StateCallback) error {
	clm.mu.Lock()
	defer clm.mu.Unlock()

	if _, exists := clm.components[componentID]; !exists {
		return fmt.Errorf("component %s not found", componentID)
	}

	clm.stateCallbacks[componentID] = append(clm.stateCallbacks[componentID], callback)
	return nil
}

// Internal methods

// updateStartOrder updates the component start order based on priorities and dependencies
func (clm *ComponentLifecycleManager) updateStartOrder() {
	type priorityComponent struct {
		id       string
		priority int
		deps     []string
	}

	components := make([]priorityComponent, 0, len(clm.components))
	for id, info := range clm.components {
		components = append(components, priorityComponent{
			id:       id,
			priority: info.Priority,
			deps:     clm.dependencies[id],
		})
	}

	// Sort by priority first, then by dependency depth
	sort.Slice(components, func(i, j int) bool {
		if components[i].priority != components[j].priority {
			return components[i].priority < components[j].priority
		}
		return len(components[i].deps) < len(components[j].deps)
	})

	clm.startOrder = make([]string, len(components))
	for i, comp := range components {
		clm.startOrder[i] = comp.id
	}
}

// hasCircularDependency checks for circular dependencies
func (clm *ComponentLifecycleManager) hasCircularDependency(componentID, newDependencyID string) bool {
	visited := make(map[string]bool)
	
	var checkDeps func(id string) bool
	checkDeps = func(id string) bool {
		if id == componentID {
			return true // Found cycle
		}
		if visited[id] {
			return false
		}
		visited[id] = true
		
		for _, dep := range clm.dependencies[id] {
			if checkDeps(dep) {
				return true
			}
		}
		return false
	}
	
	return checkDeps(newDependencyID)
}

// startComponentsSequential starts components one by one in order
func (clm *ComponentLifecycleManager) startComponentsSequential(ctx context.Context, orderedComponents []string) error {
	for _, componentID := range orderedComponents {
		if err := clm.startSingleComponent(ctx, componentID); err != nil {
			return fmt.Errorf("failed to start component %s: %w", componentID, err)
		}
	}
	return nil
}

// startComponentsParallel starts components in parallel, respecting dependencies
func (clm *ComponentLifecycleManager) startComponentsParallel(ctx context.Context, orderedComponents []string) error {
	// Group components by dependency level
	levels := clm.groupComponentsByLevel(orderedComponents)
	
	// Start each level in parallel
	for level := 0; level < len(levels); level++ {
		clm.logger.Info("Starting component level",
			zap.Int("level", level),
			zap.Strings("components", levels[level]),
		)
		
		if err := clm.startComponentLevel(ctx, levels[level]); err != nil {
			return fmt.Errorf("failed to start component level %d: %w", level, err)
		}
	}
	
	return nil
}

// groupComponentsByLevel groups components by their dependency level
func (clm *ComponentLifecycleManager) groupComponentsByLevel(orderedComponents []string) [][]string {
	levels := make([][]string, 0)
	remaining := make(map[string]bool)
	
	// Initialize remaining components
	for _, id := range orderedComponents {
		remaining[id] = true
	}
	
	for len(remaining) > 0 {
		currentLevel := make([]string, 0)
		
		// Find components with no unstarted dependencies
		for id := range remaining {
			canStart := true
			for _, dep := range clm.dependencies[id] {
				if remaining[dep] {
					canStart = false
					break
				}
			}
			if canStart {
				currentLevel = append(currentLevel, id)
			}
		}
		
		if len(currentLevel) == 0 {
			// Deadlock - shouldn't happen with proper dependency checking
			clm.logger.Error("Dependency deadlock detected", zap.Strings("remaining", keys(remaining)))
			break
		}
		
		// Remove current level from remaining
		for _, id := range currentLevel {
			delete(remaining, id)
		}
		
		levels = append(levels, currentLevel)
	}
	
	return levels
}

// startComponentLevel starts all components in a level in parallel
func (clm *ComponentLifecycleManager) startComponentLevel(ctx context.Context, levelComponents []string) error {
	errChan := make(chan error, len(levelComponents))
	
	for _, componentID := range levelComponents {
		go func(id string) {
			err := clm.startSingleComponent(ctx, id)
			errChan <- err
		}(componentID)
	}
	
	// Wait for all components in this level to complete
	var firstError error
	for i := 0; i < len(levelComponents); i++ {
		if err := <-errChan; err != nil && firstError == nil {
			firstError = err
		}
	}
	
	return firstError
}

// startSingleComponent starts a single component with retry logic
func (clm *ComponentLifecycleManager) startSingleComponent(ctx context.Context, componentID string) error {
	clm.mu.Lock()
	info, exists := clm.components[componentID]
	if !exists {
		clm.mu.Unlock()
		return fmt.Errorf("component %s not found", componentID)
	}
	component := info.Component
	clm.mu.Unlock()

	// Wait for dependencies to be ready
	if err := clm.waitForDependencies(ctx, componentID); err != nil {
		return fmt.Errorf("dependency wait failed for %s: %w", componentID, err)
	}

	clm.logger.Info("Starting component", zap.String("component_id", componentID))
	clm.updateComponentState(componentID, ComponentStateStarting)

	// Retry logic
	for attempt := 0; attempt < clm.config.MaxRetries; attempt++ {
		startCtx, cancel := context.WithTimeout(ctx, clm.config.StartTimeout)
		
		err := component.Start(startCtx)
		cancel()

		if err == nil {
			now := time.Now()
			clm.mu.Lock()
			info.StartedAt = &now
			info.ErrorCount = 0
			info.LastError = ""
			clm.mu.Unlock()
			
			clm.updateComponentState(componentID, ComponentStateRunning)
			clm.logger.Info("Component started successfully", zap.String("component_id", componentID))
			return nil
		}

		clm.logger.Warn("Component start attempt failed",
			zap.String("component_id", componentID),
			zap.Int("attempt", attempt+1),
			zap.Int("max_attempts", clm.config.MaxRetries),
			zap.Error(err),
		)

		clm.mu.Lock()
		info.ErrorCount++
		info.LastError = err.Error()
		clm.mu.Unlock()

		if attempt < clm.config.MaxRetries-1 {
			select {
			case <-ctx.Done():
				clm.updateComponentState(componentID, ComponentStateError)
				return ctx.Err()
			case <-time.After(clm.config.RetryInterval):
				continue
			}
		}
	}

	clm.updateComponentState(componentID, ComponentStateError)
	return fmt.Errorf("failed to start component %s after %d attempts", componentID, clm.config.MaxRetries)
}

// waitForDependencies waits for all dependencies to be running
func (clm *ComponentLifecycleManager) waitForDependencies(ctx context.Context, componentID string) error {
	clm.mu.RLock()
	dependencies := make([]string, len(clm.dependencies[componentID]))
	copy(dependencies, clm.dependencies[componentID])
	clm.mu.RUnlock()

	if len(dependencies) == 0 {
		return nil
	}

	waitCtx, cancel := context.WithTimeout(ctx, clm.config.DependencyWait)
	defer cancel()

	clm.logger.Info("Waiting for dependencies",
		zap.String("component_id", componentID),
		zap.Strings("dependencies", dependencies),
	)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("dependency wait timeout for component %s", componentID)
		case <-ticker.C:
			allReady := true
			for _, depID := range dependencies {
				clm.mu.RLock()
				depInfo, exists := clm.components[depID]
				clm.mu.RUnlock()
				
				if !exists || depInfo.State != ComponentStateRunning {
					allReady = false
					break
				}
			}
			
			if allReady {
				clm.logger.Info("All dependencies ready",
					zap.String("component_id", componentID),
				)
				return nil
			}
		}
	}
}

// stopComponent stops a single component
func (clm *ComponentLifecycleManager) stopComponent(componentID string) error {
	clm.mu.Lock()
	info, exists := clm.components[componentID]
	if !exists {
		clm.mu.Unlock()
		return fmt.Errorf("component %s not found", componentID)
	}
	component := info.Component
	clm.mu.Unlock()

	if info.State != ComponentStateRunning {
		return nil // Already stopped or not running
	}

	clm.logger.Info("Stopping component", zap.String("component_id", componentID))
	clm.updateComponentState(componentID, ComponentStateStopping)

	err := component.Stop()
	if err != nil {
		clm.updateComponentState(componentID, ComponentStateError)
		return err
	}

	clm.updateComponentState(componentID, ComponentStateStopped)
	clm.logger.Info("Component stopped", zap.String("component_id", componentID))
	return nil
}

// updateComponentState updates the state of a component and triggers callbacks
func (clm *ComponentLifecycleManager) updateComponentState(componentID string, newState ComponentState) {
	clm.mu.Lock()
	info, exists := clm.components[componentID]
	if !exists {
		clm.mu.Unlock()
		return
	}
	
	oldState := info.State
	info.State = newState
	callbacks := make([]StateCallback, len(clm.stateCallbacks[componentID]))
	copy(callbacks, clm.stateCallbacks[componentID])
	clm.mu.Unlock()

	// Trigger callbacks
	for _, callback := range callbacks {
		go func(cb StateCallback) {
			defer func() {
				if r := recover(); r != nil {
					clm.logger.Error("State callback panicked",
						zap.String("component_id", componentID),
						zap.Any("panic", r),
					)
				}
			}()
			cb(componentID, oldState, newState)
		}(callback)
	}
}

// Helper function to get keys from a map
func keys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}