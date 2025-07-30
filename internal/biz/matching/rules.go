package matching

import (
	"pin_intent_broadcast_network/internal/biz/common"
)

// Rules implements various matching rules
// This file will contain the implementation for task 6.3

// ExactMatchRule implements exact matching between intents
type ExactMatchRule struct {
	name     string
	priority int
}

// NewExactMatchRule creates a new exact match rule
func NewExactMatchRule() *ExactMatchRule {
	return &ExactMatchRule{
		name:     "exact_match",
		priority: 100,
	}
}

// Match performs exact matching between two intents
func (emr *ExactMatchRule) Match(intent1, intent2 *common.Intent) (common.MatchResult, error) {
	// TODO: Implement in task 6.3

	// Check for exact matches
	isExactMatch := intent1.Type == intent2.Type &&
		string(intent1.Payload) == string(intent2.Payload)

	confidence := 0.0
	if isExactMatch {
		confidence = 1.0
	}

	return common.MatchResult{
		IsMatch:    isExactMatch,
		Confidence: confidence,
		MatchType:  common.MatchTypeExact,
		Details: map[string]interface{}{
			"rule":          emr.name,
			"type_match":    intent1.Type == intent2.Type,
			"payload_match": string(intent1.Payload) == string(intent2.Payload),
		},
	}, nil
}

// GetRuleName returns the rule name
func (emr *ExactMatchRule) GetRuleName() string {
	return emr.name
}

// GetPriority returns the rule priority
func (emr *ExactMatchRule) GetPriority() int {
	return emr.priority
}

// TypeMatchRule implements type-based matching
type TypeMatchRule struct {
	name     string
	priority int
	weights  map[string]float64
}

// NewTypeMatchRule creates a new type match rule
func NewTypeMatchRule() *TypeMatchRule {
	return &TypeMatchRule{
		name:     "type_match",
		priority: 90,
		weights: map[string]float64{
			"exact":      1.0,
			"compatible": 0.8,
			"related":    0.6,
		},
	}
}

// Match performs type-based matching between two intents
func (tmr *TypeMatchRule) Match(intent1, intent2 *common.Intent) (common.MatchResult, error) {
	// TODO: Implement in task 6.3

	confidence := 0.0
	matchType := common.MatchTypePartial

	if intent1.Type == intent2.Type {
		confidence = tmr.weights["exact"]
		matchType = common.MatchTypeExact
	} else if tmr.areTypesCompatible(intent1.Type, intent2.Type) {
		confidence = tmr.weights["compatible"]
		matchType = common.MatchTypeSemantic
	} else if tmr.areTypesRelated(intent1.Type, intent2.Type) {
		confidence = tmr.weights["related"]
		matchType = common.MatchTypePattern
	}

	return common.MatchResult{
		IsMatch:    confidence > 0,
		Confidence: confidence,
		MatchType:  matchType,
		Details: map[string]interface{}{
			"rule":         tmr.name,
			"type1":        intent1.Type,
			"type2":        intent2.Type,
			"relationship": tmr.getTypeRelationship(intent1.Type, intent2.Type),
		},
	}, nil
}

// GetRuleName returns the rule name
func (tmr *TypeMatchRule) GetRuleName() string {
	return tmr.name
}

// GetPriority returns the rule priority
func (tmr *TypeMatchRule) GetPriority() int {
	return tmr.priority
}

// areTypesCompatible checks if two intent types are compatible
func (tmr *TypeMatchRule) areTypesCompatible(type1, type2 string) bool {
	compatibleTypes := map[string][]string{
		"trade":    {"swap", "exchange"},
		"transfer": {"send", "payment"},
		"lending":  {"borrow", "loan"},
		"swap":     {"trade", "exchange"},
	}

	if compatible, exists := compatibleTypes[type1]; exists {
		for _, compatType := range compatible {
			if type2 == compatType {
				return true
			}
		}
	}

	return false
}

// areTypesRelated checks if two intent types are related
func (tmr *TypeMatchRule) areTypesRelated(type1, type2 string) bool {
	relatedTypes := map[string][]string{
		"trade":    {"lending", "transfer"},
		"transfer": {"trade", "payment"},
		"lending":  {"trade", "investment"},
	}

	if related, exists := relatedTypes[type1]; exists {
		for _, relatedType := range related {
			if type2 == relatedType {
				return true
			}
		}
	}

	return false
}

// getTypeRelationship returns the relationship between two types
func (tmr *TypeMatchRule) getTypeRelationship(type1, type2 string) string {
	if type1 == type2 {
		return "exact"
	} else if tmr.areTypesCompatible(type1, type2) {
		return "compatible"
	} else if tmr.areTypesRelated(type1, type2) {
		return "related"
	}
	return "unrelated"
}

// MetadataMatchRule implements metadata-based matching
type MetadataMatchRule struct {
	name     string
	priority int
	weights  map[string]float64
}

// NewMetadataMatchRule creates a new metadata match rule
func NewMetadataMatchRule() *MetadataMatchRule {
	return &MetadataMatchRule{
		name:     "metadata_match",
		priority: 80,
		weights: map[string]float64{
			"high":   0.9,
			"medium": 0.7,
			"low":    0.5,
		},
	}
}

// Match performs metadata-based matching between two intents
func (mmr *MetadataMatchRule) Match(intent1, intent2 *common.Intent) (common.MatchResult, error) {
	// TODO: Implement in task 6.3

	if intent1.Metadata == nil || intent2.Metadata == nil {
		return common.MatchResult{
			IsMatch:    false,
			Confidence: 0.0,
			MatchType:  common.MatchTypePartial,
			Details: map[string]interface{}{
				"rule":   mmr.name,
				"reason": "missing metadata",
			},
		}, nil
	}

	// Calculate metadata similarity
	commonFields := 0
	totalFields := len(intent1.Metadata)
	if len(intent2.Metadata) > totalFields {
		totalFields = len(intent2.Metadata)
	}

	for key, value1 := range intent1.Metadata {
		if value2, exists := intent2.Metadata[key]; exists && value1 == value2 {
			commonFields++
		}
	}

	similarity := 0.0
	if totalFields > 0 {
		similarity = float64(commonFields) / float64(totalFields)
	}

	// Determine confidence level
	confidence := 0.0
	matchType := common.MatchTypePartial

	if similarity >= 0.8 {
		confidence = mmr.weights["high"]
		matchType = common.MatchTypeSemantic
	} else if similarity >= 0.6 {
		confidence = mmr.weights["medium"]
		matchType = common.MatchTypePattern
	} else if similarity >= 0.4 {
		confidence = mmr.weights["low"]
		matchType = common.MatchTypePartial
	}

	return common.MatchResult{
		IsMatch:    confidence > 0,
		Confidence: confidence,
		MatchType:  matchType,
		Details: map[string]interface{}{
			"rule":          mmr.name,
			"similarity":    similarity,
			"common_fields": commonFields,
			"total_fields":  totalFields,
		},
	}, nil
}

// GetRuleName returns the rule name
func (mmr *MetadataMatchRule) GetRuleName() string {
	return mmr.name
}

// GetPriority returns the rule priority
func (mmr *MetadataMatchRule) GetPriority() int {
	return mmr.priority
}

// CompositeMatchRule combines multiple matching rules
type CompositeMatchRule struct {
	name     string
	priority int
	rules    []common.MatchingRule
	weights  map[string]float64
}

// NewCompositeMatchRule creates a new composite match rule
func NewCompositeMatchRule(rules []common.MatchingRule) *CompositeMatchRule {
	return &CompositeMatchRule{
		name:     "composite_match",
		priority: 70,
		rules:    rules,
		weights: map[string]float64{
			"exact_match":    0.4,
			"type_match":     0.3,
			"metadata_match": 0.3,
		},
	}
}

// Match performs composite matching using multiple rules
func (cmr *CompositeMatchRule) Match(intent1, intent2 *common.Intent) (common.MatchResult, error) {
	// TODO: Implement in task 6.3

	var totalConfidence float64
	var totalWeight float64
	details := make(map[string]interface{})

	for _, rule := range cmr.rules {
		result, err := rule.Match(intent1, intent2)
		if err != nil {
			continue // Skip failed rules
		}

		weight := cmr.weights[rule.GetRuleName()]
		if weight == 0 {
			weight = 1.0 / float64(len(cmr.rules)) // Equal weight if not specified
		}

		totalConfidence += result.Confidence * weight
		totalWeight += weight

		details[rule.GetRuleName()] = result
	}

	finalConfidence := 0.0
	if totalWeight > 0 {
		finalConfidence = totalConfidence / totalWeight
	}

	// Determine match type based on final confidence
	matchType := common.MatchTypePartial
	if finalConfidence >= 0.9 {
		matchType = common.MatchTypeExact
	} else if finalConfidence >= 0.7 {
		matchType = common.MatchTypeSemantic
	} else if finalConfidence >= 0.5 {
		matchType = common.MatchTypePattern
	}

	return common.MatchResult{
		IsMatch:    finalConfidence > 0.5,
		Confidence: finalConfidence,
		MatchType:  matchType,
		Details:    details,
	}, nil
}

// GetRuleName returns the rule name
func (cmr *CompositeMatchRule) GetRuleName() string {
	return cmr.name
}

// GetPriority returns the rule priority
func (cmr *CompositeMatchRule) GetPriority() int {
	return cmr.priority
}
