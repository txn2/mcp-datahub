package client

import (
	"context"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// GraphQL query for data contracts (DataHub 1.3.x+).
const (
	// GetDataContractQuery retrieves the data contract status for a dataset.
	GetDataContractQuery = `
query getDataContract($urn: String!) {
  dataset(urn: $urn) {
    contract {
      result(refresh: false) {
        type
        assertionResults {
          assertion {
            urn
          }
          type
          result {
            type
            nativeResults {
              key
              value
            }
          }
        }
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
			Contract *struct {
				Result *contractResultEntry `json:"result"`
			} `json:"contract"`
		} `json:"dataset"`
	}

	if err := c.Execute(ctx, GetDataContractQuery, variables, &response); err != nil {
		// Return nil when data contracts are not supported
		c.logger.Debug("GetDataContract graceful fallback", "urn", datasetURN, "error", err.Error())
		return nil, nil
	}

	if response.Dataset.Contract == nil || response.Dataset.Contract.Result == nil {
		return nil, nil
	}

	return response.Dataset.Contract.Result.toContract(), nil
}

// contractResultEntry maps the GraphQL contract result response.
type contractResultEntry struct {
	Type             string                    `json:"type"`
	AssertionResults []assertionResultGQLEntry `json:"assertionResults"`
}

// assertionResultGQLEntry maps a single assertion result from GraphQL.
type assertionResultGQLEntry struct {
	Assertion struct {
		URN string `json:"urn"`
	} `json:"assertion"`
	Type   string `json:"type"`
	Result struct {
		Type          string `json:"type"`
		NativeResults []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"nativeResults"`
	} `json:"result"`
}

func (r *contractResultEntry) toContract() *types.DataContract {
	contract := &types.DataContract{
		Status: r.Type,
	}

	for _, ar := range r.AssertionResults {
		result := types.AssertionResult{
			AssertionURN: ar.Assertion.URN,
			Type:         ar.Type,
			ResultType:   ar.Result.Type,
		}

		if len(ar.Result.NativeResults) > 0 {
			result.NativeResults = make(map[string]string, len(ar.Result.NativeResults))
			for _, nr := range ar.Result.NativeResults {
				result.NativeResults[nr.Key] = nr.Value
			}
		}

		contract.AssertionResults = append(contract.AssertionResults, result)
	}

	return contract
}
