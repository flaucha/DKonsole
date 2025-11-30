# Testing Guidelines

This document defines the testing strategy for [Project Name].

## 1. Backend Testing

### 1.1 Unit Tests
- **Scope**: Test individual functions and methods in isolation.
- **Mocking**: Use interfaces and mock implementations for dependencies.
- **Naming**: Follow the language's standard naming convention.

### 1.2 Integration Tests
- **Scope**: Test interactions between components (e.g., Service + Database).
- **Environment**: Use a test environment or containerized dependencies.

### 1.3 Test Coverage
- **Goal**: Aim for >80% code coverage for business logic.
- **Critical Paths**: Auth and core logic MUST have tests.

## 2. Frontend Testing (If applicable)

### 2.1 Component Tests
- **Scope**: Render components and verify UI elements and interactions.
- **Mocking**: Mock API calls.

## 3. AI Responsibility

- **New Features**: You MUST write tests for any new feature.
- **Bug Fixes**: You MUST write a regression test for any bug fix.
- **Refactoring**: Ensure existing tests pass after refactoring.
