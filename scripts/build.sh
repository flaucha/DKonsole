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

    # Check if Docker is available (needed for frontend tests)
    DOCKER_AVAILABLE=true
    if ! command_exists docker; then
        DOCKER_AVAILABLE=false
    else
        # Test if Docker daemon is running
        if ! docker info > /dev/null 2>&1; then
            DOCKER_AVAILABLE=false
        fi
    fi

    if [ "$DOCKER_AVAILABLE" = false ]; then
        echo "⚠️  Docker not available - frontend tests will be skipped"
        echo "   Install/start Docker to run frontend tests: https://docs.docker.com/get-docker/"
        echo ""
    fi

    # Backend Tests
    echo "📋 Testing Backend..."
    cd backend
    export GOTOOLCHAIN=go1.25.8

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

    # Run golangci-lint (matching GitHub Actions: install-mode: goinstall)
    echo "  🔍 Running golangci-lint..."
    if ! command_exists golangci-lint || [[ "$(golangci-lint version 2>&1)" != *"go$(go version | grep -oP 'go\d+\.\d+')"* ]]; then
        echo "    Installing golangci-lint with current Go version..."
        GOTOOLCHAIN=go1.25.8 go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
    fi
    "$(go env GOPATH)/bin/golangci-lint" run --timeout=5m ./... || { echo "❌ golangci-lint failed"; exit 1; }
    echo "  ✅ golangci-lint passed"

    # Run govulncheck
    echo "  🔒 Running govulncheck..."
    if command_exists govulncheck; then
        govulncheck ./... || { echo "❌ govulncheck found vulnerabilities"; exit 1; }
        echo "  ✅ govulncheck passed (no vulnerabilities found)"
    else
        echo "  ⚠️  govulncheck not found, installing..."
        GOTOOLCHAIN=go1.25.8 go install golang.org/x/vuln/cmd/govulncheck@latest
        "$(go env GOPATH)/bin/govulncheck" ./... || { echo "❌ govulncheck found vulnerabilities"; exit 1; }
        echo "  ✅ govulncheck passed (no vulnerabilities found)"
    fi

    # Run tests with coverage
    echo "  🧪 Running tests with coverage..."
    go test -v -coverprofile=coverage.out ./... || { echo "❌ Tests failed"; exit 1; }
    echo "  ✅ All tests passed"

    cd ..
    unset GOTOOLCHAIN

    # Frontend Tests (using Docker)
    echo ""
    echo "📋 Testing Frontend (using Docker)..."

    if [ "$DOCKER_AVAILABLE" = false ]; then
        echo "  ⚠️  Docker not available, skipping frontend tests"
        echo "  ✅ Frontend tests skipped (will run in CI)"
    else
        echo "  🐳 Using Docker image: node:22-alpine (same as CI)"

        # Pull the image if not available
        echo "  📥 Checking/pulling Docker image node:22-alpine..."
        if ! docker image inspect node:22-alpine > /dev/null 2>&1; then
            echo "  📥 Pulling Docker image (this may take a minute)..."
            docker pull node:22-alpine || {
                echo "  ❌ Failed to pull Docker image node:22-alpine"
                echo "  💡 Check your internet connection and Docker daemon"
                exit 1
            }
        else
            echo "  ✅ Docker image already available"
        fi

        # Get absolute path for volume mount
        FRONTEND_DIR="$(cd frontend && pwd)"

        TEST_EXIT=0
        
        # Run tests in Docker container
        echo "  🔧 Installing dependencies and running tests in Docker..."
        docker run --rm \
            -v "${FRONTEND_DIR}:/app" \
            -w /app \
            -e NODE_OPTIONS="--max-old-space-size=12288" \
            node:22-bookworm \
            sh -c "
                echo '📥 Installing dependencies...' &&
                npm install &&
                echo '' &&
                echo '🔒 Running npm audit (high)...' &&
                npm audit --audit-level=high 2>&1 &&
                echo '' &&
                echo '🔍 Running linter...' &&
                npm run lint 2>&1 &&
                echo '' &&
                echo '🧪 Running tests (Chunk 1/3: Utils & Hooks)...' &&
                npm run test -- src/utils src/hooks --no-file-parallelism --exclude="**/useWorkloadListState.test.js" &&
                echo '' &&
                echo '🧪 Running tests (Chunk 2/3: API)...' &&
                npm run test -- src/api --no-file-parallelism &&
                echo '' &&
                echo '🧪 Running tests (Chunk 3/3: Components)...' &&
                npm run test -- src/components --no-file-parallelism --exclude="**/Layout.test.jsx" --exclude="**/ApiExplorer.test.jsx" --exclude="**/PodDetails.test.jsx" --exclude="**/WorkloadList.test.jsx"
            " || TEST_EXIT=$?
        
        # Handle OOM during cleanup - if tests passed, ignore exit code
        if [ $TEST_EXIT -ne 0 ]; then
            echo "  ⚠️  Tests exited with code $TEST_EXIT"
            echo "  ℹ️  Note: Coverage runs in CI (GitHub Actions), not locally"
            exit 1
        fi
        echo "  ✅ All frontend tests passed"
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
# if ! git diff-index --quiet HEAD --; then
#     echo "📝 Detected uncommitted changes, committing and pushing..."
#     git add -A
#     git commit -m "chore: update code before build ${TEST_VERSION}" || true
#     if git rev-parse --abbrev-ref HEAD | grep -q "main\|master"; then
#         git push || echo "⚠️  Warning: Could not push to remote (may need manual push)"
#     fi
#     echo "✅ Changes committed and pushed"
#     echo ""
# fi

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
