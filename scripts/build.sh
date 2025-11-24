#!/bin/bash
set -e

VERSION="1.1.3"

echo "=========================================="
echo "🔨 DKonsole Build v${VERSION}"
echo "=========================================="
echo ""

# Build Unified Docker Image (Backend + Frontend)
echo "📦 Building Unified Image (Backend + Frontend)..."
docker build -t dkonsole/dkonsole:$VERSION .
echo "✅ Unified image built successfully"
echo ""

# Also tag as latest for local testing
echo "🏷️  Tagging as 'latest' for local testing..."
docker tag dkonsole/dkonsole:$VERSION dkonsole/dkonsole:latest
echo "✅ Tagged as latest"
echo ""

echo "=========================================="
echo "✨ Build Complete!"
echo "=========================================="
echo ""
echo "📦 Docker Image:"
echo "   - dkonsole/dkonsole:${VERSION}"
echo "   - dkonsole/dkonsole:latest"
echo ""
echo "🧪 To test locally:"
echo "   docker run -p 8080:8080 dkonsole/dkonsole:${VERSION}"
echo ""
