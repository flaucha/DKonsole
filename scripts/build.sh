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

# Use same version (no test suffix)
BUILD_VERSION="${VERSION}"

echo "=========================================="
echo "🔨 DKonsole Build v${BUILD_VERSION}"
echo "=========================================="
echo ""

# Check for uncommitted changes and commit/push if needed
if ! git diff-index --quiet HEAD --; then
    echo "📝 Detected uncommitted changes, committing and pushing..."
    git add -A
    git commit -m "chore: update code before build ${BUILD_VERSION}" || true
    if git rev-parse --abbrev-ref HEAD | grep -q "main\|master"; then
        git push || echo "⚠️  Warning: Could not push to remote (may need manual push)"
    fi
    echo "✅ Changes committed and pushed"
    echo ""
fi

# Build Unified Docker Image (Backend + Frontend)
echo "📦 Building Unified Image (Backend + Frontend)..."
docker build -t dkonsole/dkonsole:$BUILD_VERSION .
echo "✅ Unified image built successfully"
echo ""

# Push to Docker Hub
echo "🚀 Pushing Unified Image to Docker Hub..."
docker push dkonsole/dkonsole:$BUILD_VERSION
echo "✅ Unified image pushed successfully"
echo ""

echo "=========================================="
echo "✨ Build Complete!"
echo "=========================================="
echo ""
echo "📦 Docker Image:"
echo "   - dkonsole/dkonsole:${BUILD_VERSION}"
echo ""
echo "🧪 To test locally:"
echo "   docker run -p 8080:8080 dkonsole/dkonsole:${BUILD_VERSION}"
echo ""
