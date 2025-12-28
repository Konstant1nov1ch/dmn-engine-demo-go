#!/bin/bash

# Generate DMN files in Camunda format (based on bundle_with_promotion_ver_from_AT_drg.dmn.backup)
# Just randomize IDs

set -e

OUTPUT_DIR="testdata/dmn/stress"
mkdir -p "$OUTPUT_DIR"

echo "Generating DMN decision tables in Camunda format..."

# Generate N decision tables
NUM_TABLES=${1:-50}

# Function to generate random ID (Camunda style)
random_id() {
    local prefix=$1
    echo "${prefix}_$(openssl rand -hex 7)"
}

for i in $(seq 1 $NUM_TABLES); do
    DECISION_ID=$(random_id "Decision")
    TABLE_ID=$(random_id "DecisionTable")
    INPUT_ID=$(random_id "InputClause")
    INPUT_EXPR_ID=$(random_id "LiteralExpression")
    OUTPUT_ID=$(random_id "OutputClause")
    
    cat > "${OUTPUT_DIR}/decision_${i}.dmn" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/" 
             xmlns:dmndi="https://www.omg.org/spec/DMN/20191111/DMNDI/" 
             xmlns:dc="http://www.omg.org/spec/DMN/20180521/DC/" 
             xmlns:camunda="http://camunda.org/schema/1.0/dmn" 
             xmlns:modeler="http://camunda.org/schema/modeler/1.0" 
             xmlns:biodi="http://bpmn.io/schema/dmn/biodi/2.0" 
             id="def_${i}" 
             name="Decision Table ${i}" 
             namespace="http://example.org/dmn" 
             exporter="Camunda Modeler" 
             exporterVersion="5.34.0" 
             modeler:executionPlatform="Camunda Platform" 
             modeler:executionPlatformVersion="7.17.0">
  <decision id="${DECISION_ID}" name="Decision_${i}" camunda:historyTimeToLive="200">
    <decisionTable id="${TABLE_ID}" hitPolicy="UNIQUE">
      <input id="${INPUT_ID}" label="Age">
        <inputExpression id="${INPUT_EXPR_ID}" typeRef="integer">
          <text>age</text>
        </inputExpression>
      </input>
      <output id="${OUTPUT_ID}" name="result" typeRef="string" />
      <rule id="DecisionRule_$(openssl rand -hex 7)">
        <description></description>
        <inputEntry id="UnaryTests_$(openssl rand -hex 7)">
          <text>&lt; 18</text>
        </inputEntry>
        <outputEntry id="LiteralExpression_$(openssl rand -hex 7)">
          <text>"rejected_${i}"</text>
        </outputEntry>
      </rule>
      <rule id="DecisionRule_$(openssl rand -hex 7)">
        <description></description>
        <inputEntry id="UnaryTests_$(openssl rand -hex 7)">
          <text>[18..65]</text>
        </inputEntry>
        <outputEntry id="LiteralExpression_$(openssl rand -hex 7)">
          <text>"approved_${i}"</text>
        </outputEntry>
      </rule>
      <rule id="DecisionRule_$(openssl rand -hex 7)">
        <description></description>
        <inputEntry id="UnaryTests_$(openssl rand -hex 7)">
          <text>&gt; 65</text>
        </inputEntry>
        <outputEntry id="LiteralExpression_$(openssl rand -hex 7)">
          <text>"pending_${i}"</text>
        </outputEntry>
      </rule>
    </decisionTable>
  </decision>
</definitions>
EOF
done

echo "✅ Generated $NUM_TABLES DMN decision tables in Camunda format in $OUTPUT_DIR"

