#!/bin/bash
# DEPRECATED: Use build.sh instead
# This script is kept for backward compatibility
set -e

VERSION="1.0.4"

echo "⚠️  Note: This script is deprecated. Please use './build.sh' instead."
echo ""
echo "Building Backend ($VERSION)..."
docker build -t dkonsole/dkonsole-backend:$VERSION ./backend
docker push dkonsole/dkonsole-backend:$VERSION

echo "Building Frontend ($VERSION)..."
docker build -t dkonsole/dkonsole-frontend:$VERSION ./frontend
docker push dkonsole/dkonsole-frontend:$VERSION

echo "Done! Images pushed with tag: $VERSION"
