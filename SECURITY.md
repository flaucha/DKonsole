# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.3.x   | :white_check_mark: |
| < 1.3   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please email security@example.com.

**Do NOT open a public GitHub issue.**

We will respond within 48 hours and provide a timeline for a fix.

## Security Features

### Automated Scanning
- ✅ Trivy container and filesystem scans (daily)
- ✅ govulncheck for Go vulnerabilities
- ✅ npm audit for Node.js dependencies
- ✅ Dependabot for dependency updates

### LDAP Security
- ✅ Input sanitization with `ldap.EscapeFilter()`
- ✅ Username validation (rejects special chars)
- ✅ DN format validation
- ✅ Comprehensive logging of auth attempts

## Changelog

### 2025-11-29: Security Hardening
- Fixed LDAP injection vulnerability (CRITICAL)
- Implemented automated dependency scanning
- Added pre-commit security checks
