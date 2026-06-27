package client

import "github.com/txn2/mcp-datahub/pkg/types"

// searchResultItem is the GraphQL response shape for a single search result.
type searchResultItem struct {
	Entity struct {
		URN         string `json:"urn"`
		Type        string `json:"type"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Platform    struct {
			Name string `json:"name"`
		} `json:"platform"`
		Properties struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"properties"`
		SubType string `json:"subType"`
		Info    struct {
			Title    string `json:"title"`
			Contents struct {
				Text string `json:"text"`
			} `json:"contents"`
			Status struct {
				State string `json:"state"`
			} `json:"status"`
		} `json:"info"`
		Ownership struct {
			Owners []struct {
				Owner struct {
					URN      string `json:"urn"`
					Username string `json:"username"`
					Name     string `json:"name"`
				} `json:"owner"`
				Type string `json:"type"`
			} `json:"owners"`
		} `json:"ownership"`
		Tags struct {
			Tags []struct {
				Tag struct {
					URN         string `json:"urn"`
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"tag"`
			} `json:"tags"`
		} `json:"tags"`
		Domain struct {
			Domain struct {
				URN        string `json:"urn"`
				Properties struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"properties"`
			} `json:"domain"`
		} `json:"domain"`
	} `json:"entity"`
	MatchedFields []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"matchedFields"`
}

// parseSearchResult converts a GraphQL search result item into a SearchEntity.
func parseSearchResult(sr searchResultItem) types.SearchEntity {
	name := sr.Entity.Name
	description := sr.Entity.Description

	// For DataProduct, GlossaryTerm, Tag the name/description come from properties
	if sr.Entity.Properties.Name != "" {
		name = sr.Entity.Properties.Name
	}
	if sr.Entity.Properties.Description != "" {
		description = sr.Entity.Properties.Description
	}
	// For Document the title comes from info, content serves as description.
	// Guard behind type check to avoid mis-attributing info fields from
	// non-document entities in the polymorphic GraphQL union response.
	if sr.Entity.Type == EntityTypeDocument {
		if sr.Entity.Info.Title != "" {
			name = sr.Entity.Info.Title
		}
		if sr.Entity.Info.Contents.Text != "" && description == "" {
			description = sr.Entity.Info.Contents.Text
		}
	}

	entity := types.SearchEntity{
		URN:         sr.Entity.URN,
		Type:        sr.Entity.Type,
		Name:        name,
		Description: description,
		Platform:    sr.Entity.Platform.Name,
	}

	for _, o := range sr.Entity.Ownership.Owners {
		ownerName := o.Owner.Username
		if o.Owner.Name != "" {
			ownerName = o.Owner.Name
		}
		entity.Owners = append(entity.Owners, types.Owner{
			URN:  o.Owner.URN,
			Name: ownerName,
			Type: types.OwnershipType(o.Type),
		})
	}

	for _, t := range sr.Entity.Tags.Tags {
		entity.Tags = append(entity.Tags, types.Tag{
			URN:         t.Tag.URN,
			Name:        t.Tag.Name,
			Description: t.Tag.Description,
		})
	}

	if sr.Entity.Domain.Domain.URN != "" {
		entity.Domain = &types.Domain{
			URN:         sr.Entity.Domain.Domain.URN,
			Name:        sr.Entity.Domain.Domain.Properties.Name,
			Description: sr.Entity.Domain.Domain.Properties.Description,
		}
	}

	for _, mf := range sr.MatchedFields {
		entity.MatchedFields = append(entity.MatchedFields, types.MatchedField{
			Name:  mf.Name,
			Value: mf.Value,
		})
	}

	return entity
}
