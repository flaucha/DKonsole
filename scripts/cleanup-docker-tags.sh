#!/bin/bash
# Script to delete test tags from Docker Hub
# Usage: ./cleanup-docker-tags.sh [DOCKER_HUB_USERNAME] [DOCKER_HUB_TOKEN]

set -e

REPO="dkonsole/dkonsole"
USERNAME="${1:-}"
TOKEN="${2:-}"

if [ -z "$USERNAME" ] || [ -z "$TOKEN" ]; then
    echo "Usage: $0 <docker_hub_username> <docker_hub_token>"
    echo ""
    echo "This script deletes all tags containing 'test' from Docker Hub."
    echo "You can get a token from: https://hub.docker.com/settings/security"
    echo ""
    echo "Example:"
    echo "  $0 myusername mytoken123"
    exit 1
fi

echo "=========================================="
echo "🧹 Cleaning up test tags from Docker Hub"
echo "=========================================="
echo ""

# Get all tags from Docker Hub API
echo "📋 Fetching tags from Docker Hub..."
TAGS_JSON=$(curl -s -u "$USERNAME:$TOKEN" \
    "https://hub.docker.com/v2/repositories/$REPO/tags/?page_size=100")

# Extract tags containing "test" (case insensitive)
TEST_TAGS=$(echo "$TAGS_JSON" | grep -o '"name":"[^"]*test[^"]*"' | sed 's/"name":"\([^"]*\)"/\1/' | sort -u)

if [ -z "$TEST_TAGS" ]; then
    echo "✅ No test tags found to delete."
    exit 0
fi

echo "Found test tags to delete:"
echo "$TEST_TAGS" | sed 's/^/  - /'
echo ""

# Confirm deletion
read -p "Do you want to delete these tags? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "❌ Deletion cancelled."
    exit 0
fi

# Delete each tag
DELETED=0
FAILED=0

for TAG in $TEST_TAGS; do
    echo -n "Deleting $TAG... "
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE \
        -u "$USERNAME:$TOKEN" \
        "https://hub.docker.com/v2/repositories/$REPO/tags/$TAG/")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

    if [ "$HTTP_CODE" = "204" ]; then
        echo "✅ Deleted"
        DELETED=$((DELETED + 1))
    else
        echo "❌ Failed (HTTP $HTTP_CODE)"
        FAILED=$((FAILED + 1))
    fi
done

echo ""
echo "=========================================="
echo "✨ Cleanup Complete!"
echo "=========================================="
echo "Deleted: $DELETED tags"
if [ $FAILED -gt 0 ]; then
    echo "Failed: $FAILED tags"
fi
echo ""
