package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/konstantin/dmn-engine-go/internal/dmn"
	"github.com/konstantin/dmn-engine-go/internal/storage"
)

// Engine is the main DMN evaluation engine
type Engine struct {
	repo        storage.DefinitionRepository
	hitPolicies map[string]HitPolicyStrategy
}

// NewEngine creates a new evaluation engine
func NewEngine(repo storage.DefinitionRepository) *Engine {
	e := &Engine{
		repo:        repo,
		hitPolicies: make(map[string]HitPolicyStrategy),
	}

	// Register hit policies
	e.hitPolicies[dmn.HitPolicyUnique] = &UniqueHitPolicy{}
	e.hitPolicies[dmn.HitPolicyFirst] = &FirstHitPolicy{}
	e.hitPolicies[dmn.HitPolicyAny] = &AnyHitPolicy{}
	e.hitPolicies[dmn.HitPolicyCollect] = &CollectHitPolicy{}
	e.hitPolicies[dmn.HitPolicyPriority] = &PriorityHitPolicy{}
	e.hitPolicies[dmn.HitPolicyRuleOrder] = &RuleOrderHitPolicy{}
	e.hitPolicies[dmn.HitPolicyOutputOrder] = &OutputOrderHitPolicy{}

	return e
}

// EvaluateRequest is a request to evaluate a decision
type EvaluateRequest struct {
	DecisionKey string                 `json:"decisionKey"`
	Version     *int                   `json:"version,omitempty"`
	Variables   map[string]interface{} `json:"variables"`
	TenantID    string                 `json:"tenantId,omitempty"`
}

// EvaluateResult is the result of a decision evaluation
type EvaluateResult struct {
	DecisionKey  string                   `json:"decisionKey"`
	DecisionName string                   `json:"decisionName"`
	Version      int                      `json:"version"`
	Outputs      []map[string]interface{} `json:"outputs"`
	MatchedRules []string                 `json:"matchedRules"`
	EvaluatedAt  time.Time                `json:"evaluatedAt"`
	DurationNs   int64                    `json:"durationNs"`
}

// Evaluate evaluates a decision
func (e *Engine) Evaluate(ctx context.Context, req *EvaluateRequest) (*EvaluateResult, error) {
	start := time.Now()

	// 1. Get definition
	var def *storage.Definition
	var err error

	if req.Version != nil && *req.Version > 0 {
		def, err = e.repo.GetByKeyAndVersion(ctx, req.DecisionKey, *req.Version, req.TenantID)
	} else {
		def, err = e.repo.GetByKey(ctx, req.DecisionKey, req.TenantID)
	}

	if err != nil {
		return nil, fmt.Errorf("definition not found: %w", err)
	}

	// 2. Find the decision
	decision := def.ParsedModel.GetDecision(req.DecisionKey)
	if decision == nil {
		// If not found by decision ID, use the first decision
		if len(def.ParsedModel.Decisions) > 0 {
			decision = &def.ParsedModel.Decisions[0]
		} else {
			return nil, fmt.Errorf("no decision found in definition")
		}
	}

	// 3. Evaluate the decision
	outputs, matchedRules, err := e.evaluateDecision(ctx, decision, req.Variables)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	// 4. Build result
	result := &EvaluateResult{
		DecisionKey:  def.Key,
		DecisionName: decision.Name,
		Version:      def.Version,
		Outputs:      outputs,
		MatchedRules: matchedRules,
		EvaluatedAt:  time.Now(),
		DurationNs:   time.Since(start).Nanoseconds(),
	}

	return result, nil
}

// evaluateDecision evaluates a single decision
func (e *Engine) evaluateDecision(ctx context.Context, decision *dmn.Decision, variables map[string]interface{}) ([]map[string]interface{}, []string, error) {
	// For now, only support decision tables
	if decision.DecisionTable == nil {
		return nil, nil, fmt.Errorf("decision must have a decision table (literal expressions not yet supported)")
	}

	return e.evaluateDecisionTable(ctx, decision.DecisionTable, variables)
}

// evaluateDecisionTable evaluates a decision table
func (e *Engine) evaluateDecisionTable(ctx context.Context, table *dmn.DecisionTable, variables map[string]interface{}) ([]map[string]interface{}, []string, error) {
	// Find matching rules
	var matchedRules []MatchedRule

	for _, rule := range table.Rules {
		matched, outputs, err := e.evaluateRule(ctx, &rule, table.Inputs, table.Outputs, variables)
		if err != nil {
			return nil, nil, fmt.Errorf("error evaluating rule %s: %w", rule.ID, err)
		}

		if matched {
			matchedRules = append(matchedRules, MatchedRule{
				RuleID:  rule.ID,
				Outputs: outputs,
			})

			// Stop on first match for FIRST policy
			hitPolicy := table.HitPolicy
			if hitPolicy == "" {
				hitPolicy = dmn.HitPolicyUnique
			}
			if hitPolicy == dmn.HitPolicyFirst {
				break
			}
		}
	}

	// Apply hit policy
	hitPolicy := table.HitPolicy
	if hitPolicy == "" {
		hitPolicy = dmn.HitPolicyUnique
	}

	strategy, ok := e.hitPolicies[hitPolicy]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported hit policy: %s", hitPolicy)
	}

	outputs, err := strategy.Apply(matchedRules, table.Aggregation)
	if err != nil {
		return nil, nil, fmt.Errorf("hit policy error: %w", err)
	}

	// Collect matched rule IDs
	ruleIDs := make([]string, len(matchedRules))
	for i, r := range matchedRules {
		ruleIDs[i] = r.RuleID
	}

	return outputs, ruleIDs, nil
}

