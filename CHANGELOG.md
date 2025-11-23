# Changelog

All notable changes to DKonsole will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.7] - 2025-11-23

### 🔒 Security Hardening Release

This release focuses on addressing critical security vulnerabilities and implementing enterprise-grade security measures.

### Critical Security Fixes

- **PromQL Injection Prevention**: Added strict input validation for all Prometheus queries using regex patterns to prevent injection attacks
- **Cookie-Based Authentication**: Migrated from localStorage to HttpOnly cookies for JWT tokens, eliminating XSS token theft risks
  - Cookies now use `HttpOnly`, `Secure`, and `SameSite=Strict` flags
  - Removed token from response body in LoginHandler
  - Added `/api/logout` and `/api/me` endpoints for proper session management
- **Kubernetes Name Validation**: Implemented RFC 1123 validation for all resource names and namespaces to prevent command injection
  - Applied to: GetResourceYAML, DeleteResource, ScaleResource, StreamPodLogs, ExecIntoPod
  - Uses native regex validation for better portability
- **Input Sanitization**: Added comprehensive validation across all API endpoints

### Security Enhancements

- **Audit Logging**: Implemented comprehensive audit middleware logging all API requests
  - Logs include: user, method, path, status code, duration
  - Enables tracking of "who did what and when" for compliance
- **Rate Limiting**: Added intelligent rate limiting (300 req/min per IP)
  - Prevents DoS attacks and brute force attempts
  - Automatically skips WebSocket upgrade requests
- **CSP Updates**: Enhanced Content Security Policy to allow Monaco Editor while maintaining security
  - Added `https://cdn.jsdelivr.net` to allowed sources
  - Added `worker-src 'self' blob:` for Monaco workers
- **JWT Secret Enforcement**: Strict validation requiring JWT_SECRET to be set and minimum 32 characters

### Bug Fixes

- Fixed table sorting functionality in WorkloadList component (was using filteredResources instead of sortedResources)
- Fixed YAML editor infinite loading issue caused by CSP restrictions
- Fixed Secret data display (now properly decodes byte arrays to strings)
- Removed duplicate mux initialization in main.go

### Architecture Improvements

- Unified middleware chain for consistent security policy application
- Separated public and secure route handlers with dedicated wrapper functions
- Enhanced error handling and validation across all handlers
- Created dedicated `middleware.go` file for better code organization

### Developer Experience

- Improved code organization with dedicated middleware.go
- Removed external validation dependency (`k8s.io/apimachinery/pkg/util/validation`)
- Better separation of concerns between authentication, auditing, and rate limiting
- Cleaner route registration with `secure()` and `public()` helpers

### Changed

- Frontend now exclusively uses cookie-based authentication
- AuthContext refactored to use `/api/me` for session validation
- All API calls now use `authFetch` wrapper for consistent auth handling

---

## [1.0.6] - 2025-11-22

### Added

- Enhanced Pod metrics with Network RX/TX and PVC usage
- Improved Cluster Overview with Prometheus integration
- Added real-time node metrics table with CPU, Memory, Disk, and Network stats
- Added cluster-wide statistics (avg CPU/Memory, network traffic, trends)
- Added fadeIn animation for smooth tab transitions

### Fixed

- Fixed visual issues in Pod metrics tabs (removed extra border, smooth transitions)
- Conditional rendering of metrics based on data availability

---

## [1.0.5] - 2025-11-22

### Added

- Added `build.sh` for simple Docker builds
- Added `release.sh` for automated releases with git tagging
- Added `SCRIPTS.md` documentation

### Changed

- Improved build and release scripts

### Deprecated

- Deprecated `deploy.sh` in favor of new scripts

---

## [1.0.4] - 2025-11-22

### Added

- Prometheus integration for Pod metrics
- Historical metrics with time range selector (1h, 6h, 12h, 1d, 7d, 15d)
- Metrics tab to Pod details

### Fixed

- Fixed namespace display for cluster-scoped resources (Nodes, ClusterRoles, etc.)

### Removed

- Removed Metrics tab from Deployment details (kept for future use)

---

## [1.0.3] - 2025-11-21

### Added

- Initial stable release
- Resource management for Deployments, Pods, Services, ConfigMaps, Secrets
- Live log streaming
- Terminal access to pods
- YAML editor
- Argon2 password hashing
- JWT-based authentication
- Multi-cluster support

---

[1.0.7]: https://github.com/flaucha/DKonsole/compare/v1.0.6...v1.0.7
[1.0.6]: https://github.com/flaucha/DKonsole/compare/v1.0.5...v1.0.6
[1.0.5]: https://github.com/flaucha/DKonsole/compare/v1.0.4...v1.0.5
[1.0.4]: https://github.com/flaucha/DKonsole/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/flaucha/DKonsole/releases/tag/v1.0.3
