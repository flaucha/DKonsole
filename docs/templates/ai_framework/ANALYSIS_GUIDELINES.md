# Analysis Guidelines

This document guides the AI in performing comprehensive code and project analysis.

## 1. Analysis Framework

When requested to "Analyze the project" or a specific component, follow this structure:

### 1.1 Code Quality Analysis
- **Readability**: Is the code self-documenting? Are variable names clear?
- **Structure**: Does it follow the defined architecture?
- **Complexity**: Identify complex functions.
- **Adherence**: Does it follow `docs/CODING_GUIDELINES.md`?

### 1.2 Security Analysis
- **OWASP Top 10**: Check against `SECURITY_GUIDELINES.md`.
- **Vulnerabilities**: List potential security risks.

### 1.3 SWOT Analysis (FODA)
Create a table or list:
- **Strengths**: What is done well?
- **Weaknesses**: What needs improvement?
- **Opportunities**: What can be added/improved?
- **Threats**: External risks.

## 2. Scoring System

Assign a score from 0 to 100 based on the following criteria:

| Category | Weight | Criteria |
| :--- | :--- | :--- |
| **Architecture** | 30% | Adherence to patterns, Clean Code. |
| **Security** | 30% | OWASP compliance, Secret management. |
| **Reliability** | 20% | Error handling, Logging, Stability. |
| **Maintainability** | 10% | Documentation, Comments, Simplicity. |
| **Testing** | 10% | Test coverage, Test quality. |

**Total Score = Sum of (Category Score * Weight)**

## 3. Reporting Format

Generate a Markdown report with:
1.  **Executive Summary**: High-level overview and Total Score.
2.  **Detailed Findings**: Breakdown by category.
3.  **SWOT Analysis**: The SWOT table.
4.  **Remediation Plan**: Prioritized list of actions to improve the score.
