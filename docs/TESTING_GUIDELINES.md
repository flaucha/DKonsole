# Testing Guidelines

This document defines the testing strategy for DKonsole.

## 1. Backend Testing (Go)

### 1.1 Unit Tests
- **Location**: `_test.go` files next to the code being tested.
- **Scope**: Test individual functions and methods in isolation.
- **Mocking**: Use interfaces and mock implementations for dependencies (e.g., K8s client, other services).
- **Naming**: `Test<FunctionName>_<Scenario>`.

### 1.2 Integration Tests
- **Location**: `backend/tests/integration/` (if applicable) or specifically marked tests.
- **Scope**: Test interactions between components (e.g., Service + Repository).
- **Environment**: May require a running K8s cluster (use `envtest` or Kind).

### 1.3 Test Coverage
- **Goal**: Aim for >80% code coverage for business logic (`service.go`).
- **Critical Paths**: Auth, K8s interaction, and critical data flows MUST have tests.

### 1.4 Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 2. Frontend Testing (React)

### 2.1 Component Tests
- **Tool**: Vitest / Jest + React Testing Library.
- **Scope**: Render components and verify UI elements and interactions.
- **Mocking**: Mock API calls (`msw` or manual mocks).

### 2.2 Running Tests
```bash
cd frontend
npm test
```

## 3. AI Responsibility

- **New Features**: You MUST write tests for any new feature.
- **Bug Fixes**: You MUST write a regression test for any bug fix.
- **Refactoring**: Ensure existing tests pass after refactoring.
