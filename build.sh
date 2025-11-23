#!/bin/bash
set -e

VERSION="1.0.7"

echo "=========================================="
echo "Building DKonsole v${VERSION}"
echo "=========================================="
echo ""

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

echo "=========================================="
echo "✨ Build Complete!"
echo "Images pushed with tag: v${VERSION}"
echo "=========================================="
