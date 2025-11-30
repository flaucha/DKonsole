# Security Guidelines

This document outlines the security standards and practices for [Project Name].

## 1. OWASP Top 10 Compliance

The AI must evaluate all code changes against the OWASP Top 10 vulnerabilities.

1.  **Broken Access Control**: Ensure users can only act on resources they own or have permission for.
2.  **Cryptographic Failures**: Protect sensitive data. No hardcoded secrets.
3.  **Injection**: Prevent SQL, Command, and other injections.
4.  **Insecure Design**: Threat modeling during the planning phase.
5.  **Security Misconfiguration**: Secure defaults.
6.  **Vulnerable and Outdated Components**: Dependency management.
7.  **Identification and Authentication Failures**: Secure login.
8.  **Software and Data Integrity Failures**: CI/CD security.
9.  **Security Logging and Monitoring Failures**: Audit logs.
10. **Server-Side Request Forgery (SSRF)**: Validate external URLs.

## 2. Code Security Standards

- **Secrets**: NEVER commit secrets to Git. Use environment variables or secret managers.
- **Input Validation**: Validate all inputs at the API boundary.
- **Output Encoding**: Encode data before rendering to prevent XSS.
- **Least Privilege**: Run with minimum necessary permissions.

## 3. Security Analysis Process

When asked to perform a security analysis:
1.  **Scan**: Review code for the OWASP Top 10.
2.  **Report**: Create a report listing vulnerabilities, severity, and remediation.
3.  **Score**: Assign a security score (0-100).

## 4. SWOT Analysis (Security Focus)

Include a SWOT analysis in security reports:
- **Strengths**: What is secure?
- **Weaknesses**: Vulnerabilities?
- **Opportunities**: Improvements?
- **Threats**: External risks?
