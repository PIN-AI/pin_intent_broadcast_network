package matching

import (
	"context"

	"pin_intent_broadcast_network/internal/biz/common"
)

// Matcher implements the IntentMatcher interface
// This file will contain the implementation for task 6.2
type Matcher struct {
	engine *Engine
	config *MatcherConfig
}

// MatcherConfig holds configuration for the matcher
type MatcherConfig struct {
	EnableContentMatching  bool    `yaml:"enable_content_matching"`
	EnableMetadataMatching bool    `yaml:"enable_metadata_matching"`
	ContentWeight          float64 `yaml:"content_weight"`
	MetadataWeight         float64 `yaml:"metadata_weight"`
	TypeWeight             float64 `yaml:"type_weight"`
}

// NewMatcher creates a new intent matcher
func NewMatcher(engine *Engine, config *MatcherConfig) *Matcher {
	return &Matcher{
		engine: engine,
		config: config,
	}
}

// FindMatches finds matching intents using the matching engine
func (m *Matcher) FindMatches(ctx context.Context, intent *common.Intent, candidates []*common.Intent) ([]*common.MatchResult, error) {
	// TODO: Implement in task 6.2
	return m.engine.FindMatches(ctx, intent, candidates)
}

// AddMatchingRule adds a matching rule
func (m *Matcher) AddMatchingRule(rule common.MatchingRule) error {
	// TODO: Implement in task 6.2
	return m.engine.AddMatchingRule(rule)
}

// RemoveMatchingRule removes a matching rule
func (m *Matcher) RemoveMatchingRule(ruleName string) error {
	// TODO: Implement in task 6.2
	return m.engine.RemoveMatchingRule(ruleName)
}

// GetMatchingRules returns all matching rules
func (m *Matcher) GetMatchingRules() []common.MatchingRule {
	// TODO: Implement in task 6.2
	return m.engine.GetMatchingRules()
}

// MatchByContent matches intents based on content similarity
func (m *Matcher) MatchByContent(intent1, intent2 *common.Intent) (float64, error) {
	// TODO: Implement content-based matching algorithm
	if !m.config.EnableContentMatching {
		return 0.0, nil
	}

	// Implement content similarity algorithm
	// This could use techniques like:
	// - Cosine similarity
	// - Jaccard similarity
	// - Edit distance
	// - Semantic similarity using embeddings

	return 0.0, nil
}

// MatchByMetadata matches intents based on metadata similarity
func (m *Matcher) MatchByMetadata(intent1, intent2 *common.Intent) (float64, error) {
	// TODO: Implement metadata-based matching algorithm
	if !m.config.EnableMetadataMatching {
		return 0.0, nil
	}

	// Compare metadata fields
	commonFields := 0
	totalFields := 0

	// Count common metadata fields
	for key, value1 := range intent1.Metadata {
		totalFields++
		if value2, exists := intent2.Metadata[key]; exists && value1 == value2 {
			commonFields++
		}
	}

	for key := range intent2.Metadata {
		if _, exists := intent1.Metadata[key]; !exists {
			totalFields++
		}
	}

	if totalFields == 0 {
		return 0.0, nil
	}

	return float64(commonFields) / float64(totalFields), nil
}

// MatchByType matches intents based on type compatibility
func (m *Matcher) MatchByType(intent1, intent2 *common.Intent) float64 {
	// TODO: Implement type-based matching
	if intent1.Type == intent2.Type {
		return 1.0
	}

	// Check for compatible types
	// This could be extended to support type hierarchies
	compatibleTypes := map[string][]string{
		"trade":    {"swap", "exchange"},
		"transfer": {"send", "payment"},
		"lending":  {"borrow", "loan"},
	}

	if compatible, exists := compatibleTypes[intent1.Type]; exists {
		for _, compatType := range compatible {
			if intent2.Type == compatType {
				return 0.8 // High compatibility but not exact match
			}
		}
	}

	return 0.0
}

// CalculateOverallMatch calculates the overall match score
func (m *Matcher) CalculateOverallMatch(intent1, intent2 *common.Intent) (common.MatchResult, error) {
	// TODO: Implement overall matching calculation

	// Calculate individual match scores
	contentScore, err := m.MatchByContent(intent1, intent2)
	if err != nil {
		return common.MatchResult{}, err
	}

	metadataScore, err := m.MatchByMetadata(intent1, intent2)
	if err != nil {
		return common.MatchResult{}, err
	}

	typeScore := m.MatchByType(intent1, intent2)

	// Calculate weighted average
	totalWeight := m.config.ContentWeight + m.config.MetadataWeight + m.config.TypeWeight
	if totalWeight == 0 {
		totalWeight = 1.0
	}

	overallScore := (contentScore*m.config.ContentWeight +
		metadataScore*m.config.MetadataWeight +
		typeScore*m.config.TypeWeight) / totalWeight

	// Determine match type
	matchType := common.MatchTypePartial
	if overallScore >= 0.9 {
		matchType = common.MatchTypeExact
	} else if overallScore >= 0.7 {
		matchType = common.MatchTypeSemantic
	} else if overallScore >= 0.5 {
		matchType = common.MatchTypePattern
	}

	return common.MatchResult{
		IsMatch:    overallScore > 0.5, // Threshold for considering it a match
		Confidence: overallScore,
		MatchType:  matchType,
		Details: map[string]interface{}{
			"content_score":  contentScore,
			"metadata_score": metadataScore,
			"type_score":     typeScore,
		},
	}, nil
}
