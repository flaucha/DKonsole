#!/bin/bash
set -e

VERSION="1.0.7"

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

# 2. Build and Push Docker Images
echo "📦 Building Backend..."
docker build -t dkonsole/dkonsole-backend:$VERSION ./backend
echo "✅ Backend built successfully"

echo "📦 Building Frontend..."
docker build -t dkonsole/dkonsole-frontend:$VERSION ./frontend
echo "✅ Frontend built successfully"

echo "🚀 Pushing Backend to Docker Hub..."
docker push dkonsole/dkonsole-backend:$VERSION
echo "✅ Backend pushed successfully"

echo "🚀 Pushing Frontend to Docker Hub..."
docker push dkonsole/dkonsole-frontend:$VERSION
echo "✅ Frontend pushed successfully"
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

Security Fixes:
- Critical: Fixed Secrets exposure in API
- Critical: Implemented strict CORS
- Critical: Enforced JWT_SECRET validation
- Critical: Added YAML import validation and limits
- Critical: Strengthened WebSocket origin check
- Critical: Reduced RBAC permissions
- High: Added Prometheus timeouts
- High: Validated file uploads
- High: Added security headers (Nginx)

Features:
- Prometheus integration for Pod metrics
- Historical metrics with time range selector
- Metrics tab in Pod details
- Fixed namespace display for cluster-scoped resources

Docker Images:
- dkonsole/dkonsole-backend:${VERSION}
- dkonsole/dkonsole-frontend:${VERSION}"

echo "✅ Git tag created"

echo "📤 Pushing Git tag to remote..."
git push origin "v${VERSION}"
echo "✅ Git tag pushed successfully"
echo ""

echo "=========================================="
echo "✨ Release v${VERSION} Complete!"
echo "=========================================="
echo ""
echo "📦 Docker Images:"
echo "   - dkonsole/dkonsole-backend:${VERSION}"
echo "   - dkonsole/dkonsole-frontend:${VERSION}"
echo ""
echo "🏷️  Git Tag:"
echo "   - v${VERSION}"
echo ""
echo "🎉 All done!"
