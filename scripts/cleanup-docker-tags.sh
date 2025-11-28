#!/bin/bash
# Script to delete test tags from Docker Hub (docker.io) and local Docker
# Usage: ./cleanup-docker-tags.sh [DOCKER_HUB_USERNAME] [DOCKER_HUB_TOKEN]

set -e

REPO="dkonsole/dkonsole"
FULL_REPO="docker.io/$REPO"
USERNAME="${1:-}"
TOKEN="${2:-}"

if [ -z "$USERNAME" ] || [ -z "$TOKEN" ]; then
    echo "Usage: $0 <docker_hub_username> <docker_hub_token>"
    echo ""
    echo "This script deletes all tags containing 'test' from Docker Hub (docker.io) and local Docker."
    echo "You can get a token from: https://hub.docker.com/settings/security"
    echo ""
    echo "Example:"
    echo "  $0 myusername mytoken123"
    exit 1
fi

echo "=========================================="
echo "🧹 Cleaning up test tags from Docker Hub (docker.io) and local"
echo "=========================================="
echo ""

# Get all tags from Docker Hub API (handle pagination)
echo "📋 Fetching tags from Docker Hub (docker.io)..."
ALL_TAGS=""
PAGE=1
PAGE_SIZE=100

while true; do
    TAGS_JSON=$(curl -s -u "$USERNAME:$TOKEN" \
        "https://hub.docker.com/v2/repositories/$REPO/tags/?page=$PAGE&page_size=$PAGE_SIZE")

    PAGE_TAGS=$(echo "$TAGS_JSON" | grep -o '"name":"[^"]*"' | sed 's/"name":"\([^"]*\)"/\1/')

    if [ -z "$PAGE_TAGS" ]; then
        break
    fi

    ALL_TAGS="$ALL_TAGS"$'\n'"$PAGE_TAGS"

    # Check if there's a next page
    HAS_NEXT=$(echo "$TAGS_JSON" | grep -o '"next":"[^"]*"' || echo "")
    if [ -z "$HAS_NEXT" ]; then
        break
    fi

    PAGE=$((PAGE + 1))
done

# Extract tags containing "test" (case insensitive)
TEST_TAGS=$(echo "$ALL_TAGS" | grep -i "test" | sort -u)

if [ -z "$TEST_TAGS" ]; then
    echo "✅ No test tags found to delete on Docker Hub."
else
    echo "Found test tags on Docker Hub to delete:"
    echo "$TEST_TAGS" | sed 's/^/  - /'
    echo ""
fi

# Also check local Docker images
echo "📋 Checking local Docker images..."
LOCAL_TEST_TAGS=$(docker images "$FULL_REPO" --format "{{.Tag}}" 2>/dev/null | grep -i "test" | sort -u || true)

if [ -n "$LOCAL_TEST_TAGS" ]; then
    echo "Found test tags locally to delete:"
    echo "$LOCAL_TEST_TAGS" | sed 's/^/  - /'
    echo ""
fi

if [ -z "$TEST_TAGS" ] && [ -z "$LOCAL_TEST_TAGS" ]; then
    echo "✅ No test tags found to delete."
    exit 0
fi

# Confirm deletion
read -p "Do you want to delete these tags? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "❌ Deletion cancelled."
    exit 0
fi

# Delete tags from Docker Hub
DELETED_HUB=0
FAILED_HUB=0

if [ -n "$TEST_TAGS" ]; then
    echo ""
    echo "🗑️  Deleting tags from Docker Hub (docker.io)..."
    for TAG in $TEST_TAGS; do
        echo -n "  Deleting $REPO:$TAG from Docker Hub... "
        RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE \
            -u "$USERNAME:$TOKEN" \
            "https://hub.docker.com/v2/repositories/$REPO/tags/$TAG/")

        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

        if [ "$HTTP_CODE" = "204" ]; then
            echo "✅ Deleted"
            DELETED_HUB=$((DELETED_HUB + 1))
        else
            echo "❌ Failed (HTTP $HTTP_CODE)"
            FAILED_HUB=$((FAILED_HUB + 1))
        fi
    done
fi

# Delete tags from local Docker
DELETED_LOCAL=0
FAILED_LOCAL=0

if [ -n "$LOCAL_TEST_TAGS" ]; then
    echo ""
    echo "🗑️  Deleting tags from local Docker..."
    for TAG in $LOCAL_TEST_TAGS; do
        echo -n "  Deleting $FULL_REPO:$TAG locally... "
        if docker rmi "$FULL_REPO:$TAG" 2>/dev/null; then
            echo "✅ Deleted"
            DELETED_LOCAL=$((DELETED_LOCAL + 1))
        else
            echo "❌ Failed (may not exist or in use)"
            FAILED_LOCAL=$((FAILED_LOCAL + 1))
        fi
    done
fi

echo ""
echo "=========================================="
echo "✨ Cleanup Complete!"
echo "=========================================="
echo "Docker Hub (docker.io):"
echo "  Deleted: $DELETED_HUB tags"
if [ $FAILED_HUB -gt 0 ]; then
    echo "  Failed: $FAILED_HUB tags"
fi
echo "Local Docker:"
echo "  Deleted: $DELETED_LOCAL tags"
if [ $FAILED_LOCAL -gt 0 ]; then
    echo "  Failed: $FAILED_LOCAL tags"
fi
echo ""
