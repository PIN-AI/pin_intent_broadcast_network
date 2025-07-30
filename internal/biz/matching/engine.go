package matching

import (
	"context"
	"sort"
	"sync"

	"pin_intent_broadcast_network/internal/biz/common"
)

// Engine implements the intent matching engine
// This file will contain the implementation for task 6.1
type Engine struct {
	rules    []common.MatchingRule
	patterns map[string]*MatchingPattern
	cache    *MatchingCache
	config   *EngineConfig
	mu       sync.RWMutex
}

// EngineConfig holds configuration for the matching engine
type EngineConfig struct {
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	MaxMatchesPerIntent int     `yaml:"max_matches_per_intent"`
	MatchingTimeout     int     `yaml:"matching_timeout"`
	EnableCaching       bool    `yaml:"enable_caching"`
	CacheSize           int     `yaml:"cache_size"`
}

// MatchingPattern represents a matching pattern
type MatchingPattern struct {
	Name    string
	Pattern string
	Weight  float64
	Enabled bool
}

// MatchingCache provides caching for matching results
type MatchingCache struct {
	cache map[string][]*common.MatchResult
	mu    sync.RWMutex
	size  int
}

// NewEngine creates a new matching engine
func NewEngine(config *EngineConfig) *Engine {
	return &Engine{
		rules:    make([]common.MatchingRule, 0),
		patterns: make(map[string]*MatchingPattern),
		cache:    NewMatchingCache(config.CacheSize),
		config:   config,
	}
}

// NewMatchingCache creates a new matching cache
func NewMatchingCache(size int) *MatchingCache {
	return &MatchingCache{
		cache: make(map[string][]*common.MatchResult),
		size:  size,
	}
}

// FindMatches finds matching intents for a given intent
func (e *Engine) FindMatches(ctx context.Context, intent *common.Intent, candidates []*common.Intent) ([]*common.MatchResult, error) {
	// TODO: Implement in task 6.1

	// Check cache first
	if e.config.EnableCaching {
		if cached := e.cache.Get(intent.ID); cached != nil {
			return cached, nil
		}
	}

	var results []*common.MatchResult

	for _, candidate := range candidates {
		for _, rule := range e.rules {
			result, err := rule.Match(intent, candidate)
			if err != nil {
				continue // Skip failed matches
			}

			if result.IsMatch && result.Confidence >= e.config.ConfidenceThreshold {
				results = append(results, &result)
			}
		}
	}

	// Sort by confidence (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})

	// Limit results
	if len(results) > e.config.MaxMatchesPerIntent {
		results = results[:e.config.MaxMatchesPerIntent]
	}

	// Cache results
	if e.config.EnableCaching {
		e.cache.Set(intent.ID, results)
	}

	return results, nil
}

// AddMatchingRule adds a matching rule to the engine
func (e *Engine) AddMatchingRule(rule common.MatchingRule) error {
	// TODO: Implement in task 6.1
	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = append(e.rules, rule)

	// Sort rules by priority
	sort.Slice(e.rules, func(i, j int) bool {
		return e.rules[i].GetPriority() > e.rules[j].GetPriority()
	})

	return nil
}

// RemoveMatchingRule removes a matching rule from the engine
func (e *Engine) RemoveMatchingRule(ruleName string) error {
	// TODO: Implement in task 6.1
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, rule := range e.rules {
		if rule.GetRuleName() == ruleName {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			return nil
		}
	}

	return nil
}

// GetMatchingRules returns all matching rules
func (e *Engine) GetMatchingRules() []common.MatchingRule {
	// TODO: Implement in task 6.1
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy to prevent external modification
	rules := make([]common.MatchingRule, len(e.rules))
	copy(rules, e.rules)
	return rules
}

// AddPattern adds a matching pattern
func (e *Engine) AddPattern(name string, pattern *MatchingPattern) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.patterns[name] = pattern
}

// RemovePattern removes a matching pattern
func (e *Engine) RemovePattern(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.patterns, name)
}

// GetPattern retrieves a matching pattern
func (e *Engine) GetPattern(name string) *MatchingPattern {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.patterns[name]
}

// Get retrieves cached matching results
func (mc *MatchingCache) Get(intentID string) []*common.MatchResult {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.cache[intentID]
}

// Set stores matching results in cache
func (mc *MatchingCache) Set(intentID string, results []*common.MatchResult) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Simple cache eviction: remove oldest if at capacity
	if len(mc.cache) >= mc.size {
		// Remove one random entry (in production, use LRU)
		for k := range mc.cache {
			delete(mc.cache, k)
			break
		}
	}

	mc.cache[intentID] = results
}

// Clear clears the cache
func (mc *MatchingCache) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.cache = make(map[string][]*common.MatchResult)
}
