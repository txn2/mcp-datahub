package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleAddGlossaryTerm(t *testing.T) {
	var capturedURN, capturedTerm string
	mock := &mockClient{
		addGlossaryTermFunc: func(_ context.Context, urn, termURN string) error {
			capturedURN = urn
			capturedTerm = termURN
			return nil
		},
	}

	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddGlossaryTerm(context.Background(), nil, AddGlossaryTermInput{
		URN:     "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		TermURN: "urn:li:glossaryTerm:Classification",
	})

	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	if capturedURN != "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)" {
		t.Errorf("unexpected URN: %s", capturedURN)
	}
	if capturedTerm != "urn:li:glossaryTerm:Classification" {
		t.Errorf("unexpected term: %s", capturedTerm)
	}
}

func TestHandleAddGlossaryTerm_EmptyURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddGlossaryTerm(context.Background(), nil, AddGlossaryTermInput{
		TermURN: "urn:li:glossaryTerm:Classification",
	})

	if !result.IsError {
		t.Error("expected error for empty URN")
	}
}

func TestHandleAddGlossaryTerm_EmptyTermURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddGlossaryTerm(context.Background(), nil, AddGlossaryTermInput{
		URN: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
	})

	if !result.IsError {
		t.Error("expected error for empty term URN")
	}
}

func TestHandleAddGlossaryTerm_WriteDisabled(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, _ := toolkit.handleAddGlossaryTerm(context.Background(), nil, AddGlossaryTermInput{
		URN:     "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		TermURN: "urn:li:glossaryTerm:Classification",
	})

	if !result.IsError {
		t.Error("expected error when write is disabled")
	}
}

func TestHandleAddGlossaryTerm_ClientError(t *testing.T) {
	mock := &mockClient{
		addGlossaryTermFunc: func(_ context.Context, _, _ string) error {
			return errors.New("api error")
		},
	}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddGlossaryTerm(context.Background(), nil, AddGlossaryTermInput{
		URN:     "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		TermURN: "urn:li:glossaryTerm:Classification",
	})

	if !result.IsError {
		t.Error("expected error on client failure")
	}
}

func TestHandleRemoveGlossaryTerm(t *testing.T) {
	var capturedURN, capturedTerm string
	mock := &mockClient{
		removeGlossaryTermFunc: func(_ context.Context, urn, termURN string) error {
			capturedURN = urn
			capturedTerm = termURN
			return nil
		},
	}

	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleRemoveGlossaryTerm(context.Background(), nil, RemoveGlossaryTermInput{
		URN:     "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		TermURN: "urn:li:glossaryTerm:Classification",
	})

	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	if capturedURN != "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)" {
		t.Errorf("unexpected URN: %s", capturedURN)
	}
	if capturedTerm != "urn:li:glossaryTerm:Classification" {
		t.Errorf("unexpected term: %s", capturedTerm)
	}
}

func TestHandleRemoveGlossaryTerm_EmptyURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleRemoveGlossaryTerm(context.Background(), nil, RemoveGlossaryTermInput{
		TermURN: "urn:li:glossaryTerm:Classification",
	})

	if !result.IsError {
		t.Error("expected error for empty URN")
	}
}

func TestHandleRemoveGlossaryTerm_EmptyTermURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleRemoveGlossaryTerm(context.Background(), nil, RemoveGlossaryTermInput{
		URN: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
	})

	if !result.IsError {
		t.Error("expected error for empty term URN")
	}
}

func TestRegisterGlossaryTermTools(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.Register(server, ToolAddGlossaryTerm, ToolRemoveGlossaryTerm)

	if !toolkit.registeredTools[ToolAddGlossaryTerm] {
		t.Error("ToolAddGlossaryTerm should be registered")
	}
	if !toolkit.registeredTools[ToolRemoveGlossaryTerm] {
		t.Error("ToolRemoveGlossaryTerm should be registered")
	}
}
