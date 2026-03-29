#!/usr/bin/env bash
# sync.sh — Download DataHub GraphQL schema files from the upstream repository.
#
# Usage:
#   ./testdata/datahub-schema/sync.sh              # uses version from VERSION file
#   ./testdata/datahub-schema/sync.sh v1.5.0.1     # explicit version
#
# Downloads all .graphql files from:
#   https://github.com/datahub-project/datahub/tree/<version>/datahub-graphql-core/src/main/resources/
#
# The downloaded files are checked into the repo as the source of truth for
# schema validation tests.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
VERSION_FILE="$SCRIPT_DIR/VERSION"

VERSION="${1:-}"
if [ -z "$VERSION" ] && [ -f "$VERSION_FILE" ]; then
    VERSION="$(cat "$VERSION_FILE" | tr -d '[:space:]')"
fi

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <datahub-version>" >&2
    echo "  e.g. $0 v1.5.0.1" >&2
    exit 1
fi

BASE_URL="https://raw.githubusercontent.com/datahub-project/datahub/${VERSION}/datahub-graphql-core/src/main/resources"

# All .graphql files in the DataHub GraphQL core module.
# This list matches the contents of datahub-graphql-core/src/main/resources/.
# If DataHub adds new schema files, add them here.
SCHEMA_FILES=(
    analytics.graphql
    app.graphql
    assertions.graphql
    auth.graphql
    common.graphql
    connection.graphql
    contract.graphql
    documents.graphql
    entity.graphql
    files.graphql
    forms.graphql
    incident.graphql
    ingestion.graphql
    lineage.graphql
    logical.graphql
    module.graphql
    operations.graphql
    patch.graphql
    properties.graphql
    query.graphql
    recommendation.graphql
    runs.graphql
    search.graphql
    semantic-search.graphql
    settings.graphql
    step.graphql
    template.graphql
    tests.graphql
    timeline.graphql
    timeseries.graphql
    versioning.graphql
)

echo "Syncing DataHub GraphQL schema ${VERSION}..."

FAILED=0
for f in "${SCHEMA_FILES[@]}"; do
    HTTP_CODE=$(curl -sL -w '%{http_code}' -o "$SCRIPT_DIR/$f" "${BASE_URL}/$f")
    if [ "$HTTP_CODE" -ne 200 ]; then
        echo "  WARN: $f returned HTTP $HTTP_CODE (may not exist in this version)"
        rm -f "$SCRIPT_DIR/$f"
        FAILED=$((FAILED + 1))
    fi
done

# Update version file
echo "$VERSION" > "$VERSION_FILE"

DOWNLOADED=$(ls "$SCRIPT_DIR"/*.graphql 2>/dev/null | wc -l | tr -d ' ')
echo "Downloaded ${DOWNLOADED} schema files for ${VERSION} (${FAILED} missing)"
echo "Version written to ${VERSION_FILE}"
