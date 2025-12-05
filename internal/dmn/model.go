package dmn

import "encoding/xml"

// Definitions is the root element of a DMN model
type Definitions struct {
	XMLName                 xml.Name                 `xml:"definitions"`
	ID                      string                   `xml:"id,attr"`
	Name                    string                   `xml:"name,attr"`
	Namespace               string                   `xml:"namespace,attr"`
	ExpressionLanguage      string                   `xml:"expressionLanguage,attr,omitempty"`
	Decisions               []Decision               `xml:"decision"`
	InputData               []InputData              `xml:"inputData"`
	BusinessKnowledgeModels []BusinessKnowledgeModel `xml:"businessKnowledgeModel"`
}

// Decision represents a DMN decision element
type Decision struct {
	ID                      string                   `xml:"id,attr"`
	Name                    string                   `xml:"name,attr"`
	Variable                *Variable                `xml:"variable"`
	InformationRequirements []InformationRequirement `xml:"informationRequirement"`
	DecisionTable           *DecisionTable           `xml:"decisionTable"`
	LiteralExpression       *LiteralExpression       `xml:"literalExpression"`
}

// Variable represents the output variable of a decision
type Variable struct {
	ID      string `xml:"id,attr,omitempty"`
	Name    string `xml:"name,attr"`
	TypeRef string `xml:"typeRef,attr,omitempty"`
}

// InformationRequirement represents a dependency on another decision or input
type InformationRequirement struct {
	ID               string            `xml:"id,attr,omitempty"`
	RequiredDecision *RequiredDecision `xml:"requiredDecision"`
	RequiredInput    *RequiredInput    `xml:"requiredInput"`
}

// RequiredDecision is a reference to another decision
type RequiredDecision struct {
	Href string `xml:"href,attr"` // e.g., "#otherDecision"
}

// RequiredInput is a reference to an input data element
type RequiredInput struct {
	Href string `xml:"href,attr"` // e.g., "#inputData1"
}

// InputData represents input data for decisions
type InputData struct {
	ID       string    `xml:"id,attr"`
	Name     string    `xml:"name,attr"`
	Variable *Variable `xml:"variable"`
}

// BusinessKnowledgeModel represents a BKM element
type BusinessKnowledgeModel struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// DecisionTable represents a DMN decision table
type DecisionTable struct {
	ID              string   `xml:"id,attr,omitempty"`
	HitPolicy       string   `xml:"hitPolicy,attr,omitempty"` // UNIQUE, FIRST, PRIORITY, ANY, COLLECT, RULE ORDER, OUTPUT ORDER
	Aggregation     string   `xml:"aggregation,attr,omitempty"` // SUM, COUNT, MIN, MAX (for COLLECT)
	PreferredOrientation string `xml:"preferredOrientation,attr,omitempty"` // Rule-as-Row, Rule-as-Column
	Inputs          []Input  `xml:"input"`
	Outputs         []Output `xml:"output"`
	Rules           []Rule   `xml:"rule"`
}

// Input represents an input column in a decision table
type Input struct {
	ID              string          `xml:"id,attr,omitempty"`
	Label           string          `xml:"label,attr,omitempty"`
	InputExpression InputExpression `xml:"inputExpression"`
	InputValues     *InputValues    `xml:"inputValues,omitempty"`
}

// InputExpression defines the expression for an input
type InputExpression struct {
	ID      string `xml:"id,attr,omitempty"`
	TypeRef string `xml:"typeRef,attr,omitempty"`
	Text    string `xml:"text"` // FEEL expression
}

// InputValues defines allowed values for an input
type InputValues struct {
	Text string `xml:"text"` // FEEL expression like "1,2,3" or '"a","b","c"'
}

// Output represents an output column in a decision table
type Output struct {
	ID           string        `xml:"id,attr,omitempty"`
	Label        string        `xml:"label,attr,omitempty"`
	Name         string        `xml:"name,attr,omitempty"`
	TypeRef      string        `xml:"typeRef,attr,omitempty"`
	OutputValues *OutputValues `xml:"outputValues,omitempty"`
}

// OutputValues defines allowed values for an output
type OutputValues struct {
	Text string `xml:"text"`
}

// Rule represents a rule in a decision table
type Rule struct {
	ID            string        `xml:"id,attr,omitempty"`
	Description   string        `xml:"description,omitempty"`
	InputEntries  []InputEntry  `xml:"inputEntry"`
	OutputEntries []OutputEntry `xml:"outputEntry"`
}

// InputEntry represents a condition cell in a rule
type InputEntry struct {
	ID   string `xml:"id,attr,omitempty"`
	Text string `xml:"text"` // FEEL unary test: ">= 18", "[1..100]", "-" (any)
}

// OutputEntry represents an output cell in a rule
type OutputEntry struct {
	ID   string `xml:"id,attr,omitempty"`
	Text string `xml:"text"` // FEEL expression or literal
}

// LiteralExpression represents a decision defined by a FEEL expression
type LiteralExpression struct {
	ID      string `xml:"id,attr,omitempty"`
	TypeRef string `xml:"typeRef,attr,omitempty"`
	Text    string `xml:"text"` // FEEL expression
}

// HitPolicy constants
const (
	HitPolicyUnique      = "UNIQUE"
	HitPolicyFirst       = "FIRST"
	HitPolicyPriority    = "PRIORITY"
	HitPolicyAny         = "ANY"
	HitPolicyCollect     = "COLLECT"
	HitPolicyRuleOrder   = "RULE ORDER"
	HitPolicyOutputOrder = "OUTPUT ORDER"
)

// Aggregation constants for COLLECT hit policy
const (
	AggregationSum   = "SUM"
	AggregationCount = "COUNT"
	AggregationMin   = "MIN"
	AggregationMax   = "MAX"
)

