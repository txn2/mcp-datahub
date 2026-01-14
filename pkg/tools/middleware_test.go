package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const testKey contextKey = "test"

func TestBeforeFunc(t *testing.T) {
	called := false
	var receivedTC *ToolContext

	beforeFn := BeforeFunc(func(ctx context.Context, tc *ToolContext) (context.Context, error) {
		called = true
		receivedTC = tc
		return context.WithValue(ctx, testKey, "value"), nil
	})

	tc := NewToolContext(ToolSearch, nil)
	ctx := context.Background()

	newCtx, err := beforeFn.Before(ctx, tc)

	if !called {
		t.Error("Before() should call the function")
	}
	if receivedTC != tc {
		t.Error("Before() should pass ToolContext")
	}
	if err != nil {
		t.Errorf("Before() unexpected error: %v", err)
	}
	if newCtx.Value(testKey) != "value" {
		t.Error("Before() should return modified context")
	}
}

func TestBeforeFuncAfterNoOp(t *testing.T) {
	beforeFn := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		return ctx, nil
	})

	tc := NewToolContext(ToolSearch, nil)
	result := TextResult("test")

	returnedResult, err := beforeFn.After(context.Background(), tc, result, nil)

	if err != nil {
		t.Errorf("After() unexpected error: %v", err)
	}
	if returnedResult != result {
		t.Error("After() should return the same result")
	}
}

func TestAfterFunc(t *testing.T) {
	called := false
	var receivedTC *ToolContext
	var receivedResult *mcp.CallToolResult
	var receivedErr error

	afterFn := AfterFunc(func(ctx context.Context, tc *ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
		called = true
		receivedTC = tc
		receivedResult = result
		receivedErr = err
		return TextResult("modified"), nil
	})

	tc := NewToolContext(ToolSearch, nil)
	originalResult := TextResult("original")
	originalErr := errors.New("test error")

	modifiedResult, err := afterFn.After(context.Background(), tc, originalResult, originalErr)

	if !called {
		t.Error("After() should call the function")
	}
	if receivedTC != tc {
		t.Error("After() should pass ToolContext")
	}
	if receivedResult != originalResult {
		t.Error("After() should pass original result")
	}
	if receivedErr != originalErr {
		t.Error("After() should pass original error")
	}
	if err != nil {
		t.Errorf("After() unexpected error: %v", err)
	}
	if modifiedResult == originalResult {
		t.Error("After() should return modified result")
	}
}

func TestAfterFuncBeforeNoOp(t *testing.T) {
	afterFn := AfterFunc(func(_ context.Context, _ *ToolContext, result *mcp.CallToolResult, _ error) (*mcp.CallToolResult, error) {
		return result, nil
	})

	tc := NewToolContext(ToolSearch, nil)
	ctx := context.Background()

	newCtx, err := afterFn.Before(ctx, tc)

	if err != nil {
		t.Errorf("Before() unexpected error: %v", err)
	}
	if newCtx != ctx {
		t.Error("Before() should return the same context")
	}
}

func TestBeforeFuncError(t *testing.T) {
	expectedErr := errors.New("before error")

	beforeFn := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		return ctx, expectedErr
	})

	tc := NewToolContext(ToolSearch, nil)
	_, err := beforeFn.Before(context.Background(), tc)

	if err != expectedErr {
		t.Errorf("Before() error = %v, want %v", err, expectedErr)
	}
}

func TestAfterFuncError(t *testing.T) {
	expectedErr := errors.New("after error")

	afterFn := AfterFunc(func(_ context.Context, _ *ToolContext, _ *mcp.CallToolResult, _ error) (*mcp.CallToolResult, error) {
		return nil, expectedErr
	})

	tc := NewToolContext(ToolSearch, nil)
	_, err := afterFn.After(context.Background(), tc, nil, nil)

	if err != expectedErr {
		t.Errorf("After() error = %v, want %v", err, expectedErr)
	}
}

// TestToolMiddlewareInterface verifies that BeforeFunc and AfterFunc implement ToolMiddleware.
func TestToolMiddlewareInterface(t *testing.T) {
	var _ ToolMiddleware = BeforeFunc(nil)
	var _ ToolMiddleware = AfterFunc(nil)
}
