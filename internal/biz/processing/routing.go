package processing

import (
	"context"
	"fmt"

	"pin_intent_broadcast_network/internal/biz/common"
)

// RoutingEngine implements intent routing logic
// This file will contain the implementation for task 5.4
type RoutingEngine struct {
	strategies map[string]RoutingStrategy
	config     *RoutingConfig
}

// RoutingConfig holds configuration for the routing engine
type RoutingConfig struct {
	DefaultStrategy     string `yaml:"default_strategy"`
	EnableLoadBalancing bool   `yaml:"enable_load_balancing"`
	MaxRetries          int    `yaml:"max_retries"`
	FallbackEnabled     bool   `yaml:"fallback_enabled"`
}

// RoutingStrategy defines how intents should be routed
type RoutingStrategy interface {
	Route(ctx context.Context, intent *common.Intent) ([]string, error)
	GetName() string
	GetPriority() int
}

// NewRoutingEngine creates a new routing engine
func NewRoutingEngine(config *RoutingConfig) *RoutingEngine {
	return &RoutingEngine{
		strategies: make(map[string]RoutingStrategy),
		config:     config,
	}
}

// RouteIntent routes an intent based on its type and content
func (re *RoutingEngine) RouteIntent(ctx context.Context, intent *common.Intent) ([]string, error) {
	// TODO: Implement in task 5.4

	// Determine routing strategy
	strategy, err := re.getRoutingStrategy(intent.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get routing strategy: %w", err)
	}

	// Execute routing
	routes, err := strategy.Route(ctx, intent)
	if err != nil {
		// Try fallback strategy if enabled
		if re.config.FallbackEnabled {
			return re.routeWithFallback(ctx, intent)
		}
		return nil, fmt.Errorf("routing failed: %w", err)
	}

	return routes, nil
}

// AddStrategy adds a routing strategy
func (re *RoutingEngine) AddStrategy(strategy RoutingStrategy) {
	re.strategies[strategy.GetName()] = strategy
}

// RemoveStrategy removes a routing strategy
func (re *RoutingEngine) RemoveStrategy(name string) {
	delete(re.strategies, name)
}

// getRoutingStrategy gets the appropriate routing strategy for an intent type
func (re *RoutingEngine) getRoutingStrategy(intentType string) (RoutingStrategy, error) {
	// Try to find type-specific strategy
	if strategy, exists := re.strategies[intentType]; exists {
		return strategy, nil
	}

	// Use default strategy
	if strategy, exists := re.strategies[re.config.DefaultStrategy]; exists {
		return strategy, nil
	}

	return nil, fmt.Errorf("no routing strategy found for intent type: %s", intentType)
}

// routeWithFallback attempts routing with fallback strategies
func (re *RoutingEngine) routeWithFallback(ctx context.Context, intent *common.Intent) ([]string, error) {
	// TODO: Implement fallback routing logic
	return nil, fmt.Errorf("fallback routing not implemented")
}

// TypeBasedStrategy implements routing based on intent type
type TypeBasedStrategy struct {
	name     string
	routes   map[string][]string
	priority int
}

// NewTypeBasedStrategy creates a new type-based routing strategy
func NewTypeBasedStrategy() *TypeBasedStrategy {
	return &TypeBasedStrategy{
		name:     "type_based",
		routes:   make(map[string][]string),
		priority: 100,
	}
}

// Route routes an intent based on its type
func (tbs *TypeBasedStrategy) Route(ctx context.Context, intent *common.Intent) ([]string, error) {
	routes, exists := tbs.routes[intent.Type]
	if !exists {
		return nil, fmt.Errorf("no routes defined for intent type: %s", intent.Type)
	}

	return routes, nil
}

// GetName returns the strategy name
func (tbs *TypeBasedStrategy) GetName() string {
	return tbs.name
}

// GetPriority returns the strategy priority
func (tbs *TypeBasedStrategy) GetPriority() int {
	return tbs.priority
}

// AddRoute adds a route for an intent type
func (tbs *TypeBasedStrategy) AddRoute(intentType string, destinations []string) {
	tbs.routes[intentType] = destinations
}

// RemoveRoute removes routes for an intent type
func (tbs *TypeBasedStrategy) RemoveRoute(intentType string) {
	delete(tbs.routes, intentType)
}

// LoadBalancingStrategy implements load-balanced routing
type LoadBalancingStrategy struct {
	name     string
	routes   map[string][]string
	counters map[string]int
	priority int
}

// NewLoadBalancingStrategy creates a new load-balancing routing strategy
func NewLoadBalancingStrategy() *LoadBalancingStrategy {
	return &LoadBalancingStrategy{
		name:     "load_balancing",
		routes:   make(map[string][]string),
		counters: make(map[string]int),
		priority: 90,
	}
}

// Route routes an intent using round-robin load balancing
func (lbs *LoadBalancingStrategy) Route(ctx context.Context, intent *common.Intent) ([]string, error) {
	routes, exists := lbs.routes[intent.Type]
	if !exists || len(routes) == 0 {
		return nil, fmt.Errorf("no routes defined for intent type: %s", intent.Type)
	}

	// Round-robin selection
	counter := lbs.counters[intent.Type]
	selectedRoute := routes[counter%len(routes)]
	lbs.counters[intent.Type] = counter + 1

	return []string{selectedRoute}, nil
}

// GetName returns the strategy name
func (lbs *LoadBalancingStrategy) GetName() string {
	return lbs.name
}

// GetPriority returns the strategy priority
func (lbs *LoadBalancingStrategy) GetPriority() int {
	return lbs.priority
}
