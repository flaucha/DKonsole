#!/bin/bash
set -e

VERSION="0.0.4"

echo "Building Backend ($VERSION)..."
docker build -t flaucha/dkonsole-backend:$VERSION ./backend
docker push flaucha/dkonsole-backend:$VERSION

echo "Building Frontend ($VERSION)..."
docker build -t flaucha/dkonsole-frontend:$VERSION ./frontend
docker push flaucha/dkonsole-frontend:$VERSION

echo "Done! Images pushed with tag: $VERSION"
