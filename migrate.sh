#!/bin/bash

# Migration script from kview to DKonsole
# Run this script to copy the source code

SOURCE_DIR="../kview"
DEST_DIR="."

echo "🔄 Migrating code from kview to DKonsole..."

# Create directories
mkdir -p backend frontend

# Copy backend files
echo "📦 Copying backend..."
cp -r $SOURCE_DIR/backend/*.go backend/
cp $SOURCE_DIR/backend/go.mod backend/
cp $SOURCE_DIR/backend/go.sum backend/

# Copy frontend files
echo "🎨 Copying frontend..."
cp -r $SOURCE_DIR/frontend/src frontend/
cp -r $SOURCE_DIR/frontend/public frontend/
cp $SOURCE_DIR/frontend/package.json frontend/
cp $SOURCE_DIR/frontend/index.html frontend/
cp $SOURCE_DIR/frontend/vite.config.js frontend/
cp $SOURCE_DIR/frontend/tailwind.config.js frontend/
cp $SOURCE_DIR/frontend/postcss.config.js frontend/
[ -f $SOURCE_DIR/frontend/.env.example ] && cp $SOURCE_DIR/frontend/.env.example frontend/

echo "✅ Migration complete!"
echo "📝 Next steps:"
echo "   1. Review the copied files"
echo "   2. Update package.json and go.mod if needed"
echo "   3. Build Docker images"
echo "   4. Deploy using Helm chart"
