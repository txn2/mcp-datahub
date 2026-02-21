//go:build integration

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

// skipIfNoEnv skips the test if required environment variables are not set.
func skipIfNoEnv(t *testing.T) {
	t.Helper()
	if os.Getenv("DATAHUB_URL") == "" || os.Getenv("DATAHUB_TOKEN") == "" {
		t.Skip("DATAHUB_URL and DATAHUB_TOKEN must be set for integration tests")
	}
}

// testClient creates a Client from environment variables and registers cleanup.
func testClient(t *testing.T) *Client {
	t.Helper()
	c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// testEntityURN returns TEST_ENTITY_URN or discovers a dataset via Search.
func testEntityURN(t *testing.T, c *Client) string {
	t.Helper()
	if urn := os.Getenv("TEST_ENTITY_URN"); urn != "" {
		return urn
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := c.Search(ctx, "*", WithEntityType("DATASET"), WithLimit(1))
	if err != nil {
		t.Fatalf("Search for test entity: %v", err)
	}
	if len(result.Entities) == 0 {
		t.Fatal("no datasets found; set TEST_ENTITY_URN explicitly")
	}
	urn := result.Entities[0].URN
	t.Logf("discovered test entity: %s", urn)
	return urn
}

// testCtx returns a context with a 30-second timeout and cleanup.
func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// containsTag checks if the globalTags aspect contains a tag with the given URN.
func containsTag(tags *globalTagsAspect, urn string) bool {
	for _, t := range tags.Tags {
		if t.Tag == urn {
			return true
		}
	}
	return false
}

// containsTerm checks if the glossaryTerms aspect contains a term with the given URN.
func containsTerm(terms *glossaryTermsAspect, urn string) bool {
	for _, t := range terms.Terms {
		if t.URN == urn {
			return true
		}
	}
	return false
}

// containsLink checks if the institutionalMemory aspect contains a link with the given URL.
func containsLink(memory *institutionalMemoryAspect, url string) bool {
	for _, e := range memory.Elements {
		if e.URL == url {
			return true
		}
	}
	return false
}

// nanos returns a unique suffix based on UnixNano for test isolation.
func nanos() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func TestIntegrationUpdateDescription(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	urn := testEntityURN(t, c)
	ctx := testCtx(t)

	// Read original description via REST so we can restore it
	var originalDesc string
	raw, err := c.getAspect(ctx, urn, "editableDatasetProperties")
	if err == nil {
		var props struct {
			Description string `json:"description"`
		}
		if jsonErr := json.Unmarshal(raw, &props); jsonErr == nil {
			originalDesc = props.Description
		}
	}

	uniqueDesc := fmt.Sprintf("Integration test desc %s", nanos())

	// Register cleanup before the write
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanCancel()
		if restoreErr := c.UpdateDescription(cleanCtx, urn, originalDesc); restoreErr != nil {
			t.Logf("cleanup: failed to restore description: %v", restoreErr)
		}
	})

	// Write
	if err := c.UpdateDescription(ctx, urn, uniqueDesc); err != nil {
		t.Fatalf("UpdateDescription: %v", err)
	}

	// Verify via REST
	raw, err = c.getAspect(ctx, urn, "editableDatasetProperties")
	if err != nil {
		t.Fatalf("getAspect after write: %v", err)
	}
	var props struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal(raw, &props); err != nil {
		t.Fatalf("unmarshal editableDatasetProperties: %v", err)
	}
	if props.Description != uniqueDesc {
		t.Errorf("REST verify: got description %q, want %q", props.Description, uniqueDesc)
	}

	// Verify via GraphQL (GetEntity reads editableProperties)
	entity, err := c.GetEntity(ctx, urn)
	if err != nil {
		t.Fatalf("GetEntity after write: %v", err)
	}
	if entity.Description != uniqueDesc {
		t.Errorf("GraphQL verify: got description %q, want %q", entity.Description, uniqueDesc)
	}
}

func TestIntegrationAddTag(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	urn := testEntityURN(t, c)
	ctx := testCtx(t)

	tagURN := fmt.Sprintf("urn:li:tag:inttest_%s", nanos())

	// Register cleanup before the write
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanCancel()
		if err := c.RemoveTag(cleanCtx, urn, tagURN); err != nil {
			t.Logf("cleanup: failed to remove tag: %v", err)
		}
	})

	// Add tag
	if err := c.AddTag(ctx, urn, tagURN); err != nil {
		t.Fatalf("AddTag: %v", err)
	}

	// Verify
	tags, err := c.readGlobalTags(ctx, urn)
	if err != nil {
		t.Fatalf("readGlobalTags after add: %v", err)
	}
	if !containsTag(tags, tagURN) {
		t.Errorf("tag %s not found after AddTag", tagURN)
	}

	// Idempotency: second add should be a no-op
	t.Run("idempotent", func(t *testing.T) {
		if err := c.AddTag(ctx, urn, tagURN); err != nil {
			t.Fatalf("AddTag (idempotent): %v", err)
		}
		tags2, err := c.readGlobalTags(ctx, urn)
		if err != nil {
			t.Fatalf("readGlobalTags after idempotent add: %v", err)
		}
		count := 0
		for _, tg := range tags2.Tags {
			if tg.Tag == tagURN {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected exactly 1 occurrence of tag, got %d", count)
		}
	})
}

func TestIntegrationRemoveTag(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	urn := testEntityURN(t, c)
	ctx := testCtx(t)

	tagURN := fmt.Sprintf("urn:li:tag:inttest_%s", nanos())

	// Setup: add the tag first
	if err := c.AddTag(ctx, urn, tagURN); err != nil {
		t.Fatalf("setup AddTag: %v", err)
	}

	// Register cleanup (RemoveTag is idempotent)
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanCancel()
		if err := c.RemoveTag(cleanCtx, urn, tagURN); err != nil {
			t.Logf("cleanup: failed to remove tag: %v", err)
		}
	})

	// Remove tag
	if err := c.RemoveTag(ctx, urn, tagURN); err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}

	// Verify removed
	tags, err := c.readGlobalTags(ctx, urn)
	if err != nil {
		t.Fatalf("readGlobalTags after remove: %v", err)
	}
	if containsTag(tags, tagURN) {
		t.Errorf("tag %s still present after RemoveTag", tagURN)
	}
}

func TestIntegrationAddGlossaryTerm(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	urn := testEntityURN(t, c)
	ctx := testCtx(t)

	termURN := fmt.Sprintf("urn:li:glossaryTerm:inttest_%s", nanos())

	// Register cleanup before the write
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanCancel()
		if err := c.RemoveGlossaryTerm(cleanCtx, urn, termURN); err != nil {
			t.Logf("cleanup: failed to remove glossary term: %v", err)
		}
	})

	// Add glossary term
	if err := c.AddGlossaryTerm(ctx, urn, termURN); err != nil {
		t.Fatalf("AddGlossaryTerm: %v", err)
	}

	// Verify
	terms, err := c.readGlossaryTerms(ctx, urn)
	if err != nil {
		t.Fatalf("readGlossaryTerms after add: %v", err)
	}
	if !containsTerm(terms, termURN) {
		t.Errorf("term %s not found after AddGlossaryTerm", termURN)
	}

	// Idempotency
	t.Run("idempotent", func(t *testing.T) {
		if err := c.AddGlossaryTerm(ctx, urn, termURN); err != nil {
			t.Fatalf("AddGlossaryTerm (idempotent): %v", err)
		}
		terms2, err := c.readGlossaryTerms(ctx, urn)
		if err != nil {
			t.Fatalf("readGlossaryTerms after idempotent add: %v", err)
		}
		count := 0
		for _, tm := range terms2.Terms {
			if tm.URN == termURN {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected exactly 1 occurrence of term, got %d", count)
		}
	})
}

func TestIntegrationRemoveGlossaryTerm(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	urn := testEntityURN(t, c)
	ctx := testCtx(t)

	termURN := fmt.Sprintf("urn:li:glossaryTerm:inttest_%s", nanos())

	// Setup: add the term first
	if err := c.AddGlossaryTerm(ctx, urn, termURN); err != nil {
		t.Fatalf("setup AddGlossaryTerm: %v", err)
	}

	// Register cleanup (RemoveGlossaryTerm is idempotent)
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanCancel()
		if err := c.RemoveGlossaryTerm(cleanCtx, urn, termURN); err != nil {
			t.Logf("cleanup: failed to remove glossary term: %v", err)
		}
	})

	// Remove
	if err := c.RemoveGlossaryTerm(ctx, urn, termURN); err != nil {
		t.Fatalf("RemoveGlossaryTerm: %v", err)
	}

	// Verify removed
	terms, err := c.readGlossaryTerms(ctx, urn)
	if err != nil {
		t.Fatalf("readGlossaryTerms after remove: %v", err)
	}
	if containsTerm(terms, termURN) {
		t.Errorf("term %s still present after RemoveGlossaryTerm", termURN)
	}
}

func TestIntegrationAddLink(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	urn := testEntityURN(t, c)
	ctx := testCtx(t)

	linkURL := fmt.Sprintf("https://inttest-%s.example.com", nanos())
	linkDesc := "Integration test link"

	// Register cleanup before the write
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanCancel()
		if err := c.RemoveLink(cleanCtx, urn, linkURL); err != nil {
			t.Logf("cleanup: failed to remove link: %v", err)
		}
	})

	// Add link
	if err := c.AddLink(ctx, urn, linkURL, linkDesc); err != nil {
		t.Fatalf("AddLink: %v", err)
	}

	// Verify
	memory, err := c.readInstitutionalMemory(ctx, urn)
	if err != nil {
		t.Fatalf("readInstitutionalMemory after add: %v", err)
	}
	if !containsLink(memory, linkURL) {
		t.Errorf("link %s not found after AddLink", linkURL)
	}

	// Idempotency
	t.Run("idempotent", func(t *testing.T) {
		if err := c.AddLink(ctx, urn, linkURL, linkDesc); err != nil {
			t.Fatalf("AddLink (idempotent): %v", err)
		}
		memory2, err := c.readInstitutionalMemory(ctx, urn)
		if err != nil {
			t.Fatalf("readInstitutionalMemory after idempotent add: %v", err)
		}
		count := 0
		for _, e := range memory2.Elements {
			if e.URL == linkURL {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected exactly 1 occurrence of link, got %d", count)
		}
	})
}

func TestIntegrationRemoveLink(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	urn := testEntityURN(t, c)
	ctx := testCtx(t)

	linkURL := fmt.Sprintf("https://inttest-%s.example.com", nanos())
	linkDesc := "Integration test link"

	// Setup: add the link first
	if err := c.AddLink(ctx, urn, linkURL, linkDesc); err != nil {
		t.Fatalf("setup AddLink: %v", err)
	}

	// Register cleanup (RemoveLink is idempotent)
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanCancel()
		if err := c.RemoveLink(cleanCtx, urn, linkURL); err != nil {
			t.Logf("cleanup: failed to remove link: %v", err)
		}
	})

	// Remove
	if err := c.RemoveLink(ctx, urn, linkURL); err != nil {
		t.Fatalf("RemoveLink: %v", err)
	}

	// Verify removed
	memory, err := c.readInstitutionalMemory(ctx, urn)
	if err != nil {
		t.Fatalf("readInstitutionalMemory after remove: %v", err)
	}
	if containsLink(memory, linkURL) {
		t.Errorf("link %s still present after RemoveLink", linkURL)
	}
}

func TestIntegrationCreateQuery(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	ctx := testCtx(t)

	queryName := fmt.Sprintf("inttest_query_%s", nanos())

	var createdURN string

	// Register cleanup before the write
	t.Cleanup(func() {
		if createdURN == "" {
			return
		}
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanCancel()
		if err := c.DeleteQuery(cleanCtx, createdURN); err != nil {
			t.Logf("cleanup: failed to delete query: %v", err)
		}
	})

	// Create query
	query, err := c.CreateQuery(ctx, CreateQueryInput{
		Name:      queryName,
		Statement: "SELECT 1 AS integration_test",
	})
	if err != nil {
		t.Fatalf("CreateQuery: %v", err)
	}

	createdURN = query.URN
	if createdURN == "" {
		t.Fatal("expected non-empty URN from CreateQuery")
	}
	if query.Name != queryName {
		t.Errorf("expected Name %q, got %q", queryName, query.Name)
	}
	if query.Statement != "SELECT 1 AS integration_test" {
		t.Errorf("unexpected Statement: %q", query.Statement)
	}
}

func TestIntegrationUpdateQuery(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	ctx := testCtx(t)

	queryName := fmt.Sprintf("inttest_query_%s", nanos())

	// Setup: create a query first
	query, err := c.CreateQuery(ctx, CreateQueryInput{
		Name:      queryName,
		Statement: "SELECT 1",
	})
	if err != nil {
		t.Fatalf("setup CreateQuery: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		cleanCtx, cleanCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanCancel()
		if err := c.DeleteQuery(cleanCtx, query.URN); err != nil {
			t.Logf("cleanup: failed to delete query: %v", err)
		}
	})

	// Update the query
	updatedName := queryName + "_updated"
	updated, err := c.UpdateQuery(ctx, UpdateQueryInput{
		URN:       query.URN,
		Name:      updatedName,
		Statement: "SELECT 2",
	})
	if err != nil {
		t.Fatalf("UpdateQuery: %v", err)
	}
	if updated.URN != query.URN {
		t.Errorf("expected URN %q, got %q", query.URN, updated.URN)
	}
}

func TestIntegrationDeleteQuery(t *testing.T) {
	skipIfNoEnv(t)
	c := testClient(t)
	ctx := testCtx(t)

	queryName := fmt.Sprintf("inttest_query_%s", nanos())

	// Create a query to delete
	query, err := c.CreateQuery(ctx, CreateQueryInput{
		Name:      queryName,
		Statement: "SELECT 1",
	})
	if err != nil {
		t.Fatalf("setup CreateQuery: %v", err)
	}

	// Delete the query
	if err := c.DeleteQuery(ctx, query.URN); err != nil {
		t.Fatalf("DeleteQuery: %v", err)
	}
}
