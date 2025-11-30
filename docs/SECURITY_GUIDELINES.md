# Security Guidelines

This document outlines the security standards and practices for DKonsole.

## 1. OWASP Top 10 Compliance

The AI must evaluate all code changes against the OWASP Top 10 vulnerabilities.

1.  **Broken Access Control**: Ensure users can only act on resources they own or have permission for.
    - *Check*: Verify Kubernetes RBAC and internal permission checks in `service.go`.
2.  **Cryptographic Failures**: Protect sensitive data.
    - *Check*: No hardcoded secrets. Use K8s Secrets. Ensure TLS for external communication.
3.  **Injection**: Prevent SQL, Command, and LDAP injection.
    - *Check*: Sanitize inputs. Use parameterized queries/commands.
4.  **Insecure Design**: Threat modeling during the planning phase.
    - *Check*: Review `implementation_plan.md` for security flaws.
5.  **Security Misconfiguration**: Secure defaults.
    - *Check*: Helm chart default values should be secure.
6.  **Vulnerable and Outdated Components**: Dependency management.
    - *Check*: Regularly scan `go.mod` and `package.json`.
7.  **Identification and Authentication Failures**: Secure login.
    - *Check*: JWT handling, session management.
8.  **Software and Data Integrity Failures**: CI/CD security.
    - *Check*: Verify signatures (if applicable), secure pipeline.
9.  **Security Logging and Monitoring Failures**: Audit logs.
    - *Check*: Ensure critical actions are logged via `utils.Logger`.
10. **Server-Side Request Forgery (SSRF)**: Validate external URLs.
    - *Check*: Validate user-supplied URLs before making requests.

## 2. Code Security Standards

- **Secrets**: NEVER commit secrets to Git. Use environment variables or K8s Secrets.
- **Input Validation**: Validate all inputs at the API boundary (Handlers).
- **Output Encoding**: Encode data before rendering in the frontend to prevent XSS.
- **Least Privilege**: The application should run with the minimum necessary K8s permissions.

## 3. Security Analysis Process

When asked to perform a security analysis:
1.  **Scan**: Review code for the OWASP Top 10.
2.  **Report**: Create a report listing vulnerabilities, severity (High/Medium/Low), and remediation.
3.  **Score**: Assign a security score (0-100) based on findings.

## 4. SWOT Analysis (Security Focus)

Include a SWOT analysis in security reports:
- **Strengths**: Robust auth, secure defaults.
- **Weaknesses**: Missing tests, legacy code.
- **Opportunities**: New security features, dependency updates.
- **Threats**: Zero-day exploits, misconfigured K8s clusters.
