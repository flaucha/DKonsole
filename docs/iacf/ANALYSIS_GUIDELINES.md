# Analysis Guidelines

This document establishes the strict industry standards for performing comprehensive code and project analysis within the organization. All AI agents and human reviewers must adhere to this framework to ensure software excellence.

## 1. Analysis Framework

When requested to "Analyze the project" or a specific component, execute the following rigorous assessment:

### 1.1 Code Quality & Metrics
- **Complexity**:
    - **Cyclomatic Complexity**: Must be < 10 per function.
    - **Cognitive Complexity**: Must be < 15 per function.
    - **Nesting Depth**: Maximum 3 levels of nesting.
- **Size**:
    - **Function Length**: Should not exceed 30 lines (excluding comments).
    - **File Length**: Should not exceed 300 lines.
- **Readability**:
    - Variable names must be descriptive and context-aware (no `x`, `data`, `temp`).
    - Public API must have GoDoc/JSDoc comments.
    - Code must be self-documenting; comments should explain *why*, not *what*.

### 1.2 Architecture & Design
- **Pattern Adherence**:
    - Verify strict adherence to **Hexagonal Architecture (Ports & Adapters)** or **Layered Architecture**.
    - **Dependency Rule**: Dependencies must point inwards. Domain logic must depend on nothing.
- **Coupling & Cohesion**:
    - **High Cohesion**: Classes/Packages should have a single responsibility (SRP).
    - **Low Coupling**: Interfaces should be used for dependencies to facilitate testing (DIP).
- **Anti-Patterns**: Identify God Classes, Spaghetti Code, Shotgun Surgery, and Feature Envy.

### 1.3 Security (DevSecOps)
- **OWASP Top 10**: Verify protection against Injection, Broken Auth, Sensitive Data Exposure, etc.
- **Secret Management**: Ensure no hardcoded secrets. Use environment variables or Vault.
- **Input Validation**: All external inputs must be sanitized and validated at the boundary.
- **Dependencies**: Check for known vulnerabilities (CVEs) in third-party libraries.

### 1.4 Reliability & Performance
- **Error Handling**: Errors must be wrapped with context, not swallowed. No empty catch blocks.
- **Concurrency**: Check for race conditions, deadlocks, and proper resource cleanup (defer/finally).
- **Performance**:
    - Identify N+1 query problems.
    - Check for inefficient algorithms (O(n^2) or worse) in hot paths.
    - Review memory usage and potential leaks.

### 1.5 Testing Strategy
- **Unit Tests**: Minimum **80% coverage** for business logic. Mocks must be used for external dependencies.
- **Integration Tests**: Verify interaction between layers (e.g., Service -> Repository).
- **Test Quality**: Tests must be deterministic, independent, and fast. Assertions must be meaningful.

## 2. Scoring System

Assign a score from 0 to 100 based on the weighted criteria below. Be strict; do not inflate scores.

| Category | Weight | Criteria for Deduction |
| :--- | :--- | :--- |
| **Security** | **25%** | -5 per High severity issue, -2 per Medium. 0 if critical CVE found. |
| **Architecture** | **20%** | -5 for dependency rule violation, -2 for tight coupling. |
| **Reliability** | **20%** | -5 for swallowed errors, -5 for race conditions. |
| **Testing** | **20%** | 0 if coverage < 50%. Pro-rated between 50%-80%. Full points only if > 80% & high quality. |
| **Maintainability**| **15%** | -1 per function with complexity > 10. -1 per file > 300 lines. |

**Total Score = Sum of (Category Score * Weight)**

## 3. Recommended Tools

Use or recommend the following industry-standard tools to automate analysis:

- **Static Analysis (SAST)**: SonarQube, GolangCI-Lint (Go), ESLint (JS/TS).
- **Security**: Trivy (Container/FS), OPA (Policy), Gitleaks (Secrets).
- **Formatting**: gofmt/goimports, Prettier.
- **Testing**: standard `go test`, Jest/Vitest.

## 4. Reporting Format

Generate a Markdown report (e.g., `ANALYSIS_REPORT.md`) containing:

1.  **Executive Summary**:
    - Total Score (0-100).
    - Critical Risks (Red/Amber/Green status).
2.  **Detailed Findings**:
    - Grouped by category (Security, Architecture, etc.).
    - Include specific file paths and line numbers.
3.  **SWOT Analysis**:
    - **Strengths**: Proven architectural decisions.
    - **Weaknesses**: Technical debt and violations.
    - **Opportunities**: Refactoring and modernization paths.
    - **Threats**: Security risks and scalability bottlenecks.
4.  **Remediation Plan**:
    - Prioritized list of actions (P0: Critical, P1: High, P2: Medium).
    - Estimated effort for each item (e.g., "Low - 2h", "High - 3d").
