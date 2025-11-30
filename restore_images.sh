#!/bin/bash
set -e

# Versions to restore (in order)
VERSIONS=("1.2.8" "1.3.0" "1.3.2" "1.3.3" "1.3.4")

# Docker repository
REPO="dkonsole/dkonsole"

# Save current branch/commit
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
CURRENT_COMMIT=$(git rev-parse HEAD)

echo "==========================================="
echo "🔄 Restoring Docker Images from Git Tags"
echo "==========================================="
echo ""
echo "Current branch: $CURRENT_BRANCH"
echo "Versions to restore: ${VERSIONS[@]}"
echo ""

# Function to build and push a version
build_and_push() {
    local VERSION=$1
    local GIT_TAG="v${VERSION}"

    echo "==========================================="
    echo "📦 Building ${VERSION}"
    echo "==========================================="

    # Checkout the tag
    echo "🔄 Checking out tag ${GIT_TAG}..."
    git checkout "$GIT_TAG" 2>/dev/null || {
        echo "❌ Failed to checkout ${GIT_TAG}"
        return 1
    }

    # Build the image
    echo "🔨 Building Docker image..."
    docker build -t "${REPO}:${VERSION}" . || {
        echo "❌ Build failed for ${VERSION}"
        return 1
    }

    echo "✅ Build successful"

    # Push the primary version tag
    echo "🚀 Pushing ${REPO}:${VERSION}..."
    docker push "${REPO}:${VERSION}" || {
        echo "❌ Push failed for ${VERSION}"
        return 1
    }

    echo "✅ Pushed ${VERSION}"

    # Handle special tags
    if [ "$VERSION" == "1.3.4" ]; then
        echo "🏷️  Tagging as 'latest' and '1.3'..."
        docker tag "${REPO}:${VERSION}" "${REPO}:latest"
        docker tag "${REPO}:${VERSION}" "${REPO}:1.3"

        docker push "${REPO}:latest"
        docker push "${REPO}:1.3"
        echo "✅ Pushed latest and 1.3"
    elif [ "$VERSION" == "1.2.8" ]; then
        echo "🏷️  Tagging as '1.2'..."
        docker tag "${REPO}:${VERSION}" "${REPO}:1.2"

        docker push "${REPO}:1.2"
        echo "✅ Pushed 1.2"
    fi

    echo ""
    return 0
}

# Process each version
SUCCESS_COUNT=0
FAILED_COUNT=0

for VERSION in "${VERSIONS[@]}"; do
    if build_and_push "$VERSION"; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        FAILED_COUNT=$((FAILED_COUNT + 1))
    fi
done

# Return to original branch/commit
echo "🔄 Returning to original branch..."
git checkout "$CURRENT_BRANCH"

echo "==========================================="
echo "✨ Restoration Complete!"
echo "==========================================="
echo "Success: $SUCCESS_COUNT/$((SUCCESS_COUNT + FAILED_COUNT))"
echo "Failed: $FAILED_COUNT"
echo ""
