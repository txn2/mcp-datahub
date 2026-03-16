package client

import "testing"

func TestParseSearchResult_DocumentEntity(t *testing.T) {
	sr := searchResultItem{}
	sr.Entity.URN = "urn:li:document:runbook-1"
	sr.Entity.Type = "DOCUMENT"
	sr.Entity.Info.Title = "Incident Runbook"
	sr.Entity.Info.Contents.Text = "Step 1: Check the logs"
	sr.Entity.Info.Status.State = "PUBLISHED"
	sr.MatchedFields = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "title", Value: "Incident Runbook"},
	}

	entity := parseSearchResult(sr)

	if entity.URN != "urn:li:document:runbook-1" {
		t.Errorf("URN = %q, want urn:li:document:runbook-1", entity.URN)
	}
	if entity.Type != "DOCUMENT" {
		t.Errorf("Type = %q, want DOCUMENT", entity.Type)
	}
	if entity.Name != "Incident Runbook" {
		t.Errorf("Name = %q, want Incident Runbook (from info.title)", entity.Name)
	}
	if entity.Description != "Step 1: Check the logs" {
		t.Errorf("Description = %q, want document content as description", entity.Description)
	}
	if len(entity.MatchedFields) != 1 {
		t.Errorf("MatchedFields len = %d, want 1", len(entity.MatchedFields))
	}
}

func TestParseSearchResult_DocumentWithExistingDescription(t *testing.T) {
	sr := searchResultItem{}
	sr.Entity.URN = "urn:li:document:doc-1"
	sr.Entity.Type = "DOCUMENT"
	sr.Entity.Description = "Explicit description"
	sr.Entity.Info.Title = "My Doc"
	sr.Entity.Info.Contents.Text = "Full content body"

	entity := parseSearchResult(sr)

	// When the entity already has a description, content should not override it.
	if entity.Description != "Explicit description" {
		t.Errorf("Description = %q, want Explicit description (content should not override)", entity.Description)
	}
	if entity.Name != "My Doc" {
		t.Errorf("Name = %q, want My Doc", entity.Name)
	}
}

func TestParseSearchResult_NonDocumentEntity(t *testing.T) {
	sr := searchResultItem{}
	sr.Entity.URN = "urn:li:dataset:test"
	sr.Entity.Type = "DATASET"
	sr.Entity.Name = "test_table"
	sr.Entity.Description = "A test table"
	sr.Entity.Platform.Name = "snowflake"
	sr.Entity.Ownership.Owners = []struct {
		Owner struct {
			URN      string `json:"urn"`
			Username string `json:"username"`
			Name     string `json:"name"`
		} `json:"owner"`
		Type string `json:"type"`
	}{
		{
			Owner: struct {
				URN      string `json:"urn"`
				Username string `json:"username"`
				Name     string `json:"name"`
			}{URN: "urn:li:corpuser:alice", Username: "alice", Name: "Alice"},
			Type: "DATAOWNER",
		},
	}
	sr.Entity.Tags.Tags = []struct {
		Tag struct {
			URN         string `json:"urn"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"tag"`
	}{
		{Tag: struct {
			URN         string `json:"urn"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}{URN: "urn:li:tag:prod", Name: "prod"}},
	}
	sr.Entity.Domain.Domain.URN = "urn:li:domain:eng"
	sr.Entity.Domain.Domain.Properties.Name = "Engineering"

	entity := parseSearchResult(sr)

	if entity.Name != "test_table" {
		t.Errorf("Name = %q, want test_table", entity.Name)
	}
	if entity.Platform != "snowflake" {
		t.Errorf("Platform = %q, want snowflake", entity.Platform)
	}
	if len(entity.Owners) != 1 || entity.Owners[0].Name != "Alice" {
		t.Errorf("Owners = %+v, want Alice", entity.Owners)
	}
	if len(entity.Tags) != 1 || entity.Tags[0].Name != "prod" {
		t.Errorf("Tags = %+v, want prod", entity.Tags)
	}
	if entity.Domain == nil || entity.Domain.Name != "Engineering" {
		t.Errorf("Domain = %+v, want Engineering", entity.Domain)
	}
}

func TestParseSearchResult_PropertiesOverride(t *testing.T) {
	sr := searchResultItem{}
	sr.Entity.URN = "urn:li:glossaryTerm:pii"
	sr.Entity.Type = "GLOSSARY_TERM"
	sr.Entity.Name = "generic-name"
	sr.Entity.Description = "generic-desc"
	sr.Entity.Properties.Name = "PII"
	sr.Entity.Properties.Description = "Personally Identifiable Information"

	entity := parseSearchResult(sr)

	if entity.Name != "PII" {
		t.Errorf("Name = %q, want PII (from properties)", entity.Name)
	}
	if entity.Description != "Personally Identifiable Information" {
		t.Errorf("Description = %q, want from properties", entity.Description)
	}
}
