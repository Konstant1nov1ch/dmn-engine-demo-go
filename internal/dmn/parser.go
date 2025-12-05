package dmn

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

// Parser parses DMN XML files into Go structures
type Parser struct{}

// NewParser creates a new DMN parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses DMN XML from a reader
func (p *Parser) Parse(r io.Reader) (*Definitions, error) {
	var defs Definitions
	decoder := xml.NewDecoder(r)
	
	if err := decoder.Decode(&defs); err != nil {
		return nil, fmt.Errorf("failed to parse DMN XML: %w", err)
	}
	
	// Set default hit policy if not specified
	for i := range defs.Decisions {
		if defs.Decisions[i].DecisionTable != nil && defs.Decisions[i].DecisionTable.HitPolicy == "" {
			defs.Decisions[i].DecisionTable.HitPolicy = HitPolicyUnique
		}
	}
	
	return &defs, nil
}

// ParseFile parses a DMN XML file
func (p *Parser) ParseFile(path string) (*Definitions, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()
	
	return p.Parse(f)
}

// ParseBytes parses DMN XML from bytes
func (p *Parser) ParseBytes(data []byte) (*Definitions, error) {
	var defs Definitions
	if err := xml.Unmarshal(data, &defs); err != nil {
		return nil, fmt.Errorf("failed to parse DMN XML: %w", err)
	}
	
	// Set default hit policy if not specified
	for i := range defs.Decisions {
		if defs.Decisions[i].DecisionTable != nil && defs.Decisions[i].DecisionTable.HitPolicy == "" {
			defs.Decisions[i].DecisionTable.HitPolicy = HitPolicyUnique
		}
	}
	
	return &defs, nil
}

// GetDecision returns a decision by ID
func (d *Definitions) GetDecision(id string) *Decision {
	for i := range d.Decisions {
		if d.Decisions[i].ID == id {
			return &d.Decisions[i]
		}
	}
	return nil
}

// GetInputData returns input data by ID
func (d *Definitions) GetInputData(id string) *InputData {
	for i := range d.InputData {
		if d.InputData[i].ID == id {
			return &d.InputData[i]
		}
	}
	return nil
}

