package tools

import "encoding/json"

// defaultOutputSchemas holds the default JSON Schema output descriptors for each built-in tool.
// These declare the structure of the JSON objects returned by each tool to MCP clients.
// Schemas are top-level objects; not exhaustive — they describe the primary response shape.
var defaultOutputSchemas = map[ToolName]json.RawMessage{
	ToolSearch:          schemaSearch,
	ToolGetEntity:       schemaGetEntity,
	ToolGetSchema:       schemaGetSchema,
	ToolGetLineage:      schemaGetLineage,
	ToolGetQueries:      schemaGetQueries,
	ToolBrowse:          schemaBrowse,
	ToolGetGlossaryTerm: schemaGetGlossaryTerm,
	ToolGetDataProduct:  schemaGetDataProduct,
	ToolListConnections: schemaListConnections,

	// Write tools (CRUD pattern)
	ToolCreate: schemaCreate,
	ToolUpdate: schemaUpdate,
	ToolDelete: schemaDelete,
}

// DefaultOutputSchema returns the default output JSON Schema for a tool.
// Returns nil if the tool name is not recognized.
func DefaultOutputSchema(name ToolName) json.RawMessage {
	return defaultOutputSchemas[name]
}

// getOutputSchema resolves the output schema for a tool using the priority chain:
//  1. Per-registration cfg.outputSchema (highest)
//  2. Toolkit-level t.outputSchemas map
//  3. defaultOutputSchemas map (lowest/default)
func (t *Toolkit) getOutputSchema(name ToolName, cfg *toolConfig) any {
	// Highest priority: per-registration override
	if cfg != nil && cfg.outputSchema != nil {
		return cfg.outputSchema
	}

	// Middle priority: toolkit-level override
	if schema, ok := t.outputSchemas[name]; ok {
		return schema
	}

	// Lowest priority: default
	return defaultOutputSchemas[name]
}

// Individual output schema definitions for each tool.
// Keeping them as package-level variables avoids an oversized init() function.

var schemaSearch = json.RawMessage(`{
  "type": "object",
  "properties": {
    "total":    {"type": "integer", "description": "Total number of matching entities"},
    "entities": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "urn":         {"type": "string"},
          "name":        {"type": "string"},
          "type":        {"type": "string"},
          "description": {"type": "string"},
          "platform":    {"type": "string"}
        }
      }
    },
    "query_context": {
      "type": "object",
      "description": "Optional: query engine availability per entity URN",
      "additionalProperties": {
        "type": "object",
        "properties": {
          "available": {"type": "boolean"},
          "table":     {"type": "string"}
        }
      }
    }
  }
}`)

var schemaGetEntity = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":         {"type": "string"},
    "name":        {"type": "string"},
    "type":        {"type": "string"},
    "description": {"type": "string"},
    "platform":    {"type": "string"},
    "owners": {
      "type": ["array", "null"],
      "items": {
        "type": "object",
        "properties": {
          "urn":  {"type": "string"},
          "name": {"type": "string"},
          "type": {"type": "string"}
        }
      }
    },
    "tags": {
      "type": ["array", "null"],
      "items": {
        "type": "object",
        "properties": {
          "urn":         {"type": "string"},
          "name":        {"type": "string"},
          "description": {"type": "string"}
        }
      }
    },
    "domain": {
      "type": ["object", "null"],
      "properties": {
        "urn":         {"type": "string"},
        "name":        {"type": "string"},
        "description": {"type": "string"}
      }
    },
    "deprecation": {
      "type": ["object", "null"],
      "properties": {
        "deprecated":        {"type": "boolean"},
        "note":              {"type": "string"},
        "actor":             {"type": "string"},
        "decommission_time": {"type": "integer"}
      }
    },
    "query_table":        {"type": "string", "description": "Optional: fully-qualified query engine table path"},
    "query_availability": {"type": "object", "description": "Optional: query engine availability details"},
    "query_examples":     {"type": "array",  "description": "Optional: example SQL queries"}
  }
}`)

var schemaGetSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":   {"type": "string"},
    "fields": {
      "type": ["array", "null"],
      "items": {
        "type": "object",
        "properties": {
          "fieldPath":   {"type": "string"},
          "type":        {"type": "string"},
          "description": {"type": "string"},
          "nullable":    {"type": "boolean"}
        }
      }
    },
    "query_table": {"type": "string", "description": "Optional: resolved query engine table path"}
  }
}`)

var schemaGetLineage = json.RawMessage(`{
  "type": "object",
  "description": "Dataset-level (default): start/direction/depth/nodes/edges. Column-level (level=column): urn/columns.",
  "properties": {
    "start":     {"type": "string", "description": "URN of the queried entity (dataset level)"},
    "direction": {"type": "string", "description": "Lineage direction: UPSTREAM or DOWNSTREAM (dataset level)"},
    "depth":     {"type": "integer", "description": "Depth of lineage traversal (dataset level)"},
    "nodes": {
      "type": ["array", "null"],
      "description": "Lineage nodes (dataset level)",
      "items": {
        "type": "object",
        "properties": {
          "urn":      {"type": "string"},
          "name":     {"type": "string"},
          "type":     {"type": "string"},
          "platform": {"type": "string"},
          "level":    {"type": "integer"}
        }
      }
    },
    "edges": {
      "type": ["array", "null"],
      "description": "Lineage edges (dataset level)",
      "items": {
        "type": "object",
        "properties": {
          "source": {"type": "string"},
          "target": {"type": "string"},
          "type":   {"type": "string"}
        }
      }
    },
    "execution_context": {
      "type": "object",
      "description": "Optional: query engine execution context for lineage bridging (dataset level)"
    },
    "urn": {"type": "string", "description": "Dataset URN (column level)"},
    "columns": {
      "type": "array",
      "description": "Column-level lineage mappings (column level)",
      "items": {
        "type": "object",
        "properties": {
          "downstreamColumn": {"type": "string"},
          "upstreamColumns": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "datasetUrn": {"type": "string"},
                "column":     {"type": "string"}
              }
            }
          }
        }
      }
    }
  }
}`)

var schemaGetQueries = json.RawMessage(`{
  "type": "object",
  "properties": {
    "total": {"type": "integer", "description": "Total number of queries"},
    "queries": {
      "type": ["array", "null"],
      "items": {
        "type": "object",
        "properties": {
          "name":        {"type": "string"},
          "description": {"type": "string"},
          "statement":   {"type": "string"},
          "language":    {"type": "string"}
        }
      }
    }
  }
}`)

var schemaGetGlossaryTerm = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":        {"type": "string"},
    "name":       {"type": "string"},
    "definition": {"type": "string"},
    "entities": {
      "type": "array",
      "description": "Datasets and columns linked to this term",
      "items": {
        "type": "object",
        "properties": {
          "urn":    {"type": "string"},
          "column": {"type": "string"}
        }
      }
    }
  }
}`)

var schemaBrowse = json.RawMessage(`{
  "type": "object",
  "properties": {
    "tags": {
      "type": "array",
      "description": "Tags (present when what=tags)",
      "items": {
        "type": "object",
        "properties": {
          "urn":         {"type": "string"},
          "name":        {"type": "string"},
          "description": {"type": "string"}
        }
      }
    },
    "domains": {
      "type": "array",
      "description": "Domains (present when what=domains)",
      "items": {
        "type": "object",
        "properties": {
          "urn":         {"type": "string"},
          "name":        {"type": "string"},
          "description": {"type": "string"}
        }
      }
    },
    "data_products": {
      "type": "array",
      "description": "Data products (present when what=data_products)",
      "items": {
        "type": "object",
        "properties": {
          "urn":         {"type": "string"},
          "name":        {"type": "string"},
          "description": {"type": "string"},
          "domain":      {"type": "string"}
        }
      }
    }
  }
}`)

var schemaGetDataProduct = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":         {"type": "string"},
    "name":        {"type": "string"},
    "description": {"type": "string"},
    "domain": {
      "type": ["object", "null"],
      "properties": {
        "urn":         {"type": "string"},
        "name":        {"type": "string"},
        "description": {"type": "string"}
      }
    },
    "owners": {
      "type": ["array", "null"],
      "items": {
        "type": "object",
        "properties": {
          "urn":  {"type": "string"},
          "name": {"type": "string"},
          "type": {"type": "string"}
        }
      }
    },
    "assets": {
      "type": ["array", "null"],
      "description": "URNs of constituent datasets",
      "items": {"type": "string"}
    },
    "properties": {
      "type": ["object", "null"],
      "additionalProperties": {"type": "string"},
      "description": "Optional: additional metadata properties"
    }
  }
}`)

var schemaListConnections = json.RawMessage(`{
  "type": "object",
  "properties": {
    "connections": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "name":       {"type": "string"},
          "url":        {"type": "string"},
          "is_default": {"type": "boolean"}
        }
      }
    }
  }
}`)

var schemaCreate = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":    {"type": "string", "description": "URN of the created entity"},
    "what":   {"type": "string", "description": "Entity type that was created"},
    "action": {"type": "string", "description": "Always 'created'"}
  }
}`)

var schemaUpdate = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":        {"type": "string", "description": "URN of the updated entity"},
    "what":       {"type": "string", "description": "What was updated"},
    "action":     {"type": "string", "description": "Action performed (added, removed, updated)"},
    "target_urn": {"type": "string", "description": "URN of the target (tag, term, owner, etc.) when applicable"}
  }
}`)

var schemaDelete = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":    {"type": "string", "description": "URN of the deleted entity"},
    "what":   {"type": "string", "description": "Entity type that was deleted"},
    "action": {"type": "string", "description": "Always 'deleted'"}
  }
}`)
