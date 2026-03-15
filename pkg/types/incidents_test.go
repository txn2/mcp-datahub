package types

import (
	"encoding/json"
	"testing"
)

func TestIncident_JSON(t *testing.T) {
	inc := Incident{
		URN:           "urn:li:incident:123",
		Type:          "OPERATIONAL",
		Title:         "Pipeline failure",
		Description:   "ETL pipeline failed at step 3",
		State:         "ACTIVE",
		Source:        "MANUAL",
		Created:       1700000000000,
		CreatedBy:     "urn:li:corpuser:admin",
		LastUpdated:   1700001000000,
		LastUpdatedBy: "urn:li:corpuser:admin",
	}

	got, err := json.Marshal(inc)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var roundTrip Incident
	if err := json.Unmarshal(got, &roundTrip); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if roundTrip.URN != inc.URN {
		t.Errorf("URN = %q, want %q", roundTrip.URN, inc.URN)
	}
	if roundTrip.State != "ACTIVE" {
		t.Errorf("State = %q, want ACTIVE", roundTrip.State)
	}
	if roundTrip.Title != inc.Title {
		t.Errorf("Title = %q, want %q", roundTrip.Title, inc.Title)
	}
}

func TestIncident_OmitEmpty(t *testing.T) {
	inc := Incident{
		URN:   "urn:li:incident:123",
		Type:  "OPERATIONAL",
		Title: "Test",
		State: "ACTIVE",
	}

	got, err := json.Marshal(inc)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if _, ok := m["description"]; ok {
		t.Error("description should be omitted when empty")
	}
	if _, ok := m["custom_type"]; ok {
		t.Error("custom_type should be omitted when empty")
	}
}

func TestIncidentResult_JSON(t *testing.T) {
	result := IncidentResult{
		Total: 1,
		Incidents: []Incident{
			{URN: "urn:li:incident:1", Type: "OPERATIONAL", Title: "Down", State: "ACTIVE"},
		},
	}

	got, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var roundTrip IncidentResult
	if err := json.Unmarshal(got, &roundTrip); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if roundTrip.Total != 1 {
		t.Errorf("Total = %d, want 1", roundTrip.Total)
	}
	if len(roundTrip.Incidents) != 1 {
		t.Fatalf("Incidents count = %d, want 1", len(roundTrip.Incidents))
	}
}

func TestRaiseIncidentInput_JSON(t *testing.T) {
	input := RaiseIncidentInput{
		Type:         "OPERATIONAL",
		Title:        "Pipeline failure",
		Description:  "ETL failed",
		ResourceURNs: []string{"urn:li:dataset:test"},
	}

	got, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var roundTrip RaiseIncidentInput
	if err := json.Unmarshal(got, &roundTrip); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if roundTrip.Type != "OPERATIONAL" {
		t.Errorf("Type = %q", roundTrip.Type)
	}
	if len(roundTrip.ResourceURNs) != 1 {
		t.Errorf("ResourceURNs count = %d", len(roundTrip.ResourceURNs))
	}
}
