package client

import (
	"context"
	"fmt"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL queries and mutations for incidents (DataHub 1.4.x+).
const (
	// GetIncidentsQuery retrieves active incidents for an entity.
	GetIncidentsQuery = `
query getIncidents($urn: String!, $start: Int!, $count: Int!) {
  entity(urn: $urn) {
    ... on Dataset {
      incidents(state: ACTIVE, start: $start, count: $count) {
        total
        incidents {
          urn
          type
          customType
          title
          description
          status {
            state
            lastUpdated {
              time
              actor
            }
          }
          source {
            type
          }
          created {
            time
            actor
          }
        }
      }
    }
    ... on Dashboard {
      incidents(state: ACTIVE, start: $start, count: $count) {
        total
        incidents {
          urn
          type
          customType
          title
          description
          status {
            state
            lastUpdated {
              time
              actor
            }
          }
          source {
            type
          }
          created {
            time
            actor
          }
        }
      }
    }
    ... on Chart {
      incidents(state: ACTIVE, start: $start, count: $count) {
        total
        incidents {
          urn
          type
          customType
          title
          description
          status {
            state
            lastUpdated {
              time
              actor
            }
          }
          source {
            type
          }
          created {
            time
            actor
          }
        }
      }
    }
    ... on DataFlow {
      incidents(state: ACTIVE, start: $start, count: $count) {
        total
        incidents {
          urn
          type
          customType
          title
          description
          status {
            state
            lastUpdated {
              time
              actor
            }
          }
          source {
            type
          }
          created {
            time
            actor
          }
        }
      }
    }
    ... on DataJob {
      incidents(state: ACTIVE, start: $start, count: $count) {
        total
        incidents {
          urn
          type
          customType
          title
          description
          status {
            state
            lastUpdated {
              time
              actor
            }
          }
          source {
            type
          }
          created {
            time
            actor
          }
        }
      }
    }
  }
}
`

	// RaiseIncidentMutation creates a new incident on entities.
	RaiseIncidentMutation = `
mutation raiseIncident($input: RaiseIncidentInput!) {
  raiseIncident(input: $input)
}
`

	// UpdateIncidentStatusMutation resolves or reactivates an incident.
	UpdateIncidentStatusMutation = `
mutation updateIncidentStatus($urn: String!, $input: UpdateIncidentStatusInput!) {
  updateIncidentStatus(urn: $urn, input: $input)
}
`
)

// GetIncidents retrieves active incidents for an entity.
// Returns empty results (not an error) when incidents are not available,
// which is common on DataHub versions before 1.4.x.
func (c *Client) GetIncidents(ctx context.Context, urn string) (*types.IncidentResult, error) {
	variables := map[string]any{
		"urn":   urn,
		"start": 0,
		"count": c.config.DefaultLimit,
	}

	var response struct {
		Entity struct {
			Incidents *incidentsResponse `json:"incidents"`
		} `json:"entity"`
	}

	if err := c.Execute(ctx, GetIncidentsQuery, variables, &response); err != nil {
		// Return empty result when incidents are not supported (DataHub < 1.4.x)
		return &types.IncidentResult{}, nil
	}

	if response.Entity.Incidents == nil {
		return &types.IncidentResult{}, nil
	}

	return response.Entity.Incidents.toResult(), nil
}

// RaiseIncident creates a new incident on entities.
func (c *Client) RaiseIncident(ctx context.Context, input types.RaiseIncidentInput) (string, error) {
	resourceURNs := make([]map[string]string, 0, len(input.ResourceURNs))
	for _, urn := range input.ResourceURNs {
		resourceURNs = append(resourceURNs, map[string]string{"urn": urn})
	}

	gqlInput := map[string]any{
		"type":         input.Type,
		"title":        input.Title,
		"resourceUrns": resourceURNs,
	}
	if input.CustomType != "" {
		gqlInput["customType"] = input.CustomType
	}
	if input.Description != "" {
		gqlInput["description"] = input.Description
	}

	variables := map[string]any{
		"input": gqlInput,
	}

	var response struct {
		RaiseIncident string `json:"raiseIncident"`
	}

	if err := c.Execute(ctx, RaiseIncidentMutation, variables, &response); err != nil {
		return "", fmt.Errorf("RaiseIncident: %w", err)
	}

	return response.RaiseIncident, nil
}

// ResolveIncident marks an incident as resolved.
func (c *Client) ResolveIncident(ctx context.Context, incidentURN, message string) error {
	gqlInput := map[string]any{
		"state": "RESOLVED",
	}
	if message != "" {
		gqlInput["message"] = message
	}

	variables := map[string]any{
		"urn":   incidentURN,
		"input": gqlInput,
	}

	var response struct {
		UpdateIncidentStatus bool `json:"updateIncidentStatus"`
	}

	if err := c.Execute(ctx, UpdateIncidentStatusMutation, variables, &response); err != nil {
		return fmt.Errorf("ResolveIncident: %w", err)
	}

	return nil
}

// incidentsResponse maps the GraphQL incidents response.
type incidentsResponse struct {
	Total     int                `json:"total"`
	Incidents []incidentGQLEntry `json:"incidents"`
}

// incidentGQLEntry maps a single incident from GraphQL.
type incidentGQLEntry struct {
	URN        string `json:"urn"`
	Type       string `json:"type"`
	CustomType string `json:"customType"`
	Title      string `json:"title"`
	Desc       string `json:"description"`
	Status     struct {
		State       string `json:"state"`
		LastUpdated struct {
			Time  int64  `json:"time"`
			Actor string `json:"actor"`
		} `json:"lastUpdated"`
	} `json:"status"`
	Source struct {
		Type string `json:"type"`
	} `json:"source"`
	Created struct {
		Time  int64  `json:"time"`
		Actor string `json:"actor"`
	} `json:"created"`
}

func (r *incidentsResponse) toResult() *types.IncidentResult {
	result := &types.IncidentResult{Total: r.Total}
	for _, inc := range r.Incidents {
		result.Incidents = append(result.Incidents, types.Incident{
			URN:           inc.URN,
			Type:          inc.Type,
			CustomType:    inc.CustomType,
			Title:         inc.Title,
			Description:   inc.Desc,
			State:         inc.Status.State,
			Source:        inc.Source.Type,
			Created:       inc.Created.Time,
			CreatedBy:     inc.Created.Actor,
			LastUpdated:   inc.Status.LastUpdated.Time,
			LastUpdatedBy: inc.Status.LastUpdated.Actor,
		})
	}
	return result
}
