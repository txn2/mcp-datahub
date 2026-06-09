package types

// DataProduct represents a DataHub data product.
// Data products group datasets for specific business use cases.
type DataProduct struct {
	URN         string            `json:"urn"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Domain      *Domain           `json:"domain,omitempty"`
	Owners      []Owner           `json:"owners,omitempty"`
	Assets      []Entity          `json:"assets,omitempty"` // constituent datasets (member entities)
	Properties  map[string]string `json:"properties,omitempty"`
}
