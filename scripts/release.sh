#!/bin/bash
set -e

VERSION="1.1.3"

echo "=========================================="
echo "🚀 DKonsole Release v${VERSION}"
echo "=========================================="
echo ""

# 1. Commit and Push Changes
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

Resource Quota Manager Improvements:
- Added namespace selector with All/Selected toggle
- Automatic refresh after create/edit/delete operations
- UI consistency improvements matching other pages
- Streamlined menu (Edit YAML and Delete only)
- YAML editor now uses kubectl apply (Server-Side Apply)
- Create menu changed from hover to click
- New resource templates use selected namespace
- Consistent gray-scale color scheme

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
