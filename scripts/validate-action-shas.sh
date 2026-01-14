#!/bin/bash
# validate-action-shas.sh
# Validates that all GitHub Actions in workflows use pinned SHA references
# instead of mutable tags like @v1 or @main

set -euo pipefail

WORKFLOWS_DIR=".github/workflows"
EXIT_CODE=0

echo "Validating GitHub Actions SHA pinning..."
echo ""

# Find all workflow files
for workflow in "$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml; do
    [ -f "$workflow" ] || continue

    echo "Checking: $workflow"

    # Extract uses: lines and check for unpinned references
    while IFS= read -r line; do
        # Skip empty lines and comments
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue

        # Extract the action reference
        if [[ "$line" =~ uses:[[:space:]]*([^[:space:]]+) ]]; then
            action="${BASH_REMATCH[1]}"

            # Skip local actions (starting with ./)
            [[ "$action" == ./* ]] && continue

            # Check if it uses a SHA (40 hex characters)
            if [[ ! "$action" =~ @[a-f0-9]{40} ]]; then
                echo "  ERROR: Unpinned action: $action"
                EXIT_CODE=1
            fi
        fi
    done < <(grep -E "uses:" "$workflow" || true)
done

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "All actions are properly pinned to SHA references."
else
    echo "Some actions are not pinned to SHA references."
    echo "Please update them to use full SHA hashes instead of tags."
    echo ""
    echo "Example:"
    echo "  Bad:  uses: actions/checkout@v4"
    echo "  Good: uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2"
fi

exit $EXIT_CODE
