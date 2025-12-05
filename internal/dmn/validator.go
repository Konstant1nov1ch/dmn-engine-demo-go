package dmn

import (
	"fmt"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validator validates DMN models
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a DMN definitions model and returns any errors found
func (v *Validator) Validate(defs *Definitions) []ValidationError {
	var errors []ValidationError

	// Check definitions has an ID
	if defs.ID == "" {
		errors = append(errors, ValidationError{
			Field:   "definitions.id",
			Message: "definitions must have an id",
		})
	}

	// Check for at least one decision
	if len(defs.Decisions) == 0 {
		errors = append(errors, ValidationError{
			Field:   "definitions.decisions",
			Message: "definitions must have at least one decision",
		})
	}

	// Check unique IDs
	seenIDs := make(map[string]bool)
	
	for _, d := range defs.Decisions {
		if d.ID == "" {
			errors = append(errors, ValidationError{
				Field:   "decision.id",
				Message: "decision must have an id",
			})
			continue
		}
		
		if seenIDs[d.ID] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("decision[%s].id", d.ID),
				Message: "duplicate decision id",
			})
		}
		seenIDs[d.ID] = true

		// Validate decision
		errors = append(errors, v.validateDecision(&d)...)
	}

	for _, input := range defs.InputData {
		if input.ID == "" {
			errors = append(errors, ValidationError{
				Field:   "inputData.id",
				Message: "inputData must have an id",
			})
			continue
		}
		
		if seenIDs[input.ID] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("inputData[%s].id", input.ID),
				Message: "duplicate id",
			})
		}
		seenIDs[input.ID] = true
	}

	// Check for cyclic dependencies
	if cycleErr := v.checkCyclicDependencies(defs); cycleErr != nil {
		errors = append(errors, *cycleErr)
	}

	return errors
}

// validateDecision validates a single decision
func (v *Validator) validateDecision(d *Decision) []ValidationError {
	var errors []ValidationError
	prefix := fmt.Sprintf("decision[%s]", d.ID)

	// Decision must have either a decision table or literal expression
	if d.DecisionTable == nil && d.LiteralExpression == nil {
		errors = append(errors, ValidationError{
			Field:   prefix,
			Message: "decision must have either a decisionTable or literalExpression",
		})
		return errors
	}

	// Validate decision table if present
	if d.DecisionTable != nil {
		errors = append(errors, v.validateDecisionTable(d.DecisionTable, prefix)...)
	}

	// Validate literal expression if present
	if d.LiteralExpression != nil {
		if d.LiteralExpression.Text == "" {
			errors = append(errors, ValidationError{
				Field:   prefix + ".literalExpression.text",
				Message: "literal expression must have text",
			})
		}
	}

	return errors
}

// validateDecisionTable validates a decision table
func (v *Validator) validateDecisionTable(dt *DecisionTable, prefix string) []ValidationError {
	var errors []ValidationError
	prefix = prefix + ".decisionTable"

	// Validate hit policy
	if !isValidHitPolicy(dt.HitPolicy) {
		errors = append(errors, ValidationError{
			Field:   prefix + ".hitPolicy",
			Message: fmt.Sprintf("invalid hit policy: %s", dt.HitPolicy),
		})
	}

	// Must have at least one output
	if len(dt.Outputs) == 0 {
		errors = append(errors, ValidationError{
			Field:   prefix + ".outputs",
			Message: "decision table must have at least one output",
		})
	}

	// Validate rules
	for i, rule := range dt.Rules {
		rulePrefix := fmt.Sprintf("%s.rules[%d]", prefix, i)
		
		// Rule must have same number of input entries as inputs
		if len(rule.InputEntries) != len(dt.Inputs) {
			errors = append(errors, ValidationError{
				Field:   rulePrefix + ".inputEntries",
				Message: fmt.Sprintf("expected %d input entries, got %d", len(dt.Inputs), len(rule.InputEntries)),
			})
		}
		
		// Rule must have same number of output entries as outputs
		if len(rule.OutputEntries) != len(dt.Outputs) {
			errors = append(errors, ValidationError{
				Field:   rulePrefix + ".outputEntries",
				Message: fmt.Sprintf("expected %d output entries, got %d", len(dt.Outputs), len(rule.OutputEntries)),
			})
		}
	}

	// Validate aggregation for COLLECT policy
	if dt.HitPolicy == HitPolicyCollect && dt.Aggregation != "" {
		if !isValidAggregation(dt.Aggregation) {
			errors = append(errors, ValidationError{
				Field:   prefix + ".aggregation",
				Message: fmt.Sprintf("invalid aggregation: %s", dt.Aggregation),
			})
		}
	}

	return errors
}

// checkCyclicDependencies checks for cyclic dependencies in the DRG
func (v *Validator) checkCyclicDependencies(defs *Definitions) *ValidationError {
	// Build dependency graph
	graph := make(map[string][]string)
	for _, d := range defs.Decisions {
		deps := make([]string, 0)
		for _, req := range d.InformationRequirements {
			if req.RequiredDecision != nil {
				// Remove the # prefix from href
				depID := strings.TrimPrefix(req.RequiredDecision.Href, "#")
				deps = append(deps, depID)
			}
		}
		graph[d.ID] = deps
	}

	// Check for cycles using DFS
	visited := make(map[string]int) // 0: unvisited, 1: visiting, 2: visited
	
	var hasCycle func(node string) bool
	hasCycle = func(node string) bool {
		if visited[node] == 1 {
			return true // cycle detected
		}
		if visited[node] == 2 {
			return false // already processed
		}
		
		visited[node] = 1
		for _, dep := range graph[node] {
			if hasCycle(dep) {
				return true
			}
		}
		visited[node] = 2
		return false
	}

	for id := range graph {
		if hasCycle(id) {
			return &ValidationError{
				Field:   "decisions",
				Message: fmt.Sprintf("cyclic dependency detected involving decision: %s", id),
			}
		}
	}

	return nil
}

func isValidHitPolicy(hp string) bool {
	switch hp {
	case "", HitPolicyUnique, HitPolicyFirst, HitPolicyPriority, 
		HitPolicyAny, HitPolicyCollect, HitPolicyRuleOrder, HitPolicyOutputOrder:
		return true
	default:
		return false
	}
}

func isValidAggregation(agg string) bool {
	switch agg {
	case "", AggregationSum, AggregationCount, AggregationMin, AggregationMax:
		return true
	default:
		return false
	}
}

