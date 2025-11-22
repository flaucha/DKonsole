#!/bin/bash
set -e

VERSION="1.0.4"

echo "=========================================="
echo "🚀 DKonsole Release v${VERSION}"
echo "=========================================="
echo ""

# Check if there are uncommitted changes
if [[ -n $(git status -s) ]]; then
    echo "⚠️  Warning: You have uncommitted changes"
    echo "Please commit or stash your changes before releasing"
    echo ""
    git status -s
    echo ""
    read -p "Do you want to continue anyway? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Release cancelled"
        exit 1
    fi
fi

echo "📦 Building Backend..."
docker build -t dkonsole/dkonsole-backend:$VERSION ./backend
echo "✅ Backend built successfully"
echo ""

echo "📦 Building Frontend..."
docker build -t dkonsole/dkonsole-frontend:$VERSION ./frontend
echo "✅ Frontend built successfully"
echo ""

echo "🚀 Pushing Backend to Docker Hub..."
docker push dkonsole/dkonsole-backend:$VERSION
echo "✅ Backend pushed successfully"
echo ""

echo "🚀 Pushing Frontend to Docker Hub..."
docker push dkonsole/dkonsole-frontend:$VERSION
echo "✅ Frontend pushed successfully"
echo ""

echo "🏷️  Creating Git tag v${VERSION}..."
if git rev-parse "v${VERSION}" >/dev/null 2>&1; then
    echo "⚠️  Tag v${VERSION} already exists"
    read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d "v${VERSION}"
        git push origin ":refs/tags/v${VERSION}" 2>/dev/null || true
        echo "🗑️  Old tag deleted"
    else
        echo "❌ Release cancelled"
        exit 1
    fi
fi

git tag -a "v${VERSION}" -m "Release v${VERSION}

Features:
- Prometheus integration for Pod metrics
- Historical metrics with time range selector
- Metrics tab in Pod details
- Fixed namespace display for cluster-scoped resources

Docker Images:
- dkonsole/dkonsole-backend:${VERSION}
- dkonsole/dkonsole-frontend:${VERSION}"

echo "✅ Git tag created"
echo ""

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
