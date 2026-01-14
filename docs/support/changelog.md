# Changelog

All notable changes to mcp-datahub are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of mcp-datahub
- DataHub search tool (`datahub_search`)
- Entity metadata tool (`datahub_get_entity`)
- Schema exploration tool (`datahub_get_schema`)
- Lineage tool (`datahub_get_lineage`)
- Query retrieval tool (`datahub_get_queries`)
- Glossary term tool (`datahub_get_glossary_term`)
- Tag listing tool (`datahub_list_tags`)
- Domain listing tool (`datahub_list_domains`)
- Data products tools (`datahub_list_data_products`, `datahub_get_data_product`)
- Composable Go library architecture
- Middleware support for custom authentication and logging
- Integration hooks for URN resolution and access filtering
- Comprehensive documentation site

### Security
- SLSA Level 3 provenance for all releases
- Cosign keyless signing for binaries and images
- Token redaction in logs and error messages

---

For the latest changes, see the [GitHub releases](https://github.com/txn2/mcp-datahub/releases).
