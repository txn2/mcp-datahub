package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleUpdateDescription(t *testing.T) {
	var capturedURN, capturedDesc string
	mock := &mockClient{
		updateDescriptionFunc: func(_ context.Context, urn, description string) error {
			capturedURN = urn
			capturedDesc = description
			return nil
		},
	}

	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleUpdateDescription(context.Background(), nil, UpdateDescriptionInput{
		URN:         "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		Description: "Updated description",
	})

	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	if capturedURN != "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)" {
		t.Errorf("unexpected URN: %s", capturedURN)
	}
	if capturedDesc != "Updated description" {
		t.Errorf("unexpected description: %s", capturedDesc)
	}
}

func TestHandleUpdateDescription_EmptyURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleUpdateDescription(context.Background(), nil, UpdateDescriptionInput{
		Description: "desc",
	})

	if !result.IsError {
		t.Error("expected error for empty URN")
	}
}

func TestHandleUpdateDescription_WriteDisabled(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig()) // WriteEnabled defaults to false

	result, _, _ := toolkit.handleUpdateDescription(context.Background(), nil, UpdateDescriptionInput{
		URN:         "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		Description: "desc",
	})

	if !result.IsError {
		t.Error("expected error when write is disabled")
	}
}

func TestHandleUpdateDescription_ClientError(t *testing.T) {
	mock := &mockClient{
		updateDescriptionFunc: func(_ context.Context, _, _ string) error {
			return errors.New("api error")
		},
	}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleUpdateDescription(context.Background(), nil, UpdateDescriptionInput{
		URN:         "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		Description: "desc",
	})

	if !result.IsError {
		t.Error("expected error on client failure")
	}
}

func TestRegisterUpdateDescriptionTool(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.Register(server, ToolUpdateDescription)

	if !toolkit.registeredTools[ToolUpdateDescription] {
		t.Error("ToolUpdateDescription should be registered")
	}
}
