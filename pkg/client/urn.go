package client

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// BuildDatasetURN constructs a dataset URN.
func BuildDatasetURN(platform, qualifiedName, env string) string {
	if env == "" {
		env = "PROD"
	}
	encodedName := url.PathEscape(qualifiedName)
	return fmt.Sprintf("urn:li:dataset:(urn:li:dataPlatform:%s,%s,%s)", platform, encodedName, env)
}

// BuildDashboardURN constructs a dashboard URN.
func BuildDashboardURN(platform, dashboardID string) string {
	return fmt.Sprintf("urn:li:dashboard:(%s,%s)", platform, dashboardID)
}

// BuildChartURN constructs a chart URN.
func BuildChartURN(platform, chartID string) string {
	return fmt.Sprintf("urn:li:chart:(%s,%s)", platform, chartID)
}

// BuildDataFlowURN constructs a data flow URN.
func BuildDataFlowURN(orchestrator, flowID, cluster string) string {
	return fmt.Sprintf("urn:li:dataFlow:(%s,%s,%s)", orchestrator, flowID, cluster)
}

// BuildDataJobURN constructs a data job URN.
func BuildDataJobURN(dataFlowURN, jobID string) string {
	return fmt.Sprintf("urn:li:dataJob:(%s,%s)", dataFlowURN, jobID)
}

// BuildGlossaryTermURN constructs a glossary term URN.
func BuildGlossaryTermURN(termPath string) string {
	return fmt.Sprintf("urn:li:glossaryTerm:%s", termPath)
}

// BuildTagURN constructs a tag URN.
func BuildTagURN(tagName string) string {
	return fmt.Sprintf("urn:li:tag:%s", tagName)
}

// BuildDomainURN constructs a domain URN.
func BuildDomainURN(domainID string) string {
	return fmt.Sprintf("urn:li:domain:%s", domainID)
}

// ParseURN parses a DataHub URN into its components.
func ParseURN(urn string) (*types.ParsedURN, error) {
	if !strings.HasPrefix(urn, "urn:li:") {
		return nil, fmt.Errorf("%w: must start with 'urn:li:'", ErrInvalidURN)
	}

	// Remove prefix
	rest := strings.TrimPrefix(urn, "urn:li:")

	// Find entity type (everything before the first : or ()
	colonIdx := strings.Index(rest, ":")
	parenIdx := strings.Index(rest, "(")

	var entityType, remainder string

	switch {
	case colonIdx == -1 && parenIdx == -1:
		// Simple URN like urn:li:tag:MyTag
		entityType = rest
		remainder = ""
	case parenIdx != -1 && (colonIdx == -1 || parenIdx < colonIdx):
		// Compound URN like urn:li:dataset:(...)
		entityType = rest[:parenIdx]
		remainder = rest[parenIdx:]
	default:
		// Simple URN with value like urn:li:glossaryTerm:path.to.term
		entityType = rest[:colonIdx]
		remainder = rest[colonIdx+1:]
	}

	parsed := &types.ParsedURN{
		Raw:        urn,
		EntityType: entityType,
	}

	// Parse entity-specific parts
	switch entityType {
	case "dataset":
		if err := parseDatasetURN(remainder, parsed); err != nil {
			return nil, err
		}
	case "dashboard", "chart":
		if err := parseTupleURN(remainder, parsed); err != nil {
			return nil, err
		}
	default:
		// For simple URNs, the remainder is the name
		parsed.Name = remainder
	}

	return parsed, nil
}

// parseDatasetURN parses dataset URN specifics.
func parseDatasetURN(remainder string, parsed *types.ParsedURN) error {
	// Format: (urn:li:dataPlatform:platform,name,env)
	if !strings.HasPrefix(remainder, "(") || !strings.HasSuffix(remainder, ")") {
		return fmt.Errorf("%w: dataset URN must have parentheses", ErrInvalidURN)
	}

	inner := remainder[1 : len(remainder)-1]
	parts := strings.SplitN(inner, ",", 3)
	if len(parts) != 3 {
		return fmt.Errorf("%w: dataset URN must have 3 parts", ErrInvalidURN)
	}

	// Parse platform
	platformURN := parts[0]
	if !strings.HasPrefix(platformURN, "urn:li:dataPlatform:") {
		return fmt.Errorf("%w: invalid platform URN", ErrInvalidURN)
	}
	parsed.Platform = strings.TrimPrefix(platformURN, "urn:li:dataPlatform:")

	// Decode name
	name, err := url.PathUnescape(parts[1])
	if err != nil {
		parsed.Name = parts[1]
	} else {
		parsed.Name = name
	}

	parsed.Env = parts[2]

	return nil
}

// parseTupleURN parses tuple-style URNs like dashboard and chart.
func parseTupleURN(remainder string, parsed *types.ParsedURN) error {
	if !strings.HasPrefix(remainder, "(") || !strings.HasSuffix(remainder, ")") {
		return fmt.Errorf("%w: tuple URN must have parentheses", ErrInvalidURN)
	}

	inner := remainder[1 : len(remainder)-1]
	parts := strings.SplitN(inner, ",", 2)
	if len(parts) != 2 {
		return fmt.Errorf("%w: tuple URN must have 2 parts", ErrInvalidURN)
	}

	parsed.Platform = parts[0]
	parsed.Name = parts[1]

	return nil
}
