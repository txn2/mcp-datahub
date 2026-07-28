package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const descTestURN = "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)"

// resultText returns the text of a tool result's first content block.
func resultText(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return ""
	}
	return text.Text
}

// recordingDescClient captures the description text that reaches the client, so a
// test can tell an empty write apart from no write at all.
type recordingDescClient struct {
	mockClient
	called bool
	got    string
}

func newRecordingDescClient() *recordingDescClient {
	rec := &recordingDescClient{}
	rec.updateDescriptionFunc = func(_ context.Context, _, description string) error {
		rec.called = true
		rec.got = description
		return nil
	}
	rec.updateColumnDescriptionFunc = func(_ context.Context, _, _, description string) error {
		rec.called = true
		rec.got = description
		return nil
	}
	return rec
}

// TestHandleUpdate_DescriptionFromDescriptionField covers the text arriving in
// `description` rather than `value`. `description` is a real field for other what
// values, so the schema accepts it here; reading only `value` wrote an empty
// description - erasing any description the target carried - and still answered
// "updated" (#194).
func TestHandleUpdate_DescriptionFromDescriptionField(t *testing.T) {
	tests := []struct {
		name  string
		input UpdateInput
	}{
		{
			"entity description",
			UpdateInput{What: "description", URN: descTestURN, Description: "Gross margin before returns"},
		},
		{
			"column description",
			UpdateInput{
				What: "column_description", URN: descTestURN, FieldPath: "amount",
				Description: "Gross margin before returns",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecordingDescClient()
			toolkit := NewToolkit(rec, Config{WriteEnabled: true})

			result, _, _ := toolkit.handleUpdate(context.Background(), nil, tt.input)

			if result.IsError {
				t.Fatalf("expected success, got error result")
			}
			if !rec.called {
				t.Fatal("client was never called")
			}
			if rec.got != "Gross margin before returns" {
				t.Errorf("client received %q, want the description text", rec.got)
			}
		})
	}
}

// TestHandleUpdate_DescriptionValueWins keeps `value` authoritative when both
// fields carry the same text.
func TestHandleUpdate_DescriptionValueWins(t *testing.T) {
	rec := newRecordingDescClient()
	toolkit := NewToolkit(rec, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{
		What: "description", URN: descTestURN, Value: "Same text", Description: "Same text",
	})

	if result.IsError {
		t.Fatalf("expected success, got error result")
	}
	if rec.got != "Same text" {
		t.Errorf("client received %q, want %q", rec.got, "Same text")
	}
}

// TestHandleUpdate_DescriptionConflict rejects a genuine disagreement rather than
// silently picking one of the two.
func TestHandleUpdate_DescriptionConflict(t *testing.T) {
	rec := newRecordingDescClient()
	toolkit := NewToolkit(rec, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{
		What: "description", URN: descTestURN, Value: "One text", Description: "Other text",
	})

	if !result.IsError {
		t.Fatal("expected an error result when value and description differ")
	}
	if rec.called {
		t.Error("client must not be called when the two fields conflict")
	}
}

// TestHandleUpdate_DescriptionRequiresText refuses a description update carrying
// no text at all, instead of writing an empty description and reporting "updated".
func TestHandleUpdate_DescriptionRequiresText(t *testing.T) {
	tests := []struct {
		name  string
		input UpdateInput
	}{
		{"entity description", UpdateInput{What: "description", URN: descTestURN}},
		{"column description", UpdateInput{What: "column_description", URN: descTestURN, FieldPath: "amount"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecordingDescClient()
			toolkit := NewToolkit(rec, Config{WriteEnabled: true})

			result, _, _ := toolkit.handleUpdate(context.Background(), nil, tt.input)

			if !result.IsError {
				t.Fatal("expected an error result when no description text is given")
			}
			if rec.called {
				t.Error("client must not be called when no description text is given")
			}
			if text := resultText(result); !strings.Contains(text, "value") {
				t.Errorf("error should name the value parameter, got %q", text)
			}
		})
	}
}

// TestHandleUpdate_ColumnDescriptionRequiresFieldPathFirst keeps the missing
// field_path error ahead of the description check, so the caller is told the more
// specific problem.
func TestHandleUpdate_ColumnDescriptionRequiresFieldPathFirst(t *testing.T) {
	rec := newRecordingDescClient()
	toolkit := NewToolkit(rec, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{
		What: "column_description", URN: descTestURN, Description: "text",
	})

	if !result.IsError {
		t.Fatal("expected an error result when field_path is missing")
	}
	if text := resultText(result); !strings.Contains(text, "field_path") {
		t.Errorf("error should name field_path, got %q", text)
	}
}
