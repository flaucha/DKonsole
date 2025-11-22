#!/bin/bash
set -e

VERSION="0.0.5"

echo "Building Backend ($VERSION)..."
docker build -t dkonsole/dkonsole-backend:$VERSION ./backend
docker push dkonsole/dkonsole-backend:$VERSION

echo "Building Frontend ($VERSION)..."
docker build -t dkonsole/dkonsole-frontend:$VERSION ./frontend
docker push dkonsole/dkonsole-frontend:$VERSION

echo "Done! Images pushed with tag: $VERSION"
