package engine

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/konstantin/dmn-engine-go/internal/dmn"
)

// evaluateRule evaluates a single rule against the input variables
func (e *Engine) evaluateRule(
	ctx context.Context,
	rule *dmn.Rule,
	inputs []dmn.Input,
	outputs []dmn.Output,
	variables map[string]interface{},
) (bool, map[string]interface{}, error) {

	// Check all input conditions
	for i, inputEntry := range rule.InputEntries {
		if i >= len(inputs) {
			return false, nil, fmt.Errorf("input entry index %d out of bounds", i)
		}

		input := inputs[i]
		
		// Get the input expression (variable name)
		inputExpr := input.InputExpression.Text
		
		// Get the value from variables
		value, ok := variables[inputExpr]
		if !ok {
			// Variable not provided, treat as no match
			return false, nil, nil
		}

		// Evaluate the unary test (input condition)
		matched, err := evaluateUnaryTest(inputEntry.Text, value)
		if err != nil {
			return false, nil, fmt.Errorf("error in input entry %d: %w", i, err)
		}

		if !matched {
			return false, nil, nil
		}
	}

	// All conditions matched - evaluate outputs
	outputValues := make(map[string]interface{})
	for i, outputEntry := range rule.OutputEntries {
		if i >= len(outputs) {
			return false, nil, fmt.Errorf("output entry index %d out of bounds", i)
		}

		output := outputs[i]
		outputName := output.Name
		if outputName == "" {
			outputName = output.ID
		}

		// Parse the output value
		value, err := parseOutputValue(outputEntry.Text)
		if err != nil {
			return false, nil, fmt.Errorf("error parsing output %s: %w", outputName, err)
		}

		outputValues[outputName] = value
	}

	return true, outputValues, nil
}

// evaluateUnaryTest evaluates a FEEL unary test expression
// This is a simplified implementation - full FEEL support would require a proper parser
func evaluateUnaryTest(expression string, value interface{}) (bool, error) {
	expr := strings.TrimSpace(expression)

	// Handle empty or "-" (any value matches)
	if expr == "" || expr == "-" {
		return true, nil
	}

	// Try to convert value to float for numeric comparisons
	var numValue float64
	var isNumeric bool

	switch v := value.(type) {
	case int:
		numValue = float64(v)
		isNumeric = true
	case int64:
		numValue = float64(v)
		isNumeric = true
	case float64:
		numValue = v
		isNumeric = true
	case float32:
		numValue = float64(v)
		isNumeric = true
	}

	// Handle quoted strings (exact match)
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		expectedValue := strings.Trim(expr, "\"")
		actualValue := fmt.Sprintf("%v", value)
		return actualValue == expectedValue, nil
	}

	// Handle ranges: [10..20], ]10..20[, etc.
	if strings.Contains(expr, "..") {
		if !isNumeric {
			return false, fmt.Errorf("range comparison requires numeric value")
		}
		return evaluateRange(expr, numValue)
	}

	// Handle comparison operators: <, >, <=, >=, =, !=
	if isNumeric {
		if strings.HasPrefix(expr, "<=") {
			threshold, err := strconv.ParseFloat(strings.TrimSpace(expr[2:]), 64)
			if err != nil {
				return false, err
			}
			return numValue <= threshold, nil
		}
		if strings.HasPrefix(expr, ">=") {
			threshold, err := strconv.ParseFloat(strings.TrimSpace(expr[2:]), 64)
			if err != nil {
				return false, err
			}
			return numValue >= threshold, nil
		}
		if strings.HasPrefix(expr, "<") {
			threshold, err := strconv.ParseFloat(strings.TrimSpace(expr[1:]), 64)
			if err != nil {
				return false, err
			}
			return numValue < threshold, nil
		}
		if strings.HasPrefix(expr, ">") {
			threshold, err := strconv.ParseFloat(strings.TrimSpace(expr[1:]), 64)
			if err != nil {
				return false, err
			}
			return numValue > threshold, nil
		}
		if strings.HasPrefix(expr, "=") {
			threshold, err := strconv.ParseFloat(strings.TrimSpace(expr[1:]), 64)
			if err != nil {
				return false, err
			}
			return numValue == threshold, nil
		}
	}

	// Handle lists: "value1", "value2", ...
	if strings.Contains(expr, ",") {
		values := strings.Split(expr, ",")
		actualValue := fmt.Sprintf("%v", value)
		for _, v := range values {
			v = strings.TrimSpace(v)
			v = strings.Trim(v, "\"")
			if actualValue == v {
				return true, nil
			}
		}
		return false, nil
	}

	// Default: try exact match
	if isNumeric {
		threshold, err := strconv.ParseFloat(expr, 64)
		if err == nil {
			return numValue == threshold, nil
		}
	}

	// String comparison
	return fmt.Sprintf("%v", value) == expr, nil
}

// evaluateRange evaluates a range expression like [10..20], ]10..20[, etc.
func evaluateRange(expr string, value float64) (bool, error) {
	// Determine if endpoints are inclusive
	leftInclusive := strings.HasPrefix(expr, "[")
	rightInclusive := strings.HasSuffix(expr, "]")

	// Remove brackets
	expr = strings.Trim(expr, "[]")
	expr = strings.Trim(expr, "()")

	// Split on ".."
	parts := strings.Split(expr, "..")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid range format: %s", expr)
	}

	// Parse bounds
	lower, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return false, fmt.Errorf("invalid lower bound: %w", err)
	}

	upper, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return false, fmt.Errorf("invalid upper bound: %w", err)
	}

	// Check bounds
	if leftInclusive && rightInclusive {
		return value >= lower && value <= upper, nil
	} else if leftInclusive && !rightInclusive {
		return value >= lower && value < upper, nil
	} else if !leftInclusive && rightInclusive {
		return value > lower && value <= upper, nil
	} else {
		return value > lower && value < upper, nil
	}
}

// parseOutputValue parses an output expression to a Go value
// This is simplified - full FEEL would require proper parsing
func parseOutputValue(expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)

	// Empty
	if expr == "" {
		return nil, nil
	}

	// String literal
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return strings.Trim(expr, "\""), nil
	}

	// Boolean
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}

	// Number
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		// Return as int if it's a whole number
		if num == float64(int64(num)) {
			return int64(num), nil
		}
		return num, nil
	}

	// Default: return as string
	return expr, nil
}

