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

echo "=========================================="
echo "🚀 DKonsole Release v${VERSION}"
echo "=========================================="
echo ""

# 1. Pull latest changes first
echo "📥 Pulling latest changes from remote..."
git pull --rebase origin main 2>&1 || git pull origin main
echo "✅ Remote changes synced"
echo ""

# 2. Commit and Push Changes
echo "📦 Preparing Git..."
if [[ -n $(git status -s) ]]; then
    echo "📝 Committing changes..."
    git add .
    git commit -m "chore: release v${VERSION}"
    echo "✅ Changes committed"
else
    echo "✨ No changes to commit"
fi

echo "⬆️  Pushing code to remote..."
git push origin main || git push
echo "✅ Code pushed"
echo ""

# 2. Build and Push Docker Image (Unified)
echo "📦 Building Unified Image (Backend + Frontend)..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."
docker build -t dkonsole/dkonsole:$VERSION .
echo "✅ Unified image built successfully"

echo "🚀 Pushing Unified Image to Docker Hub..."
docker push dkonsole/dkonsole:$VERSION
echo "✅ Unified image pushed successfully"
echo ""

# 3. Handle Git Tag
echo "🏷️  Handling Git tag v${VERSION}..."
if git rev-parse "v${VERSION}" >/dev/null 2>&1; then
    echo "⚠️  Tag v${VERSION} already exists. Deleting..."
    git tag -d "v${VERSION}"
    git push origin ":refs/tags/v${VERSION}" 2>/dev/null || true
    echo "🗑️  Old tag deleted"
fi

echo "🏷️  Creating new tag v${VERSION}..."
git tag -a "v${VERSION}" -m "Release v${VERSION}

Major Refactoring & Code Cleanup:
- Extracted resource detail components into modular structure
- Removed obsolete components (DeploymentTable, PodTable)
- Cleaned up legacy code and unused functions
- Removed duplicate files (AcercaDe.jsx)
- Cleaned up ingress and certificate example files
- Improved code organization and maintainability
- Removed unused imports and functions
- Backend refactoring: removed GetResources_OLD function (~760 lines)
- Frontend refactoring: componentized all resource details

Docker Image:
- dkonsole/dkonsole:${VERSION}"

echo "✅ Git tag created"

echo "📤 Pushing Git tag to remote..."
git push origin "v${VERSION}"
echo "✅ Git tag pushed successfully"
echo ""

echo "=========================================="
echo "✨ Release v${VERSION} Complete!"
echo "=========================================="
echo ""
echo "📦 Docker Image:"
echo "   - dkonsole/dkonsole:${VERSION}"
echo ""
echo "🏷️  Git Tag:"
echo "   - v${VERSION}"
echo ""
echo "🎉 All done!"
