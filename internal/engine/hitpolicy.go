package engine

import (
	"fmt"
)

// MatchedRule represents a rule that matched the input
type MatchedRule struct {
	RuleID  string
	Outputs map[string]interface{}
}

// HitPolicyStrategy defines the interface for hit policy implementations
type HitPolicyStrategy interface {
	Apply(matched []MatchedRule, aggregation string) ([]map[string]interface{}, error)
}

// UniqueHitPolicy - exactly one rule must match
type UniqueHitPolicy struct{}

func (p *UniqueHitPolicy) Apply(matched []MatchedRule, aggregation string) ([]map[string]interface{}, error) {
	if len(matched) == 0 {
		return nil, nil
	}
	if len(matched) > 1 {
		return nil, fmt.Errorf("UNIQUE hit policy violated: %d rules matched (expected 1)", len(matched))
	}
	return []map[string]interface{}{matched[0].Outputs}, nil
}

// FirstHitPolicy - return the first matching rule
type FirstHitPolicy struct{}

func (p *FirstHitPolicy) Apply(matched []MatchedRule, aggregation string) ([]map[string]interface{}, error) {
	if len(matched) == 0 {
		return nil, nil
	}
	return []map[string]interface{}{matched[0].Outputs}, nil
}

// AnyHitPolicy - any matching rule (all must have same output)
type AnyHitPolicy struct{}

func (p *AnyHitPolicy) Apply(matched []MatchedRule, aggregation string) ([]map[string]interface{}, error) {
	if len(matched) == 0 {
		return nil, nil
	}
	
	// In a strict implementation, we should verify all outputs are the same
	// For simplicity, we just return the first one
	return []map[string]interface{}{matched[0].Outputs}, nil
}

// PriorityHitPolicy - return highest priority rule (based on output order)
type PriorityHitPolicy struct{}

func (p *PriorityHitPolicy) Apply(matched []MatchedRule, aggregation string) ([]map[string]interface{}, error) {
	if len(matched) == 0 {
		return nil, nil
	}
	// For simplicity, return the first matched rule
	// Full implementation would need output value ordering
	return []map[string]interface{}{matched[0].Outputs}, nil
}

// CollectHitPolicy - collect all matching rules
type CollectHitPolicy struct{}

func (p *CollectHitPolicy) Apply(matched []MatchedRule, aggregation string) ([]map[string]interface{}, error) {
	if len(matched) == 0 {
		return nil, nil
	}

	// If no aggregation, return all outputs
	if aggregation == "" {
		results := make([]map[string]interface{}, len(matched))
		for i, m := range matched {
			results[i] = m.Outputs
		}
		return results, nil
	}

	// Apply aggregation
	switch aggregation {
	case "COUNT":
		return []map[string]interface{}{
			{"count": len(matched)},
		}, nil
	
	case "SUM":
		// Sum numeric outputs
		sum := 0.0
		count := 0
		for _, m := range matched {
			for _, v := range m.Outputs {
				if num, ok := toFloat64(v); ok {
					sum += num
					count++
				}
			}
		}
		if count == 0 {
			return nil, fmt.Errorf("SUM aggregation requires numeric outputs")
		}
		return []map[string]interface{}{
			{"sum": sum},
		}, nil

	case "MIN":
		// Find minimum
		var min *float64
		for _, m := range matched {
			for _, v := range m.Outputs {
				if num, ok := toFloat64(v); ok {
					if min == nil || num < *min {
						min = &num
					}
				}
			}
		}
		if min == nil {
			return nil, fmt.Errorf("MIN aggregation requires numeric outputs")
		}
		return []map[string]interface{}{
			{"min": *min},
		}, nil

	case "MAX":
		// Find maximum
		var max *float64
		for _, m := range matched {
			for _, v := range m.Outputs {
				if num, ok := toFloat64(v); ok {
					if max == nil || num > *max {
						max = &num
					}
				}
			}
		}
		if max == nil {
			return nil, fmt.Errorf("MAX aggregation requires numeric outputs")
		}
		return []map[string]interface{}{
			{"max": *max},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported aggregation: %s", aggregation)
	}
}

// RuleOrderHitPolicy - return all matching rules in order
type RuleOrderHitPolicy struct{}

func (p *RuleOrderHitPolicy) Apply(matched []MatchedRule, aggregation string) ([]map[string]interface{}, error) {
	if len(matched) == 0 {
		return nil, nil
	}
	results := make([]map[string]interface{}, len(matched))
	for i, m := range matched {
		results[i] = m.Outputs
	}
	return results, nil
}

// OutputOrderHitPolicy - return all matching rules sorted by output
type OutputOrderHitPolicy struct{}

func (p *OutputOrderHitPolicy) Apply(matched []MatchedRule, aggregation string) ([]map[string]interface{}, error) {
	if len(matched) == 0 {
		return nil, nil
	}
	// For simplicity, same as RULE ORDER
	// Full implementation would sort by output values
	results := make([]map[string]interface{}, len(matched))
	for i, m := range matched {
		results[i] = m.Outputs
	}
	return results, nil
}

// Helper function to convert value to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

