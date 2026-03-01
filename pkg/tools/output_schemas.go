package tools

import "encoding/json"

// defaultOutputSchemas holds the default JSON Schema output descriptors for each built-in tool.
// These declare the structure of the JSON objects returned by each tool to MCP clients.
// Schemas are top-level objects; not exhaustive — they describe the primary response shape.
var defaultOutputSchemas = map[ToolName]json.RawMessage{
	ToolSearch:           schemaSearch,
	ToolGetEntity:        schemaGetEntity,
	ToolGetSchema:        schemaGetSchema,
	ToolGetLineage:       schemaGetLineage,
	ToolGetColumnLineage: schemaGetColumnLineage,
	ToolGetQueries:       schemaGetQueries,
	ToolGetGlossaryTerm:  schemaGetGlossaryTerm,
	ToolListTags:         schemaListTags,
	ToolListDomains:      schemaListDomains,
	ToolListDataProducts: schemaListDataProducts,
	ToolGetDataProduct:   schemaGetDataProduct,
	ToolListConnections:  schemaListConnections,
	// Write tools
	ToolUpdateDescription:  schemaUpdateDescription,
	ToolAddTag:             schemaAddTag,
	ToolRemoveTag:          schemaRemoveTag,
	ToolAddGlossaryTerm:    schemaAddGlossaryTerm,
	ToolRemoveGlossaryTerm: schemaRemoveGlossaryTerm,
	ToolAddLink:            schemaAddLink,
	ToolRemoveLink:         schemaRemoveLink,
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
  "properties": {
    "start":     {"type": "string", "description": "URN of the queried entity"},
    "direction": {"type": "string", "description": "Lineage direction: UPSTREAM or DOWNSTREAM"},
    "depth":     {"type": "integer", "description": "Depth of lineage traversal"},
    "nodes": {
      "type": ["array", "null"],
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
      "description": "Optional: query engine execution context for lineage bridging"
    }
  }
}`)

var schemaGetColumnLineage = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn": {"type": "string"},
    "columns": {
      "type": "array",
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

var schemaListTags = json.RawMessage(`{
  "type": "object",
  "properties": {
    "tags": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "urn":         {"type": "string"},
          "name":        {"type": "string"},
          "description": {"type": "string"}
        }
      }
    }
  }
}`)

var schemaListDomains = json.RawMessage(`{
  "type": "object",
  "properties": {
    "domains": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "urn":         {"type": "string"},
          "name":        {"type": "string"},
          "description": {"type": "string"}
        }
      }
    }
  }
}`)

var schemaListDataProducts = json.RawMessage(`{
  "type": "object",
  "properties": {
    "data_products": {
      "type": "array",
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

var schemaUpdateDescription = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":         {"type": "string"},
    "description": {"type": "string"},
    "aspect":      {"type": "string"},
    "action":      {"type": "string"}
  }
}`)

var schemaAddTag = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":    {"type": "string"},
    "tag":    {"type": "string"},
    "aspect": {"type": "string"},
    "action": {"type": "string"}
  }
}`)

var schemaRemoveTag = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":    {"type": "string"},
    "tag":    {"type": "string"},
    "aspect": {"type": "string"},
    "action": {"type": "string"}
  }
}`)

var schemaAddGlossaryTerm = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":           {"type": "string"},
    "glossary_term": {"type": "string"},
    "aspect":        {"type": "string"},
    "action":        {"type": "string"}
  }
}`)

var schemaRemoveGlossaryTerm = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":           {"type": "string"},
    "glossary_term": {"type": "string"},
    "aspect":        {"type": "string"},
    "action":        {"type": "string"}
  }
}`)

var schemaAddLink = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":    {"type": "string"},
    "url":    {"type": "string"},
    "label":  {"type": "string"},
    "action": {"type": "string"}
  }
}`)

var schemaRemoveLink = json.RawMessage(`{
  "type": "object",
  "properties": {
    "urn":    {"type": "string"},
    "url":    {"type": "string"},
    "action": {"type": "string"}
  }
}`)
