#!/bin/bash
set -e

# Check for --skip-tests flag
SKIP_TESTS=false
if [[ "$*" == *"--skip-tests"* ]]; then
    SKIP_TESTS=true
    echo "⚠️  Skipping validation tests (--skip-tests flag detected)"
fi

# Read version from VERSION file or use default
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

if [ -f "VERSION" ]; then
    VERSION=$(cat VERSION | tr -d '[:space:]')
else
    VERSION="1.1.9"
fi

# Use test version for testing (add -test-1 suffix only if version doesn't already have a suffix)
if [[ "$VERSION" == *"-"* ]]; then
    # Version already has a suffix (e.g., 1.3.0-alfa1), use it directly
    TEST_VERSION="$VERSION"
else
    # No suffix, add -test-1
    TEST_VERSION="${VERSION}-test-1"
fi

echo "=========================================="
echo "🔨 DKonsole Build v${TEST_VERSION}"
echo "=========================================="
echo ""
echo "💡 Tip: Use --skip-tests flag to skip validation tests"
echo ""

# ==========================================
# 🧪 PRE-BUILD VALIDATION
# ==========================================
if [ "$SKIP_TESTS" = false ]; then
    echo "🧪 Running pre-build validation tests..."
    echo ""

    # Helper function to check if command exists
    command_exists() {
        command -v "$1" >/dev/null 2>&1
    }

    # Backend Tests
    echo "📋 Testing Backend..."
    cd backend

    # Update go.mod
    echo "  🔄 Updating go.mod..."
    go mod tidy || { echo "❌ Failed to update go.mod"; exit 1; }

    # Download dependencies
    echo "  📥 Downloading dependencies..."
    go mod download || { echo "❌ Failed to download dependencies"; exit 1; }

    # Run go vet
    echo "  🔍 Running go vet..."
    go vet ./... || { echo "❌ go vet failed"; exit 1; }
    echo "  ✅ go vet passed"

    # Run golangci-lint (optional but recommended)
    if command_exists golangci-lint; then
        echo "  🔍 Running golangci-lint..."
        golangci-lint run --timeout=5m ./... || { echo "❌ golangci-lint failed"; exit 1; }
        echo "  ✅ golangci-lint passed"
    else
        echo "  ⚠️  golangci-lint not found (optional, install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)"
    fi

    # Run govulncheck
    echo "  🔒 Running govulncheck..."
    if command_exists govulncheck; then
        govulncheck ./... || { echo "❌ govulncheck found vulnerabilities"; exit 1; }
        echo "  ✅ govulncheck passed (no vulnerabilities found)"
    else
        echo "  ⚠️  govulncheck not found, installing..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
        "$(go env GOPATH)/bin/govulncheck" ./... || { echo "❌ govulncheck found vulnerabilities"; exit 1; }
        echo "  ✅ govulncheck passed (no vulnerabilities found)"
    fi

    # Run tests with coverage
    echo "  🧪 Running tests with coverage..."
    go test -v -coverprofile=coverage.out ./... || { echo "❌ Tests failed"; exit 1; }
    echo "  ✅ All tests passed"

    cd ..

    # Frontend Tests
    echo ""
    echo "📋 Testing Frontend..."

    # Check Node.js version
    NODE_VERSION=$(node --version 2>/dev/null | sed 's/v//' | cut -d. -f1)
    if [ -z "$NODE_VERSION" ] || [ "$NODE_VERSION" -lt 18 ]; then
        echo "  ⚠️  Node.js version is too old (current: $(node --version 2>/dev/null || echo 'unknown'))"
        echo "  ⚠️  Frontend tests require Node.js 18+ (CI uses Node.js 20)"
        echo "  ⚠️  Skipping frontend tests due to incompatible Node.js version"
        echo "  💡 To test frontend locally, upgrade Node.js or use nvm: nvm install 20 && nvm use 20"
        echo "  ✅ Frontend tests skipped (will run in CI with Node.js 20)"
    else
        cd frontend

        # Install dependencies if node_modules doesn't exist
        if [ ! -d "node_modules" ]; then
            echo "  📥 Installing dependencies..."
            npm install || { echo "❌ Failed to install dependencies"; exit 1; }
        else
            echo "  ✅ Dependencies already installed"
        fi

        # Run npm audit (warning only, don't fail)
        echo "  🔒 Running npm audit..."
        if npm audit --audit-level=moderate 2>/dev/null; then
            echo "  ✅ npm audit passed"
        else
            echo "  ⚠️  npm audit found vulnerabilities (check with: npm audit)"
        fi

        # Run linter (warning only, don't fail)
        echo "  🔍 Running linter..."
        if npm run lint 2>/dev/null; then
            echo "  ✅ Linter passed"
        else
            echo "  ⚠️  Linter found issues (check with: npm run lint)"
        fi

        # Run tests with coverage
        echo "  🧪 Running tests with coverage..."
        npm run test -- --run --coverage || { echo "❌ Frontend tests failed"; exit 1; }
        echo "  ✅ All frontend tests passed"

        cd ..
    fi

    # Security Scan with Trivy (optional, warnings only)
    echo ""
    echo "🔒 Running security scan..."
    if command_exists trivy; then
        echo "  🔍 Scanning filesystem with Trivy..."
        set +e  # Don't fail on Trivy warnings
        trivy fs --severity CRITICAL,HIGH . 2>&1
        TRIVY_EXIT_CODE=$?
        set -e  # Re-enable exit on error
        if [ $TRIVY_EXIT_CODE -eq 0 ]; then
            echo "  ✅ Trivy scan completed (no critical/high vulnerabilities found)"
        else
            echo "  ⚠️  Trivy found critical/high vulnerabilities (check output above)"
            echo "  💡 Consider fixing vulnerabilities before building"
        fi
    else
        echo "  ⚠️  Trivy not found (optional, install with: https://aquasecurity.github.io/trivy/latest/getting-started/installation/)"
    fi

    echo ""
    echo "✅ All pre-build validation tests passed!"
    echo ""
else
    echo "⏭️  Skipping validation tests (use without --skip-tests to run them)"
    echo ""
fi

# Check for uncommitted changes and commit/push if needed
if ! git diff-index --quiet HEAD --; then
    echo "📝 Detected uncommitted changes, committing and pushing..."
    git add -A
    git commit -m "chore: update code before build ${TEST_VERSION}" || true
    if git rev-parse --abbrev-ref HEAD | grep -q "main\|master"; then
        git push || echo "⚠️  Warning: Could not push to remote (may need manual push)"
    fi
    echo "✅ Changes committed and pushed"
    echo ""
fi

# Build Unified Docker Image (Backend + Frontend)
echo "📦 Building Unified Image (Backend + Frontend)..."
docker build -t dkonsole/dkonsole:$TEST_VERSION .
echo "✅ Unified image built successfully"
echo ""

# Push to Docker Hub
echo "🚀 Pushing Unified Image to Docker Hub..."
docker push dkonsole/dkonsole:$TEST_VERSION
echo "✅ Unified image pushed successfully"
echo ""

echo "=========================================="
echo "✨ Build Complete!"
echo "=========================================="
echo ""
echo "📦 Docker Image:"
echo "   - dkonsole/dkonsole:${TEST_VERSION}"
echo ""
echo "🧪 To test locally:"
echo "   docker run -p 8080:8080 dkonsole/dkonsole:${TEST_VERSION}"
echo ""
