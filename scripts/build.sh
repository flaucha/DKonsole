#!/bin/bash
set -e

# Read version from VERSION file or use default
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

if [ -f "VERSION" ]; then
    VERSION=$(cat VERSION | tr -d '[:space:]')
else
    VERSION="1.1.9"
fi

# Use test version for testing (add -test-1 suffix only if version doesn't already have a suffix)
if [[ "$VERSION" == *"-"* ]]; then
    # Version already has a suffix (e.g., 1.3.0-alfa1), use it directly
    TEST_VERSION="$VERSION"
else
    # No suffix, add -test-1
    TEST_VERSION="${VERSION}-test-1"
fi

echo "=========================================="
echo "🔨 DKonsole Build v${TEST_VERSION}"
echo "=========================================="
echo ""

# Check for uncommitted changes and commit/push if needed
if ! git diff-index --quiet HEAD --; then
    echo "📝 Detected uncommitted changes, committing and pushing..."
    git add -A
    git commit -m "chore: update code before build ${TEST_VERSION}" || true
    if git rev-parse --abbrev-ref HEAD | grep -q "main\|master"; then
        git push || echo "⚠️  Warning: Could not push to remote (may need manual push)"
    fi
    echo "✅ Changes committed and pushed"
    echo ""
fi

# Build Unified Docker Image (Backend + Frontend)
echo "📦 Building Unified Image (Backend + Frontend)..."
docker build -t dkonsole/dkonsole:$TEST_VERSION .
echo "✅ Unified image built successfully"
echo ""

# Push to Docker Hub
echo "🚀 Pushing Unified Image to Docker Hub..."
docker push dkonsole/dkonsole:$TEST_VERSION
echo "✅ Unified image pushed successfully"
echo ""

echo "=========================================="
echo "✨ Build Complete!"
echo "=========================================="
echo ""
echo "📦 Docker Image:"
echo "   - dkonsole/dkonsole:${TEST_VERSION}"
echo ""
echo "🧪 To test locally:"
echo "   docker run -p 8080:8080 dkonsole/dkonsole:${TEST_VERSION}"
echo ""
