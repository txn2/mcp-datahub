package types

import (
	"encoding/json"
	"testing"
)

func TestDataContract_JSON(t *testing.T) {
	contract := DataContract{
		Status: "PASSING",
		AssertionResults: []AssertionResult{
			{
				AssertionURN: "urn:li:assertion:freshness-check",
				Type:         "FRESHNESS",
			},
			{
				AssertionURN: "urn:li:assertion:schema-check",
				Type:         "SCHEMA",
			},
		},
	}

	got, err := json.Marshal(contract)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var roundTrip DataContract
	if err := json.Unmarshal(got, &roundTrip); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if roundTrip.Status != "PASSING" {
		t.Errorf("Status = %q, want PASSING", roundTrip.Status)
	}
	if len(roundTrip.AssertionResults) != 2 {
		t.Fatalf("AssertionResults count = %d, want 2", len(roundTrip.AssertionResults))
	}
	if roundTrip.AssertionResults[1].Type != "SCHEMA" {
		t.Error("Type not preserved through round-trip")
	}
}

func TestDataContract_OmitEmpty(t *testing.T) {
	contract := DataContract{Status: "PASSING"}

	got, err := json.Marshal(contract)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if _, ok := m["assertion_results"]; ok {
		t.Error("assertion_results should be omitted when empty")
	}
}
