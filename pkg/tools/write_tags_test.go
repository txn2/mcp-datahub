package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleAddTag(t *testing.T) {
	var capturedURN, capturedTag string
	mock := &mockClient{
		addTagFunc: func(_ context.Context, urn, tagURN string) error {
			capturedURN = urn
			capturedTag = tagURN
			return nil
		},
	}

	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddTag(context.Background(), nil, AddTagInput{
		URN:    "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		TagURN: "urn:li:tag:PII",
	})

	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	if capturedURN != "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)" {
		t.Errorf("unexpected URN: %s", capturedURN)
	}
	if capturedTag != "urn:li:tag:PII" {
		t.Errorf("unexpected tag: %s", capturedTag)
	}
}

func TestHandleAddTag_EmptyURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddTag(context.Background(), nil, AddTagInput{
		TagURN: "urn:li:tag:PII",
	})

	if !result.IsError {
		t.Error("expected error for empty URN")
	}
}

func TestHandleAddTag_EmptyTagURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddTag(context.Background(), nil, AddTagInput{
		URN: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
	})

	if !result.IsError {
		t.Error("expected error for empty tag URN")
	}
}

func TestHandleAddTag_WriteDisabled(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, _ := toolkit.handleAddTag(context.Background(), nil, AddTagInput{
		URN:    "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		TagURN: "urn:li:tag:PII",
	})

	if !result.IsError {
		t.Error("expected error when write is disabled")
	}
}

func TestHandleAddTag_ClientError(t *testing.T) {
	mock := &mockClient{
		addTagFunc: func(_ context.Context, _, _ string) error {
			return errors.New("api error")
		},
	}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddTag(context.Background(), nil, AddTagInput{
		URN:    "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		TagURN: "urn:li:tag:PII",
	})

	if !result.IsError {
		t.Error("expected error on client failure")
	}
}

func TestHandleRemoveTag(t *testing.T) {
	var capturedURN, capturedTag string
	mock := &mockClient{
		removeTagFunc: func(_ context.Context, urn, tagURN string) error {
			capturedURN = urn
			capturedTag = tagURN
			return nil
		},
	}

	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleRemoveTag(context.Background(), nil, RemoveTagInput{
		URN:    "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		TagURN: "urn:li:tag:PII",
	})

	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	if capturedURN != "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)" {
		t.Errorf("unexpected URN: %s", capturedURN)
	}
	if capturedTag != "urn:li:tag:PII" {
		t.Errorf("unexpected tag: %s", capturedTag)
	}
}

func TestHandleRemoveTag_EmptyURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleRemoveTag(context.Background(), nil, RemoveTagInput{
		TagURN: "urn:li:tag:PII",
	})

	if !result.IsError {
		t.Error("expected error for empty URN")
	}
}

func TestHandleRemoveTag_EmptyTagURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleRemoveTag(context.Background(), nil, RemoveTagInput{
		URN: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
	})

	if !result.IsError {
		t.Error("expected error for empty tag URN")
	}
}

func TestRegisterTagTools(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.Register(server, ToolAddTag, ToolRemoveTag)

	if !toolkit.registeredTools[ToolAddTag] {
		t.Error("ToolAddTag should be registered")
	}
	if !toolkit.registeredTools[ToolRemoveTag] {
		t.Error("ToolRemoveTag should be registered")
	}
}
