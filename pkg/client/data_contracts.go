package client

import (
	"context"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL query for data contracts (DataHub 1.3.x+).
const (
	// GetDataContractQuery retrieves the data contract for a dataset.
	// DataContract has properties (assertion references by category) and
	// status (overall state). Each category contains assertions with URNs.
	GetDataContractQuery = `
query getDataContract($urn: String!) {
  dataset(urn: $urn) {
    contract {
      urn
      properties {
        entityUrn
        freshness {
          assertion {
            urn
          }
        }
        schema {
          assertion {
            urn
          }
        }
        dataQuality {
          assertion {
            urn
          }
        }
      }
      status {
        state
      }
    }
  }
}
`
)

// GetDataContract retrieves the data contract status for a dataset.
// Returns nil (not an error) when data contracts are not available.
func (c *Client) GetDataContract(ctx context.Context, datasetURN string) (*types.DataContract, error) {
	variables := map[string]any{
		"urn": datasetURN,
	}

	var response struct {
		Dataset struct {
			Contract *contractGQLResponse `json:"contract"`
		} `json:"dataset"`
	}

	if err := c.Execute(ctx, GetDataContractQuery, variables, &response); err != nil {
		// Return nil when data contracts are not supported
		c.logger.Debug("GetDataContract graceful fallback", "urn", datasetURN, "error", err.Error())
		return nil, nil
	}

	if response.Dataset.Contract == nil {
		return nil, nil
	}

	return response.Dataset.Contract.toContract(), nil
}

// contractGQLResponse maps the GraphQL DataContract response.
type contractGQLResponse struct {
	URN        string `json:"urn"`
	Properties *struct {
		EntityURN   string                 `json:"entityUrn"`
		Freshness   []contractAssertionRef `json:"freshness"`
		Schema      []contractAssertionRef `json:"schema"`
		DataQuality []contractAssertionRef `json:"dataQuality"`
	} `json:"properties"`
	Status *struct {
		State string `json:"state"`
	} `json:"status"`
}

// contractAssertionRef maps an assertion reference within a contract category.
type contractAssertionRef struct {
	Assertion struct {
		URN string `json:"urn"`
	} `json:"assertion"`
}

func (r *contractGQLResponse) toContract() *types.DataContract {
	contract := &types.DataContract{}

	if r.Status != nil {
		contract.Status = r.Status.State
	}

	if r.Properties == nil {
		return contract
	}

	for _, a := range r.Properties.Freshness {
		contract.AssertionResults = append(contract.AssertionResults, types.AssertionResult{
			AssertionURN: a.Assertion.URN,
			Type:         "FRESHNESS",
		})
	}

	for _, a := range r.Properties.Schema {
		contract.AssertionResults = append(contract.AssertionResults, types.AssertionResult{
			AssertionURN: a.Assertion.URN,
			Type:         "SCHEMA",
		})
	}

	for _, a := range r.Properties.DataQuality {
		contract.AssertionResults = append(contract.AssertionResults, types.AssertionResult{
			AssertionURN: a.Assertion.URN,
			Type:         "DATA_QUALITY",
		})
	}

	return contract
}
