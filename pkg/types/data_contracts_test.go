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
				ResultType:   "SUCCESS",
			},
			{
				AssertionURN: "urn:li:assertion:schema-check",
				Type:         "SCHEMA",
				ResultType:   "FAILURE",
				NativeResults: map[string]string{
					"expected_columns": "10",
					"actual_columns":   "9",
				},
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
	if roundTrip.AssertionResults[1].NativeResults["expected_columns"] != "10" {
		t.Error("NativeResults not preserved through round-trip")
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

func TestAssertionResult_OmitEmpty(t *testing.T) {
	ar := AssertionResult{
		AssertionURN: "urn:li:assertion:test",
		Type:         "FRESHNESS",
		ResultType:   "SUCCESS",
	}

	got, err := json.Marshal(ar)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if _, ok := m["native_results"]; ok {
		t.Error("native_results should be omitted when empty")
	}
}
