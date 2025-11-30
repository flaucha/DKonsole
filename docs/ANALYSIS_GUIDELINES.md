# Analysis Guidelines

This document guides the AI in performing comprehensive code and project analysis.

## 1. Analysis Framework

When requested to "Analyze the project" or a specific component, follow this structure:

### 1.1 Code Quality Analysis
- **Readability**: Is the code self-documenting? Are variable names clear?
- **Structure**: Does it follow the Layered Architecture (Factory -> Service -> Repository)?
- **Complexity**: Identify complex functions (Cyclomatic Complexity).
- **Adherence**: Does it follow `docs/CODING_GUIDELINES.md`?

### 1.2 Security Analysis
- **OWASP Top 10**: Check against `SECURITY_GUIDELINES.md`.
- **Vulnerabilities**: List potential security risks.

### 1.3 SWOT Analysis (FODA)
Create a table or list:
- **Strengths**: What is done well? (e.g., "Clean architecture", "Fast build").
- **Weaknesses**: What needs improvement? (e.g., "Low test coverage", "Legacy code").
- **Opportunities**: What can be added/improved? (e.g., "New K8s features", "Performance optimization").
- **Threats**: External risks (e.g., "Dependency vulnerabilities", "API changes").

## 2. Scoring System

Assign a score from 0 to 100 based on the following criteria:

| Category | Weight | Criteria |
| :--- | :--- | :--- |
| **Architecture** | 30% | Adherence to DDD, Layered Pattern, Clean Code. |
| **Security** | 30% | OWASP compliance, Secret management, RBAC. |
| **Reliability** | 20% | Error handling, Logging, Stability. |
| **Maintainability** | 10% | Documentation, Comments, Simplicity. |
| **Testing** | 10% | Test coverage, Test quality. |

**Total Score = Sum of (Category Score * Weight)**

## 3. Reporting Format

Generate a Markdown report (e.g., `ANALYSIS_REPORT.md`) with:
1.  **Executive Summary**: High-level overview and Total Score.
2.  **Detailed Findings**: Breakdown by category.
3.  **SWOT Analysis**: The SWOT table.
4.  **Remediation Plan**: Prioritized list of actions to improve the score.
