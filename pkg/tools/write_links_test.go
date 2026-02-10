package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleAddLink(t *testing.T) {
	var capturedURN, capturedURL, capturedDesc string
	mock := &mockClient{
		addLinkFunc: func(_ context.Context, urn, linkURL, description string) error {
			capturedURN = urn
			capturedURL = linkURL
			capturedDesc = description
			return nil
		},
	}

	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddLink(context.Background(), nil, AddLinkInput{
		URN:         "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		URL:         "https://docs.example.com",
		Description: "Documentation",
	})

	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	if capturedURN != "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)" {
		t.Errorf("unexpected URN: %s", capturedURN)
	}
	if capturedURL != "https://docs.example.com" {
		t.Errorf("unexpected URL: %s", capturedURL)
	}
	if capturedDesc != "Documentation" {
		t.Errorf("unexpected description: %s", capturedDesc)
	}
}

func TestHandleAddLink_EmptyURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddLink(context.Background(), nil, AddLinkInput{
		URL: "https://docs.example.com",
	})

	if !result.IsError {
		t.Error("expected error for empty URN")
	}
}

func TestHandleAddLink_EmptyURL(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddLink(context.Background(), nil, AddLinkInput{
		URN: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
	})

	if !result.IsError {
		t.Error("expected error for empty URL")
	}
}

func TestHandleAddLink_WriteDisabled(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, _ := toolkit.handleAddLink(context.Background(), nil, AddLinkInput{
		URN: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		URL: "https://docs.example.com",
	})

	if !result.IsError {
		t.Error("expected error when write is disabled")
	}
}

func TestHandleAddLink_ClientError(t *testing.T) {
	mock := &mockClient{
		addLinkFunc: func(_ context.Context, _, _, _ string) error {
			return errors.New("api error")
		},
	}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleAddLink(context.Background(), nil, AddLinkInput{
		URN: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		URL: "https://docs.example.com",
	})

	if !result.IsError {
		t.Error("expected error on client failure")
	}
}

func TestHandleRemoveLink(t *testing.T) {
	var capturedURN, capturedURL string
	mock := &mockClient{
		removeLinkFunc: func(_ context.Context, urn, linkURL string) error {
			capturedURN = urn
			capturedURL = linkURL
			return nil
		},
	}

	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleRemoveLink(context.Background(), nil, RemoveLinkInput{
		URN: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
		URL: "https://docs.example.com",
	})

	if result.IsError {
		t.Errorf("expected success, got error: %v", result.Content)
	}
	if capturedURN != "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)" {
		t.Errorf("unexpected URN: %s", capturedURN)
	}
	if capturedURL != "https://docs.example.com" {
		t.Errorf("unexpected URL: %s", capturedURL)
	}
}

func TestHandleRemoveLink_EmptyURN(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleRemoveLink(context.Background(), nil, RemoveLinkInput{
		URL: "https://docs.example.com",
	})

	if !result.IsError {
		t.Error("expected error for empty URN")
	}
}

func TestHandleRemoveLink_EmptyURL(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleRemoveLink(context.Background(), nil, RemoveLinkInput{
		URN: "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
	})

	if !result.IsError {
		t.Error("expected error for empty URL")
	}
}

func TestRegisterLinkTools(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.Register(server, ToolAddLink, ToolRemoveLink)

	if !toolkit.registeredTools[ToolAddLink] {
		t.Error("ToolAddLink should be registered")
	}
	if !toolkit.registeredTools[ToolRemoveLink] {
		t.Error("ToolRemoveLink should be registered")
	}
}
