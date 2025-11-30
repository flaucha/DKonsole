#!/bin/bash
set -e

# AI Sanity Check Script
# This script runs all necessary checks to verify the project state.
# Usage: ./scripts/ai-check.sh

echo "🤖 AI Sanity Check Initiated..."

# 1. Backend Checks
echo "🔍 Checking Backend..."
cd backend
echo "  > Running go vet..."
go vet ./...
echo "  > Running staticcheck (if installed)..."
if command -v staticcheck &> /dev/null; then
    staticcheck ./...
fi
cd ..

# 2. Frontend Checks
echo "🔍 Checking Frontend..."
cd frontend
if [ -d "node_modules" ]; then
    echo "  > Running lint..."
    npm run lint
else
    echo "⚠️  node_modules missing, skipping frontend lint."
fi
cd ..

# 3. Run All Tests
echo "🧪 Running Tests..."
./scripts/test-all.sh

echo "✅ AI Sanity Check Passed! You are good to go."
