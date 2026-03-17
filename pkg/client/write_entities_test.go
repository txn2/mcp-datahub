package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestCreateTag(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		desc     string
		response string
		wantURN  string
		wantErr  bool
	}{
		{
			name:     "success",
			tagName:  "PII",
			desc:     "Personally identifiable information",
			response: `{"data": {"createTag": "urn:li:tag:PII"}}`,
			wantURN:  "urn:li:tag:PII",
		},
		{
			name:     "graphql error",
			tagName:  "PII",
			desc:     "desc",
			response: `{"errors": [{"message": "duplicate tag"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["name"] != tt.tagName {
					t.Errorf("name = %v, want %v", input["name"], tt.tagName)
				}
				if input["description"] != tt.desc {
					t.Errorf("description = %v, want %v", input["description"], tt.desc)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			urn, err := c.CreateTag(context.Background(), tt.tagName, tt.desc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}
		})
	}
}

func TestCreateDomain(t *testing.T) {
	tests := []struct {
		name       string
		domainName string
		desc       string
		response   string
		wantURN    string
		wantErr    bool
	}{
		{
			name:       "success",
			domainName: "Engineering",
			desc:       "Engineering domain",
			response:   `{"data": {"createDomain": "urn:li:domain:Engineering"}}`,
			wantURN:    "urn:li:domain:Engineering",
		},
		{
			name:       "graphql error",
			domainName: "Engineering",
			desc:       "desc",
			response:   `{"errors": [{"message": "duplicate domain"}]}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["name"] != tt.domainName {
					t.Errorf("name = %v, want %v", input["name"], tt.domainName)
				}
				if input["description"] != tt.desc {
					t.Errorf("description = %v, want %v", input["description"], tt.desc)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			urn, err := c.CreateDomain(context.Background(), tt.domainName, tt.desc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}
		})
	}
}

func TestCreateGlossaryTerm(t *testing.T) {
	tests := []struct {
		name       string
		termName   string
		desc       string
		parentNode string
		response   string
		wantURN    string
		wantErr    bool
		checkInput func(t *testing.T, input map[string]any)
	}{
		{
			name:     "success without parent",
			termName: "Revenue",
			desc:     "Total revenue",
			response: `{"data": {"createGlossaryTerm": "urn:li:glossaryTerm:Revenue"}}`,
			wantURN:  "urn:li:glossaryTerm:Revenue",
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if _, hasParent := input["parentNode"]; hasParent {
					t.Error("expected no parentNode when empty")
				}
			},
		},
		{
			name:       "success with parent",
			termName:   "Revenue",
			desc:       "Total revenue",
			parentNode: "urn:li:glossaryNode:finance",
			response:   `{"data": {"createGlossaryTerm": "urn:li:glossaryTerm:Revenue"}}`,
			wantURN:    "urn:li:glossaryTerm:Revenue",
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["parentNode"] != "urn:li:glossaryNode:finance" {
					t.Errorf("parentNode = %v, want urn:li:glossaryNode:finance", input["parentNode"])
				}
			},
		},
		{
			name:     "graphql error",
			termName: "Revenue",
			desc:     "desc",
			response: `{"errors": [{"message": "creation failed"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["name"] != tt.termName {
					t.Errorf("name = %v, want %v", input["name"], tt.termName)
				}
				if tt.checkInput != nil {
					tt.checkInput(t, input)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			urn, err := c.CreateGlossaryTerm(context.Background(), tt.termName, tt.desc, tt.parentNode)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}
		})
	}
}

func TestCreateDataProduct(t *testing.T) {
	tests := []struct {
		name      string
		prodName  string
		desc      string
		domainURN string
		response  string
		wantURN   string
		wantErr   bool
	}{
		{
			name:      "success",
			prodName:  "Analytics",
			desc:      "Analytics data product",
			domainURN: "urn:li:domain:engineering",
			response:  `{"data": {"createDataProduct": {"urn": "urn:li:dataProduct:Analytics"}}}`,
			wantURN:   "urn:li:dataProduct:Analytics",
		},
		{
			name:      "empty domainURN returns error",
			prodName:  "Analytics",
			desc:      "Analytics data product",
			domainURN: "",
			wantErr:   true,
		},
		{
			name:      "graphql error",
			prodName:  "Analytics",
			desc:      "desc",
			domainURN: "urn:li:domain:eng",
			response:  `{"errors": [{"message": "failed"}]}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if input["domainUrn"] != tt.domainURN {
					t.Errorf("domainUrn = %v, want %v", input["domainUrn"], tt.domainURN)
				}
				props, _ := input["properties"].(map[string]any)
				if props["name"] != tt.prodName {
					t.Errorf("properties.name = %v, want %v", props["name"], tt.prodName)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			urn, err := c.CreateDataProduct(context.Background(), tt.prodName, tt.desc, tt.domainURN)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}
		})
	}
}

func TestCreateDocument(t *testing.T) {
	tests := []struct {
		name       string
		input      types.CreateDocumentInput
		response   string
		wantURN    string
		wantErr    bool
		checkInput func(t *testing.T, input map[string]any)
	}{
		{
			name: "success with all fields",
			input: types.CreateDocumentInput{
				Title:            "Design Doc",
				Content:          "Architecture overview",
				Status:           "PUBLISHED",
				SubType:          "DESIGN_DOC",
				RelatedAssetURNs: []string{"urn:li:dataset:ds1"},
				GlobalContext:    true,
			},
			response: `{"data": {"createDocument": "urn:li:document:design-doc"}}`,
			wantURN:  "urn:li:document:design-doc",
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["title"] != "Design Doc" {
					t.Errorf("title = %v", input["title"])
				}
				contents, _ := input["contents"].(map[string]any)
				if contents["text"] != "Architecture overview" {
					t.Errorf("contents.text = %v", contents["text"])
				}
				if input["state"] != "PUBLISHED" {
					t.Errorf("state = %v", input["state"])
				}
				if input["subType"] != "DESIGN_DOC" {
					t.Errorf("subType = %v", input["subType"])
				}
				assets, _ := input["relatedAssets"].([]any)
				if len(assets) != 1 || assets[0] != "urn:li:dataset:ds1" {
					t.Errorf("relatedAssets = %v", input["relatedAssets"])
				}
				settings, _ := input["settings"].(map[string]any)
				if settings["showInGlobalContext"] != true {
					t.Errorf("settings = %v", input["settings"])
				}
			},
		},
		{
			name: "success with minimal fields",
			input: types.CreateDocumentInput{
				Title:   "Simple Doc",
				Content: "Body text",
			},
			response: `{"data": {"createDocument": "urn:li:document:simple"}}`,
			wantURN:  "urn:li:document:simple",
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if _, hasState := input["state"]; hasState {
					t.Error("expected no state when empty")
				}
				if _, hasSubType := input["subType"]; hasSubType {
					t.Error("expected no subType when empty")
				}
				if _, hasAssets := input["relatedAssets"]; hasAssets {
					t.Error("expected no relatedAssets when empty")
				}
				if _, hasSettings := input["settings"]; hasSettings {
					t.Error("expected no settings when GlobalContext is false")
				}
			},
		},
		{
			name: "empty title returns error",
			input: types.CreateDocumentInput{
				Content: "Body text",
			},
			wantErr: true,
		},
		{
			name: "graphql error",
			input: types.CreateDocumentInput{
				Title:   "Fails",
				Content: "Body",
			},
			response: `{"errors": [{"message": "creation failed"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if tt.checkInput != nil {
					tt.checkInput(t, input)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			urn, err := c.CreateDocument(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}
		})
	}
}

func TestCreateApplication(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		desc     string
		response string
		wantURN  string
		wantErr  bool
	}{
		{
			name:     "success",
			appName:  "MyApp",
			desc:     "My application",
			response: `{"data": {"createApplication": {"urn": "urn:li:application:MyApp"}}}`,
			wantURN:  "urn:li:application:MyApp",
		},
		{
			name:     "graphql error",
			appName:  "MyApp",
			desc:     "desc",
			response: `{"errors": [{"message": "failed"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				props, _ := input["properties"].(map[string]any)
				if props["name"] != tt.appName {
					t.Errorf("properties.name = %v, want %v", props["name"], tt.appName)
				}
				if props["description"] != tt.desc {
					t.Errorf("properties.description = %v, want %v", props["description"], tt.desc)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			urn, err := c.CreateApplication(context.Background(), tt.appName, tt.desc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}
		})
	}
}

func TestCreateStructuredProperty(t *testing.T) {
	tests := []struct {
		name       string
		input      types.CreateStructuredPropertyInput
		response   string
		wantURN    string
		wantErr    bool
		checkInput func(t *testing.T, input map[string]any)
	}{
		{
			name: "success with all fields",
			input: types.CreateStructuredPropertyInput{
				QualifiedName: "io.acryl.privacy.retentionTime",
				DisplayName:   "Retention Time",
				Description:   "How long data is retained",
				ValueType:     "string",
				Cardinality:   "SINGLE",
				EntityTypes:   []string{"dataset", "dataFlow"},
				AllowedValues: []types.AllowedValue{
					{Value: "30d", Description: "30 days"},
					{Value: "90d", Description: "90 days"},
				},
			},
			response: `{"data": {"createStructuredProperty": {"urn": "urn:li:structuredProperty:io.acryl.privacy.retentionTime"}}}`,
			wantURN:  "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["qualifiedName"] != "io.acryl.privacy.retentionTime" {
					t.Errorf("qualifiedName = %v", input["qualifiedName"])
				}
				if input["valueType"] != "string" {
					t.Errorf("valueType = %v", input["valueType"])
				}
				if input["displayName"] != "Retention Time" {
					t.Errorf("displayName = %v", input["displayName"])
				}
				if input["cardinality"] != "SINGLE" {
					t.Errorf("cardinality = %v", input["cardinality"])
				}
				// Verify AllowedValues mapping uses flat structure
				avList, _ := input["allowedValues"].([]any)
				if len(avList) != 2 {
					t.Fatalf("allowedValues length = %d, want 2", len(avList))
				}
				av0, _ := avList[0].(map[string]any)
				if av0["stringValue"] != "30d" {
					t.Errorf("allowedValues[0].stringValue = %v, want 30d", av0["stringValue"])
				}
				if av0["description"] != "30 days" {
					t.Errorf("allowedValues[0].description = %v, want 30 days", av0["description"])
				}
			},
		},
		{
			name: "missing qualifiedName returns error",
			input: types.CreateStructuredPropertyInput{
				ValueType: "string",
			},
			wantErr: true,
		},
		{
			name: "missing valueType returns error",
			input: types.CreateStructuredPropertyInput{
				QualifiedName: "io.acryl.test",
				EntityTypes:   []string{"dataset"},
			},
			wantErr: true,
		},
		{
			name: "missing entityTypes returns error",
			input: types.CreateStructuredPropertyInput{
				QualifiedName: "io.acryl.test",
				ValueType:     "string",
			},
			wantErr: true,
		},
		{
			name: "graphql error",
			input: types.CreateStructuredPropertyInput{
				QualifiedName: "io.acryl.test",
				ValueType:     "string",
				EntityTypes:   []string{"dataset"},
			},
			response: `{"errors": [{"message": "creation failed"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if tt.checkInput != nil {
					tt.checkInput(t, input)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			urn, err := c.CreateStructuredProperty(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}
		})
	}
}

func TestUpsertDataContract(t *testing.T) {
	tests := []struct {
		name       string
		input      types.UpsertDataContractInput
		response   string
		wantURN    string
		wantErr    bool
		checkInput func(t *testing.T, input map[string]any)
	}{
		{
			name: "success with all assertion types",
			input: types.UpsertDataContractInput{
				DatasetURN:               "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
				SchemaAssertionURNs:      []string{"urn:li:assertion:schema1"},
				FreshnessAssertionURNs:   []string{"urn:li:assertion:fresh1"},
				DataQualityAssertionURNs: []string{"urn:li:assertion:dq1", "urn:li:assertion:dq2"},
			},
			response: `{"data": {"upsertDataContract": {"urn": "urn:li:dataContract:contract1"}}}`,
			wantURN:  "urn:li:dataContract:contract1",
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["entityUrn"] != "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)" {
					t.Errorf("entityUrn = %v", input["entityUrn"])
				}
				schema, _ := input["schema"].([]any)
				if len(schema) != 1 {
					t.Fatalf("schema length = %d, want 1", len(schema))
				}
				s0, _ := schema[0].(map[string]any)
				if s0["assertionUrn"] != "urn:li:assertion:schema1" {
					t.Errorf("schema[0].assertionUrn = %v", s0["assertionUrn"])
				}
				dq, _ := input["dataQuality"].([]any)
				if len(dq) != 2 {
					t.Fatalf("dataQuality length = %d, want 2", len(dq))
				}
			},
		},
		{
			name: "success with only dataset URN",
			input: types.UpsertDataContractInput{
				DatasetURN: "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			},
			response: `{"data": {"upsertDataContract": {"urn": "urn:li:dataContract:contract2"}}}`,
			wantURN:  "urn:li:dataContract:contract2",
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if _, has := input["schema"]; has {
					t.Error("expected no schema when empty")
				}
				if _, has := input["freshness"]; has {
					t.Error("expected no freshness when empty")
				}
				if _, has := input["dataQuality"]; has {
					t.Error("expected no dataQuality when empty")
				}
			},
		},
		{
			name: "empty datasetURN returns error",
			input: types.UpsertDataContractInput{
				DatasetURN: "",
			},
			wantErr: true,
		},
		{
			name: "graphql error",
			input: types.UpsertDataContractInput{
				DatasetURN: "urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
			},
			response: `{"errors": [{"message": "upsert failed"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				input := extractGraphQLInput(t, r)
				if tt.checkInput != nil {
					tt.checkInput(t, input)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			urn, err := c.UpsertDataContract(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}
		})
	}
}
