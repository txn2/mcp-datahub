package integration

import (
	"context"
	"testing"
)

func TestNoOpQueryProvider(t *testing.T) {
	p := &NoOpQueryProvider{}

	if p.Name() != "noop" {
		t.Errorf("Name() = %s, want noop", p.Name())
	}

	table, err := p.ResolveTable(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Errorf("ResolveTable error: %v", err)
	}
	if table != nil {
		t.Error("ResolveTable should return nil")
	}

	avail, err := p.GetTableAvailability(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Errorf("GetTableAvailability error: %v", err)
	}
	if avail != nil {
		t.Error("GetTableAvailability should return nil")
	}

	examples, err := p.GetQueryExamples(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Errorf("GetQueryExamples error: %v", err)
	}
	if examples != nil {
		t.Error("GetQueryExamples should return nil")
	}

	execCtx, err := p.GetExecutionContext(context.Background(), []string{"urn:li:dataset:test"})
	if err != nil {
		t.Errorf("GetExecutionContext error: %v", err)
	}
	if execCtx != nil {
		t.Error("GetExecutionContext should return nil")
	}

	if err := p.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}

func TestNoOpQueryProvider_ImplementsInterface(t *testing.T) {
	var _ QueryProvider = (*NoOpQueryProvider)(nil)
}

func TestQueryProviderFunc_AllMethods(t *testing.T) {
	ctx := context.Background()
	testURN := "urn:li:dataset:test"

	p := &QueryProviderFunc{
		NameFn: func() string { return "custom" },
		ResolveTableFn: func(_ context.Context, urn string) (*TableIdentifier, error) {
			return &TableIdentifier{Catalog: "cat", Schema: "sch", Table: urn}, nil
		},
		GetTableAvailabilityFn: func(_ context.Context, _ string) (*TableAvailability, error) {
			return &TableAvailability{Available: true}, nil
		},
		GetQueryExamplesFn: func(_ context.Context, _ string) ([]QueryExample, error) {
			return []QueryExample{{Name: "test", SQL: "SELECT 1", Category: "sample"}}, nil
		},
		GetExecutionContextFn: func(_ context.Context, urns []string) (*ExecutionContext, error) {
			return &ExecutionContext{Source: "test", Tables: map[string]*TableIdentifier{}}, nil
		},
		CloseFn: func() error { return nil },
	}

	if p.Name() != "custom" {
		t.Errorf("Name() = %s, want custom", p.Name())
	}

	table, err := p.ResolveTable(ctx, testURN)
	if err != nil {
		t.Errorf("ResolveTable error: %v", err)
	}
	if table == nil || table.Table != testURN {
		t.Error("ResolveTable returned wrong result")
	}

	avail, err := p.GetTableAvailability(ctx, testURN)
	if err != nil {
		t.Errorf("GetTableAvailability error: %v", err)
	}
	if avail == nil || !avail.Available {
		t.Error("GetTableAvailability returned wrong result")
	}

	examples, err := p.GetQueryExamples(ctx, testURN)
	if err != nil {
		t.Errorf("GetQueryExamples error: %v", err)
	}
	if len(examples) != 1 || examples[0].Name != "test" {
		t.Error("GetQueryExamples returned wrong result")
	}

	execCtx, err := p.GetExecutionContext(ctx, []string{testURN})
	if err != nil {
		t.Errorf("GetExecutionContext error: %v", err)
	}
	if execCtx == nil || execCtx.Source != "test" {
		t.Error("GetExecutionContext returned wrong result")
	}

	if err := p.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}

func TestQueryProviderFunc_NilFunctions(t *testing.T) {
	ctx := context.Background()
	p := &QueryProviderFunc{}

	if p.Name() != "func" {
		t.Errorf("Name() = %s, want func (default)", p.Name())
	}

	table, err := p.ResolveTable(ctx, "urn")
	if err != nil || table != nil {
		t.Error("ResolveTable should return nil, nil for nil function")
	}

	avail, err := p.GetTableAvailability(ctx, "urn")
	if err != nil || avail != nil {
		t.Error("GetTableAvailability should return nil, nil for nil function")
	}

	examples, err := p.GetQueryExamples(ctx, "urn")
	if err != nil || examples != nil {
		t.Error("GetQueryExamples should return nil, nil for nil function")
	}

	execCtx, err := p.GetExecutionContext(ctx, []string{"urn"})
	if err != nil || execCtx != nil {
		t.Error("GetExecutionContext should return nil, nil for nil function")
	}

	if err := p.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}
}

func TestQueryProviderFunc_ImplementsInterface(t *testing.T) {
	var _ QueryProvider = (*QueryProviderFunc)(nil)
}

func TestTableIdentifier_String(t *testing.T) {
	tests := []struct {
		name string
		ti   TableIdentifier
		want string
	}{
		{
			name: "basic",
			ti:   TableIdentifier{Catalog: "catalog", Schema: "schema", Table: "table"},
			want: "catalog.schema.table",
		},
		{
			name: "with connection",
			ti:   TableIdentifier{Connection: "trino", Catalog: "hive", Schema: "default", Table: "users"},
			want: "trino:hive.default.users",
		},
		{
			name: "empty connection",
			ti:   TableIdentifier{Connection: "", Catalog: "cat", Schema: "sch", Table: "tbl"},
			want: "cat.sch.tbl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ti.String(); got != tt.want {
				t.Errorf("TableIdentifier.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTableAvailability_Fields(t *testing.T) {
	rowCount := int64(1000)
	avail := &TableAvailability{
		Available:  true,
		Table:      &TableIdentifier{Catalog: "cat", Schema: "sch", Table: "tbl"},
		Connection: "trino",
		RowCount:   &rowCount,
	}

	if !avail.Available {
		t.Error("Available should be true")
	}
	if avail.Table == nil {
		t.Error("Table should not be nil")
	}
	if avail.Connection != "trino" {
		t.Errorf("Connection = %q, want trino", avail.Connection)
	}
	if avail.RowCount == nil || *avail.RowCount != 1000 {
		t.Error("RowCount should be 1000")
	}
}

func TestQueryExample_Fields(t *testing.T) {
	example := QueryExample{
		Name:        "Sample Query",
		Description: "Get sample data",
		SQL:         "SELECT * FROM users LIMIT 10",
		Category:    "sample",
	}

	if example.Name != "Sample Query" {
		t.Errorf("Name = %q, want Sample Query", example.Name)
	}
	if example.Description != "Get sample data" {
		t.Errorf("Description = %q, want Get sample data", example.Description)
	}
	if example.SQL != "SELECT * FROM users LIMIT 10" {
		t.Errorf("SQL = %q, want SELECT * FROM users LIMIT 10", example.SQL)
	}
	if example.Category != "sample" {
		t.Errorf("Category = %q, want sample", example.Category)
	}
}

func TestExecutionContext_Fields(t *testing.T) {
	execCtx := &ExecutionContext{
		Tables: map[string]*TableIdentifier{
			"urn:1": {Catalog: "cat", Schema: "sch", Table: "t1"},
		},
		Source:      "trino-prod",
		Connections: []string{"trino"},
		Queries: []ExecutionQuery{
			{SQL: "SELECT * FROM t1 JOIN t2", Sources: []string{"urn:1"}},
		},
	}

	if len(execCtx.Tables) != 1 {
		t.Errorf("Tables count = %d, want 1", len(execCtx.Tables))
	}
	if execCtx.Source != "trino-prod" {
		t.Errorf("Source = %q, want trino-prod", execCtx.Source)
	}
	if len(execCtx.Queries) != 1 {
		t.Errorf("Queries count = %d, want 1", len(execCtx.Queries))
	}
	if len(execCtx.Connections) != 1 {
		t.Errorf("Connections count = %d, want 1", len(execCtx.Connections))
	}
}

func TestExecutionQuery_Fields(t *testing.T) {
	query := ExecutionQuery{
		SQL:     "SELECT * FROM users",
		Sources: []string{"urn:1"},
		Targets: []string{"urn:2"},
		QueryID: "query-123",
	}

	if query.SQL != "SELECT * FROM users" {
		t.Errorf("SQL = %q, want SELECT * FROM users", query.SQL)
	}
	if len(query.Sources) != 1 {
		t.Errorf("Sources count = %d, want 1", len(query.Sources))
	}
	if len(query.Targets) != 1 {
		t.Errorf("Targets count = %d, want 1", len(query.Targets))
	}
	if query.QueryID != "query-123" {
		t.Errorf("QueryID = %q, want query-123", query.QueryID)
	}
}
